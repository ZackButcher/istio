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
	"path"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"istio.io/istio/pilot/pkg/model"
)

func SocketAddress(address string) *core.Address {
	return &core.Address{
		Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Address: address,
			},
		},
	}
}

func TlsContext(certDir string, sans []string) *auth.CommonTlsContext {
	return &auth.CommonTlsContext{
		TlsCertificates: []*auth.TlsCertificate{
			{
				CertificateChain: FileDataSource(path.Join(certDir, model.CertChainFilename)),
				PrivateKey:       FileDataSource(path.Join(certDir, model.KeyFilename)),
			},
		},
		ValidationContext: &auth.CertificateValidationContext{
			TrustedCa:            FileDataSource(path.Join(certDir, model.RootCertFilename)),
			VerifySubjectAltName: sans,
		},
		AlpnProtocols: ListenersALPNProtocols,
	}
}

func FileDataSource(path string) *core.DataSource {
	return &core.DataSource{
		Specifier: &core.DataSource_Filename{
			Filename: path,
		},
	}
}
