package mcp

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/grpc"

	mcpapi "istio.io/api/mcp/v1alpha1"
	"istio.io/istio/galley/pkg/metadata"
	"istio.io/istio/galley/pkg/runtime"
	"istio.io/istio/galley/pkg/runtime/resource"
	"istio.io/istio/pilot/pkg/bootstrap"
	"istio.io/istio/pkg/log"
	"istio.io/istio/pkg/mcp/client"
	"istio.io/istio/pkg/mcp/creds"
)

var (
	scope = log.RegisterScope("mcp", "mcp debugging", 0)
)

// This should line up with the MCP client's batch size so we can enqueue every event in a batch without blocking.
const defaultChanSize = 1000

type source struct {
	nodeID  string
	address string
	ctx     context.Context
	dial    func(ctx context.Context, address string) (mcpapi.AggregatedMeshConfigServiceClient, error)
	c       chan resource.Event
	stop    func()
	lCache  map[string]map[string]*Cache
}

type Cache struct {
	version string
	prevVer string
	ek      resource.EventKind
	entry   *resource.Entry
}

var _ runtime.Source = &source{}
var _ client.Updater = &source{}

// New returns nil error. error is added to be consistent with other source's New functions
func New(ctx context.Context, copts *creds.Options, mcpAddress, nodeID string) (runtime.Source, error) {
	s := &source{
		nodeID:  nodeID,
		address: mcpAddress,
		ctx:     ctx,
		c:       make(chan resource.Event, defaultChanSize),
		lCache:  make(map[string]map[string]*Cache),
		dial: func(ctx context.Context, address string) (mcpapi.AggregatedMeshConfigServiceClient, error) {

			// Copied from pilot/pkg/bootstrap/server.go
			requiredMCPCertCheckFreq := 500 * time.Millisecond
			u, err := url.Parse(mcpAddress)
			if err != nil {
				return nil, err
			}

			securityOption := grpc.WithInsecure()
			if u.Scheme == "mcps" {
				// TODO - Check for presence of cert files - this code could be extracted into function
				// and shared with pilot/pkg/bootstrap/server.go- initMCPConfigController
				// https://github.com/istio/istio/blob/master/pilot/pkg/bootstrap/server.go#L574-L591

				requiredFiles := []string{
					copts.CertificateFile,
					copts.KeyFile,
					copts.CACertificateFile,
				}
				scope.Infof("Secure MCP Client configured. Waiting for required certificate files to become available: %v",
					requiredFiles)
				for len(requiredFiles) > 0 {
					if _, err := os.Stat(requiredFiles[0]); os.IsNotExist(err) {
						log.Infof("%v not found. Checking again in %v", requiredFiles[0], requiredMCPCertCheckFreq)
						select {
						case <-ctx.Done():
							return nil, nil
						case <-time.After(requiredMCPCertCheckFreq):
							// retry
						}
						continue
					}
					log.Infof("%v found", requiredFiles[0])
					requiredFiles = requiredFiles[1:]
				}
				//

				watcher, err := creds.WatchFiles(ctx.Done(), copts)
				if err != nil {
					return nil, err
				}
				credentials := creds.CreateForClient(u.Hostname(), watcher)
				securityOption = grpc.WithTransportCredentials(credentials)
			}

			msgSizeOption := grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(bootstrap.DefaultMCPMaxMsgSize))
			conn, err := grpc.DialContext(ctx, u.Hostname()+":"+u.Port(), securityOption, msgSizeOption)
			if err != nil {
				scope.Errorf("Unable to dial MCP Server %q: %v", address, err)
				return nil, err
			}
			return mcpapi.NewAggregatedMeshConfigServiceClient(conn), nil
		},
	}
	return s, nil
}

