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
		dial: func(ctx context.Context, address string) (mcpapi.AggregatedMeshConfigServiceClient, error) {

			// Using the same credentials that Galley uses as a server for the Client Connection as well
			// This would mean the same root of trust for all the peers of Galley - be it server or client
			// TODO - May need client specific credentials

			// Copied from pilot/pkg/bootstrap/server.go
			requiredMCPCertCheckFreq := 500 * time.Millisecond
			u, err := url.Parse(mcpAddress)
			if err != nil {
				return nil, err
			}

			securityOption := grpc.WithInsecure()
			if u.Scheme == "mcps" {
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
	i, found := metadata.Types.Lookup(c.TypeURL)
	if !found {
		return fmt.Errorf("invalid type: %v", c.TypeURL)
	}
	scope.Debugf("received object %+v and length %d", c, len(c.Objects))
	var errs error
	ea := make([]resource.Event, 0)
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
			ea = append(ea, resource.Event{
				Kind:  resource.Added,
				Entry: e})
		}
	}

	// At this point we have all resources for that TypeURL
	if errs == nil {
		// Is it worth having another cache to find the delta? Instead
		// remove existing state entries and recreate for the TypeURL
		s.c <- resource.Event{
			Kind: resource.DeletedTypeURL,
			Entry: resource.Entry{
				ID: resource.VersionedKey{
					Key: resource.Key{
						TypeURL: i.TypeURL,
					},
				},
			},
		}
		// If there are no objects for the TypeURL, mcp server would send
		// an empty object slice. So ea would be empty
		for _, e := range ea {
			s.c <- e
		}
		s.c <- resource.Event{Kind: resource.FullSync}
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
