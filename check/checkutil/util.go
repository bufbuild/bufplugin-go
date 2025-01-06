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

package checkutil

import (
	"fmt"
	"sort"

	"buf.build/go/bufplugin/descriptor"
	"buf.build/go/bufplugin/internal/pkg/xslices"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type container interface {
	Enums() protoreflect.EnumDescriptors
	Messages() protoreflect.MessageDescriptors
	Extensions() protoreflect.ExtensionDescriptors
}

func getPathToFileDescriptor(fileDescriptors []descriptor.FileDescriptor) (map[string]descriptor.FileDescriptor, error) {
	pathToFileDescriptorMap := make(map[string]descriptor.FileDescriptor, len(fileDescriptors))
	for _, fileDescriptor := range fileDescriptors {
		path := fileDescriptor.ProtoreflectFileDescriptor().Path()
		if _, ok := pathToFileDescriptorMap[path]; ok {
			return nil, fmt.Errorf("duplicate file: %q", path)
		}
		pathToFileDescriptorMap[path] = fileDescriptor
	}
	return pathToFileDescriptorMap, nil
}

func getFullNameToEnumDescriptor(fileDescriptors []descriptor.FileDescriptor) (map[protoreflect.FullName]protoreflect.EnumDescriptor, error) {
	fullNameToEnumDescriptorMap := make(map[protoreflect.FullName]protoreflect.EnumDescriptor)
	for _, fileDescriptor := range fileDescriptors {
		if err := forEachEnum(
			fileDescriptor.ProtoreflectFileDescriptor(),
			func(enumDescriptor protoreflect.EnumDescriptor) error {
				fullName := enumDescriptor.FullName()
				if _, ok := fullNameToEnumDescriptorMap[fullName]; ok {
					return fmt.Errorf("duplicate enum: %q", fullName)
				}
				fullNameToEnumDescriptorMap[fullName] = enumDescriptor
				return nil
			},
		); err != nil {
			return nil, err
		}
	}
	return fullNameToEnumDescriptorMap, nil
}

// Keeping this function around for now, this is to suppress lint unused.
var _ = getNumberToEnumValueDescriptors

func getNumberToEnumValueDescriptors(enumDescriptor protoreflect.EnumDescriptor) (map[protoreflect.EnumNumber][]protoreflect.EnumValueDescriptor, error) {
	numberToEnumValueDescriptorsMap := make(map[protoreflect.EnumNumber][]protoreflect.EnumValueDescriptor)
	if err := forEachEnumValue(
		enumDescriptor,
		func(enumValueDescriptor protoreflect.EnumValueDescriptor) error {
			numberToEnumValueDescriptorsMap[enumValueDescriptor.Number()] = append(
				numberToEnumValueDescriptorsMap[enumValueDescriptor.Number()],
				enumValueDescriptor,
			)
			return nil
		},
	); err != nil {
		return nil, err
	}
	for _, enumValueDescriptors := range numberToEnumValueDescriptorsMap {
		sort.Slice(
			enumValueDescriptors,
			func(i int, j int) bool {
				return enumValueDescriptors[i].Name() < enumValueDescriptors[j].Name()
			},
		)
	}
	return numberToEnumValueDescriptorsMap, nil
}

func getFullNameToMessageDescriptor(fileDescriptors []descriptor.FileDescriptor) (map[protoreflect.FullName]protoreflect.MessageDescriptor, error) {
	fullNameToMessageDescriptorMap := make(map[protoreflect.FullName]protoreflect.MessageDescriptor)
	for _, fileDescriptor := range fileDescriptors {
		if err := forEachMessage(
			fileDescriptor.ProtoreflectFileDescriptor(),
			func(messageDescriptor protoreflect.MessageDescriptor) error {
				fullName := messageDescriptor.FullName()
				if _, ok := fullNameToMessageDescriptorMap[fullName]; ok {
					return fmt.Errorf("duplicate message: %q", fullName)
				}
				fullNameToMessageDescriptorMap[fullName] = messageDescriptor
				return nil
			},
		); err != nil {
			return nil, err
		}
	}
	return fullNameToMessageDescriptorMap, nil
}

func getContainingMessageFullNameToNumberToFieldDescriptor(
	fileDescriptors []descriptor.FileDescriptor,
) (map[protoreflect.FullName]map[protoreflect.FieldNumber]protoreflect.FieldDescriptor, error) {
	containingMessageFullNameToNumberToFieldDescriptorMap := make(
		map[protoreflect.FullName]map[protoreflect.FieldNumber]protoreflect.FieldDescriptor,
	)
	for _, fileDescriptor := range fileDescriptors {
		if err := forEachField(
			fileDescriptor.ProtoreflectFileDescriptor(),
			func(fieldDescriptor protoreflect.FieldDescriptor) error {
				number := fieldDescriptor.Number()
				containingMessage := fieldDescriptor.ContainingMessage()
				if containingMessage == nil {
					return fmt.Errorf("containing message was nil for field %d", number)
				}
				fullName := containingMessage.FullName()
				numberToFieldDescriptor, ok := containingMessageFullNameToNumberToFieldDescriptorMap[fullName]
				if !ok {
					numberToFieldDescriptor = make(map[protoreflect.FieldNumber]protoreflect.FieldDescriptor)
					containingMessageFullNameToNumberToFieldDescriptorMap[fullName] = numberToFieldDescriptor
				}
				if _, ok := numberToFieldDescriptor[number]; ok {
					return fmt.Errorf("duplicate field on message %q: %d", fullName, number)
				}
				numberToFieldDescriptor[number] = fieldDescriptor
				return nil
			},
		); err != nil {
			return nil, err
		}
	}
	return containingMessageFullNameToNumberToFieldDescriptorMap, nil
}

func getFullNameToServiceDescriptor(fileDescriptors []descriptor.FileDescriptor) (map[protoreflect.FullName]protoreflect.ServiceDescriptor, error) {
	fullNameToServiceDescriptorMap := make(map[protoreflect.FullName]protoreflect.ServiceDescriptor)
	for _, fileDescriptor := range fileDescriptors {
		if err := forEachService(
			fileDescriptor.ProtoreflectFileDescriptor(),
			func(serviceDescriptor protoreflect.ServiceDescriptor) error {
				fullName := serviceDescriptor.FullName()
				if _, ok := fullNameToServiceDescriptorMap[fullName]; ok {
					return fmt.Errorf("duplicate service: %q", fullName)
				}
				fullNameToServiceDescriptorMap[fullName] = serviceDescriptor
				return nil
			},
		); err != nil {
			return nil, err
		}
	}
	return fullNameToServiceDescriptorMap, nil
}

func getNameToMethodDescriptor(serviceDescriptor protoreflect.ServiceDescriptor) (map[protoreflect.Name]protoreflect.MethodDescriptor, error) {
	nameToMethodDescriptorMap := make(map[protoreflect.Name]protoreflect.MethodDescriptor)
	if err := forEachMethod(
		serviceDescriptor,
		func(methodDescriptor protoreflect.MethodDescriptor) error {
			name := methodDescriptor.Name()
			if _, ok := nameToMethodDescriptorMap[name]; ok {
				return fmt.Errorf("duplicate method on service %q: %q", serviceDescriptor.FullName(), name)
			}
			nameToMethodDescriptorMap[name] = methodDescriptor
			return nil
		},
	); err != nil {
		return nil, err
	}
	return nameToMethodDescriptorMap, nil
}

func forEachFileImport(
	fileDescriptor protoreflect.FileDescriptor,
	f func(protoreflect.FileImport) error,
) error {
	fileImports := fileDescriptor.Imports()
	for i := 0; i < fileImports.Len(); i++ {
		if err := f(fileImports.Get(i)); err != nil {
			return err
		}
	}
	return nil
}

func forEachEnum(
	container container,
	f func(protoreflect.EnumDescriptor) error,
) error {
	enums := container.Enums()
	for i := 0; i < enums.Len(); i++ {
		if err := f(enums.Get(i)); err != nil {
			return err
		}
	}
	messages := container.Messages()
	for i := 0; i < messages.Len(); i++ {
		// Nested enums.
		if err := forEachEnum(messages.Get(i), f); err != nil {
			return err
		}
	}
	return nil
}

func forEachEnumValue(
	enumDescriptor protoreflect.EnumDescriptor,
	f func(protoreflect.EnumValueDescriptor) error,
) error {
	enumValues := enumDescriptor.Values()
	for i := 0; i < enumValues.Len(); i++ {
		if err := f(enumValues.Get(i)); err != nil {
			return err
		}
	}
	return nil
}

func forEachMessage(
	container container,
	f func(protoreflect.MessageDescriptor) error,
) error {
	messages := container.Messages()
	for i := 0; i < messages.Len(); i++ {
		messageDescriptor := messages.Get(i)
		if err := f(messageDescriptor); err != nil {
			return err
		}
		// Nested messages.
		if err := forEachMessage(messageDescriptor, f); err != nil {
			return err
		}
	}
	return nil
}

func forEachField(
	container container,
	f func(protoreflect.FieldDescriptor) error,
) error {
	if err := forEachMessage(
		container,
		func(messageDescriptor protoreflect.MessageDescriptor) error {
			fields := messageDescriptor.Fields()
			for i := 0; i < fields.Len(); i++ {
				if err := f(fields.Get(i)); err != nil {
					return err
				}
			}
			extensions := messageDescriptor.Extensions()
			for i := 0; i < extensions.Len(); i++ {
				if err := f(extensions.Get(i)); err != nil {
					return err
				}
			}
			return nil
		},
	); err != nil {
		return err
	}
	extensions := container.Extensions()
	for i := 0; i < extensions.Len(); i++ {
		if err := f(extensions.Get(i)); err != nil {
			return err
		}
	}
	return nil
}

func forEachOneof(
	messageDescriptor protoreflect.MessageDescriptor,
	f func(protoreflect.OneofDescriptor) error,
) error {
	oneofs := messageDescriptor.Oneofs()
	for i := 0; i < oneofs.Len(); i++ {
		if err := f(oneofs.Get(i)); err != nil {
			return err
		}
	}
	return nil
}

func forEachService(
	fileDescriptor protoreflect.FileDescriptor,
	f func(protoreflect.ServiceDescriptor) error,
) error {
	services := fileDescriptor.Services()
	for i := 0; i < services.Len(); i++ {
		if err := f(services.Get(i)); err != nil {
			return err
		}
	}
	return nil
}

func forEachMethod(
	serviceDescriptor protoreflect.ServiceDescriptor,
	f func(protoreflect.MethodDescriptor) error,
) error {
	methods := serviceDescriptor.Methods()
	for i := 0; i < methods.Len(); i++ {
		if err := f(methods.Get(i)); err != nil {
			return err
		}
	}
	return nil
}

func filterFileDescriptors(fileDescriptors []descriptor.FileDescriptor, withoutImports bool) []descriptor.FileDescriptor {
	if !withoutImports {
		return fileDescriptors
	}
	return xslices.Filter(fileDescriptors, func(fileDescriptor descriptor.FileDescriptor) bool { return !fileDescriptor.IsImport() })
}
