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

// Package main implements a plugin that implements two Rules:
//
//   - A lint Rule that checks that every field has the option (acme.option.v1.safe_for_ml) explicitly set.
//   - A breaking Rule that verifes that no field goes from having option (acme.option.v1.safe_for_ml) going
//     from true to false. That is, if a field is marked as safe, it can not then be moved to unsafe.
//
// This is an example of a plugin that will check a custom option, which is a very typical
// case for a custom lint or breaking change plugin. In this case, we're saying that an organization
// wants to explicitly mark every field in its schemas as either safe to train ML models on, or
// unsafe to train models on. This plugin enforces that all fields have such markings, and that
// those fields do not transition from safe to unsafe.
//
// This plugin also demonstrates the usage of categories. The Rules have IDs:
//
//   - FIELD_OPTION_SAFE_FOR_ML_SET
//   - FIELD_OPTION_SAFE_FOR_ML_STAYS_TRUE
//
// However, the Rules both belong to category FIELD_OPTION_SAFE_FOR_ML. This means that you
// do not need to specify the individual rules in your configuration. You can just specify
// the Category, and all Rules in this Category will be included.
//
// To use this plugin:
//
//	# buf.yaml
//	version: v2
//	lint:
//	  use:
//	   - STANDARD # omit if you do not want to use the rules builtin to buf
//	   - FIELD_OPTION_SAFE_FOR_ML
//	breaking:
//	  use:
//	   - WIRE_JSON # omit if you do not want to use the rules builtin to buf
//	   - FIELD_OPTION_SAFE_FOR_ML
//	plugins:
//	  - plugin: buf-plugin-field-option-safe-for-ml
package main

import (
	"context"
	"fmt"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	optionv1 "buf.build/go/bufplugin/check/internal/example/gen/acme/option/v1"
	"buf.build/go/bufplugin/info"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	// fieldOptionSafeForMLSetRuleID is the Rule ID of the field option safe for ML set Rule.
	fieldOptionSafeForMLSetRuleID = "FIELD_OPTION_SAFE_FOR_ML_SET"
	// fieldOptionSafeForMLStaysTrueRuleID is the Rule ID of the field option safe for ML stays true Rule.
	fieldOptionSafeForMLStaysTrueRuleID = "FIELD_OPTION_SAFE_FOR_ML_STAYS_TRUE"
	// fieldOptionSafeForMLCategoryID is the Category ID for the rules concerning (acme.option.v1.safe_for_ml).
	fieldOptionSafeForMLCategoryID = "FIELD_OPTION_SAFE_FOR_ML"
)

