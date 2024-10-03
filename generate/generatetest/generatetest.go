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

// Package generatetest provides testing helpers when writing generate plugins.
//
// The easiest entry point is TestCase. This allows you to set up a test and run it extremely
// easily. Other functions provide lower-level primitives if TestCase doesn't meet your needs.
package generatetest

import (
	"context"
	"errors"
	"testing"

	"buf.build/go/bufplugin/descriptor/descriptortest"
	"buf.build/go/bufplugin/generate"
	"buf.build/go/bufplugin/option"
	"github.com/stretchr/testify/require"
)

// SpecTest tests your spec with generate.ValidateSpec.
//
// Almost every plugin should run a test with SpecTest.
//
//	func TestSpec(t *testing.T) {
//	  t.Parallel()
//	  generatetest.SpecTest(t, yourSpec)
//	}
func SpecTest(t *testing.T, spec *generate.Spec) {
	require.NoError(t, generate.ValidateSpec(spec))
}

// GenerateTest is a single Generate test to run against a Spec.
type GenerateTest struct {
	// Request is the request spec to test.
	Request *RequestSpec
	// Spec is the Spec to test.
	//
	// Required.
	Spec *generate.Spec
}

// Run runs the test.
//
// This will:
//
//   - Build the Files and AgainstFiles.
//   - Create a new Request.
//   - Create a new Client based on the Spec.
//   - Call Generate on the Client.
//   - Compare the resulting Annotations with the ExpectedAnnotations, failing if there is a mismatch.
func (c GenerateTest) Run(t *testing.T) {
	ctx := context.Background()

	require.NotNil(t, c.Request)
	require.NotNil(t, c.Spec)

	request, err := c.Request.ToRequest(ctx)
	require.NoError(t, err)
	client, err := generate.NewClientForSpec(c.Spec)
	require.NoError(t, err)
	response, err := client.Generate(ctx, request)
	require.NoError(t, err)
	require.NoError(t, "TODO")
}

// RequestSpec specifies request parameters to be compiled for testing.
//
// This allows a Request to be built from a directory of .proto files.
type RequestSpec struct {
	// Files specifies the input files.
	//
	// Required.
	Files *descriptortest.ProtoFileSetSpec
	// Options are any options to pass to the plugin.
	Options map[string]any
}

// ToRequest converts the spec into a generate.Request.
//
// If r is nil, this returns nil.
func (r *RequestSpec) ToRequest(ctx context.Context) (generate.Request, error) {
	if r == nil {
		return nil, nil
	}

	if r.Files == nil {
		return nil, errors.New("RequestSpec.Files not set")
	}

	options, err := option.NewOptions(r.Options)
	if err != nil {
		return nil, err
	}
	requestOptions := []generate.RequestOption{
		generate.WithOptions(options),
	}

	fileDescriptors, err := r.Files.ToFileDescriptors(ctx)
	if err != nil {
		return nil, err
	}
	return generate.NewRequest(fileDescriptors, requestOptions...)
}
