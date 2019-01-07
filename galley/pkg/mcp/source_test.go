package mcp

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"testing"
	"time"

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
}

func allConfigsSnapshot(infos []testData) snapshot.Snapshot {
	createTime := time.Now()
	b := snapshot.NewInMemoryBuilder()
	for _, i := range infos {
		b.SetEntry(i.info.TypeURL.String(), i.info.TypeURL.String(), "1", createTime, i.entry)
	}
	return b.Build()
}

func Test(t *testing.T) {
	data := []testData{
		{
			info:  metadata.Types.Get("type.googleapis.com/istio.networking.v1alpha3.VirtualService"),
			entry: &v1alpha3.VirtualService{},
		},
	}

	server, err := mcptest.NewServer(0, metadata.Types.TypeURLs())
	if err != nil {
		t.Fatalf("failed to start MCP test server: %v", err)
	}

	infos := []resource.Info{metadata.Types.Get("type.googleapis.com/istio.networking.v1alpha3.VirtualService")}
	bytype := make(map[string]resource.Info, len(infos))
	for _, i := range infos {
		bytype[i.TypeURL.String()] = i
	}
	server.Cache.SetSnapshot(snapshot.DefaultGroup, allConfigsSnapshot(data))

	url := callableURL(server.URL)
	fmt.Printf("connecting to: %q\n", url)
	underTest := New(context.Background(), url, "")
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
	go func() {
		for e := range events {
			fmt.Printf("processing event: %v", e.Entry.ID)
			es = append(es, e)
			if len(es) == len(infos) {
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
		if len(r) != len(infos) {
			t.Fatalf("too few results: %v", r)
		}
	}
}

func callableURL(u *url.URL) string {
	return fmt.Sprintf("[%s]:%s", u.Hostname(), u.Port())
}
