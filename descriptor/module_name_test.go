// Copyright 2024-2025 Buf Technologies, Inc.
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

package descriptor

import (
	"strings"
	"testing"

	descriptorv1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/descriptor/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestParseModuleName(t *testing.T) {
	t.Parallel()

	moduleName, err := ParseModuleName("buf.build/acme/weather")
	require.NoError(t, err)
	assert.Equal(t, "buf.build", moduleName.Registry())
	assert.Equal(t, "acme", moduleName.Owner())
	assert.Equal(t, "weather", moduleName.Module())
	assert.Equal(t, "buf.build/acme/weather", moduleName.String())

	for _, invalid := range []string{
		"",
		"buf.build",
		"buf.build/acme",
		"buf.build/acme/weather/extra",
		"/acme/weather",
		"buf.build//weather",
		"buf.build/acme/",
		"buf.build/acme/w",
		"buf.build/" + strings.Repeat("a", 49) + "/weather",
		"buf.build/acme/" + strings.Repeat("a", 101),
	} {
		t.Run(invalid, func(t *testing.T) {
			t.Parallel()
			_, err := ParseModuleName(invalid)
			assert.Error(t, err)
		})
	}
}

func TestModuleNameForProtoModuleName(t *testing.T) {
	t.Parallel()

	assert.Nil(t, ModuleNameForProtoModuleName(nil))

	protoModuleName := descriptorv1.ModuleName_builder{
		Registry: "buf.build",
		Owner:    "acme",
		Module:   "weather",
	}.Build()
	moduleName := ModuleNameForProtoModuleName(protoModuleName)
	require.NotNil(t, moduleName)
	assert.Equal(t, "buf.build/acme/weather", moduleName.String())
	assert.True(t, proto.Equal(protoModuleName, moduleName.ToProto()))
}
