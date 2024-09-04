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
	"slices"
	"strings"
)

// CompareAnnotations returns -1 if one < two, 1 if one > two, 0 otherwise.
func CompareAnnotations(one Annotation, two Annotation) int {
	if one == nil && two == nil {
		return 0
	}
	if one == nil && two != nil {
		return -1
	}
	if one != nil && two == nil {
		return 1
	}
	if compare := strings.Compare(one.RuleID(), two.RuleID()); compare != 0 {
		return compare
	}

	if compare := CompareLocations(one.Location(), two.Location()); compare != 0 {
		return compare
	}

	if compare := CompareLocations(one.AgainstLocation(), two.AgainstLocation()); compare != 0 {
		return compare
	}
	return strings.Compare(one.Message(), two.Message())
}

// CompareLocations returns -1 if one < two, 1 if one > two, 0 otherwise.
func CompareLocations(one Location, two Location) int {
	if one == nil && two == nil {
		return 0
	}
	if one == nil && two != nil {
		return -1
	}
	if one != nil && two == nil {
		return 1
	}
	if compare := strings.Compare(one.File().FileDescriptor().Path(), two.File().FileDescriptor().Path()); compare != 0 {
		return compare
	}

	if compare := intCompare(one.StartLine(), two.StartLine()); compare != 0 {
		return compare
	}

	if compare := intCompare(one.StartColumn(), two.StartColumn()); compare != 0 {
		return compare
	}

	if compare := intCompare(one.EndLine(), two.EndLine()); compare != 0 {
		return compare
	}

	if compare := intCompare(one.EndColumn(), two.EndColumn()); compare != 0 {
		return compare
	}

	if compare := slices.Compare(one.unclonedSourcePath(), two.unclonedSourcePath()); compare != 0 {
		return compare
	}

	if compare := strings.Compare(one.LeadingComments(), two.LeadingComments()); compare != 0 {
		return compare
	}

	if compare := strings.Compare(one.TrailingComments(), two.TrailingComments()); compare != 0 {
		return compare
	}
	return slices.Compare(one.unclonedLeadingDetachedComments(), two.unclonedLeadingDetachedComments())
}

// CompareRules returns -1 if one < two, 1 if one > two, 0 otherwise.
func CompareRules(one Rule, two Rule) int {
	if one == nil && two == nil {
		return 0
	}
	if one == nil && two != nil {
		return -1
	}
	if one != nil && two == nil {
		return 1
	}
	return strings.Compare(one.ID(), two.ID())
}

// CompareCategories returns -1 if one < two, 1 if one > two, 0 otherwise.
func CompareCategories(one Category, two Category) int {
	if one == nil && two == nil {
		return 0
	}
	if one == nil && two != nil {
		return -1
	}
	if one != nil && two == nil {
		return 1
	}
	return strings.Compare(one.ID(), two.ID())
}

// *** PRIVATE ***

// compareRuleSpecs returns -1 if one < two, 1 if one > two, 0 otherwise.
func compareRuleSpecs(one *RuleSpec, two *RuleSpec) int {
	if one == nil && two == nil {
		return 0
	}
	if one == nil && two != nil {
		return -1
	}
	if one != nil && two == nil {
		return 1
	}
	return strings.Compare(one.ID, two.ID)
}

// compareCategorySpecs returns -1 if one < two, 1 if one > two, 0 otherwise.
func compareCategorySpecs(one *CategorySpec, two *CategorySpec) int {
	if one == nil && two == nil {
		return 0
	}
	if one == nil && two != nil {
		return -1
	}
	if one != nil && two == nil {
		return 1
	}
	return strings.Compare(one.ID, two.ID)
}

func intCompare(one int, two int) int {
	if one < two {
		return -1
	}
	if one > two {
		return 1
	}
	return 0
}
