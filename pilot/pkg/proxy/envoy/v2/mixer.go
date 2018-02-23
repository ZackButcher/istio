// Copyright 2018 Istio Authors.
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
	"net"
	"net/url"

	xds "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/cluster"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/gogo/protobuf/types"

	"istio.io/api/mesh/v1alpha1"
	mixerpb "istio.io/api/mixer/v1"
	mixc "istio.io/api/mixer/v1/config/client"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/log"
	"istio.io/istio/pkg/proto"
)

const (
	// MixerCheckClusterName is the name of the mixer cluster used for policy checks
	MixerCheckClusterName = "mixer_check_server"

	// MixerReportClusterName is the name of the mixer cluster used for telemetry
	MixerReportClusterName = "mixer_report_server"

	// MixerFilter name and its attributes
	MixerFilter = "mixer"

	// AttrSourcePrefix all source attributes start with this prefix
	AttrSourcePrefix = "source"

	// AttrSourceIP is client source IP
	AttrSourceIP = "source.ip"

	// AttrSourceUID is platform-specific unique identifier for the client instance of the source service
	AttrSourceUID = "source.uid"

	// AttrDestinationPrefix all destination attributes start with this prefix
	AttrDestinationPrefix = "destination"

	// AttrDestinationIP is the server source IP
	AttrDestinationIP = "destination.ip"

	// AttrDestinationUID is platform-specific unique identifier for the server instance of the target service
	AttrDestinationUID = "destination.uid"

	// AttrDestinationLabels is Labels associated with the destination
	AttrDestinationLabels = "destination.labels"

	// AttrDestinationService is name of the target service
	AttrDestinationService = "destination.service"

	// AttrIPSuffix represents IP address suffix.
	AttrIPSuffix = "ip"

	// AttrUIDSuffix is the uid suffix of with source or destination.
	AttrUIDSuffix = "uid"

	// AttrLabelsSuffix is the suffix for labels associated with source or destination.
	AttrLabelsSuffix = "labels"

	// keyConfigMixer is a key in the opaque_config. It is base64(json.Marshal(ServiceConfig)).
	keyConfigMixer = "mixer"

	// keyConfigMixerSha is the sha of keyConfigMixer. It is used for equality check.
	// MixerClient uses it to avoid decoding and processing keyConfigMixer on every request.
	keyConfigMixerSha = "mixer_sha"

	// MixerRequestCount is the quota bucket name
	MixerRequestCount = "RequestCount"

	// MixerCheck switches Check call on and off
	MixerCheck = "mixer_check"

	// MixerReport switches Report call on and off
	MixerReport = "mixer_report"

	// MixerForward switches attribute forwarding on and off
	MixerForward = "mixer_forward"

	// OutboundJWTURIClusterPrefix is the prefix for jwt_uri service
	// clusters external to the proxy instance
	OutboundJWTURIClusterPrefix = "jwt."
)

var (
	orCheckDefault  = altIfEmpty(MixerCheckClusterName)
	orReportDefault = altIfEmpty(MixerReportClusterName)
)

func mixerClusters(ctx context, mixerSAN []string) (clusters []*xds.Cluster) {
	mixerCluster := func(ctx context, mixerSAN []string, mixerAddress *v1alpha1.ServerAddress, clusterName string) *xds.Cluster {
		address := mixerAddress.PlainText
		var tlsContext *auth.UpstreamTlsContext
		if ctx.Mesh.DefaultConfig.ControlPlaneAuthPolicy == v1alpha1.AuthenticationPolicy_MUTUAL_TLS {
			tlsContext = &auth.UpstreamTlsContext{CommonTlsContext: TlsContext(model.AuthCertsPath, mixerSAN)}
			address = mixerAddress.MutualTls
		}

		return &xds.Cluster{
			Name:           clusterName,
			Type:           xds.Cluster_STRICT_DNS,
			ConnectTimeout: proto.ToDuration(ctx.Mesh.ConnectTimeout),
			LbPolicy:       xds.Cluster_ROUND_ROBIN,
			Hosts:          []*core.Address{SocketAddress("tcp://" + address)},
			CircuitBreakers: &cluster.CircuitBreakers{
				Thresholds: []*cluster.CircuitBreakers_Thresholds{{
					MaxPendingRequests: &types.UInt32Value{Value: 10000},
					MaxRequests:        &types.UInt32Value{Value: 10000},
				}},
			},
			Http2ProtocolOptions: &core.Http2ProtocolOptions{}, // Mark mixer as supporting h2
			TlsContext:           tlsContext,
		}
	}

	if ctx.Mesh.MixerCheck != nil {
		clusters = append(clusters, mixerCluster(ctx, mixerSAN, ctx.Mesh.MixerCheck, MixerCheckClusterName))
	}
	if ctx.Mesh.MixerReport != nil {
		clusters = append(clusters, mixerCluster(ctx, mixerSAN, ctx.Mesh.MixerReport, MixerReportClusterName))
	}
	return
}

