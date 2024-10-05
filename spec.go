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
	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/generate"
	"buf.build/go/bufplugin/info"
	"buf.build/go/bufplugin/internal/bufpluginutil"
)

// Spec is the spec for a plugin. This defines all plugin operations.
//
// It is used to construct a plugin on the server-side (i.e. within the plugin).
//
// Generally, this is provided to Main.
type Spec struct {
	// Check contains the Spec for lint and breaking change checks.
	//
	// At least one of Check or Generate must be present.
	//
	// If not set, the resulting Server will not implement the CheckService.
	Check *check.Spec
	// Generate contains the Spec for generation.
	//
	// At least one of Check or Generate must be present.
	//
	// If not set, the resulting Server will not implement the GenerateService.
	Generate *generate.Spec
	// Info contains information about a plugin.
	//
	// Optional.
	//
	// If not set, the resulting Server will not implement the PluginInfoService.
	Info *info.Spec
}

// ValidateSpec validates all values on a Spec.
//
// This is exposed publicly so it can be run as part of plugin tests. This will verify
// that your Spec will result in a valid plugin.
func ValidateSpec(spec *Spec) error {
	if spec.Check == nil && spec.Generate == nil {
		return bufpluginutil.NewValidateSpecError(spec, "at least one of Check and Generate must be set")
	}
	if spec.Check != nil {
		if err := check.ValidateSpec(spec.Check); err != nil {
			return err
		}
	}
	if spec.Generate != nil {
		if err := generate.ValidateSpec(spec.Generate); err != nil {
			return err
		}
	}
	if spec.Info != nil {
		if err := info.ValidateSpec(spec.Info); err != nil {
			return err
		}
	}
	return nil
}
