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

package check

import (
	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/info"
	checkv1pluginrpc "buf.build/go/bufplugin/internal/gen/buf/plugin/check/v1/v1pluginrpc"
	infov1pluginrpc "buf.build/go/bufplugin/internal/gen/buf/plugin/info/v1/v1pluginrpc"
	"pluginrpc.com/pluginrpc"
)

// NewServer is a convenience function that creates a new pluginrpc.Server for
// the given Spec.
//
// This registers:
//
// - The Check RPC on the command "check" (if spec.Check is present).
// - The ListRules RPC on the command "list-rules" (if spec.Check is present).
// - The ListCategories RPC on the command "list-categories" (if spec.Check is present).
// - The Generate RPC on the command "generate" (if spec.Generate is present).
// - The GetPluginInfo RPC on the command "info" (if spec.Info is present).
func NewServer(spec *Spec, options ...ServerOption) (pluginrpc.Server, error) {
	serverOptions := newServerOptions()
	for _, option := range options {
		option(serverOptions)
	}

	if err := ValidateSpec(spec); err != nil {
		return nil, err
	}

	var pluginrpcSpecs []pluginrpc.Spec

	if spec.Check != nil {
		pluginrpcSpec, err := checkv1pluginrpc.CheckServiceSpecBuilder{
			Check:          []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("check")},
			ListRules:      []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("list-rules")},
			ListCategories: []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("list-categories")},
		}.Build()
		pluginrpcSpecs = append(pluginrpcSpecs, pluginrpcSpec)

		checkServiceHandler, err := check.NewCheckServiceHandler(
			spec,
			check.CheckServiceHandlerWithParallelism(serverOptions.parallelism),
		)
		if err != nil {
			return nil, err
		}
		handler := pluginrpc.NewHandler(pluginrpcSpec)
		checkServiceServer := checkv1pluginrpc.NewCheckServiceServer(handler, checkServiceHandler)
		checkv1pluginrpc.RegisterCheckServiceServer(serverRegistrar, checkServiceServer)
	}

	if spec.Generate != nil {
		pluginrpcSpec, err := generatev1pluginrpc.GenerateServiceSpecBuilder{
			Generate: []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("generate")},
		}.Build()
		pluginrpcSpecs = append(pluginrpcSpecs, pluginrpcSpec)

		generateServiceHandler, err := generate.NewGenerateServiceHandler(
			spec,
			generate.GenerateServiceHandlerWithParallelism(serverOptions.parallelism),
		)
		if err != nil {
			return nil, err
		}
		handler := pluginrpc.NewHandler(pluginrpcSpec)
		generateServiceServer := generatev1pluginrpc.NewGenerateServiceServer(handler, generateServiceHandler)
		generatev1pluginrpc.RegisterGenerateServiceServer(serverRegistrar, generateServiceServer)
	}

	if spec.Info != nil {
		pluginrpcSpec, err := infov1pluginrpc.PluginInfoServiceSpecBuilder{
			GetPluginInfo: []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("info")},
		}.Build()
		if err != nil {
			return nil, err
		}
		pluginrpcSpecs = append(pluginrpcSpecs, pluginrpcSpec)

		pluginInfoServiceHandler, err := info.NewPluginInfoServiceHandler(spec.Info)
		if err != nil {
			return nil, err
		}
		handler := pluginrpc.NewHandler(pluginrpcSpec)
		pluginInfoServiceServer := infov1pluginrpc.NewPluginInfoServiceServer(handler, pluginInfoServiceHandler)
		infov1pluginrpc.RegisterPluginInfoServiceServer(serverRegistrar, pluginInfoServiceServer)
	}
	if err != nil {
		return nil, err
	}

	pluginrpcSpec, err = pluginrpc.MergeSpecs(pluginrpcSpecs...)
	if err != nil {
		return nil, err
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

// ServerWithParallelism returns a new ServerOption that sets the parallelism
// by which Rules will be run.
//
// If this is set to a value >= 1, this many concurrent Rules can be run at the same time.
// A value of 0 indicates the default behavior, which is to use runtime.GOMAXPROCS(0).
//
// A value if < 0 has no effect.
func ServerWithParallelism(parallelism int) ServerOption {
	return func(serverOptions *serverOptions) {
		if parallelism < 0 {
			parallelism = 0
		}
		serverOptions.parallelism = parallelism
	}
}

type serverOptions struct {
	parallelism int
}

func newServerOptions() *serverOptions {
	return &serverOptions{}
}
