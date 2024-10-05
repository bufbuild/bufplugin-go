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

	"buf.build/go/bufplugin/internal/gen/buf/plugin/generate/v1/v1pluginrpc"
	"buf.build/go/bufplugin/internal/pkg/cache"
	"pluginrpc.com/pluginrpc"
)

// Client is a client for a custom lint or breaking change plugin.
//
// All calls with pluginrpc.Error with CodeUnimplemented if any procedure is not implemented.
type Client interface {
	// Generate invokes a generate using the plugin..
	Generate(ctx context.Context, request Request, options ...GenerateCallOption) (Response, error)

	isClient()
}

// NewClient returns a new Client for the given pluginrpc.Client.
func NewClient(pluginrpcClient pluginrpc.Client, _ ...ClientOption) Client {
	return newClient(pluginrpcClient, clientOptions.caching)
}

// ClientOption is an option for a new Client.
type ClientOption interface {
	ClientForSpecOption

	applyToClient(opts *clientOptions)
}

// NewClientForSpec return a new Client that directly uses the given Spec.
//
// This should primarily be used for testing.
func NewClientForSpec(spec *Spec, _ ...ClientForSpecOption) (Client, error) {
	server, err := NewServer(spec)
	if err != nil {
		return nil, err
	}
	return newClient(
		pluginrpc.NewClient(
			pluginrpc.NewServerRunner(server),
		),
	), nil
}

// ClientForSpecOption is an option for a new Client constructed with NewClientForSpec.
type ClientForSpecOption interface {
	applyToClientForSpec(opts *clientForSpecOptions)
}

// GenerateCallOption is an option for a Client.Generate call.
type GenerateCallOption func(*generateCallOptions)

// *** PRIVATE ***

type client struct {
	pluginrpcClient pluginrpc.Client

	generateServiceClient *cache.Singleton[v1pluginrpc.GenerateServiceClient]
}

func newClient(
	pluginrpcClient pluginrpc.Client,
) *client {
	client := &client{
		pluginrpcClient: pluginrpcClient,
	}
	client.generateServiceClient = cache.NewSingleton(client.getGenerateServiceClientUncached)
	return client
}

func (c *client) Generate(ctx context.Context, request Request, _ ...GenerateCallOption) (Response, error) {
	generateServiceClient, err := c.generateServiceClient.Get(ctx)
	if err != nil {
		return nil, err
	}
	protoRequest, err := request.toProto()
	if err != nil {
		return nil, err
	}
	protoResponse, err := generateServiceClient.Generate(ctx, protoRequest)
	if err != nil {
		return nil, err
	}
	responseWriter := newResponseWriter()
	for _, protoFile := range protoResponse.GetFiles() {
		writer, err := responseWriter.Put(protoFile.GetPath())
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write(protoFile.GetContent()); err != nil {
			return nil, err
		}
	}
	return responseWriter.toResponse()
}

func (c *client) getGenerateServiceClientUncached(ctx context.Context) (v1pluginrpc.GenerateServiceClient, error) {
	spec, err := c.pluginrpcClient.Spec(ctx)
	if err != nil {
		return nil, err
	}
	// All of these procedures are required for a plugin to be considered a buf plugin.
	for _, procedurePath := range []string{
		v1pluginrpc.GenerateServiceGeneratePath,
	} {
		if spec.ProcedureForPath(procedurePath) == nil {
			return nil, pluginrpc.NewErrorf(pluginrpc.CodeUnimplemented, "procedure unimplemented: %q", procedurePath)
		}
	}
	return v1pluginrpc.NewGenerateServiceClient(c.pluginrpcClient)
}

func (*client) isClient() {}

type clientOptions struct{}

type clientForSpecOptions struct{}

type generateCallOptions struct{}
