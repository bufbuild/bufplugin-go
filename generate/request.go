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

package generate

import (
	"slices"

	generatev1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/generate/v1"
	"buf.build/go/bufplugin/descriptor"
	"buf.build/go/bufplugin/internal/bufpluginutil"
	"buf.build/go/bufplugin/internal/pkg/xslices"
	"buf.build/go/bufplugin/option"
)

// Request is a request to a plugin to run generates.
type Request interface {
	// FileDescriptors contains the FileDescriptors to generate.
	//
	// Will never be nil or empty.
	//
	// FileDescriptors are guaranteed to be unique with respect to their name.
	FileDescriptors() []descriptor.FileDescriptor
	// Options contains any options passed to the plugin.
	//
	// Will never be nil, but may have no values.
	Options() option.Options

	toProto() (*generatev1.GenerateRequest, error)

	isRequest()
}

// NewRequest returns a new Request for the given FileDescriptors.
func NewRequest(
	fileDescriptors []descriptor.FileDescriptor,
	options ...RequestOption,
) (Request, error) {
	return newRequest(fileDescriptors, options...)
}

// RequestOption is an option for a new Request.
type RequestOption func(*requestOptions)

// WithOption adds the given Options to the Request.
func WithOptions(options option.Options) RequestOption {
	return func(requestOptions *requestOptions) {
		requestOptions.options = options
	}
}

// RequestForProtoRequest returns a new Request for the given generatev1.Request.
func RequestForProtoRequest(protoRequest *generatev1.GenerateRequest) (Request, error) {
	fileDescriptors, err := descriptor.FileDescriptorsForProtoFileDescriptors(protoRequest.GetFileDescriptors())
	if err != nil {
		return nil, err
	}
	options, err := option.OptionsForProtoOptions(protoRequest.GetOptions())
	if err != nil {
		return nil, err
	}
	return NewRequest(
		fileDescriptors,
		WithOptions(options),
	)
}

// *** PRIVATE ***

type request struct {
	fileDescriptors []descriptor.FileDescriptor
	options         option.Options
}

func newRequest(
	fileDescriptors []descriptor.FileDescriptor,
	options ...RequestOption,
) (*request, error) {
	requestOptions := newRequestOptions()
	for _, option := range options {
		option(requestOptions)
	}
	if requestOptions.options == nil {
		requestOptions.options = option.EmptyOptions
	}
	if err := bufpluginutil.ValidateFileDescriptors(fileDescriptors); err != nil {
		return nil, err
	}
	return &request{
		fileDescriptors: fileDescriptors,
		options:         requestOptions.options,
	}, nil
}

func (r *request) FileDescriptors() []descriptor.FileDescriptor {
	return slices.Clone(r.fileDescriptors)
}

func (r *request) Options() option.Options {
	return r.options
}

func (r *request) toProto() (*generatev1.GenerateRequest, error) {
	if r == nil {
		return nil, nil
	}
	protoFileDescriptors := xslices.Map(r.fileDescriptors, descriptor.FileDescriptor.ToProto)
	protoOptions, err := r.options.ToProto()
	if err != nil {
		return nil, err
	}
	return &generatev1.GenerateRequest{
		FileDescriptors: protoFileDescriptors,
		Options:         protoOptions,
	}, nil
}

func (*request) isRequest() {}

type requestOptions struct {
	options option.Options
}

func newRequestOptions() *requestOptions {
	return &requestOptions{}
}
