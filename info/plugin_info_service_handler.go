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

package info

import (
	"context"

	infov1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/info/v1"
	"buf.build/go/bufplugin/internal/gen/buf/plugin/info/v1/v1pluginrpc"
	"github.com/bufbuild/protovalidate-go"
)

// NewPluginInfoServiceHandler returns a new v1pluginrpc.PluginInfoServiceHandler for the given Spec.
//
// The Spec will be validated.
func NewPluginInfoServiceHandler(spec *Spec, options ...PluginInfoServiceHandlerOption) (v1pluginrpc.PluginInfoServiceHandler, error) {
	return newPluginInfoServiceHandler(spec, options...)
}

// PluginInfoServiceHandlerOption is an option for PluginInfoServiceHandler.
type PluginInfoServiceHandlerOption func(*pluginInfoServiceHandlerOptions)

// *** PRIVATE ***

type pluginInfoServiceHandler struct {
	protoPluginInfo *infov1.PluginInfo
	spec            *Spec
	validator       *protovalidate.Validator
}

func newPluginInfoServiceHandler(spec *Spec, _ ...PluginInfoServiceHandlerOption) (*pluginInfoServiceHandler, error) {
	// Also calls ValidateSpec.
	pluginInfo, err := NewPluginInfoForSpec(spec)
	if err != nil {
		return nil, err
	}
	protoPluginInfo := pluginInfo.toProto()
	validator, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	if err := validator.Validate(protoPluginInfo); err != nil {
		return nil, err
	}
	return &pluginInfoServiceHandler{
		protoPluginInfo: protoPluginInfo,
	}, nil
}

func (c *pluginInfoServiceHandler) GetPluginInfo(context.Context, *infov1.GetPluginInfoRequest) (*infov1.GetPluginInfoResponse, error) {
	return &infov1.GetPluginInfoResponse{
		PluginInfo: c.protoPluginInfo,
	}, nil
}

type pluginInfoServiceHandlerOptions struct{}
