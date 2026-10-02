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

package check_test

import (
	"context"
	"testing"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checktest"
	"github.com/stretchr/testify/require"
)

func TestModuleName(t *testing.T) {
	t.Parallel()
	const ruleID = "TEST_MODULE_NAME"
	spec := &check.Spec{
		Rules: []*check.RuleSpec{
			{
				ID:      ruleID,
				Default: true,
				Purpose: "Purpose.",
				Type:    check.RuleTypeLint,
				Handler: check.RuleHandlerFunc(
					func(_ context.Context, responseWriter check.ResponseWriter, request check.Request) error {
						for _, fileDescriptor := range request.FileDescriptors() {
							message := "<none>"
							if moduleName := fileDescriptor.ModuleName(); moduleName != nil {
								message = moduleName.String()
							}
							responseWriter.AddAnnotation(
								check.WithMessage(message),
								check.WithFileName(fileDescriptor.ProtoreflectFileDescriptor().Path()),
							)
						}
						return nil
					},
				),
			},
		},
	}

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:   []string{"testdata/module_name"},
				FilePaths:  []string{"a.proto"},
				ModuleName: "buf.build/acme/weather",
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:       ruleID,
				Message:      "buf.build/acme/weather",
				FileLocation: &checktest.ExpectedFileLocation{FileName: "a.proto", EndLine: 8, EndColumn: 1},
			},
			{
				RuleID:       ruleID,
				Message:      "<none>",
				FileLocation: &checktest.ExpectedFileLocation{FileName: "b.proto", EndLine: 4, EndColumn: 12},
			},
		},
	}.Run(t)
}

func TestModuleNameInvalidProtoFileSpec(t *testing.T) {
	t.Parallel()

	_, err := (&checktest.ProtoFileSpec{
		DirPaths:   []string{"testdata/module_name"},
		FilePaths:  []string{"a.proto"},
		ModuleName: "buf.build/acme",
	}).ToFileDescriptors(t.Context())
	require.ErrorContains(t, err, `invalid module name "buf.build/acme"`)
}
