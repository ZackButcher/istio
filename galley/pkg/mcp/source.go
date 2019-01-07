package mcp

import (
	"context"
	"fmt"
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

	c    chan resource.Event
	stop func()
}

var _ runtime.Source = &source{}
var _ client.Updater = &source{}

func New(ctx context.Context, mcpAddress, nodeID string) runtime.Source {
	s := &source{
		nodeID:  nodeID,
		address: mcpAddress,
		ctx:     ctx,
		c:       make(chan resource.Event, defaultChanSize),
		dial: func(ctx context.Context, address string) (mcpapi.AggregatedMeshConfigServiceClient, error) {
			// TODO: configure connection creds
			securityOption := grpc.WithInsecure()

			// Copied from pilot/pkg/bootstrap/server.go
			msgSizeOption := grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(bootstrap.DefaultMCPMaxMsgSize))
			conn, err := grpc.DialContext(ctx, address, securityOption, msgSizeOption)
			if err != nil {
				scope.Errorf("Unable to dial MCP Server %q: %v", address, err)
				return nil, err
			}
			return mcpapi.NewAggregatedMeshConfigServiceClient(conn), nil
		},
	}
	return s
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
	var errs error
	for _, o := range c.Objects {
		e, err := toEntry(o)
		if err != nil {
			errs = multierror.Append(errs, err)
		}

		s.c <- resource.Event{
			// TODO: ensure FullSync is the correct kind; because we get the full resource I think this is the right thing to do
			Kind:  resource.FullSync,
			Entry: e,
		}
	}
	return errs
}

// toEntry converts the object into a resource.Entry. It returns an error if it fails to marshal the object's time,
// but the resulting Entry is still usable.
func toEntry(o *client.Object) (resource.Entry, error) {
	t, err := types.TimestampFromProto(o.Metadata.CreateTime)
	if err != nil {
		// TODO: what time do we want to use if we can't parse the object's creation time?
		t = time.Date(0, 0, 0, 0, 0, 0, 0, nil) // time.Now()
	}

	i, found := metadata.Types.Lookup(o.TypeURL)
	if !found {
		return resource.Entry{}, fmt.Errorf("invalid type: %v", o)
	}

	return resource.Entry{
		ID: resource.VersionedKey{
			Key: resource.Key{
				TypeURL:  i.TypeURL,
				FullName: resource.FullNameFromNamespaceAndName(o.Metadata.Name, ""),
			},
			Version:    resource.Version(o.Metadata.Version),
			CreateTime: t,
		},
		Item: o.Resource,
	}, err
}
