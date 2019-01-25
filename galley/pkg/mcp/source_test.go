package mcp

import (
	"context"
	"fmt"
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
	version string
}

func allConfigsSnapshot(infos []testData, version string) snapshot.Snapshot {
	createTime := time.Now()
	b := snapshot.NewInMemoryBuilder()
	// snapshot is supposed to be immutable. This is for testing
	for _, i := range infos {
		b.SetEntry(i.info.TypeURL.String(), i.name, i.version, createTime, i.entry)
		b.SetVersion(i.info.TypeURL.String(), version)
	}

	return b.Build()
}

func Test(t *testing.T) {
	data := []testData{
		{
			info: metadata.Types.Get("type.googleapis.com/istio.networking.v1alpha3.VirtualService"),
			name: "test.vservice1",
			version: "1",
			entry: &v1alpha3.VirtualService{
				Hosts: []string{"localhost"},
			},
		},
		{
			info: metadata.Types.Get("type.googleapis.com/istio.networking.v1alpha3.VirtualService"),
			name: "test.vservice2",
			version: "1",
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
			version: "1",
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

	// start test mcp server
	server, err := mcptest.NewServer(0, metadata.Types.TypeURLs())
	if err != nil {
		t.Fatalf("failed to start MCP test server: %v", err)
	}

	// run galley Start for galley mcp client to connect to mcp server
	url := schemeURL(server.URL)
	fmt.Printf("connecting to: %q\n", url)
	underTest, _ := New(context.Background(), &creds.Options{}, url, "")
	events, err := underTest.Start()
	if err != nil {
		t.Fatalf("failed to start test server with: %v", err)
	}

	// Test Add
	// We should receive 2 Add events and a FullSync for Virtual Service
	// 1 Add and a FullSync for Gateway
	server.Cache.SetSnapshot(snapshot.DefaultGroup, allConfigsSnapshot(data, "1"))
	testAdd(t, events)

	//Test Update
	// We should see one update event in the channel
	data[0].version = "2"
	server.Cache.SetSnapshot(snapshot.DefaultGroup, allConfigsSnapshot(data, "2"))
	testUpdate(t, events)

	//Test AddAndUpdate
	// We should see one Update and one Add event in the channel
	data[1].version = "2"
	a := testData {
	info: metadata.Types.Get("type.googleapis.com/istio.networking.v1alpha3.Gateway"),
		name: "test.gw2",
		version: "1",
		entry: &v1alpha3.Gateway{
		Servers:  nil,
		Selector: map[string]string{"istio": "ingressgateway"},
		},
	}
	data = append(data, a)
	server.Cache.SetSnapshot(snapshot.DefaultGroup, allConfigsSnapshot(data, "3"))
	testAddAndUpdate(t, events)

	//Test Delete
	//We should see two Delete events in the channel
	data = append(data[:1], data[2])
	server.Cache.SetSnapshot(snapshot.DefaultGroup, allConfigsSnapshot(data, "4"))
	testDelete(t, events)


	//Test NoUpdate
	// We should not see any events in the channel
	server.Cache.SetSnapshot(snapshot.DefaultGroup, allConfigsSnapshot(data, "5"))
	testNoChange(t, events)


}

func schemeURL(u *url.URL) string {
	return fmt.Sprintf("mcp://%s:%s", u.Hostname(), u.Port())
}

func testAdd(t *testing.T, events chan resource.Event){
	results := make(chan resource.Event)
	vs := 0
	gw := 0
	tot := 0
	go func() {
		for {
			if tot == 5 {
				break
			}
			e := <-events
			fmt.Println("received event ", e)
			switch e.Kind {
			case resource.Added:
				if e.Entry.ID.TypeURL.String() == "type.googleapis.com/istio.networking.v1alpha3.VirtualService" {
					vs++
				}
				if e.Entry.ID.TypeURL.String() == "type.googleapis.com/istio.networking.v1alpha3.Gateway" {
					gw++
				}
				tot++
			case resource.FullSync:
				tot++
			default:
				t.Fatalf("Unexpected event received")
			}
		}
		if gw != 1 || vs != 2 {
			t.Fatalf("Didn't receive all the Add events")
		}
		results <- resource.Event{}
	}()

	wait := time.NewTimer(1 * time.Minute).C
	select {
	case <-wait:
		t.Fatalf("timed out waiting for all events")
	case r := <-results:
		if tot != 5 {
			t.Fatalf("too few results: %v", r)
		}
	}
	fmt.Println("event testAdd pass")

}

func testUpdate(t *testing.T, events chan resource.Event){
	results := make(chan resource.Event)
	vs := 0
	tot := 0
	go func() {
		for {
			if tot == 2 {
				break
			}
			e := <-events
			fmt.Println("received event ", e)
			switch e.Kind {
			case resource.Updated:
				if e.Entry.ID.TypeURL.String() == "type.googleapis.com/istio.networking.v1alpha3.VirtualService" {
					vs++
				}
				tot++
			case resource.FullSync:
				tot++
			default:
				t.Fatalf("Unexpected event received")
			}
		}
		if vs != 1 {
			t.Fatalf("Didn't receive all the Update events")
		}
		results <- resource.Event{}
	}()

	wait := time.NewTimer(1 * time.Minute).C
	select {
	case <-wait:
		t.Fatalf("timed out waiting for all events")
	case r := <-results:
		if tot != 2 {
			t.Fatalf("too few results: %v", r)
		}
	}
	fmt.Println("event testUpdate pass")

}

func testAddAndUpdate(t *testing.T, events chan resource.Event){
	results := make(chan resource.Event)
	vs := 0
	gw := 0
	tot := 0
	go func() {
		for {
			if tot == 4 {
				break
			}
			e := <-events
			fmt.Println("received event ", e)
			switch e.Kind {
			case resource.Added:
				if e.Entry.ID.TypeURL.String() == "type.googleapis.com/istio.networking.v1alpha3.Gateway" {
					gw++
					tot++
				} else {
						t.Fatalf("Unexpected event received")
				}
			case resource.Updated:
				if e.Entry.ID.TypeURL.String() == "type.googleapis.com/istio.networking.v1alpha3.VirtualService" {
					vs++
					tot++
				} else {
					t.Fatalf("Unexpected event received")
				}
			case resource.FullSync:
				tot++
			default:
				t.Fatalf("Unexpected event received")
			}
		}
		if gw != 1 || vs != 1 {
			t.Fatalf("Didn't receive all the expected events")
		}
		results <- resource.Event{}
	}()

	wait := time.NewTimer(1 * time.Minute).C
	select {
	case <-wait:
		t.Fatalf("timed out waiting for all events")
	case r := <-results:
		if tot != 4 {
			t.Fatalf("too few results: %v", r)
		}
	}
	fmt.Println("event testAddAndUpdate pass")

}

func testDelete(t *testing.T, events chan resource.Event){
	results := make(chan resource.Event)
	vs := 0
	gw := 0
	tot := 0
	go func() {
		for {
			if tot == 4 {
				break
			}
			e := <-events
			fmt.Println("received event ", e)
			switch e.Kind {
			case resource.Deleted:
				if e.Entry.ID.TypeURL.String() == "type.googleapis.com/istio.networking.v1alpha3.VirtualService" {
					vs++
				} else if e.Entry.ID.TypeURL.String() == "type.googleapis.com/istio.networking.v1alpha3.Gateway" {
					gw++
				} else {
						t.Fatalf("Unexpected event received")
				}
				tot++
			case resource.FullSync:
				tot++
			default:
				t.Fatalf("Unexpected event received")
			}
		}
		if vs != 1 || gw != 1 {
			t.Fatalf("Didn't receive all the Deleted events")
		}
		results <- resource.Event{}
	}()

	wait := time.NewTimer(1 * time.Minute).C
	select {
	case <-wait:
		t.Fatalf("timed out waiting for all events")
	case r := <-results:
		if tot != 4 {
			t.Fatalf("too few results: %v", r)
		}
	}
	fmt.Println("event testDeleted pass")

}


func testNoChange(t *testing.T, events chan resource.Event){
	go func() {
		for {
			e := <-events
			fmt.Println("received event ", e)
			switch e.Kind {
			default:
				t.Fatalf("Unexpected event received")
			}
		}
	}()

	wait := time.NewTimer(1 * time.Minute).C
	<-wait
	fmt.Println("event testNoChange pass")
}


