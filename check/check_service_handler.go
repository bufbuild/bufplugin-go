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

	checkv1beta1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1beta1"
	"github.com/bufbuild/bufplugin-go/internal/pkg/xslices"
	"github.com/bufbuild/pluginrpc-go"
	"github.com/bufbuild/protovalidate-go"
)

const defaultPageSize = 250

// *** PRIVATE ***

type checkServiceHandler struct {
	spec             *Spec
	ruleIDToRuleSpec map[string]*RuleSpec
	ruleIDToIndex    map[string]int
}

func newCheckServiceHandler(spec *Spec) (*checkServiceHandler, error) {
	validator, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	if err := validateSpec(validator, spec); err != nil {
		return nil, err
	}
	ruleIDToRuleSpec := make(map[string]*RuleSpec, len(spec.Rules))
	ruleIDToIndex := make(map[string]int, len(spec.Rules))
	for i, ruleSpec := range spec.Rules {
		id := ruleSpec.ID
		if _, ok := ruleIDToRuleSpec[id]; ok {
			return nil, fmt.Errorf("duplicate Rule ID: %q", id)
		}
		ruleIDToRuleSpec[id] = ruleSpec
		ruleIDToIndex[id] = i
	}
	return &checkServiceHandler{
		spec:             spec,
		ruleIDToRuleSpec: ruleIDToRuleSpec,
		ruleIDToIndex:    ruleIDToIndex,
	}, nil
}

func (c *checkServiceHandler) Check(
	ctx context.Context,
	checkRequest *checkv1beta1.CheckRequest,
) (*checkv1beta1.CheckResponse, error) {
	request, err := RequestForProtoRequest(checkRequest)
	if err != nil {
		return nil, err
	}
	if c.spec.Before != nil {
		ctx, request, err = c.spec.Before(ctx, request)
		if err != nil {
			return nil, err
		}
	}
	ruleSpecs := c.spec.Rules
	if ruleIDs := request.RuleIDs(); len(ruleIDs) > 0 {
		ruleSpecs = make([]*RuleSpec, 0)
		for _, ruleID := range ruleIDs {
			ruleSpec, ok := c.ruleIDToRuleSpec[ruleID]
			if !ok {
				return nil, pluginrpc.NewErrorf(pluginrpc.CodeInvalidArgument, "unknown rule ID: %q", ruleID)
			}
			ruleSpecs = append(ruleSpecs, ruleSpec)
		}
	}
	multiResponseWriter, err := newMultiResponseWriter(request)
	if err != nil {
		return nil, err
	}
	for _, ruleSpec := range ruleSpecs {
		if err := ruleSpec.Handler.Handle(
			ctx,
			multiResponseWriter.newResponseWriter(ruleSpec.ID),
			request,
		); err != nil {
			return nil, err
		}
	}
	response, err := multiResponseWriter.toResponse()
	if err != nil {
		return nil, err
	}
	return response.toProto(), nil
}

func (c *checkServiceHandler) ListRules(_ context.Context, listRulesRequest *checkv1beta1.ListRulesRequest) (*checkv1beta1.ListRulesResponse, error) {
	ruleSpecs, nextPageToken, err := c.getRuleSpecsAndNextPageToken(
		int(listRulesRequest.GetPageSize()),
		listRulesRequest.GetPageToken(),
	)
	if err != nil {
		return nil, err
	}
	protoRules := xslices.Map(
		xslices.Map(
			ruleSpecs,
			// Assumes validated.
			ruleSpecToRule,
		),
		Rule.toProto,
	)
	return &checkv1beta1.ListRulesResponse{
		NextPageToken: nextPageToken,
		Rules:         protoRules,
	}, nil
}

func (c *checkServiceHandler) getRuleSpecsAndNextPageToken(pageSize int, pageToken string) ([]*RuleSpec, string, error) {
	index := 0
	if pageToken != "" {
		var ok bool
		index, ok = c.ruleIDToIndex[pageToken]
		if !ok {
			return nil, "", pluginrpc.NewErrorf(pluginrpc.CodeInvalidArgument, "unknown page token: %q", pageToken)
		}
	}
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	var resultRuleSpecs []*RuleSpec
	for i := 0; i < pageSize; i++ {
		if index >= len(c.spec.Rules) {
			break
		}
		resultRuleSpecs = append(resultRuleSpecs, c.spec.Rules[index])
		index++
	}
	var nextPageToken string
	if index < len(c.spec.Rules) {
		nextPageToken = c.spec.Rules[index].ID
	}
	return resultRuleSpecs, nextPageToken, nil
}
