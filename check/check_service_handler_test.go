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

package check

import (
	"context"
	"testing"

	checkv1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1"
	descriptorv1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/descriptor/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"pluginrpc.com/pluginrpc"
)

func TestCheckServiceHandlerUniqueFiles(t *testing.T) {
	t.Parallel()

	checkServiceHandler, err := NewCheckServiceHandler(
		&Spec{
			Rules: []*RuleSpec{
				testNewSimpleLintRuleSpec("RULE1", nil, true, false, nil),
			},
		},
	)
	require.NoError(t, err)

	_, err = checkServiceHandler.Check(
		context.Background(),
		&checkv1.CheckRequest{
			FileDescriptors: []*descriptorv1.FileDescriptor{
				{
					FileDescriptorProto: &descriptorpb.FileDescriptorProto{
						Name:           proto.String("foo.proto"),
						SourceCodeInfo: &descriptorpb.SourceCodeInfo{},
					},
				},
			},
			AgainstFileDescriptors: []*descriptorv1.FileDescriptor{
				{
					FileDescriptorProto: &descriptorpb.FileDescriptorProto{
						Name:           proto.String("foo.proto"),
						SourceCodeInfo: &descriptorpb.SourceCodeInfo{},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	_, err = checkServiceHandler.Check(
		context.Background(),
		&checkv1.CheckRequest{
			FileDescriptors: []*descriptorv1.FileDescriptor{
				{
					FileDescriptorProto: &descriptorpb.FileDescriptorProto{
						Name:           proto.String("foo.proto"),
						SourceCodeInfo: &descriptorpb.SourceCodeInfo{},
					},
				},
				{
					FileDescriptorProto: &descriptorpb.FileDescriptorProto{
						Name:           proto.String("foo.proto"),
						SourceCodeInfo: &descriptorpb.SourceCodeInfo{},
					},
				},
			},
		},
	)
	pluginrpcError := &pluginrpc.Error{}
	require.ErrorAs(t, err, &pluginrpcError)
	require.Equal(t, pluginrpc.CodeInvalidArgument, pluginrpcError.Code())

	_, err = checkServiceHandler.Check(
		context.Background(),
		&checkv1.CheckRequest{
			FileDescriptors: []*descriptorv1.FileDescriptor{
				{
					FileDescriptorProto: &descriptorpb.FileDescriptorProto{
						Name:           proto.String("foo.proto"),
						SourceCodeInfo: &descriptorpb.SourceCodeInfo{},
					},
				},
			},
			AgainstFileDescriptors: []*descriptorv1.FileDescriptor{
				{
					FileDescriptorProto: &descriptorpb.FileDescriptorProto{
						Name:           proto.String("bar.proto"),
						SourceCodeInfo: &descriptorpb.SourceCodeInfo{},
					},
				},
				{
					FileDescriptorProto: &descriptorpb.FileDescriptorProto{
						Name:           proto.String("bar.proto"),
						SourceCodeInfo: &descriptorpb.SourceCodeInfo{},
					},
				},
			},
		},
	)
	pluginrpcError = &pluginrpc.Error{}
	require.ErrorAs(t, err, &pluginrpcError)
	require.Equal(t, pluginrpc.CodeInvalidArgument, pluginrpcError.Code())
}

func TestCheckServiceHandlerNoSourceCodeInfo(t *testing.T) {
	t.Parallel()

	checkServiceHandler, err := NewCheckServiceHandler(
		&Spec{
			Rules: []*RuleSpec{
				testNewSimpleLintRuleSpec("RULE1", nil, true, false, nil),
			},
		},
	)
	require.NoError(t, err)

	_, err = checkServiceHandler.Check(
		context.Background(),
		&checkv1.CheckRequest{
			FileDescriptors: []*descriptorv1.FileDescriptor{
				{
					FileDescriptorProto: &descriptorpb.FileDescriptorProto{
						Name: proto.String("foo.proto"),
					},
				},
			},
			AgainstFileDescriptors: []*descriptorv1.FileDescriptor{
				{
					FileDescriptorProto: &descriptorpb.FileDescriptorProto{
						Name: proto.String("foo.proto"),
					},
				},
			},
		},
	)
	require.NoError(t, err)
}
