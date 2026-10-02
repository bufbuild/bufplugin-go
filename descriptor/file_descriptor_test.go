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
	"testing"

	descriptorv1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/descriptor/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestFileDescriptorRoundTrip(t *testing.T) {
	t.Parallel()

	protoFileDescriptors := []*descriptorv1.FileDescriptor{
		descriptorv1.FileDescriptor_builder{
			FileDescriptorProto: &descriptorpb.FileDescriptorProto{
				Name:       new("a.proto"),
				Syntax:     new("proto3"),
				Dependency: []string{"b.proto", "c.proto"},
			},
			UnusedDependency: []int32{1},
			ModuleName: descriptorv1.ModuleName_builder{
				Registry: "buf.build",
				Owner:    "acme",
				Module:   "weather",
			}.Build(),
		}.Build(),
		descriptorv1.FileDescriptor_builder{
			FileDescriptorProto: &descriptorpb.FileDescriptorProto{
				Name:   new("b.proto"),
				Syntax: new("proto3"),
			},
			IsImport: true,
		}.Build(),
		descriptorv1.FileDescriptor_builder{
			FileDescriptorProto: &descriptorpb.FileDescriptorProto{
				Name: new("c.proto"),
			},
			IsImport:            true,
			IsSyntaxUnspecified: true,
		}.Build(),
	}
	fileDescriptors, err := FileDescriptorsForProtoFileDescriptors(protoFileDescriptors)
	require.NoError(t, err)
	require.Len(t, fileDescriptors, len(protoFileDescriptors))

	fileNameToFileDescriptor := make(map[string]FileDescriptor, len(fileDescriptors))
	for _, fileDescriptor := range fileDescriptors {
		fileNameToFileDescriptor[fileDescriptor.ProtoreflectFileDescriptor().Path()] = fileDescriptor
	}

	aFileDescriptor := fileNameToFileDescriptor["a.proto"]
	require.NotNil(t, aFileDescriptor)
	assert.False(t, aFileDescriptor.IsImport())
	assert.False(t, aFileDescriptor.IsSyntaxUnspecified())
	assert.Equal(t, []int32{1}, aFileDescriptor.UnusedDependencyIndexes())
	require.NotNil(t, aFileDescriptor.ModuleName())
	assert.Equal(t, "buf.build/acme/weather", aFileDescriptor.ModuleName().String())

	bFileDescriptor := fileNameToFileDescriptor["b.proto"]
	require.NotNil(t, bFileDescriptor)
	assert.True(t, bFileDescriptor.IsImport())
	assert.False(t, bFileDescriptor.IsSyntaxUnspecified())
	assert.Empty(t, bFileDescriptor.UnusedDependencyIndexes())
	assert.Nil(t, bFileDescriptor.ModuleName())

	cFileDescriptor := fileNameToFileDescriptor["c.proto"]
	require.NotNil(t, cFileDescriptor)
	assert.True(t, cFileDescriptor.IsImport())
	assert.True(t, cFileDescriptor.IsSyntaxUnspecified())
	assert.Empty(t, cFileDescriptor.UnusedDependencyIndexes())
	assert.Nil(t, cFileDescriptor.ModuleName())

	for _, protoFileDescriptor := range protoFileDescriptors {
		fileName := protoFileDescriptor.GetFileDescriptorProto().GetName()
		fileDescriptor := fileNameToFileDescriptor[fileName]
		assert.Same(t, protoFileDescriptor.GetFileDescriptorProto(), fileDescriptor.FileDescriptorProto())
		assert.True(
			t,
			proto.Equal(protoFileDescriptor, fileDescriptor.ToProto()),
			"round trip mismatch for %s",
			fileName,
		)
	}
}
