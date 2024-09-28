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
	"buf.build/go/bufplugin/internal/gen/buf/plugin/check/v1/v1pluginrpc"
	"pluginrpc.com/pluginrpc"
)

// NewServer is a convenience function that creates a new pluginrpc.Server for
// the given v1pluginrpc.CheckServiceHandler.
//
// This registers the Check RPC on the command "check", the ListRules RPC on the command
// "list-rules", and the ListCategories RPC on the command "list-categories". No options
// are passed to any of the types necessary to create this Server. If further customization
// is necessary, this can be done manually.
func NewCheckServiceServer(checkServiceHandler v1pluginrpc.CheckServiceHandler) (pluginrpc.Server, error) {
	spec, err := v1pluginrpc.CheckServiceSpecBuilder{
		Check:          []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("check")},
		ListRules:      []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("list-rules")},
		ListCategories: []pluginrpc.ProcedureOption{pluginrpc.ProcedureWithArgs("list-categories")},
	}.Build()
	if err != nil {
		return nil, err
	}
	serverRegistrar := pluginrpc.NewServerRegistrar()
	checkServiceServer := v1pluginrpc.NewCheckServiceServer(pluginrpc.NewHandler(spec), checkServiceHandler)
	v1pluginrpc.RegisterCheckServiceServer(serverRegistrar, checkServiceServer)
	return pluginrpc.NewServer(spec, serverRegistrar)
}
