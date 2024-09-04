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

package main

import (
	"testing"

	"github.com/bufbuild/bufplugin-go/check/checktest"
)

func TestSpec(t *testing.T) {
	t.Parallel()
	checktest.SpecTest(t, spec)
}

func TestSimple(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/simple"},
				FilePaths: []string{"simple.proto"},
			},
			// This linter only has a single Rule, so this has no effect in this
			// test, however this is how you scope a test to a single Rule.
			RuleIDs: []string{timestampSuffixRuleID},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID: timestampSuffixRuleID,
				Location: &checktest.ExpectedLocation{
					FileName:    "simple.proto",
					StartLine:   8,
					StartColumn: 2,
					EndLine:     8,
					EndColumn:   50,
				},
			},
		},
	}.Run(t)
}

func TestOption(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/option"},
				FilePaths: []string{"option.proto"},
			},
			Options: map[string]any{
				timestampSuffixOptionKey: "_timestamp",
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID: timestampSuffixRuleID,
				Location: &checktest.ExpectedLocation{
					FileName:    "option.proto",
					StartLine:   8,
					StartColumn: 2,
					EndLine:     8,
					EndColumn:   45,
				},
			},
		},
	}.Run(t)
}
