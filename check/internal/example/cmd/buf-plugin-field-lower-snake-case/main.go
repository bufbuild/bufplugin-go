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

// Package main implements a simple plugin that checks that all field names are lower_snake_case.
//
// To use this plugin:
//
//	# buf.yaml
//	version: v2
//	lint:
//	  use:
//	   - STANDARD # omit if you do not want to use the rules builtin to buf
//	   - PLUGIN_FIELD_LOWER_SNAKE_CASE
//	plugins:
//	  - plugin: buf-plugin-field-lower-snake-case
//
// Note that the buf CLI implements this check as a builtin Rule, but this is just for example.
package main

import (
	"context"
	"strings"
	"unicode"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"buf.build/go/bufplugin/info"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// fieldLowerSnakeCaseRuleID is the Rule ID of the timestamp suffix Rule.
//
// This has a "PLUGIN_" prefix as the buf CLI has a rule "FIELD_LOWER_SNAKE_CASE" builtin,
// and plugins/the buf CLI must have unique Rule IDs.
const fieldLowerSnakeCaseRuleID = "PLUGIN_FIELD_LOWER_SNAKE_CASE"

var (
	// fieldLowerSnakeCaseRuleSpec is the RuleSpec for the timestamp suffix Rule.
	fieldLowerSnakeCaseRuleSpec = &check.RuleSpec{
		ID:      fieldLowerSnakeCaseRuleID,
		Default: true,
		Purpose: "Checks that all field names are lower_snake_case.",
		Type:    check.RuleTypeLint,
		Handler: checkutil.NewFieldRuleHandler(checkFieldLowerSnakeCase, checkutil.WithoutImports()),
	}

	// spec is the Spec for the timestamp suffix plugin.
	spec = &check.Spec{
		Rules: []*check.RuleSpec{
			fieldLowerSnakeCaseRuleSpec,
		},
		// Optional.
		Info: &info.Spec{
			SPDXLicenseID: "apache-2.0",
			LicenseURL:    "https://github.com/bufbuild/bufplugin-go/blob/main/LICENSE",
		},
	}
)

func main() {
	check.Main(spec)
}

func checkFieldLowerSnakeCase(
	_ context.Context,
	responseWriter check.ResponseWriter,
	_ check.Request,
	fieldDescriptor protoreflect.FieldDescriptor,
) error {
	fieldName := string(fieldDescriptor.Name())
	fieldNameToLowerSnakeCase := toLowerSnakeCase(fieldName)
	if fieldName != fieldNameToLowerSnakeCase {
		responseWriter.AddAnnotation(
			check.WithMessagef("Field name %q should be lower_snake_case, such as %q.", fieldName, fieldNameToLowerSnakeCase),
			check.WithDescriptor(fieldDescriptor),
		)
	}
	return nil
}

func toLowerSnakeCase(s string) string {
	return strings.ToLower(toSnakeCase(s))
}

func toSnakeCase(s string) string {
	output := ""
	s = strings.TrimFunc(s, isDelimiter)
	for i, c := range s {
		if isDelimiter(c) {
			c = '_'
		}
		switch {
		case i == 0:
			output += string(c)
		case isSnakeCaseNewWord(c, false) &&
			output[len(output)-1] != '_' &&
			((i < len(s)-1 && !isSnakeCaseNewWord(rune(s[i+1]), true) && !isDelimiter(rune(s[i+1]))) ||
				(unicode.IsLower(rune(s[i-1])))):
			output += "_" + string(c)
		case !(isDelimiter(c) && output[len(output)-1] == '_'):
			output += string(c)
		}
	}
	return output
}

func isSnakeCaseNewWord(r rune, newWordOnDigits bool) bool {
	if newWordOnDigits {
		return unicode.IsUpper(r) || unicode.IsDigit(r)
	}
	return unicode.IsUpper(r)
}

func isDelimiter(r rune) bool {
	return r == '.' || r == '-' || r == '_' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
}
