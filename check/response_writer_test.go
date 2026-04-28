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

// TestSourceLocationFallback checks that an annotation by source path points to the nearest location if the original one is missing.
func TestSourceLocationFallback(t *testing.T) {
	t.Parallel()
	const file = "source_location_fallback.proto"
	const ruleID = "TEST_SOURCE_LOCATION_FALLBACK"
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata"},
				FilePaths: []string{file},
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
						fileDescriptor descriptor.FileDescriptor,
					) error {
						const optsNum = 7 // google.protobuf.DescriptorProto.options
						const aNum = 5000 // a
						const abNum = 1   // A.b
						const acNum = 2   // A.c
						const acdNum = 3  // C.d

						fooMsg := fileDescriptor.ProtoreflectFileDescriptor().Messages().ByName("Foo")
						foo := fileDescriptor.ProtoreflectFileDescriptor().SourceLocations().ByDescriptor(fooMsg).Path
						writer.AddAnnotation(
							check.WithMessage("Foo - message - right location"),
							check.WithFileNameAndSourcePath(file, foo),
						)
						writer.AddAnnotation(
							check.WithMessage("Foo - option (a) - fallback to (a).b"),
							check.WithFileNameAndSourcePath(file, append(foo, optsNum, aNum)),
						)
						writer.AddAnnotation(
							check.WithMessage("Foo - option (a).b - right location"),
							check.WithFileNameAndSourcePath(file, append(foo, optsNum, aNum, abNum)),
						)
						writer.AddAnnotation(
							check.WithMessage("Foo - option (a).c - fallback to (a).c.d"),
							check.WithFileNameAndSourcePath(file, append(foo, optsNum, aNum, acNum)),
						)
						writer.AddAnnotation(
							check.WithMessage("Foo - option (a).c.d - right location"),
							check.WithFileNameAndSourcePath(file, append(foo, optsNum, aNum, acNum, acdNum)),
						)

						barMsg := fileDescriptor.ProtoreflectFileDescriptor().Messages().ByName("Bar")
						bar := fileDescriptor.ProtoreflectFileDescriptor().SourceLocations().ByDescriptor(barMsg).Path
						writer.AddAnnotation(
							check.WithMessage("Bar - option (a) - fallback to (a).c"),
							check.WithFileNameAndSourcePath(file, append(bar, optsNum, aNum)),
						)
						writer.AddAnnotation(
							check.WithMessage("Bar - option (a).c - right location"),
							check.WithFileNameAndSourcePath(file, append(bar, optsNum, aNum, acNum)),
						)
						writer.AddAnnotation(
							check.WithMessage("Bar - option (a).c.d - right location"),
							check.WithFileNameAndSourcePath(file, append(bar, optsNum, aNum, acNum, acdNum)),
						)
						writer.AddAnnotation(
							check.WithMessage("Bar - option (a).b - right location"),
							check.WithFileNameAndSourcePath(file, append(bar, optsNum, aNum, abNum)),
						)
						return nil
					},
					checkutil.WithoutImports(),
				),
			}},
		},
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			// Foo
			{
				RuleID:  ruleID,
				Message: "Foo - message - right location",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    file,
					StartLine:   17,
					StartColumn: 0,
					EndLine:     21,
					EndColumn:   1,
				},
			},
			{
				RuleID:  ruleID,
				Message: "Foo - option (a) - fallback to (a).b",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    file,
					StartLine:   19,
					StartColumn: 2,
					EndLine:     19,
					EndColumn:   21,
				},
			},
			{
				RuleID:  ruleID,
				Message: "Foo - option (a).b - right location",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    file,
					StartLine:   19,
					StartColumn: 2,
					EndLine:     19,
					EndColumn:   21,
				},
			},
			{
				RuleID:  ruleID,
				Message: "Foo - option (a).c - fallback to (a).c.d",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    file,
					StartLine:   20,
					StartColumn: 2,
					EndLine:     20,
					EndColumn:   23,
				},
			},
			{
				RuleID:  ruleID,
				Message: "Foo - option (a).c.d - right location",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    file,
					StartLine:   20,
					StartColumn: 2,
					EndLine:     20,
					EndColumn:   23,
				},
			},

			// Bar
			{
				RuleID:  ruleID,
				Message: "Bar - option (a) - fallback to (a).c",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    file,
					StartLine:   25,
					StartColumn: 2,
					EndLine:     27,
					EndColumn:   4,
				},
			},
			{
				RuleID:  ruleID,
				Message: "Bar - option (a).c - right location",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    file,
					StartLine:   25,
					StartColumn: 2,
					EndLine:     27,
					EndColumn:   4,
				},
			},
			{
				RuleID:  ruleID,
				Message: "Bar - option (a).c.d - right location",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    file,
					StartLine:   26,
					StartColumn: 4,
					EndLine:     26,
					EndColumn:   10,
				},
			},
			{
				RuleID:  ruleID,
				Message: "Bar - option (a).b - right location",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    file,
					StartLine:   28,
					StartColumn: 2,
					EndLine:     28,
					EndColumn:   21,
				},
			},
		},
	}.Run(t)
}
