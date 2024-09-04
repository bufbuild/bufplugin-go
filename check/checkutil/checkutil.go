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

// Package checkutil implements helpers for the check package.
package checkutil

// IteratorOption is an option for any of the New.*RuleHandler functions in this package.
type IteratorOption func(*iteratorOptions)

// WithoutImports returns a new IteratorOption that will not call the provided function
// for any imports.
//
// For lint RuleHandlers, this is generally an option you will want to pass. For breaking
// RuleHandlers, you generally want to consider imports as part of breaking changes.
//
// The default is to call the provided function for all imports.
func WithoutImports() IteratorOption {
	return func(iteratorOptions *iteratorOptions) {
		iteratorOptions.withoutImports = true
	}
}

// *** PRIVATE ***

type iteratorOptions struct {
	withoutImports bool
}

func newIteratorOptions() *iteratorOptions {
	return &iteratorOptions{}
}