var (
	// fieldOptionSafeForMLRuleSpec is the RuleSpec for the field option safe for ML set Rule.
	fieldOptionSafeForMLSetRuleSpec = &check.RuleSpec{
		ID:      fieldOptionSafeForMLSetRuleID,
		Default: true,
		Purpose: "Checks that every field has option (acme.option.v1.safe_for_ml) explicitly set.",
		CategoryIDs: []string{
			fieldOptionSafeForMLCategoryID,
		},
		Type:    check.RuleTypeLint,
		Handler: checkutil.NewFieldRuleHandler(checkFieldOptionSafeForMLSet, checkutil.WithoutImports()),
	}
	// fieldOptionSafeForMLStaysTrueRuleSpec is the RuleSpec for the field option safe for ML stays  true Rule.
	fieldOptionSafeForMLStaysTrueRuleSpec = &check.RuleSpec{
		ID:      fieldOptionSafeForMLStaysTrueRuleID,
		Default: true,
		Purpose: "Checks that every field marked with (acme.option.v1.safe_for_ml) = true does not change to false.",
		CategoryIDs: []string{
			fieldOptionSafeForMLCategoryID,
		},
		Type:    check.RuleTypeBreaking,
		Handler: checkutil.NewFieldPairRuleHandler(checkFieldOptionSafeForMLStaysTrue, checkutil.WithoutImports()),
	}
	fieldOptionSafeForMLCategorySpec = &check.CategorySpec{
		ID:      fieldOptionSafeForMLCategoryID,
		Purpose: "Checks properties around the (acme.option.v1.safe_for_ml) option.",
	}

	// spec is the Spec for the syntax specified plugin.
	spec = &check.Spec{
		Rules: []*check.RuleSpec{
			fieldOptionSafeForMLSetRuleSpec,
			fieldOptionSafeForMLStaysTrueRuleSpec,
		},
		Categories: []*check.CategorySpec{
			fieldOptionSafeForMLCategorySpec,
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

func checkFieldOptionSafeForMLSet(
	_ context.Context,
	responseWriter check.ResponseWriter,
	_ check.Request,
	fieldDescriptor protoreflect.FieldDescriptor,
) error {
	// Ignore the actual field options - we don't need to mark safe_for_ml as safe_for_ml.
	if fieldDescriptor.ContainingMessage().FullName() == "google.protobuf.FieldOptions" {
		return nil
	}
	fieldOptions, err := getFieldOptions(fieldDescriptor)
	if err != nil {
		return err
	}
	if !proto.HasExtension(fieldOptions, optionv1.E_SafeForMl) {
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Field %q on message %q should have option (acme.option.v1.safe_for_ml) explicitly set.",
				fieldDescriptor.Name(),
				fieldDescriptor.ContainingMessage().FullName(),
			),
			check.WithDescriptor(fieldDescriptor),
		)
	}
	return nil
}

func checkFieldOptionSafeForMLStaysTrue(
	_ context.Context,
	responseWriter check.ResponseWriter,
	_ check.Request,
	fieldDescriptor protoreflect.FieldDescriptor,
	againstFieldDescriptor protoreflect.FieldDescriptor,
) error {
	// Ignore the actual field options - we don't need to mark safe_for_ml as safe_for_ml.
	if fieldDescriptor.ContainingMessage().FullName() == "google.protobuf.FieldOptions" {
		return nil
	}
	againstSafeForML, err := getSafeForML(againstFieldDescriptor)
	if err != nil {
		return err
	}
	if !againstSafeForML {
		// If the field does not have safe_for_ml or safe_for_ml is false, we are done. It is up to the
		// lint Rule to enforce whether or not every field has this option explicitly set.
		return nil
	}
	safeForML, err := getSafeForML(fieldDescriptor)
	if err != nil {
		return err
	}
	if !safeForML {
		responseWriter.AddAnnotation(
			check.WithMessagef(
				"Field %q on message %q should had option (acme.option.v1.safe_for_ml) change from true to false.",
				fieldDescriptor.Name(),
				fieldDescriptor.ContainingMessage().FullName(),
			),
			check.WithDescriptor(fieldDescriptor),
			check.WithAgainstDescriptor(againstFieldDescriptor),
		)
	}
	return nil
}

func getFieldOptions(fieldDescriptor protoreflect.FieldDescriptor) (*descriptorpb.FieldOptions, error) {
	fieldOptions, ok := fieldDescriptor.Options().(*descriptorpb.FieldOptions)
	if !ok {
		// This should never happen.
		return nil, fmt.Errorf("expected *descriptorpb.FieldOptions for FieldDescriptor %q Options but got %T", fieldDescriptor.FullName(), fieldOptions)
	}
	return fieldOptions, nil
}

func getSafeForML(fieldDescriptor protoreflect.FieldDescriptor) (bool, error) {
	fieldOptions, err := getFieldOptions(fieldDescriptor)
	if err != nil {
		return false, err
	}
	if !proto.HasExtension(fieldOptions, optionv1.E_SafeForMl) {
		return false, nil
	}
	safeForMLIface := proto.GetExtension(fieldOptions, optionv1.E_SafeForMl)
	if safeForMLIface == nil {
		return false, fmt.Errorf("expected non-nil value for FieldDescriptor %q option value", fieldDescriptor.FullName())
	}
	safeForML, ok := safeForMLIface.(bool)
	if !ok {
		// This should never happen.
		return false, fmt.Errorf("expected bool for FieldDescriptor %q option value but got %T", fieldDescriptor.FullName(), safeForMLIface)
	}
	return safeForML, nil
}
