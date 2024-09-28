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

package info

import (
	"fmt"
)

// Doc contains the documentation for a plugin.
//
// The documentation contains a required Short docuementation string, and optional
// details contained in Long.
//
// When printed, the Short and Long strings will be contatenated with two newlines.
type Doc interface {
	// fmt.Stringer will contatenate Short and Long with two newlines if Long is present, and
	// otherwise return Short.
	fmt.Stringer

	// Short contains a short description of the plugin's functionality.
	//
	// Required.
	Short() string
	// Long contains extra details of the plugin.
	//
	// Optional.
	Long() string

	isDoc()
}

// *** PRIVATE ***

type doc struct {
	short string
	long  string
}

func newDoc(
	short string,
	long string,
) *doc {
	return &doc{
		short: short,
		long:  long,
	}
}

func (d *doc) String() string {
	if d.long == "" {
		return d.short
	}
	return d.short + "\n\n" + d.long
}

func (d *doc) Short() string {
	return d.short
}

func (d *doc) Long() string {
	return d.long
}

func (*doc) isDoc() {}
