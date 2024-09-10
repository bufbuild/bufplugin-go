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

	"buf.build/go/bufplugin/check"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// NewFilePairRuleHandler returns a new RuleHandler that will call f for every file pair
// within the check.Request's Files() and AgainstFiles().
//
// The Files will be paired up by name. Files that cannot be paired up are skipped.
//
// This is typically used for breaking change Rules.
func NewFilePairRuleHandler(
	f func(
		ctx context.Context,
		responseWriter check.ResponseWriter,
		request check.Request,
		file check.File,
		againstFile check.File,
	) error,
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
			files := filterFiles(request.Files(), iteratorOptions.withoutImports)
			againstFiles := filterFiles(request.AgainstFiles(), iteratorOptions.withoutImports)
			pathToFile, err := getPathToFile(files)
			if err != nil {
				return err
			}
			againstPathToFile, err := getPathToFile(againstFiles)
			if err != nil {
				return err
			}
			for againstPath, againstFile := range againstPathToFile {
				if file, ok := pathToFile[againstPath]; ok {
					if err = f(ctx, responseWriter, request, file, againstFile); err != nil {
						return err
					}
				}
			}
			return nil
		},
	)
}

// NewEnumPairRuleHandler returns a new RuleHandler that will call f for every enum pair
// within the check.Request's Files() and AgainstFiles().
//
// The enums will be paired up by fully-qualified name. Enums that cannot be paired up are skipped.
//
// This is typically used for breaking change Rules.
func NewEnumPairRuleHandler(
	f func(
		ctx context.Context,
		responseWriter check.ResponseWriter,
		request check.Request,
		enumDescriptor protoreflect.EnumDescriptor,
		againstEnumDescriptor protoreflect.EnumDescriptor,
	) error,
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
			files := filterFiles(request.Files(), iteratorOptions.withoutImports)
			againstFiles := filterFiles(request.AgainstFiles(), iteratorOptions.withoutImports)
			fullNameToEnumDescriptor, err := getFullNameToEnumDescriptor(files)
			if err != nil {
				return err
			}
			againstFullNameToEnumDescriptor, err := getFullNameToEnumDescriptor(againstFiles)
			if err != nil {
				return err
			}
			for againstFullName, againstEnumDescriptor := range againstFullNameToEnumDescriptor {
				if enumDescriptor, ok := fullNameToEnumDescriptor[againstFullName]; ok {
					if err = f(ctx, responseWriter, request, enumDescriptor, againstEnumDescriptor); err != nil {
						return err
					}
				}
			}
			return nil
		},
	)
}

// NewMessagePairRuleHandler returns a new RuleHandler that will call f for every message pair
// within the check.Request's Files() and AgainstFiles().
//
// The messages will be paired up by fully-qualified name. Messages that cannot be paired up are skipped.
//
// This is typically used for breaking change Rules.
func NewMessagePairRuleHandler(
	f func(
		ctx context.Context,
		responseWriter check.ResponseWriter,
		request check.Request,
		messageDescriptor protoreflect.MessageDescriptor,
		againstMessageDescriptor protoreflect.MessageDescriptor,
	) error,
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
			files := filterFiles(request.Files(), iteratorOptions.withoutImports)
			againstFiles := filterFiles(request.AgainstFiles(), iteratorOptions.withoutImports)
			fullNameToMessageDescriptor, err := getFullNameToMessageDescriptor(files)
			if err != nil {
				return err
			}
			againstFullNameToMessageDescriptor, err := getFullNameToMessageDescriptor(againstFiles)
			if err != nil {
				return err
			}
			for againstFullName, againstMessageDescriptor := range againstFullNameToMessageDescriptor {
				if messageDescriptor, ok := fullNameToMessageDescriptor[againstFullName]; ok {
					if err = f(ctx, responseWriter, request, messageDescriptor, againstMessageDescriptor); err != nil {
						return err
					}
				}
			}
			return nil
		},
	)
}

