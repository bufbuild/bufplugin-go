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
	"errors"

	"buf.build/go/bufplugin/internal/gen/buf/plugin/check/v1/v1pluginrpc"
	checkv1pluginrpc "buf.build/go/bufplugin/internal/gen/buf/plugin/check/v1/v1pluginrpc"
	infov1pluginrpc "buf.build/go/bufplugin/internal/gen/buf/plugin/info/v1/v1pluginrpc"
	"pluginrpc.com/pluginrpc"
)

// ServerSpec is a specification for a new pluginrpc.Server.
type ServerSpec struct {
	// Required.
	CheckServiceHandler checkv1pluginrpc.CheckServiceHandler
	// Optional.
	PluginInfoServiceHandler infov1pluginrpc.PluginInfoServiceHandler
}

// NewServer is a convenience function that creates a new pluginrpc.Server for
// the given handlers.
//
// This registers:
//
// - The Check RPC on the command "check".
// - The ListRules RPC on the command "list-rules".
// - The ListCategories RPC on the command "list-categories".
// - The GetPluginInfo RPC on the command "info" (if the PluginInfoServiceHandler is present).
//
// No options are passed to any of the types necessary to create this Server. If further
// customization is necessary, this can be done manually.
func NewServer(serverSpec *ServerSpec) (pluginrpc.Server, error) {
	checkServiceHandler := serverSpec.CheckServiceHandler
	pluginInfoServiceHandler := serverSpec.PluginInfoServiceHandler
	if checkServiceHandler == nil {
		return nil, errors.New("ServerSpec.CheckServiceHandler is required")
	}

	spec, err := checkv1pluginrpc.CheckServiceSpecBuilder{
		Check:          []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("check")},
		ListRules:      []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("list-rules")},
		ListCategories: []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("list-categories")},
	}.Build()
	if err != nil {
		return nil, err
	}
	if pluginInfoServiceHandler != nil {
		pluginInfoSpec, err := infov1pluginrpc.PluginInfoServiceSpecBuilder{
			GetPluginInfo: []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("info")},
		}.Build()
		if err != nil {
			return nil, err
		}
		spec, err = pluginrpc.MergeSpecs(spec, pluginInfoSpec)
		if err != nil {
			return nil, err
		}
	}

	serverRegistrar := pluginrpc.NewServerRegistrar()
	handler := pluginrpc.NewHandler(spec)
	checkServiceServer := v1pluginrpc.NewCheckServiceServer(handler, serverSpec.CheckServiceHandler)
	checkv1pluginrpc.RegisterCheckServiceServer(serverRegistrar, checkServiceServer)
	if pluginInfoServiceHandler != nil {
		pluginInfoServiceServer := infov1pluginrpc.NewPluginInfoServiceServer(handler, pluginInfoServiceHandler)
		infov1pluginrpc.RegisterPluginInfoServiceServer(serverRegistrar, pluginInfoServiceServer)
	}

	return pluginrpc.NewServer(spec, serverRegistrar)
}
