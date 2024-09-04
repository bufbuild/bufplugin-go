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

package check

import (
	"fmt"
	"slices"

	checkv1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// File is an invidual file that should be checked.
//
// Both the protoreflect FileDescriptor and the raw FileDescriptorProto interacves
// are provided.
//
// Files also have the property of being imports or non-imports.
type File interface {
	// FileDescriptor returns the protoreflect FileDescriptor representing this File.
	//
	// This will always contain SourceCodeInfo.
	FileDescriptor() protoreflect.FileDescriptor
	// FileDescriptorProto returns the FileDescriptorProto representing this File.
	//
	// This is not a copy - do not modify!
	FileDescriptorProto() *descriptorpb.FileDescriptorProto
	// IsImport returns true if the File is an import.
	//
	// An import is a file that is either:
	//
	//   - A Well-Known Type included from the compiler and imported by a targeted file.
	//   - A file that was included from a Buf module dependency and imported by a targeted file.
	//   - A file that was not targeted, but was imported by a targeted file.
	//
	// We use "import" as this matches with the protoc concept of --include_imports, however
	// import is a bit of an overloaded term.
	IsImport() bool

	// IsSyntaxUnspecified denotes whether the file did not have a syntax explicitly specified.
	//
	// Per the FileDescriptorProto spec, it would be fine in this case to just leave the syntax field
	// unset to denote this and to set the syntax field to "proto2" if it is specified. However,
	// protoc does not set the syntax field if it was "proto2". Plugins may want to differentiate
	// between "proto2" and unset, and this field allows them to.
	IsSyntaxUnspecified() bool

	// UnusedDependencyIndexes are the indexes within the Dependency field on FileDescriptorProto for
	// those dependencies that are not used.
	//
	// This matches the shape of the PublicDependency and WeakDependency fields.
	UnusedDependencyIndexes() []int32

	toProto() *checkv1.File

	isFile()
}

// FilesForProtoFiles returns a new slice of Files for the given checkv1.Files.
func FilesForProtoFiles(protoFiles []*checkv1.File) ([]File, error) {
	if len(protoFiles) == 0 {
		return nil, nil
	}
	fileNameToProtoFile := make(map[string]*checkv1.File, len(protoFiles))
	fileDescriptorProtos := make([]*descriptorpb.FileDescriptorProto, len(protoFiles))
	for i, protoFile := range protoFiles {
		fileDescriptorProto := protoFile.GetFileDescriptorProto()
		fileName := fileDescriptorProto.GetName()
		if _, ok := fileNameToProtoFile[fileName]; ok {
			//  This should have been validated via protovalidate.
			return nil, fmt.Errorf("duplicate file name: %q", fileName)
		}
		fileDescriptorProtos[i] = fileDescriptorProto
		fileNameToProtoFile[fileName] = protoFile
	}

	protoregistryFiles, err := protodesc.NewFiles(
		&descriptorpb.FileDescriptorSet{
			File: fileDescriptorProtos,
		},
	)
	if err != nil {
		return nil, err
	}

	files := make([]File, 0, len(protoFiles))
	protoregistryFiles.RangeFiles(
		func(fileDescriptor protoreflect.FileDescriptor) bool {
			protoFile, ok := fileNameToProtoFile[fileDescriptor.Path()]
			if !ok {
				// If the protoreflect API is sane, this should never happen.
				// However, the protoreflect API is not sane.
				err = fmt.Errorf("unknown file: %q", fileDescriptor.Path())
				return false
			}
			files = append(
				files,
				newFile(
					fileDescriptor,
					protoFile.GetFileDescriptorProto(),
					protoFile.GetIsImport(),
					protoFile.GetIsSyntaxUnspecified(),
					protoFile.GetUnusedDependency(),
				),
			)
			return true
		},
	)
	if err != nil {
		return nil, err
	}
	if len(files) != len(protoFiles) {
		// If the protoreflect API is sane, this should never happen.
		// However, the protoreflect API is not sane.
		return nil, fmt.Errorf("expected %d files from protoregistry, got %d", len(protoFiles), len(files))
	}
	return files, nil
}

// *** PRIVATE ***

type file struct {
	fileDescriptor          protoreflect.FileDescriptor
	fileDescriptorProto     *descriptorpb.FileDescriptorProto
	isImport                bool
	isSyntaxUnspecified     bool
	unusedDependencyIndexes []int32
}

func newFile(
	fileDescriptor protoreflect.FileDescriptor,
	fileDescriptorProto *descriptorpb.FileDescriptorProto,
	isImport bool,
	isSyntaxUnspecified bool,
	unusedDependencyIndexes []int32,
) *file {
	return &file{
		fileDescriptor:          fileDescriptor,
		fileDescriptorProto:     fileDescriptorProto,
		isImport:                isImport,
		isSyntaxUnspecified:     isSyntaxUnspecified,
		unusedDependencyIndexes: unusedDependencyIndexes,
	}
}

func (f *file) FileDescriptor() protoreflect.FileDescriptor {
	return f.fileDescriptor
}

func (f *file) FileDescriptorProto() *descriptorpb.FileDescriptorProto {
	return f.fileDescriptorProto
}

func (f *file) IsImport() bool {
	return f.isImport
}

func (f *file) IsSyntaxUnspecified() bool {
	return f.isSyntaxUnspecified
}

func (f *file) UnusedDependencyIndexes() []int32 {
	return slices.Clone(f.unusedDependencyIndexes)
}

func (f *file) toProto() *checkv1.File {
	return &checkv1.File{
		FileDescriptorProto: f.fileDescriptorProto,
		IsImport:            f.isImport,
		IsSyntaxUnspecified: f.isSyntaxUnspecified,
		UnusedDependency:    f.unusedDependencyIndexes,
	}
}

func (*file) isFile() {}

func validateFiles(files []File) error {
	_, err := fileNameToFileForFiles(files)
	return err
}

func fileNameToFileForFiles(files []File) (map[string]File, error) {
	fileNameToFile := make(map[string]File, len(files))
	for _, file := range files {
		fileName := file.FileDescriptor().Path()
		if _, ok := fileNameToFile[fileName]; ok {
			return nil, fmt.Errorf("duplicate file name: %q", fileName)
		}
		fileNameToFile[fileName] = file
	}
	return fileNameToFile, nil
}
