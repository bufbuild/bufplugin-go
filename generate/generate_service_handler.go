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
	"context"

	generatev1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/generate/v1"
	"buf.build/go/bufplugin/internal/gen/buf/plugin/generate/v1/v1pluginrpc"
	"github.com/bufbuild/protovalidate-go"
	"pluginrpc.com/pluginrpc"
)

// NewGenerateServiceHandler returns a new v1pluginrpc.GenerateServiceHandler for the given Spec.
//
// The Spec will be validated.
func NewGenerateServiceHandler(spec *Spec, options ...GenerateServiceHandlerOption) (v1pluginrpc.GenerateServiceHandler, error) {
	return newGenerateServiceHandler(spec, options...)
}

// GenerateServiceHandlerOption is an option for GenerateServiceHandler.
type GenerateServiceHandlerOption func(*generateServiceHandlerOptions)

// *** PRIVATE ***

type generateServiceHandler struct {
	spec      *Spec
	validator *protovalidate.Validator
}

func newGenerateServiceHandler(spec *Spec, _ ...GenerateServiceHandlerOption) (*generateServiceHandler, error) {
	if err := ValidateSpec(spec); err != nil {
		return nil, err
	}
	validator, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &generateServiceHandler{
		spec:      spec,
		validator: validator,
	}, nil
}

func (c *generateServiceHandler) Generate(
	ctx context.Context,
	generateRequest *generatev1.GenerateRequest,
) (*generatev1.GenerateResponse, error) {
	if err := c.validator.Validate(generateRequest); err != nil {
		return nil, pluginrpc.NewError(pluginrpc.CodeInvalidArgument, err)
	}
	request, err := RequestForProtoRequest(generateRequest)
	if err != nil {
		return nil, err
	}
	if c.spec.Before != nil {
		ctx, request, err = c.spec.Before(ctx, request)
		if err != nil {
			return nil, err
		}
	}
	response, err := responseWriter.toResponse()
	if err != nil {
		return nil, err
	}
	generateResponse := response.toProto()
	if err := c.validator.Validate(generateResponse); err != nil {
		return nil, err
	}
	return generateResponse, nil
}

type generateServiceHandlerOptions struct{}
