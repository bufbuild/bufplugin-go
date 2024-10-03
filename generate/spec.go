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

package generate

import (
	"context"

	"buf.build/go/bufplugin/info"
)

// Spec is the spec for a plugin.
//
// It is used to construct a plugin on the server-side (i.e. within the plugin).
//
// Generally, this is provided to Main. This library will handle Generate and ListRules calls
// based on the provided RuleSpecs.
type Spec struct {
	// Handler is the Handler that implements the generate logic.
	//
	// Required.
	Handler Handler
	// Info contains information about a plugin.
	//
	// Optional.
	//
	// If not set, the resulting server will not implement the PluginInfoService.
	Info *info.Spec
	// Before is a function that will be executed before any RuleHandlers are
	// invoked that returns a new Context and Request. This new Context and
	// Request will be passed to the RuleHandlers. This allows for any
	// pre-processing that needs to occur.
	Before func(ctx context.Context, request Request) (context.Context, Request, error)
}

// ValidateSpec validates all values on a Spec.
//
// This is exposed publicly so it can be run as part of plugin tests. This will verify
// that your Spec will result in a valid plugin.
func ValidateSpec(spec *Spec) error {
	if spec.Handler == nil {
		return newValidateSpecError("Handler is nil")
	}
	if spec.Info != nil {
		if err := info.ValidateSpec(spec.Info); err != nil {
			return err
		}
	}
	return nil
}