// NewFieldPairRuleHandler returns a new RuleHandler that will call f for every field pair
// within the check.Request's Files() and AgainstFiles().
//
// The fields will be paired up by the fully-qualified name of the message, and the field number.
// Fields that cannot be paired up are skipped.
//
// This includes extensions.
//
// This is typically used for breaking change Rules.
func NewFieldPairRuleHandler(
	f func(
		ctx context.Context,
		responseWriter check.ResponseWriter,
		request check.Request,
		fieldDescriptor protoreflect.FieldDescriptor,
		againstFieldDescriptor protoreflect.FieldDescriptor,
	) error,
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
			files := filterFiles(request.Files(), iteratorOptions.withoutImports)
			againstFiles := filterFiles(request.AgainstFiles(), iteratorOptions.withoutImports)
			containingMessageFullNameToNumberToFieldDescriptor, err := getContainingMessageFullNameToNumberToFieldDescriptor(files)
			if err != nil {
				return err
			}
			againstContainingMessageFullNameToNumberToFieldDescriptor, err := getContainingMessageFullNameToNumberToFieldDescriptor(againstFiles)
			if err != nil {
				return err
			}
			for againstContainingMessageFullName, againstNumberToFieldDescriptor := range againstContainingMessageFullNameToNumberToFieldDescriptor {
				if numberToFieldDescriptor, ok := containingMessageFullNameToNumberToFieldDescriptor[againstContainingMessageFullName]; ok {
					for againstNumber, againstFieldDescriptor := range againstNumberToFieldDescriptor {
						if fieldDescriptor, ok := numberToFieldDescriptor[againstNumber]; ok {
							if err = f(ctx, responseWriter, request, fieldDescriptor, againstFieldDescriptor); err != nil {
								return err
							}
						}
					}
				}
			}
			return nil
		},
	)
}

// NewServicePairRuleHandler returns a new RuleHandler that will call f for every service pair
// within the check.Request's Files() and AgainstFiles().
//
// The services will be paired up by fully-qualified name. Services that cannot be paired up are skipped.
//
// This is typically used for breaking change Rules.
func NewServicePairRuleHandler(
	f func(
		ctx context.Context,
		responseWriter check.ResponseWriter,
		request check.Request,
		serviceDescriptor protoreflect.ServiceDescriptor,
		againstServiceDescriptor protoreflect.ServiceDescriptor,
	) error,
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
			files := filterFiles(request.Files(), iteratorOptions.withoutImports)
			againstFiles := filterFiles(request.AgainstFiles(), iteratorOptions.withoutImports)
			fullNameToServiceDescriptor, err := getFullNameToServiceDescriptor(files)
			if err != nil {
				return err
			}
			againstFullNameToServiceDescriptor, err := getFullNameToServiceDescriptor(againstFiles)
			if err != nil {
				return err
			}
			for againstFullName, againstServiceDescriptor := range againstFullNameToServiceDescriptor {
				if serviceDescriptor, ok := fullNameToServiceDescriptor[againstFullName]; ok {
					if err = f(ctx, responseWriter, request, serviceDescriptor, againstServiceDescriptor); err != nil {
						return err
					}
				}
			}
			return nil
		},
	)
}

// NewMethodPairRuleHandler returns a new RuleHandler that will call f for every method pair
// within the check.Request's Files() and AgainstFiles().
//
// The services will be paired up by fully-qualified name of the service, and name of the method.
// Methods that cannot be paired up are skipped.
//
// This is typically used for breaking change Rules.
func NewMethodPairRuleHandler(
	f func(
		ctx context.Context,
		responseWriter check.ResponseWriter,
		request check.Request,
		methodDescriptor protoreflect.MethodDescriptor,
		againstMethodDescriptor protoreflect.MethodDescriptor,
	) error,
	options ...IteratorOption,
) check.RuleHandler {
	return NewServicePairRuleHandler(
		func(
			ctx context.Context,
			responseWriter check.ResponseWriter,
			request check.Request,
			serviceDescriptor protoreflect.ServiceDescriptor,
			againstServiceDescriptor protoreflect.ServiceDescriptor,
		) error {
			nameToMethodDescriptor, err := getNameToMethodDescriptor(serviceDescriptor)
			if err != nil {
				return err
			}
			againstNameToMethodDescriptor, err := getNameToMethodDescriptor(againstServiceDescriptor)
			if err != nil {
				return err
			}
			for againstName, againstMethodDescriptor := range againstNameToMethodDescriptor {
				if methodDescriptor, ok := nameToMethodDescriptor[againstName]; ok {
					if err = f(ctx, responseWriter, request, methodDescriptor, againstMethodDescriptor); err != nil {
						return err
					}
				}
			}
			return nil
		},
		options...,
	)
}
