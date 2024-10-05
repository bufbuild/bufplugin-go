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
	"context"
	"testing"

	"buf.build/go/bufplugin/info"
	"github.com/stretchr/testify/require"
	"pluginrpc.com/pluginrpc"
)

func TestPluginInfo(t *testing.T) {
	t.Parallel()

	client, err := NewClientForSpec(
		&Spec{
			Rules: []*RuleSpec{
				{
					ID:      "RULE1",
					Purpose: "Test RULE1.",
					Type:    RuleTypeLint,
					Handler: nopRuleHandler,
				},
			},
			Info: &info.Spec{
				SPDXLicenseID: "apache-2.0",
				LicenseURL:    "https://foo.com/license",
			},
		},
	)
	require.NoError(t, err)
	pluginInfo, err := client.GetPluginInfo(context.Background())
	require.NoError(t, err)
	license := pluginInfo.License()
	require.NotNil(t, license)
	require.NotNil(t, license.URL())
	// Case-sensitive.
	require.Equal(t, "Apache-2.0", license.SPDXLicenseID())
	require.Equal(t, "https://foo.com/license", license.URL().String())
}

func TestPluginInfoUnimplemented(t *testing.T) {
	t.Parallel()

	client, err := NewClientForSpec(
		&Spec{
			Rules: []*RuleSpec{
				{
					ID:      "RULE1",
					Purpose: "Test RULE1.",
					Type:    RuleTypeLint,
					Handler: nopRuleHandler,
				},
			},
		},
	)
	require.NoError(t, err)
	_, err = client.GetPluginInfo(context.Background())
	pluginrpcError := &pluginrpc.Error{}
	require.Error(t, err)
	require.ErrorAs(t, err, &pluginrpcError)
	require.Equal(t, pluginrpc.CodeUnimplemented, pluginrpcError.Code())
}
