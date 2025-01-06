// Copyright 2024-2025 Buf Technologies, Inc.
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

	"buf.build/go/bufplugin/check/checktest"
)

func TestSpec(t *testing.T) {
	t.Parallel()
	checktest.SpecTest(t, spec)
}

func TestSimpleSuccess(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths: []string{
					"../../proto",
					"testdata/simple_success",
				},
				FilePaths: []string{
					"acme/option/v1/option.proto",
					"simple.proto",
				},
			},
		},
		Spec: spec,
	}.Run(t)
}

func TestSimpleFailure(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths: []string{
					"../../proto",
					"testdata/simple_failure",
				},
				FilePaths: []string{
					"acme/option/v1/option.proto",
					"simple.proto",
				},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID: fieldOptionSafeForMLSetRuleID,
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   8,
					StartColumn: 2,
					EndLine:     8,
					EndColumn:   17,
				},
			},
		},
	}.Run(t)
}

func TestChangeSuccess(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths: []string{
					"../../proto",
					"testdata/change_success/current",
				},
				FilePaths: []string{
					"acme/option/v1/option.proto",
					"simple.proto",
				},
			},
			AgainstFiles: &checktest.ProtoFileSpec{
				DirPaths: []string{
					"../../proto",
					"testdata/change_success/previous",
				},
				FilePaths: []string{
					"acme/option/v1/option.proto",
					"simple.proto",
				},
			},
		},
		Spec: spec,
	}.Run(t)
}

func TestChangeFailure(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths: []string{
					"../../proto",
					"testdata/change_failure/current",
				},
				FilePaths: []string{
					"acme/option/v1/option.proto",
					"simple.proto",
				},
			},
			AgainstFiles: &checktest.ProtoFileSpec{
				DirPaths: []string{
					"../../proto",
					"testdata/change_failure/previous",
				},
				FilePaths: []string{
					"acme/option/v1/option.proto",
					"simple.proto",
				},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID: fieldOptionSafeForMLStaysTrueRuleID,
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   8,
					StartColumn: 2,
					EndLine:     8,
					EndColumn:   56,
				},
				AgainstFileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   8,
					StartColumn: 2,
					EndLine:     8,
					EndColumn:   55,
				},
			},
		},
	}.Run(t)
}
