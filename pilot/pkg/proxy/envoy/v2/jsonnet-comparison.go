// Copyright 2016 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v2

import (
	"strconv"
	"strings"

	"github.com/envoyproxy/go-control-plane/api"

	"time"
	"fmt"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/proto"
)

func longestSuffix(a, b []string, j int) int {
	if j >= len(a) || j >= len(b) {
		return j
	} else if a[len(a)-1-j] != b[len(b)-1-j] {
		return j
	}
	return longestSuffix(a, b, j+1)
}

func domains(service model.Service, port int, domain string) []string {
	svcNames := strings.Split(service.Hostname, ".")
	ctxNames := strings.Split(domain, ".")
	j := longestSuffix(svcNames, ctxNames, 0)
	portStr := strconv.Itoa(port)

	out := make([]string, 0, 2*(j+1))
	for i := 0; i <= j; i++ {
		domain := strings.Join(svcNames[0:len(svcNames)-1], ".")
		out = append(out, domain)
		out = append(out, domain+":"+portStr)
	}
	if service.Address != "" {
		out = append(out, service.Address)
		out = append(out, service.Address+":"+portStr)
	}
	return out
}

func inboundCluster(port int, protocol model.Protocol) *api.Cluster {
	return &api.Cluster{
		Name:                 "in." + strconv.Itoa(port),
		ConnectTimeout:       5 * time.Second,
		Type:                 api.Cluster_STATIC,
		LbPolicy:             api.Cluster_ROUND_ROBIN,
		Hosts:                []*api.Address{},
		Http2ProtocolOptions: ifH2(protocol),
	}
}

func outboundCluster(hostname string, labels map[string]string, port model.Port) *api.Cluster {
	key := key(hostname, labels, port)
	return &api.Cluster{
		Name: key,
		ConnectTimeout: 5 * time.Second,
		Type: api.Cluster_EDS,
		EdsClusterConfig: &api.Cluster_EdsClusterConfig{
			ServiceName: key   ,
			EdsConfig: &api.ConfigSource{
				ConfigSourceSpecifier: &api.ConfigSource_Ads{
					Ads: &api.AggregatedConfigSource{},
				},
			},
		},
		LbPolicy: api.Cluster_ROUND_ROBIN,
		// TODO: "hostname:: hostname,"
	}
}

func defaultRoute(cluster api.Cluster, operation string) *api.Route {
	return &api.Route{
		Match: &api.RouteMatch{
			PathSpecifier:   &api.RouteMatch_Prefix{
				Prefix: "/",
			},
		},
		Action: &api.Route_Route{
			Route: &api.RouteAction{
				ClusterSpecifier: &api.RouteAction_Cluster{
					Cluster: cluster.Name,
				},
			},
		},
		Decorator: &api.Decorator{
			Operation: operation,
		},
	}
}

func inboundListeners(instances []*model.ServiceInstance) []*api.Listener {
	listeners := make([]*api.Listener, 0, len(instances))
	for _, instance := range instances {
		protocol := instance.Endpoint.ServicePort.Protocol
		port := instance.Endpoint.Port
		cluster := inboundCluster(port, protocol)
		prefix := fmt.Sprintf("in_%s_%d", protocol, port)
		name := fmt.Sprintf("in_%s_%s_%d", protocol, instance.Endpoint.Address, port)

		listeners = append(listeners, &api.Listener{
			Address: &api.Address{
				Address: &api.Address_SocketAddress{
					SocketAddress: &api.SocketAddress{
						Address: instance.Endpoint.Address,
						PortSpecifier: &api.SocketAddress_PortValue{
							PortValue: uint32(port),
						},
					},
				},
			},
			FilterChains: []*api.FilterChain{
				{
					Filters: []*api.Filter{
						{
							Name: "envoy.http_connection_manager",
							Config: proto.MessageToStructMust(&)
						},
					},
				},
			},
		})
	}
}

    inbound_listeners(instances)::
        [{
            local protocol = instance.endpoint.service_port.protocol,
            local port = instance.endpoint.port,
            local cluster = config.inbound_cluster(port, protocol),
            local prefix = "in_%s_%d" % [protocol, port],
            name: "in_%s_%s_%d" % [protocol, instance.endpoint.ip_address, port],
            cluster:: cluster,
            address: {
                socket_address: {
                    address: instance.endpoint.ip_address,
                    port_value: port,
                },
            },
            filter_chains: [
                {
                    filters: [
                        if model.is_http(protocol) then
                            {
                                name: "envoy.http_connection_manager",
                                config: {
                                    stat_prefix: prefix,
                                    codec_type: "AUTO",
                                    access_log: [{
                                        name: "envoy.file_access_log",
                                        config: { path: "/dev/stdout" },
                                    }],
                                    generate_request_id: true,
                                    route_config: {
                                        name: prefix,
                                        virtual_hosts: [{
                                            name: prefix,
                                            domains: ["*"],
                                            routes: [config.default_route(cluster, "inbound_route")],
                                        }],
                                        validate_clusters: false,
                                    },
                                    http_filters: [{
                                        name: "envoy.router",
                                    }],
                                },
                            }
                        else if !model.is_udp(protocol) then
                            {
                                name: "envoy.tcp_proxy",
                                config: {
                                    stat_prefix: prefix,
                                    cluster: cluster.name,
                                },
                            },
                    ],
                },
            ],
        } for instance in instances],

func ifH2(protocol model.Protocol) (opts *api.Http2ProtocolOptions) {
	if protocol == model.ProtocolHTTP2 || protocol == model.ProtocolGRPC {
		opts = &api.Http2ProtocolOptions{}
	}
	return
}

func key(hostname string, labels map[string]string, port model.Port) string {
	lstrs := make([]string, 0, len(labels))
	for k, v := range labels {
		lstrs = append(lstrs, k + "=" + v)
	}
	return fmt.Sprintf("%s|%s|%s", hostname, port.Name, strings.Join(lstrs, ","))
}