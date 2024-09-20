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
	"context"
	"fmt"

	checkv1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1"
	"buf.build/go/bufplugin/internal/gen/buf/plugin/check/v1/v1pluginrpc"
	"buf.build/go/bufplugin/internal/pkg/cache"
	"buf.build/go/bufplugin/internal/pkg/xslices"
	"pluginrpc.com/pluginrpc"
)

const (
	listRulesPageSize      = 250
	listCategoriesPageSize = 250
)

// Client is a client for a custom lint or breaking change plugin.
type Client interface {
	// Check invokes a check using the plugin..
	Check(ctx context.Context, request Request, options ...CheckCallOption) (Response, error)
	// ListRules lists all available Rules from the plugin.
	//
	// The Rules will be sorted by Rule ID.
	// Returns error if duplicate Rule IDs were detected from the underlying source.
	ListRules(ctx context.Context, options ...ListRulesCallOption) ([]Rule, error)
	// ListCategories lists all available Categories from the plugin.
	//
	// The Categories will be sorted by Category ID.
	// Returns error if duplicate Category IDs were detected from the underlying source.
	ListCategories(ctx context.Context, options ...ListCategoriesCallOption) ([]Category, error)

	isClient()
}

// NewClient returns a new Client for the given pluginrpc.Client.
func NewClient(pluginrpcClient pluginrpc.Client, options ...ClientOption) Client {
	clientOptions := newClientOptions()
	for _, option := range options {
		option.applyToClient(clientOptions)
	}
	return newClient(pluginrpcClient, clientOptions.cacheRulesAndCategories)
}

// ClientOption is an option for a new Client.
type ClientOption interface {
	ClientForSpecOption

	applyToClient(opts *clientOptions)
}

// ClientWithCacheRulesAndCategories returns a new ClientOption that will result in the Rules from
// ListRules and the Categories from ListCategories being cached.
//
// The default is to not cache Rules or Categories.
func ClientWithCacheRulesAndCategories() ClientOption {
	return clientWithCacheRulesAndCategoriesOption{}
}

// NewClientForSpec return a new Client that directly uses the given Spec.
//
// This should primarily be used for testing.
func NewClientForSpec(spec *Spec, options ...ClientForSpecOption) (Client, error) {
	clientForSpecOptions := newClientForSpecOptions()
	for _, option := range options {
		option.applyToClientForSpec(clientForSpecOptions)
	}
	checkServiceHandler, err := NewCheckServiceHandler(spec)
	if err != nil {
		return nil, err
	}
	checkServiceServer, err := NewCheckServiceServer(checkServiceHandler)
	if err != nil {
		return nil, err
	}
	return newClient(
		pluginrpc.NewClient(
			pluginrpc.NewServerRunner(checkServiceServer),
		),
		clientForSpecOptions.cacheRulesAndCategories,
	), nil
}

// ClientForSpecOption is an option for a new Client constructed with NewClientForSpec.
type ClientForSpecOption interface {
	applyToClientForSpec(opts *clientForSpecOptions)
}

// CheckCallOption is an option for a Client.Check call.
type CheckCallOption func(*checkCallOptions)

// ListRulesCallOption is an option for a Client.ListRules call.
type ListRulesCallOption func(*listRulesCallOptions)

// ListCategoriesCallOption is an option for a Client.ListCategories call.
type ListCategoriesCallOption func(*listCategoriesCallOptions)

// *** PRIVATE ***

type client struct {
	pluginrpcClient pluginrpc.Client

	cacheRulesAndCategories bool

	// Singleton ordering: rules -> categories -> checkServiceClient
	rules              *cache.Singleton[[]Rule]
	categories         *cache.Singleton[[]Category]
	checkServiceClient *cache.Singleton[v1pluginrpc.CheckServiceClient]
}

func newClient(
	pluginrpcClient pluginrpc.Client,
	cacheRulesAndCategories bool,
) *client {
	client := &client{
		pluginrpcClient:         pluginrpcClient,
		cacheRulesAndCategories: cacheRulesAndCategories,
	}
	client.rules = cache.NewSingleton(client.listRulesUncached)
	client.categories = cache.NewSingleton(client.listCategoriesUncached)
	client.checkServiceClient = cache.NewSingleton(client.getCheckServiceClientUncached)
	return client
}

