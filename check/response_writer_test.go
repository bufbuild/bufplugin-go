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
	"os"
	"path/filepath"
	"testing"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checktest"
	"buf.build/go/bufplugin/check/checkutil"
	"buf.build/go/bufplugin/descriptor"
	"github.com/stretchr/testify/require"
)

const protoFile = `
syntax = "proto3";
import "google/protobuf/descriptor.proto";
extend google.protobuf.MessageOptions {
  X x = 5000;
}
message X {
  string y = 1;
}
message Foo {
  option deprecated = true;
  option (x).y = "z";
}
`

// TestIssue20 checks that an annotation by source path points to the nearest parent if the original one is missing.
func TestIssue20(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	const fileName = "file.proto"
	const ruleID = "TEST_MISSING_SOURCE_LOCATION"
	require.NoError(t, os.WriteFile(filepath.Join(dir, fileName), []byte(protoFile), 0666))
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{dir},
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
					StartLine:   9,
					StartColumn: 0,
					EndLine:     12,
					EndColumn:   1,
				},
			},
			{
				RuleID:  ruleID,
				Message: "Annotation for option (x) points to first message option",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    fileName,
					StartLine:   10,
					StartColumn: 2,
					EndLine:     10,
					EndColumn:   27,
				},
			},
			{
				RuleID:  ruleID,
				Message: "Annotation for option (x).y",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    fileName,
					StartLine:   11,
					StartColumn: 2,
					EndLine:     11,
					EndColumn:   21,
				},
			},
		},
	}.Run(t)
}
