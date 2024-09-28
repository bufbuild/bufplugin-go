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
	"net/url"

	infov1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/info/v1"
)

// PluginInfo contains information about a plugin.
type PluginInfo interface {
	// URL returns the URL for a plugin.
	//
	// Optional.
	//
	// Will always be absolute.
	//
	// This typically is the source control repository that contains the plugin's implementation.
	URL() *url.URL
	// License returns the license of the plugin.
	//
	// Optional.
	License() License
	// Doc returns the documentation of the plugin.
	//
	// Optional.
	Doc() Doc

	toProto() *infov1.PluginInfo

	isPluginInfo()
}

// NewPluginInfoForSpec returns a new PluginInfo for the given Spec.
func NewPluginInfoForSpec(spec *Spec) (PluginInfo, error) {
	if err := ValidateSpec(spec); err != nil {
		return nil, err
	}

	var uri *url.URL
	var err error
	if spec.URL != "" {
		uri, err = url.Parse(spec.URL)
		if err != nil {
			return nil, err
		}
	}

	var license *license
	if spec.SPDXLicenseID != "" || spec.LicenseText != "" || spec.LicenseURL != "" {
		var licenseURI *url.URL
		if spec.LicenseURL != "" {
			licenseURI, err = url.Parse(spec.LicenseURL)
			if err != nil {
				return nil, err
			}
		}
		license, err = newLicense(
			spec.SPDXLicenseID,
			spec.LicenseText,
			licenseURI,
		)
		if err != nil {
			return nil, err
		}
	}

	var doc *doc
	if spec.DocShort != "" {
		doc, err = newDoc(spec.DocShort, spec.DocLong)
		if err != nil {
			return nil, err
		}
	}
	return newPluginInfo(uri, license, doc)
}

// *** PRIVATE ***

type pluginInfo struct {
	url *url.URL
	// Need to keep as pointer for Go nil is not nil problem.
	license *license
	// Need to keep as pointer for Go nil is not nil problem.
	doc *doc
}

func newPluginInfo(
	url *url.URL,
	license *license,
	doc *doc,
) (*pluginInfo, error) {
	if url != nil && url.Host == "" {
		return nil, fmt.Errorf("url %v must be absolute", url)
	}
	return &pluginInfo{
		url:     url,
		license: license,
		doc:     doc,
	}, nil
}

func (p *pluginInfo) URL() *url.URL {
	return p.url
}

func (p *pluginInfo) License() License {
	// Go nil is not nil problem.
	if p.license == nil {
		return nil
	}
	return p.license
}

func (p *pluginInfo) Doc() Doc {
	// Go nil is not nil problem.
	if p.doc == nil {
		return nil
	}
	return p.doc
}

func (p *pluginInfo) toProto() *infov1.PluginInfo {
	var urlString string
	if p.url != nil {
		urlString = p.url.String()
	}
	return &infov1.PluginInfo{
		Url:     urlString,
		License: p.license.toProto(),
		Doc:     p.doc.toProto(),
	}
}

func (*pluginInfo) isPluginInfo() {}

func pluginInfoForProtoPluginInfo(protoPluginInfo *infov1.PluginInfo) (PluginInfo, error) {
	if protoPluginInfo == nil {
		return nil, nil
	}
	var uri *url.URL
	if urlString := protoPluginInfo.GetUrl(); urlString != "" {
		var err error
		uri, err = url.Parse(urlString)
		if err != nil {
			return nil, err
		}
	}
	license, err := licenseForProtoLicense(protoPluginInfo.GetLicense())
	if err != nil {
		return nil, err
	}
	doc, err := docForProtoDoc(protoPluginInfo.GetDoc())
	if err != nil {
		return nil, err
	}
	return newPluginInfo(uri, license, doc)
}
