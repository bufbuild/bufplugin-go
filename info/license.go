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
	"net/url"
)

// License contains license information about a plugin.
//
// A License will either have raw text or a URL that contains the License.
// Zero or one of these will be set.
type License interface {
	// SPDXLicenseID returns the SPDX license ID.
	//
	// Optional.
	SPDXLicenseID() string
	// Text returns the raw text of the License.
	//
	// At most one of Text and URL will be set.
	Text() string
	// URL returns the URL that contains the License.
	//
	// At most one of Text and URL will be set.
	URL() *url.URL

	isLicense()
}

// *** PRIVATE ***

type license struct {
	spdxLicenseID string
	text          string
	url           *url.URL
}

func newLicense(
	spdxLicenseID string,
	text string,
	url *url.URL,
) *license {
	return &license{
		spdxLicenseID: spdxLicenseID,
		text:          text,
		url:           url,
	}
}

func (l *license) SPDXLicenseID() string {
	return l.spdxLicenseID
}

func (l *license) Text() string {
	return l.text
}

func (l *license) URL() *url.URL {
	return l.url
}

func (*license) isLicense() {}
