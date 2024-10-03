// Copyright 2024 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generate

import (
	"buf.build/go/bufplugin/info"
	generatev1pluginrpc "buf.build/go/bufplugin/internal/gen/buf/plugin/generate/v1/v1pluginrpc"
	infov1pluginrpc "buf.build/go/bufplugin/internal/gen/buf/plugin/info/v1/v1pluginrpc"
	"pluginrpc.com/pluginrpc"
)

// NewServer is a convenience function that creates a new pluginrpc.Server for
// the given Spec.
//
// This registers:
//
// - The Generate RPC on the command "generate".
// - The GetPluginInfo RPC on the command "info" (if spec.Info is present).
func NewServer(spec *Spec, options ...ServerOption) (pluginrpc.Server, error) {
	serverOptions := newServerOptions()
	for _, option := range options {
		option(serverOptions)
	}

	generateServiceHandler, err := NewGenerateServiceHandler(spec, GenerateServiceHandlerWithParallelism(serverOptions.parallelism))
	if err != nil {
		return nil, err
	}
	var pluginInfoServiceHandler infov1pluginrpc.PluginInfoServiceHandler
	if spec.Info != nil {
		pluginInfoServiceHandler, err = info.NewPluginInfoServiceHandler(spec.Info)
		if err != nil {
			return nil, err
		}
	}
	pluginrpcSpec, err := generatev1pluginrpc.GenerateServiceSpecBuilder{
		Generate: []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("generate")},
	}.Build()
	if err != nil {
		return nil, err
	}
	if pluginInfoServiceHandler != nil {
		pluginrpcInfoSpec, err := infov1pluginrpc.PluginInfoServiceSpecBuilder{
			GetPluginInfo: []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("info")},
		}.Build()
		if err != nil {
			return nil, err
		}
		pluginrpcSpec, err = pluginrpc.MergeSpecs(pluginrpcSpec, pluginrpcInfoSpec)
		if err != nil {
			return nil, err
		}
	}

	serverRegistrar := pluginrpc.NewServerRegistrar()
	handler := pluginrpc.NewHandler(pluginrpcSpec)
	generateServiceServer := generatev1pluginrpc.NewGenerateServiceServer(handler, generateServiceHandler)
	generatev1pluginrpc.RegisterGenerateServiceServer(serverRegistrar, generateServiceServer)
	if pluginInfoServiceHandler != nil {
		pluginInfoServiceServer := infov1pluginrpc.NewPluginInfoServiceServer(handler, pluginInfoServiceHandler)
		infov1pluginrpc.RegisterPluginInfoServiceServer(serverRegistrar, pluginInfoServiceServer)
	}

	// Add documentation to -h/--help.
	var pluginrpcServerOptions []pluginrpc.ServerOption
	if spec.Info != nil {
		pluginInfo, err := info.NewPluginInfoForSpec(spec.Info)
		if err != nil {
			return nil, err
		}
		if doc := pluginInfo.Doc(); doc != nil {
			pluginrpcServerOptions = append(
				pluginrpcServerOptions,
				pluginrpc.ServerWithDoc(doc.String()),
			)
		}
	}
	return pluginrpc.NewServer(pluginrpcSpec, serverRegistrar, pluginrpcServerOptions...)
}

// ServerOption is an option for Server.
type ServerOption func(*serverOptions)

type serverOptions struct{}
