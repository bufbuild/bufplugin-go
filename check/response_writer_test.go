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
	"buf.build/go/bufplugin/check/checkutil"
	"buf.build/go/bufplugin/descriptor"
)

// TestSourceLocationFallback checks that an annotation by source path points to the nearest parent if the original one is missing.
func TestSourceLocationFallback(t *testing.T) {
	t.Parallel()
	const fileName = "source_location_fallback.proto"
	const ruleID = "TEST_SOURCE_LOCATION_FALLBACK"
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata"},
				FilePaths: []string{fileName},
			},
			RuleIDs: []string{ruleID},
		},
		Spec: &check.Spec{
			Rules: []*check.RuleSpec{{
				ID:      ruleID,
				Purpose: "Purpose.",
				Type:    check.RuleTypeLint,
				Handler: checkutil.NewFileRuleHandler(
					func(
						_ context.Context,
						writer check.ResponseWriter,
						_ check.Request,
						file descriptor.FileDescriptor,
					) error {
						foo := file.ProtoreflectFileDescriptor().Messages().ByName("Foo")
						fooPath := file.ProtoreflectFileDescriptor().SourceLocations().ByDescriptor(foo).Path
						writer.AddAnnotation(
							check.WithMessage("Annotation for message"),
							check.WithFileNameAndSourcePath(
								fileName,
								fooPath,
							),
						)
						writer.AddAnnotation(
							check.WithMessage("Annotation for option (x).y"),
							check.WithFileNameAndSourcePath(
								fileName,
								append(
									fooPath,
									7,    // google.protobuf.DescriptorProto.options
									5000, // x
									1,    // X.y
								),
							),
						)
						writer.AddAnnotation(
							check.WithMessage("Annotation for option (x) points to first message option"),
							check.WithFileNameAndSourcePath(
								fileName,
								append(
									fooPath,
									7,    // google.protobuf.DescriptorProto.options
									5000, // x
								),
							),
						)
						return nil
					},
					checkutil.WithoutImports(),
				),
			}},
		},
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  ruleID,
				Message: "Annotation for message",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    fileName,
					StartLine:   12,
					StartColumn: 0,
					EndLine:     15,
					EndColumn:   1,
				},
			},
			{
				RuleID:  ruleID,
				Message: "Annotation for option (x) points to first message option",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    fileName,
					StartLine:   13,
					StartColumn: 2,
					EndLine:     13,
					EndColumn:   27,
				},
			},
			{
				RuleID:  ruleID,
				Message: "Annotation for option (x).y",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    fileName,
					StartLine:   14,
					StartColumn: 2,
					EndLine:     14,
					EndColumn:   21,
				},
			},
		},
	}.Run(t)
}
