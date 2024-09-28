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

import "errors"

// Spec is the spec for the information about a plugin.
type Spec struct {
	// URL is the URL for a plugin.
	URL string
	// SPDXLicenseID is the SDPX ID of the License.
	SPDXLicenseID string
	// LicenseText is the raw text of the License.
	//
	// Zero or one of LicenseText and LicenseURL must be set.
	LicenseText string
	// LicenseURL is the URL that contains the License.
	//
	// Zero or one of LicenseText and LicenseURL must be set.
	LicenseURL string
	// DocShort contains a short description of the plugin's functionality.
	//
	// Required if DocLong is set..
	DocShort string
	// DocLong contains extra details of the plugin.
	//
	// May not be set if DocShort is not set.
	DocLong string
}

// *** PRIVATE ***

// ValidateSpec validates all values on a Spec.
func ValidateSpec(spec *Spec) error {
	return errors.New("TODO")
}
