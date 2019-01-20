// Copyright 2018 Istio Authors
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

package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"

	mcp "istio.io/api/mcp/v1alpha1"
	"istio.io/istio/galley/cmd/shared"
	"istio.io/istio/galley/pkg/fs"
	"istio.io/istio/galley/pkg/kube"
	"istio.io/istio/galley/pkg/kube/converter"
	"istio.io/istio/galley/pkg/kube/source"
	mcps "istio.io/istio/galley/pkg/mcp"
	"istio.io/istio/galley/pkg/meshconfig"
	"istio.io/istio/galley/pkg/metadata"
	"istio.io/istio/galley/pkg/runtime"
	"istio.io/istio/pkg/ctrlz"
	"istio.io/istio/pkg/log"
	"istio.io/istio/pkg/mcp/creds"
	"istio.io/istio/pkg/mcp/server"
	"istio.io/istio/pkg/mcp/snapshot"
	"istio.io/istio/pkg/probe"
	"istio.io/istio/pkg/version"
)

var scope = log.RegisterScope("runtime", "Galley runtime", 0)

// Server is the main entry point into the Galley code.
type Server struct {
	shutdown chan error

	grpcServer *grpc.Server
	processor  *runtime.Processor
	mcp        *server.Server
	listener   net.Listener
	controlZ   *ctrlz.Server
	stopCh     chan struct{}
}

type patchTable struct {
	logConfigure                func(*log.Options) error
	newKubeFromConfigFile       func(string) (kube.Interfaces, error)
	verifyResourceTypesPresence func(kube.Interfaces) error
	newSource                   func(kube.Interfaces, time.Duration, *converter.Config) (runtime.Source, error)
	netListen                   func(network, address string) (net.Listener, error)
	newMeshConfigCache          func(path string) (meshconfig.Cache, error)
	mcpMetricReporter           func(string) server.Reporter
	fsNew                       func(string, *converter.Config) (runtime.Source, error)
	mcpSrcNew                   func(ctx context.Context, copts *creds.Options, mcpAddress, nodeID string) (runtime.Source, error)
}

func defaultPatchTable() patchTable {
	return patchTable{
		logConfigure:                log.Configure,
		newKubeFromConfigFile:       kube.NewKubeFromConfigFile,
		verifyResourceTypesPresence: source.VerifyResourceTypesPresence,
		newSource:                   source.New,
		netListen:                   net.Listen,
		mcpMetricReporter:           func(prefix string) server.Reporter { return server.NewStatsContext(prefix) },
		newMeshConfigCache:          func(path string) (meshconfig.Cache, error) { return meshconfig.NewCacheFromFile(path) },
		fsNew:                       fs.New,
		mcpSrcNew:                   mcps.New,
	}
}

// New returns a new instance of a Server.
func New(a *Args) (*Server, error) {
	return newServer(a, defaultPatchTable())
}

