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

	infov1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/info/v1"
	"github.com/stretchr/testify/require"
)

func TestPluginInfoServiceHandlerBasic(t *testing.T) {
	t.Parallel()

	pluginInfoServiceHandler, err := NewPluginInfoServiceHandler(
		&Spec{
			LicenseURL: "https://foo.com/license",
		},
	)
	require.NoError(t, err)

	_, err = pluginInfoServiceHandler.GetPluginInfo(
		context.Background(),
		&infov1.GetPluginInfoRequest{},
	)
	require.NoError(t, err)
}