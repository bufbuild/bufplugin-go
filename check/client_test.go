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
	"slices"
	"testing"

	"buf.build/go/bufplugin/internal/pkg/xslices"
	"github.com/stretchr/testify/require"
)

func TestClientListRulesCategoriesSimple(t *testing.T) {
	t.Parallel()

	testClientListRulesCategoriesSimple(t)
	testClientListRulesCategoriesSimple(t, ClientWithCacheRulesAndCategories())
}

func testClientListRulesCategoriesSimple(t *testing.T, options ...ClientForSpecOption) {
	ctx := context.Background()
	client, err := NewClientForSpec(
		&Spec{
			Rules: []*RuleSpec{
				{
					ID:      "RULE1",
					Purpose: "Test RULE1.",
					Type:    RuleTypeLint,
					Handler: nopRuleHandler,
				},
				{
					ID: "RULE2",
					CategoryIDs: []string{
						"CATEGORY1",
					},
					Purpose: "Test RULE2.",
					Type:    RuleTypeLint,
					Handler: nopRuleHandler,
				},
				{
					ID: "RULE3",
					CategoryIDs: []string{
						"CATEGORY1",
						"CATEGORY2",
					},
					Purpose: "Test RULE3.",
					Type:    RuleTypeLint,
					Handler: nopRuleHandler,
				},
			},
			Categories: []*CategorySpec{
				{
					ID:      "CATEGORY1",
					Purpose: "Test CATEGORY1.",
				},
				{
					ID:      "CATEGORY2",
					Purpose: "Test CATEGORY2.",
				},
			},
		},
		options...,
	)
	require.NoError(t, err)
	rules, err := client.ListRules(ctx)
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"RULE1",
			"RULE2",
			"RULE3",
		},
		xslices.Map(rules, Rule.ID),
	)
	categories, err := client.ListCategories(ctx)
	require.NoError(t, err)
	require.Equal(
		t,
		[]string{
			"CATEGORY1",
			"CATEGORY2",
		},
		xslices.Map(categories, Category.ID),
	)
	categories = rules[0].Categories()
	require.Empty(t, categories)
	categories = rules[1].Categories()
	require.Equal(
		t,
		[]string{
			"CATEGORY1",
		},
		xslices.Map(categories, Category.ID),
	)
	categories = rules[2].Categories()
	require.Equal(
		t,
		[]string{
			"CATEGORY1",
			"CATEGORY2",
		},
		xslices.Map(categories, Category.ID),
	)
}

func TestClientListRulesCount(t *testing.T) {
	t.Parallel()

	testClientListRulesCount(t, listRulesPageSize-1)
	testClientListRulesCount(t, listRulesPageSize)
	testClientListRulesCount(t, listRulesPageSize+1)
	testClientListRulesCount(t, listRulesPageSize*2)
	testClientListRulesCount(t, (listRulesPageSize*2)+1)
	testClientListRulesCount(t, (listRulesPageSize*4)+1)
}

func testClientListRulesCount(t *testing.T, count int) {
	require.True(t, count < 10000, "count must be less than 10000 for sorting to work properly in this test")
	ruleSpecs := make([]*RuleSpec, count)
	for i := 0; i < count; i++ {
		ruleSpecs[i] = &RuleSpec{
			ID:      fmt.Sprintf("RULE%05d", i),
			Purpose: fmt.Sprintf("Test RULE%05d.", i),
			Type:    RuleTypeLint,
			Handler: nopRuleHandler,
		}
	}
	// Make the ruleSpecs not in sorted order.
	ruleSpecsOutOfOrder := slices.Clone(ruleSpecs)
	slices.Reverse(ruleSpecsOutOfOrder)
	client, err := NewClientForSpec(&Spec{Rules: ruleSpecsOutOfOrder})
	require.NoError(t, err)
	rules, err := client.ListRules(context.Background())
	require.NoError(t, err)
	require.Equal(t, count, len(rules))
	for i := 0; i < count; i++ {
		require.Equal(t, ruleSpecs[i].ID, rules[i].ID())
	}
}