func buildHTTPMixerFilterConfig(ctx context) *mixc.HttpClientConfig {
	var forwardedAttrs *mixerpb.Attributes
	if ctx.node.Type != model.Sidecar {
		forwardedAttrs = nodeAttributes(AttrSourcePrefix, ctx.node.IPAddress, ctx.node.ID, ctx.service.Labels)
	}

	return &mixc.HttpClientConfig{
		Transport: &mixc.TransportConfig{
			// TODO: how does mTLS need to work here?
			CheckCluster:  orCheckDefault(ctx.Mesh.MixerCheck.MutualTls),
			ReportCluster: orReportDefault(ctx.Mesh.MixerReport.MutualTls),
		},
		ServiceConfigs:            map[string]*mixc.ServiceConfig{ctx.service.Service.Hostname: nil},
		DefaultDestinationService: ctx.service.Service.Hostname,
		MixerAttributes:           nodeAttributes(AttrDestinationPrefix, ctx.node.IPAddress, ctx.node.ID, ctx.service.Labels),
		ForwardAttributes:         forwardedAttrs,
	}
}

func buildTCPMixerFilterConfig(ctx context) *mixc.TcpClientConfig {
	return &mixc.TcpClientConfig{
		Transport: &mixc.TransportConfig{
			CheckCluster:  MixerCheckClusterName,
			ReportCluster: MixerReportClusterName,
		},
		MixerAttributes: &mixerpb.Attributes{
			Attributes: map[string]*mixerpb.Attributes_AttributeValue{
				AttrDestinationIP: {
					Value: &mixerpb.Attributes_AttributeValue_StringValue{
						StringValue: ctx.node.IPAddress,
					},
				},
				AttrDestinationUID: {
					Value: &mixerpb.Attributes_AttributeValue_StringValue{
						StringValue: "kubernetes://" + ctx.node.ID,
					},
				},
				AttrDestinationService: {
					Value: &mixerpb.Attributes_AttributeValue_StringValue{
						StringValue: ctx.service.Service.Hostname,
					},
				},
			},
		},
		DisableCheckCalls:  ctx.Mesh.MixerCheck != nil && ctx.Mesh.DisablePolicyChecks,
		DisableReportCalls: ctx.Mesh.MixerReport != nil,
	}
}

func nodeAttributes(prefix, IPAddress, ID string, labels map[string]string) *mixerpb.Attributes {
	attr := make(map[string]*mixerpb.Attributes_AttributeValue, 3)
	addNodeAttributes(attr, prefix, IPAddress, ID, labels)
	return &mixerpb.Attributes{Attributes: attr}
}

func addNodeAttributes(attr map[string]*mixerpb.Attributes_AttributeValue, prefix string, IPAddress string, ID string, labels map[string]string) {
	if len(IPAddress) > 0 {
		attr[prefix+"."+AttrIPSuffix] = &mixerpb.Attributes_AttributeValue{
			Value: &mixerpb.Attributes_AttributeValue_BytesValue{net.ParseIP(IPAddress)},
		}
	}

	attr[prefix+"."+AttrUIDSuffix] = &mixerpb.Attributes_AttributeValue{
		Value: &mixerpb.Attributes_AttributeValue_StringValue{"kubernetes://" + ID},
	}

	if len(labels) > 0 {
		attr[prefix+"."+AttrLabelsSuffix] = &mixerpb.Attributes_AttributeValue{
			Value: &mixerpb.Attributes_AttributeValue_StringMapValue{
				StringMapValue: &mixerpb.Attributes_StringMap{Entries: labels},
			},
		}
	}
}

