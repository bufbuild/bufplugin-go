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

package checkutil

import (
	"context"

	"github.com/bufbuild/bufplugin-go/check"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// NewFileRuleHandler returns a new RuleHandler that will call f for every file
// within the check.Request's Files().
//
// This is typically used for lint Rules. Most callers will use the WithoutImports() options.
func NewFileRuleHandler(
	f func(context.Context, check.ResponseWriter, check.Request, check.File) error,
	options ...IteratorOption,
) check.RuleHandler {
	iteratorOptions := newIteratorOptions()
	for _, option := range options {
		option(iteratorOptions)
	}
	return check.RuleHandlerFunc(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
		) error {
			for _, file := range request.Files() {
				if iteratorOptions.withoutImports && file.IsImport() {
					continue
				}
				if err := f(ctx, responseWriter, request, file); err != nil {
					return err
				}
			}
			return nil
		},
	)
}

// NewFileImportRuleHandler returns a new RuleHandler that will call f for every "import" statement
// within the check.Request's Files().
//
// Note that terms are overloaded here: check.File.IsImport denotes whether the File is an import
// itself, while this iterates over the protoreflect.FileImports within each File. The option
// WithoutImports() is a separate concern - NewFileImportRuleHandler(f, WithoutImports()) will
// iterate over all the FileImports for the non-import Files.
//
// This is typically used for lint Rules. Most callers will use the WithoutImports() options.
func NewFileImportRuleHandler(
	f func(context.Context, check.ResponseWriter, check.Request, protoreflect.FileImport) error,
	options ...IteratorOption,
) check.RuleHandler {
	return NewFileRuleHandler(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
			file check.File,
		) error {
			return forEachFileImport(
				file.FileDescriptor(),
				func(fileImport protoreflect.FileImport) error {
					return f(ctx, responseWriter, request, fileImport)
				},
			)
		},
		options...,
	)
}

// NewEnumRuleHandler returns a new RuleHandler that will call f for every enum
// within the check.Request's Files().
//
// This is typically used for lint Rules. Most callers will use the WithoutImports() options.
func NewEnumRuleHandler(
	f func(context.Context, check.ResponseWriter, check.Request, protoreflect.EnumDescriptor) error,
	options ...IteratorOption,
) check.RuleHandler {
	return NewFileRuleHandler(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
			file check.File,
		) error {
			return forEachEnum(
				file.FileDescriptor(),
				func(enumDescriptor protoreflect.EnumDescriptor) error {
					return f(ctx, responseWriter, request, enumDescriptor)
				},
			)
		},
		options...,
	)
}

// NewEnumValueRuleHandler returns a new RuleHandler that will call f for every value in every enum
// within the check.Request's Files().
//
// This is typically used for lint Rules. Most callers will use the WithoutImports() options.
func NewEnumValueRuleHandler(
	f func(context.Context, check.ResponseWriter, check.Request, protoreflect.EnumValueDescriptor) error,
	options ...IteratorOption,
) check.RuleHandler {
	return NewEnumRuleHandler(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
			enumDescriptor protoreflect.EnumDescriptor,
		) error {
			return forEachEnumValue(
				enumDescriptor,
				func(enumValueDescriptor protoreflect.EnumValueDescriptor) error {
					return f(ctx, responseWriter, request, enumValueDescriptor)
				},
			)
		},
		options...,
	)
}

// NewMessageRuleHandler returns a new RuleHandler that will call f for every message
// within the check.Request's Files().
//
// This is typically used for lint Rules. Most callers will use the WithoutImports() options.
func NewMessageRuleHandler(
	f func(context.Context, check.ResponseWriter, check.Request, protoreflect.MessageDescriptor) error,
	options ...IteratorOption,
) check.RuleHandler {
	return NewFileRuleHandler(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
			file check.File,
		) error {
			return forEachMessage(
				file.FileDescriptor(),
				func(messageDescriptor protoreflect.MessageDescriptor) error {
					return f(ctx, responseWriter, request, messageDescriptor)
				},
			)
		},
		options...,
	)
}

// NewFieldRuleHandler returns a new RuleHandler that will call f for every field in every message
// within the check.Request's Files().
//
// This includes extensions.
//
// This is typically used for lint Rules. Most callers will use the WithoutImports() options.
func NewFieldRuleHandler(
	f func(context.Context, check.ResponseWriter, check.Request, protoreflect.FieldDescriptor) error,
	options ...IteratorOption,
) check.RuleHandler {
	return NewFileRuleHandler(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
			file check.File,
		) error {
			return forEachField(
				file.FileDescriptor(),
				func(fieldDescriptor protoreflect.FieldDescriptor) error {
					return f(ctx, responseWriter, request, fieldDescriptor)
				},
			)
		},
		options...,
	)
}

// NewOneofRuleHandler returns a new RuleHandler that will call f for every oneof in every message
// within the check.Request's Files().
//
// This is typically used for lint Rules. Most callers will use the WithoutImports() options.
func NewOneofRuleHandler(
	f func(context.Context, check.ResponseWriter, check.Request, protoreflect.OneofDescriptor) error,
	options ...IteratorOption,
) check.RuleHandler {
	return NewMessageRuleHandler(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
			messageDescriptor protoreflect.MessageDescriptor,
		) error {
			return forEachOneof(
				messageDescriptor,
				func(oneofDescriptor protoreflect.OneofDescriptor) error {
					return f(ctx, responseWriter, request, oneofDescriptor)
				},
			)
		},
		options...,
	)
}

// NewServiceRuleHandler returns a new RuleHandler that will call f for every service
// within the check.Request's Files().
//
// This is typically used for lint Rules. Most callers will use the WithoutImports() options.
func NewServiceRuleHandler(
	f func(context.Context, check.ResponseWriter, check.Request, protoreflect.ServiceDescriptor) error,
	options ...IteratorOption,
) check.RuleHandler {
	return NewFileRuleHandler(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
			file check.File,
		) error {
			return forEachService(
				file.FileDescriptor(),
				func(serviceDescriptor protoreflect.ServiceDescriptor) error {
					return f(ctx, responseWriter, request, serviceDescriptor)
				},
			)
		},
		options...,
	)
}

// NewMethodRuleHandler returns a new RuleHandler that will call f for every method in every service
// within the check.Request's Files().
//
// This is typically used for lint Rules. Most callers will use the WithoutImports() options.
func NewMethodRuleHandler(
	f func(context.Context, check.ResponseWriter, check.Request, protoreflect.MethodDescriptor) error,
	options ...IteratorOption,
) check.RuleHandler {
	return NewServiceRuleHandler(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
			serviceDescriptor protoreflect.ServiceDescriptor,
		) error {
			return forEachMethod(
				serviceDescriptor,
				func(methodDescriptor protoreflect.MethodDescriptor) error {
					return f(ctx, responseWriter, request, methodDescriptor)
				},
			)
		},
		options...,
	)
}
