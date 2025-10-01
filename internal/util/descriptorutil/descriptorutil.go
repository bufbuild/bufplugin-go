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

// Package descriptorutil contains extra utilities for FileDescriptors that we don't
// want to expose publicly, but want to use across multiple packages as part of
// bufplugin-go's implementation.
//
// This is not part of internal/pkg as internal/pkg packages should not rely on anything
// outside of internal/pkg.
package descriptorutil

import (
	"fmt"

	"buf.build/go/bufplugin/descriptor"
)

// ValidateFileDescriptors validates that the FileDescriptors are all unique by name.
func ValidateFileDescriptors(fileDescriptors []descriptor.FileDescriptor) error {
	_, err := FileNameToFileDescriptorForFileDescriptors(fileDescriptors)
	return err
}

// FileNameToFileDescriptorForFileDescriptors returns a map from file name to FileDescriptor
// for the given FileDescriptors.
//
// Returns error if there are non-unique names.
func FileNameToFileDescriptorForFileDescriptors(
	fileDescriptors []descriptor.FileDescriptor,
) (map[string]descriptor.FileDescriptor, error) {
	fileNameToFileDescriptor := make(map[string]descriptor.FileDescriptor, len(fileDescriptors))
	for _, fileDescriptor := range fileDescriptors {
		fileName := fileDescriptor.ProtoreflectFileDescriptor().Path()
		if _, ok := fileNameToFileDescriptor[fileName]; ok {
			return nil, fmt.Errorf("duplicate file name: %q", fileName)
		}
		fileNameToFileDescriptor[fileName] = fileDescriptor
	}
	return fileNameToFileDescriptor, nil
}