func newServer(a *Args, p patchTable) (*Server, error) {
	s := &Server{}
	var err error
	if err = p.logConfigure(a.LoggingOptions); err != nil {
		return nil, err
	}

	mesh, err := p.newMeshConfigCache(a.MeshConfigFile)
	if err != nil {
		return nil, err
	}
	converterCfg := &converter.Config{Mesh: mesh}

	var src runtime.Source

	// Config Source is exclusive
	if a.ConfigPath != "" {
		src, err = p.fsNew(a.ConfigPath, converterCfg)
		if err != nil {
			return nil, err
		}
	} else if a.SourceMCPServerAddrs != "" {
		src, err = p.mcpSrcNew(context.Background(), a.CredentialOptions, a.SourceMCPServerAddrs, "")
		if err != nil {
			return nil, err
		}
	} else {
		k, err := p.newKubeFromConfigFile(a.KubeConfig)
		if err != nil {
			return nil, err
		}
		if !a.DisableResourceReadyCheck {
			if err := p.verifyResourceTypesPresence(k); err != nil {
				return nil, err
			}
		}
		src, err = p.newSource(k, a.ResyncPeriod, converterCfg)
		if err != nil {
			return nil, err
		}
	}

	processorCfg := runtime.Config{
		DomainSuffix: a.DomainSuffix,
		Mesh:         mesh,
	}
	distributor := snapshot.New(snapshot.DefaultGroupIndex)
	s.processor = runtime.NewProcessor(src, distributor, &processorCfg)

	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(uint32(a.MaxConcurrentStreams)))
	grpcOptions = append(grpcOptions, grpc.MaxRecvMsgSize(int(a.MaxReceivedMessageSize)))

	s.stopCh = make(chan struct{})
	checker := server.NewAllowAllChecker()
	if !a.Insecure {
		checker, err = watchAccessList(s.stopCh, a.AccessListFile)
		if err != nil {
			return nil, err
		}

		watcher, err := creds.WatchFiles(s.stopCh, a.CredentialOptions)
		if err != nil {
			return nil, err
		}
		credentials := creds.CreateForServer(watcher)

		grpcOptions = append(grpcOptions, grpc.Creds(credentials))
	}
	grpc.EnableTracing = a.EnableGRPCTracing
	s.grpcServer = grpc.NewServer(grpcOptions...)

	s.mcp = server.New(distributor, metadata.Types.TypeURLs(), checker, p.mcpMetricReporter("galley/"))

	// get the network stuff setup
	network := "tcp"
	var address string
	idx := strings.Index(a.APIAddress, "://")
	if idx < 0 {
		address = a.APIAddress
	} else {
		network = a.APIAddress[:idx]
		address = a.APIAddress[idx+3:]
	}

	if s.listener, err = p.netListen(network, address); err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("unable to listen: %v", err)
	}

	mcp.RegisterAggregatedMeshConfigServiceServer(s.grpcServer, s.mcp)

	s.controlZ, _ = ctrlz.Run(a.IntrospectionOptions, nil)

	return s, nil
}

// Run enables Galley to start receiving gRPC requests on its main API port.
func (s *Server) Run() {
	s.shutdown = make(chan error, 1)
	go func() {
		err := s.processor.Start()
		if err != nil {
			s.shutdown <- err
			return
		}

		// start serving
		err = s.grpcServer.Serve(s.listener)
		// notify closer we're done
		s.shutdown <- err
	}()
}

// Wait waits for the server to exit.
func (s *Server) Wait() error {
	if s.shutdown == nil {
		return fmt.Errorf("server not running")
	}

	err := <-s.shutdown
	s.shutdown = nil
	return err
}

// Close cleans up resources used by the server.
func (s *Server) Close() error {
	if s.stopCh != nil {
		close(s.stopCh)
		s.stopCh = nil
	}

	if s.shutdown != nil {
		s.grpcServer.GracefulStop()
		_ = s.Wait()
	}

	if s.controlZ != nil {
		s.controlZ.Close()
	}

	if s.processor != nil {
		s.processor.Stop()
	}

	if s.listener != nil {
		_ = s.listener.Close()
	}

	// final attempt to purge buffered logs
	_ = log.Sync()

	return nil
}

//RunServer start Galley Server mode
func RunServer(sa *Args, printf, fatalf shared.FormatFn, livenessProbeController,
	readinessProbeController probe.Controller) {
	printf("Galley started with\n%s", sa)
	s, err := New(sa)
	if err != nil {
		fatalf("Unable to initialize Galley Server: %v", err)
	}
	printf("Istio Galley: %s", version.Info)
	printf("Starting gRPC server on %v", sa.APIAddress)
	s.Run()
	if livenessProbeController != nil {
		serverLivenessProbe := probe.NewProbe()
		serverLivenessProbe.SetAvailable(nil)
		serverLivenessProbe.RegisterProbe(livenessProbeController, "serverLiveness")
		defer serverLivenessProbe.SetAvailable(errors.New("stopped"))
	}
	if readinessProbeController != nil {
		serverReadinessProbe := probe.NewProbe()
		serverReadinessProbe.SetAvailable(nil)
		serverReadinessProbe.RegisterProbe(readinessProbeController, "serverReadiness")
		defer serverReadinessProbe.SetAvailable(errors.New("stopped"))
	}
	err = s.Wait()
	if err != nil {
		fatalf("Galley Server unexpectedly terminated: %v", err)
	}
	_ = s.Close()
}
