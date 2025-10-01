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

// Package main implements a simple plugin that checks that all
// google.protobuf.Timestamp fields end in a specific suffix.
//
// To use this plugin:
//
//	# buf.yaml
//	version: v2
//	lint:
//	  use:
//	   - STANDARD # omit if you do not want to use the rules builtin to buf
//	   - TIMESTAMP_SUFFIX
//	plugins:
//	  - plugin: buf-plugin-timestamp-suffix
//
// The default suffix is "_time", but this can be overridden with the
// "timestamp_suffix" option key in your  buf.yaml:
//
//	# buf.yaml
//	version: v2
//	lint:
//	  use:
//	   - STANDARD # omit if you do not want to use the rules builtin to buf
//	   - TIMESTAMP_SUFFIX
//	plugins:
//	  - plugin: buf-plugin-timestamp-suffix
//	    options:
//	      timestamp_suffix: _timestamp
package main

import (
	"context"
	"strings"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"buf.build/go/bufplugin/info"
	"buf.build/go/bufplugin/option"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	// timestampSuffixRuleID is the Rule ID of the timestamp suffix Rule.
	timestampSuffixRuleID = "TIMESTAMP_SUFFIX"

	// timestampSuffixOptionKey is the option key to override the default timestamp suffix.
	timestampSuffixOptionKey = "timestamp_suffix"

	defaultTimestampSuffix = "_time"
)

var (
	// timestampSuffixRuleSpec is the RuleSpec for the timestamp suffix Rule.
	timestampSuffixRuleSpec = &check.RuleSpec{
		ID:      timestampSuffixRuleID,
		Default: true,
		Purpose: `Checks that all google.protobuf.Timestamps end in a specific suffix (default is "_time").`,
		Type:    check.RuleTypeLint,
		Handler: checkutil.NewFieldRuleHandler(checkTimestampSuffix, checkutil.WithoutImports()),
	}

	// spec is the Spec for the timestamp suffix plugin.
	spec = &check.Spec{
		Rules: []*check.RuleSpec{
			timestampSuffixRuleSpec,
		},
		// Optional.
		Info: &info.Spec{
			Documentation: `A simple plugin that checks that all google.protobuf.Timestamp fields end in a specific suffix (default is "_time").`,
			SPDXLicenseID: "apache-2.0",
			LicenseURL:    "https://github.com/bufbuild/bufplugin-go/blob/main/LICENSE",
		},
	}
)

func main() {
	check.Main(spec)
}

func checkTimestampSuffix(
	_ context.Context,
	responseWriter check.ResponseWriter,
	request check.Request,
	fieldDescriptor protoreflect.FieldDescriptor,
) error {
	timestampSuffix := defaultTimestampSuffix
	timestampSuffixOptionValue, err := option.GetStringValue(request.Options(), timestampSuffixOptionKey)
	if err != nil {
		return err
	}
	if timestampSuffixOptionValue != "" {
		timestampSuffix = timestampSuffixOptionValue
	}

	fieldDescriptorType := fieldDescriptor.Message()
	if fieldDescriptorType == nil {
		return nil
	}
	if string(fieldDescriptorType.FullName()) != "google.protobuf.Timestamp" {
		return nil
	}
	if !strings.HasSuffix(string(fieldDescriptor.Name()), timestampSuffix) {
		responseWriter.AddAnnotation(
			check.WithMessagef("Fields of type google.protobuf.Timestamp must end in %q but field name was %q.", timestampSuffix, string(fieldDescriptor.Name())),
			check.WithDescriptor(fieldDescriptor),
		)
	}
	return nil
}