func (c *client) Check(ctx context.Context, request Request, _ ...CheckCallOption) (Response, error) {
	checkServiceClient, err := c.checkServiceClient.Get(ctx)
	if err != nil {
		return nil, err
	}
	multiResponseWriter, err := newMultiResponseWriter(request)
	if err != nil {
		return nil, err
	}
	protoRequests, err := request.toProtos()
	if err != nil {
		return nil, err
	}
	for _, protoRequest := range protoRequests {
		protoResponse, err := checkServiceClient.Check(ctx, protoRequest)
		if err != nil {
			return nil, err
		}
		for _, protoAnnotation := range protoResponse.GetAnnotations() {
			multiResponseWriter.addAnnotation(
				protoAnnotation.GetRuleId(),
				WithMessage(protoAnnotation.GetMessage()),
				WithFileNameAndSourcePath(
					protoAnnotation.GetFileLocation().GetFileName(),
					protoAnnotation.GetFileLocation().GetSourcePath(),
				),
				WithAgainstFileNameAndSourcePath(
					protoAnnotation.GetAgainstFileLocation().GetFileName(),
					protoAnnotation.GetAgainstFileLocation().GetSourcePath(),
				),
			)
		}
	}
	return multiResponseWriter.toResponse()
}

func (c *client) ListRules(ctx context.Context, _ ...ListRulesCallOption) ([]Rule, error) {
	return c.rules.Get(ctx)
}

func (c *client) ListCategories(ctx context.Context, _ ...ListCategoriesCallOption) ([]Category, error) {
	return c.categories.Get(ctx)
}

func (c *client) listRulesUncached(ctx context.Context) ([]Rule, error) {
	checkServiceClient, err := c.checkServiceClient.Get(ctx)
	if err != nil {
		return nil, err
	}
	var protoRules []*checkv1.Rule
	var pageToken string
	for {
		response, err := checkServiceClient.ListRules(
			ctx,
			&checkv1.ListRulesRequest{
				PageSize:  listRulesPageSize,
				PageToken: pageToken,
			},
		)
		if err != nil {
			return nil, err
		}
		protoRules = append(protoRules, response.GetRules()...)
		pageToken = response.GetNextPageToken()
		if pageToken == "" {
			break
		}
	}

	// We acquire rules before categories.
	categories, err := c.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	categoryIDToCategory := make(map[string]Category)
	for _, category := range categories {
		// We know there are no duplicate IDs from validation.
		categoryIDToCategory[category.ID()] = category
	}
	rules, err := xslices.MapError(
		protoRules,
		func(protoRule *checkv1.Rule) (Rule, error) {
			return ruleForProtoRule(protoRule, categoryIDToCategory)
		},
	)
	if err != nil {
		return nil, err
	}
	if err := validateRules(rules); err != nil {
		return nil, err
	}
	sortRules(rules)
	return rules, nil
}

func (c *client) listCategoriesUncached(ctx context.Context) ([]Category, error) {
	checkServiceClient, err := c.checkServiceClient.Get(ctx)
	if err != nil {
		return nil, err
	}
	var protoCategories []*checkv1.Category
	var pageToken string
	for {
		response, err := checkServiceClient.ListCategories(
			ctx,
			&checkv1.ListCategoriesRequest{
				PageSize:  listCategoriesPageSize,
				PageToken: pageToken,
			},
		)
		if err != nil {
			return nil, err
		}
		protoCategories = append(protoCategories, response.GetCategories()...)
		pageToken = response.GetNextPageToken()
		if pageToken == "" {
			break
		}
	}
	categories, err := xslices.MapError(protoCategories, categoryForProtoCategory)
	if err != nil {
		return nil, err
	}
	if err := validateCategories(categories); err != nil {
		return nil, err
	}
	sortCategories(categories)
	return categories, nil
}

func (c *client) getCheckServiceClientUncached(ctx context.Context) (v1pluginrpc.CheckServiceClient, error) {
	spec, err := c.pluginrpcClient.Spec(ctx)
	if err != nil {
		return nil, err
	}
	// All of these procedures are required for a plugin to be considered a buf plugin.
	for _, procedurePath := range []string{
		v1pluginrpc.CheckServiceCheckPath,
		v1pluginrpc.CheckServiceListRulesPath,
		v1pluginrpc.CheckServiceListCategoriesPath,
	} {
		if spec.ProcedureForPath(procedurePath) == nil {
			return nil, fmt.Errorf("plugin spec not implemented: RPC %q not found", procedurePath)
		}
	}
	return v1pluginrpc.NewCheckServiceClient(c.pluginrpcClient)
}

func (*client) isClient() {}

type clientOptions struct {
	cacheRulesAndCategories bool
}

func newClientOptions() *clientOptions {
	return &clientOptions{}
}

type clientForSpecOptions struct {
	cacheRulesAndCategories bool
}

func newClientForSpecOptions() *clientForSpecOptions {
	return &clientForSpecOptions{}
}

type clientWithCacheRulesAndCategoriesOption struct{}

func (clientWithCacheRulesAndCategoriesOption) applyToClient(clientOptions *clientOptions) {
	clientOptions.cacheRulesAndCategories = true
}

func (clientWithCacheRulesAndCategoriesOption) applyToClientForSpec(clientForSpecOptions *clientForSpecOptions) {
	clientForSpecOptions.cacheRulesAndCategories = true
}

type checkCallOptions struct{}

type listRulesCallOptions struct{}

type listCategoriesCallOptions struct{}
