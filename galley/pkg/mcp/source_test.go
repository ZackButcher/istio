package mcp

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"testing"
	"time"

	"istio.io/istio/pkg/mcp/creds"

	"istio.io/api/networking/v1alpha3"

	"github.com/gogo/protobuf/proto"

	"istio.io/istio/galley/pkg/metadata"
	"istio.io/istio/galley/pkg/runtime/resource"
	"istio.io/istio/pkg/mcp/snapshot"
	"istio.io/istio/pkg/mcp/testing"
)

type testData struct {
	info  resource.Info
	entry proto.Message
	name  string
}

func allConfigsSnapshot(infos []testData) snapshot.Snapshot {
	createTime := time.Now()
	b := snapshot.NewInMemoryBuilder()
	// snapshot is supposed to be immutable. This is for testing
	for _, i := range infos {
		b.SetEntry(i.info.TypeURL.String(), i.name, "1", createTime, i.entry)
		b.SetVersion(i.info.TypeURL.String(), "1")
	}

	return b.Build()
}

func Test(t *testing.T) {
	data := []testData{
		{
			info: metadata.Types.Get("type.googleapis.com/istio.networking.v1alpha3.VirtualService"),
			name: "test.vservice1",
			entry: &v1alpha3.VirtualService{
				Hosts: []string{"localhost"},
			},
		},
		{
			info: metadata.Types.Get("type.googleapis.com/istio.networking.v1alpha3.VirtualService"),
			name: "test.vservice2",
			entry: &v1alpha3.VirtualService{
				Hosts:       []string{"somehost"},
				Gateways:    []string{"test"},
				Http:        nil,
				Tls:         nil,
				Tcp:         nil,
				ConfigScope: 0,
			},
		},
		{
			info: metadata.Types.Get("type.googleapis.com/istio.networking.v1alpha3.Gateway"),
			name: "test.gw1",
			entry: &v1alpha3.Gateway{
				Servers:  nil,
				Selector: map[string]string{"istio": "ingressgateway"},
			},
		},
	}

	infos := []resource.Info{metadata.Types.Get("type.googleapis.com/istio.networking.v1alpha3.VirtualService"),
		metadata.Types.Get("type.googleapis.com/istio.networking.v1alpha3.Gateway")}
	bytype := make(map[string]resource.Info, len(infos))
	for _, i := range infos {
		bytype[i.TypeURL.String()] = i
	}

	server, err := mcptest.NewServer(0, metadata.Types.TypeURLs())
	if err != nil {
		t.Fatalf("failed to start MCP test server: %v", err)
	}

	url := schemeURL(server.URL)
	fmt.Printf("connecting to: %q\n", url)
	underTest, _ := New(context.Background(), &creds.Options{}, url, "")
	events, err := underTest.Start()
	if err != nil {
		t.Fatalf("failed to start test server with: %v", err)
	}

	for {
		if status := server.Cache.Status(snapshot.DefaultGroup); status != nil {
			if status.Watches() > 0 {
				break
			}
		}
		log.Println("sleeping to wait for cache watch")
		time.Sleep(10 * time.Millisecond)
	}

	server.Cache.SetSnapshot(snapshot.DefaultGroup, allConfigsSnapshot(data))

	results := make(chan []resource.Event)
	es := make([]resource.Event, 0, len(infos))
	prev := resource.FullSync
	// two groups (TypeURL) in this test
	gc := 0
	tot := 2
	go func() {
		// For every TypeURL update, Added events come flanked between DeletedTypeURL and FullSync
		for {
			e := <-events
			switch e.Kind {
			case resource.DeletedTypeURL:
				if prev != resource.FullSync {
					t.Fatalf("First event in the group is %s instead of %s", e.Kind.String(), resource.DeletedTypeURL.String())
				}
				prev = resource.DeletedTypeURL
			case resource.Added:
				if prev != resource.DeletedTypeURL && prev != resource.Added {
					t.Fatalf("Middle events in the group is %s instead of %s", e.Kind.String(), resource.Added.String())
				}
				prev = resource.Added
			case resource.FullSync:
				if prev != resource.Added && prev != resource.DeletedTypeURL {
					t.Fatalf("Last event in the group is %s instead of %s", e.Kind.String(), resource.FullSync.String())
				}
				prev = resource.FullSync
				gc++
			}

			fmt.Printf("processing event: %s %v\n", e.Kind.String(), e.Entry.ID)
			es = append(es, e)
			if gc == tot {
				break
			}
		}
		results <- es
	}()

	wait := time.NewTimer(1 * time.Minute).C
	select {
	case <-wait:
		t.Fatalf("timed out waiting for all events")
	case r := <-results:
		if gc != tot {
			t.Fatalf("too few results: %v", r)
		}
	}
}

func schemeURL(u *url.URL) string {
	return fmt.Sprintf("mcp://%s:%s", u.Hostname(), u.Port())
}