func buildServiceConfig(ctx context) *mixc.ServiceConfig {
	return &mixc.ServiceConfig{
		DisableCheckCalls:  ctx.Mesh.MixerCheck != nil && ctx.Mesh.DisablePolicyChecks,
		DisableReportCalls: ctx.Mesh.MixerReport != nil,
		MixerAttributes: &mixerpb.Attributes{
			Attributes: map[string]*mixerpb.Attributes_AttributeValue{
				AttrDestinationService: {
					Value: &mixerpb.Attributes_AttributeValue_StringValue{StringValue: ctx.service.Service.Hostname},
				},
				AttrDestinationLabels: {
					Value: &mixerpb.Attributes_AttributeValue_StringMapValue{
						StringMapValue: &mixerpb.Attributes_StringMap{Entries: ctx.service.Labels},
					},
				},
			},
		},
		HttpApiSpec:      getAPISpecs(ctx),
		QuotaSpec:        getQuotaSpecs(ctx),
		EndUserAuthnSpec: getEndUserAuthSpec(ctx),
	}
}

func getAPISpecs(ctx context) (apiSpecs []*mixc.HTTPAPISpec) {
	specs := ctx.HTTPAPISpecByDestination(ctx.service)
	model.SortHTTPAPISpec(specs)
	for _, config := range specs {
		apiSpecs = append(apiSpecs, config.Spec.(*mixc.HTTPAPISpec))
	}
	return
}

func getQuotaSpecs(ctx context) (quotaSpecs []*mixc.QuotaSpec) {
	specs := ctx.QuotaSpecByDestination(ctx.service)
	model.SortQuotaSpec(specs)
	quotaSpecs = make([]*mixc.QuotaSpec, 0, len(specs))
	for _, config := range specs {
		quotaSpecs = append(quotaSpecs, config.Spec.(*mixc.QuotaSpec))
	}
	return
}

func getEndUserAuthSpec(ctx context) (spec *mixc.EndUserAuthenticationPolicySpec) {
	authSpecs := ctx.EndUserAuthenticationPolicySpecByDestination(ctx.service)
	model.SortEndUserAuthenticationPolicySpec(authSpecs)
	if len(authSpecs) == 0 {
		return
	} else if len(authSpecs) > 1 {
		// TODO - validation should catch this problem earlier at config time.
		log.Warnf("Multiple EndUserAuthenticationPolicySpec found for service %q. Selecting %v", ctx.service.Service, spec)
	}

	spec = (authSpecs[0].Spec).(*mixc.EndUserAuthenticationPolicySpec)
	// Update jwks_uri_envoy_cluster This cluster should be
	// created elsewhere using the same host-to-cluster naming
	// scheme, i.e. buildJWKSURIClusterNameAndAddress.
	for _, jwt := range spec.Jwts {
		if name, _, _, err := buildJWKSURIClusterNameAndAddress(jwt.JwksUri, ctx); err != nil {
			log.Warnf("Could not set jwks_uri_envoy and address for jwks_uri %q: %v", jwt.JwksUri, err)
		} else {
			jwt.JwksUriEnvoyCluster = name
		}
	}
	return
}

// buildJWKSURIClusterNameAndAddress builds the internal envoy cluster
// name and DNS address from the jwks_uri. The cluster name is used by
// the JWT auth filter to fetch public keys. The cluster name and
// address are used to build an envoy cluster that corresponds to the
// jwks_uri server.
// Returns: cluster name, address, whether TLS should be used, and an error
func buildJWKSURIClusterNameAndAddress(raw string, ctx context) (string, string, bool, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", false, err
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	address := host + ":" + port
	name := host + "|" + port

	return truncateClusterName(OutboundJWTURIClusterPrefix+name, ctx), address, u.Scheme == "https", nil
}

// altIfEmpty returns a function that returns defaultVal if the input string is the empty string,
// or the input string itself otherwise.
func altIfEmpty(defaultVal string) func(string) string {
	return func(a string) string {
		if a == "" {
			return defaultVal
		}
		return a
	}
}
