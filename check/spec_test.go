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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSpec(t *testing.T) {
	t.Parallel()

	validateRuleSpecError := &validateRuleSpecError{}
	validateCategorySpecError := &validateCategorySpecError{}
	validateSpecError := &validateSpecError{}

	// Simple spec that passes validation.
	spec := &Spec{
		Rules: []*RuleSpec{
			testNewSimpleLintRuleSpec("RULE1", nil, true, false, nil),
			testNewSimpleLintRuleSpec("RULE2", []string{"CATEGORY1"}, true, false, nil),
			testNewSimpleLintRuleSpec("RULE3", []string{"CATEGORY1", "CATEGORY2"}, true, false, nil),
		},
		Categories: []*CategorySpec{
			testNewSimpleCategorySpec("CATEGORY1", false, nil),
			testNewSimpleCategorySpec("CATEGORY2", false, nil),
		},
	}
	require.NoError(t, ValidateSpec(spec))

	// More complicated spec with deprecated rules and categories that passes validation.
	spec = &Spec{
		Rules: []*RuleSpec{
			testNewSimpleLintRuleSpec("RULE1", nil, true, false, nil),
			testNewSimpleLintRuleSpec("RULE2", []string{"CATEGORY1"}, true, false, nil),
			testNewSimpleLintRuleSpec("RULE3", []string{"CATEGORY1", "CATEGORY2"}, true, false, nil),
			testNewSimpleLintRuleSpec("RULE4", []string{"CATEGORY1"}, false, true, []string{"RULE1"}),
			testNewSimpleLintRuleSpec("RULE5", []string{"CATEGORY3", "CATEGORY4"}, false, true, []string{"RULE2", "RULE3"}),
		},
		Categories: []*CategorySpec{
			testNewSimpleCategorySpec("CATEGORY1", false, nil),
			testNewSimpleCategorySpec("CATEGORY2", false, nil),
			testNewSimpleCategorySpec("CATEGORY3", true, []string{"CATEGORY1"}),
			testNewSimpleCategorySpec("CATEGORY4", true, []string{"CATEGORY1", "CATEGORY2"}),
		},
	}
	require.NoError(t, ValidateSpec(spec))

	// Spec that has rules with categories with no resulting category spec.
	spec = &Spec{
		Rules: []*RuleSpec{
			testNewSimpleLintRuleSpec("RULE1", nil, true, false, nil),
			testNewSimpleLintRuleSpec("RULE2", []string{"CATEGORY1"}, true, false, nil),
			testNewSimpleLintRuleSpec("RULE3", []string{"CATEGORY1", "CATEGORY2"}, true, false, nil),
		},
		Categories: []*CategorySpec{
			testNewSimpleCategorySpec("CATEGORY1", false, nil),
		},
	}
	require.ErrorAs(t, ValidateSpec(spec), &validateRuleSpecError)

	// Spec that has categories with no rules with those categories.
	spec = &Spec{
		Rules: []*RuleSpec{
			testNewSimpleLintRuleSpec("RULE1", nil, true, false, nil),
			testNewSimpleLintRuleSpec("RULE2", []string{"CATEGORY1"}, true, false, nil),
			testNewSimpleLintRuleSpec("RULE3", []string{"CATEGORY1", "CATEGORY2"}, true, false, nil),
		},
		Categories: []*CategorySpec{
			testNewSimpleCategorySpec("CATEGORY1", false, nil),
			testNewSimpleCategorySpec("CATEGORY2", false, nil),
			testNewSimpleCategorySpec("CATEGORY3", false, nil),
			testNewSimpleCategorySpec("CATEGORY4", false, nil),
		},
	}
	require.ErrorAs(t, ValidateSpec(spec), &validateCategorySpecError)

	// Spec that has overlapping rules and categories.
	spec = &Spec{
		Rules: []*RuleSpec{
			testNewSimpleLintRuleSpec("RULE1", nil, true, false, nil),
			testNewSimpleLintRuleSpec("RULE2", []string{"CATEGORY1"}, true, false, nil),
			testNewSimpleLintRuleSpec("RULE3", []string{"CATEGORY1", "CATEGORY2"}, true, false, nil),
		},
		Categories: []*CategorySpec{
			testNewSimpleCategorySpec("CATEGORY1", false, nil),
			testNewSimpleCategorySpec("CATEGORY2", false, nil),
			testNewSimpleCategorySpec("RULE3", false, nil),
		},
	}
	require.ErrorAs(t, ValidateSpec(spec), &validateSpecError)

	// Spec that has deprecated rules that point to deprecated rules.
	spec = &Spec{
		Rules: []*RuleSpec{
			testNewSimpleLintRuleSpec("RULE1", nil, true, false, nil),
			testNewSimpleLintRuleSpec("RULE2", nil, false, true, []string{"RULE1"}),
			testNewSimpleLintRuleSpec("RULE3", nil, false, true, []string{"RULE2"}),
		},
	}
	require.ErrorAs(t, ValidateSpec(spec), &validateRuleSpecError)

	// Spec that has deprecated rules that are defaults.
	spec = &Spec{
		Rules: []*RuleSpec{
			testNewSimpleLintRuleSpec("RULE1", nil, true, true, nil),
		},
	}
	require.ErrorAs(t, ValidateSpec(spec), &validateRuleSpecError)

	// Spec that has deprecated categories that point to deprecated categories.
	spec = &Spec{
		Rules: []*RuleSpec{
			testNewSimpleLintRuleSpec("RULE1", nil, true, false, nil),
			testNewSimpleLintRuleSpec("RULE2", []string{"CATEGORY1"}, true, false, nil),
			testNewSimpleLintRuleSpec("RULE3", []string{"CATEGORY1", "CATEGORY2", "CATEGORY3"}, true, false, nil),
		},
		Categories: []*CategorySpec{
			testNewSimpleCategorySpec("CATEGORY1", false, nil),
			testNewSimpleCategorySpec("CATEGORY2", true, []string{"CATEGORY1"}),
			testNewSimpleCategorySpec("CATEGORY3", true, []string{"CATEGORY2"}),
		},
	}
	require.ErrorAs(t, ValidateSpec(spec), &validateCategorySpecError)
}

func testNewSimpleLintRuleSpec(
	id string,
	categoryIDs []string,
	isDefault bool,
	deprecated bool,
	replacementIDs []string,
) *RuleSpec {
	return &RuleSpec{
		ID:             id,
		CategoryIDs:    categoryIDs,
		Default:        isDefault,
		Purpose:        "Checks " + id + ".",
		Type:           RuleTypeLint,
		Deprecated:     deprecated,
		ReplacementIDs: replacementIDs,
		Handler:        nopRuleHandler,
	}
}

func testNewSimpleCategorySpec(
	id string,
	deprecated bool,
	replacementIDs []string,
) *CategorySpec {
	return &CategorySpec{
		ID:             id,
		Purpose:        "Checks " + id + ".",
		Deprecated:     deprecated,
		ReplacementIDs: replacementIDs,
	}
}