func (s *source) Start() (chan resource.Event, error) {
	ctx, cancel := context.WithCancel(s.ctx)
	s.stop = cancel
	cl, err := s.dial(ctx, s.address)
	if err != nil {
		scope.Debugf("failed to dial %q", s.address)
		return nil, err
	}

	mcpClient := client.New(cl, metadata.Types.TypeURLs(), s, s.nodeID, map[string]string{}, client.NewStatsContext("galley-mcp-client"))
	scope.Debugf("starting MCP client in its own goroutine")
	go mcpClient.Run(ctx)
	return s.c, nil
}

func (s *source) Stop() {
	s.stop()
}

func (s *source) Apply(c *client.Change) error {
	_, found := metadata.Types.Lookup(c.TypeURL)
	if !found {
		return fmt.Errorf("invalid type: %v", c.TypeURL)
	}
	scope.Debugf("received object %+v and length %d", c, len(c.Objects))

	// By default mark all the elements of local cache deleted, if present
	if _, ok := s.lCache[c.TypeURL]; !ok {
		s.lCache[c.TypeURL] = make(map[string]*Cache)
	} else {
		for _, c := range s.lCache[c.TypeURL] {
			c.ek = resource.Deleted
		}
	}

	var errs error
	for _, o := range c.Objects {
		if o.TypeURL != c.TypeURL {
			errs = multierror.Append(errs,
				fmt.Errorf("type %v mismatch in received object: %v", c.TypeURL, o.TypeURL))
			continue
		}
		e, err := toEntry(o)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}
		// If there is an error, ignore the entire batch.
		// but keep appending errs[] for the batch
		if errs == nil {
			lc, ok := s.lCache[o.TypeURL][o.Metadata.Name]
			if ok {
				if lc.version == o.Metadata.Version {
					lc.ek = resource.None
				} else {
					lc.prevVer = lc.version
					lc.version = o.Metadata.Version
					lc.ek = resource.Updated
				}
				lc.entry = &e

			} else {
				s.lCache[o.TypeURL][o.Metadata.Name] = &Cache{
					version: o.Metadata.Version,
					ek:      resource.Added,
					entry:   &e,
				}
			}
		}
	}

	// If there is an error, the entire batch will be ignored
	// Hence remove any new adds from the local cache
	// And restore prev version if there were updates
	if errs != nil {
		for key, lc := range s.lCache[c.TypeURL] {
			if lc.ek == resource.Added {
				delete(s.lCache[c.TypeURL], key) //this is safe
			} else if lc.ek == resource.Updated {
				lc.version = lc.prevVer // We are not worried about the entry as it will be overwritten
			}
		}
	} else {
		// At this point we have all resources for that TypeURL
		//loop through the local cache and generate the appropriate events
		es := false
		for key, lc := range s.lCache[c.TypeURL] {
			if lc.ek == resource.None {
				continue
			}
			es = true
			scope.Debugf("pushed an event %+v", lc.ek)
			s.c <- resource.Event{
				Kind:  lc.ek,
				Entry: *lc.entry,
			}
			// If its a Deleted event, remove it from local cache
			if lc.ek == resource.Deleted {
				delete(s.lCache[c.TypeURL], key)
			}
		}
		if es {
			// Do a FullSynch after all events for a publish
			s.c <- resource.Event{Kind: resource.FullSync}
		}
	}

	return errs
}

// toEntry converts the object into a resource.Entry.
func toEntry(o *client.Object) (resource.Entry, error) {

	t := time.Now()
	if o.Metadata.CreateTime != nil {
		var err error
		if t, err = types.TimestampFromProto(o.Metadata.CreateTime); err != nil {
			return resource.Entry{}, err
		}
	}

	i, _ := metadata.Types.Lookup(o.TypeURL)
	return resource.Entry{
		ID: resource.VersionedKey{
			Key: resource.Key{
				TypeURL:  i.TypeURL,
				FullName: resource.FullNameFromNamespaceAndName("", o.Metadata.Name),
			},
			Version:    resource.Version(o.Metadata.Version),
			CreateTime: t,
		},
		Item: o.Resource,
	}, nil
}
