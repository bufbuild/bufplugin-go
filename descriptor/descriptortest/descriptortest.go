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

// Package descriptortest provides testing helpers for descriptors when writing plugins.
package descriptortest

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	descriptorv1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/descriptor/v1"
	"buf.build/go/bufplugin/descriptor"
	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/protoutil"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/bufbuild/protocompile/wellknownimports"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ProtoFileSetSpec specifies files to be compiled for testing.
//
// This allows tests to effectively point at a directory, and get back descriptor.FileDesriptors.
type ProtoFileSetSpec struct {
	// DirPaths are the paths where .proto files are contained.
	//
	// Imports within .proto files should derive from one of these directories.
	// This must contain at least one element.
	//
	// This corresponds to the -I flag in protoc.
	DirPaths []string
	// FilePaths are the specific paths to build within the DirPaths.
	//
	// Any imports of the FilePaths will be built as well, and marked as imports.
	// This must contain at least one element.
	// FilePaths should be relative to DirPaths.
	//
	// This corresponds to arguments passed to protoc.
	FilePaths []string
}

// ToFileDescriptors compiles the files into descriptor.FileDescriptors.
//
// If p is nil, this returns an empty slice.
func (p *ProtoFileSetSpec) ToFileDescriptors(ctx context.Context) ([]descriptor.FileDescriptor, error) {
	if p == nil {
		return nil, nil
	}
	fileDescriptors, err := Compile(ctx, p.DirPaths, p.FilePaths)
	if err != nil {
		return nil, fmt.Errorf("could not compile ProtoFileSetSpec: %w", err)
	}
	return fileDescriptors, nil
}

// Compile compiles the given directory and file paths into descriptor.FileDescriptors.
//
// firPaths are the paths where .proto files are contained.
// Imports within .proto files should derive from one of these directories.
// This must contain at least one element.
// This corresponds to the -I flag in protoc.
//
// filePaths are the specific paths to build within the DirPaths.
//
// Any imports of the filePaths will be built as well, and marked as imports.
// This must contain at least one element.
// filePaths should be relative to dirPaths.
//
// This corresponds to arguments passed to protoc.
func Compile(ctx context.Context, dirPaths []string, filePaths []string) ([]descriptor.FileDescriptor, error) {
	if len(dirPaths) == 0 {
		return nil, errors.New("no dir paths specified")
	}
	if len(filePaths) == 0 {
		return nil, errors.New("no file paths specified")
	}
	dirPaths = fromSlashPaths(dirPaths)
	filePaths = fromSlashPaths(filePaths)
	toSlashFilePathMap := make(map[string]struct{}, len(filePaths))
	for _, filePath := range filePaths {
		toSlashFilePathMap[filepath.ToSlash(filePath)] = struct{}{}
	}

	var warningErrorsWithPos []reporter.ErrorWithPos
	compiler := protocompile.Compiler{
		Resolver: wellknownimports.WithStandardImports(
			&protocompile.SourceResolver{
				ImportPaths: dirPaths,
			},
		),
		Reporter: reporter.NewReporter(
			func(reporter.ErrorWithPos) error {
				return nil
			},
			func(errorWithPos reporter.ErrorWithPos) {
				warningErrorsWithPos = append(warningErrorsWithPos, errorWithPos)
			},
		),
		// This is what buf uses.
		SourceInfoMode: protocompile.SourceInfoExtraOptionLocations,
	}
	files, err := compiler.Compile(ctx, filePaths...)
	if err != nil {
		return nil, err
	}
	syntaxUnspecifiedFilePaths := make(map[string]struct{})
	filePathToUnusedDependencyFilePaths := make(map[string]map[string]struct{})
	for _, warningErrorWithPos := range warningErrorsWithPos {
		maybeAddSyntaxUnspecified(syntaxUnspecifiedFilePaths, warningErrorWithPos)
		maybeAddUnusedDependency(filePathToUnusedDependencyFilePaths, warningErrorWithPos)
	}
	fileDescriptorSet := fileDescriptorSetForFileDescriptors(files)

	protoFileDescriptors := make([]*descriptorv1.FileDescriptor, len(fileDescriptorSet.GetFile()))
	for i, fileDescriptorProto := range fileDescriptorSet.GetFile() {
		_, isNotImport := toSlashFilePathMap[fileDescriptorProto.GetName()]
		_, isSyntaxUnspecified := syntaxUnspecifiedFilePaths[fileDescriptorProto.GetName()]
		unusedDependencyIndexes := unusedDependencyIndexesForFilePathToUnusedDependencyFilePaths(
			fileDescriptorProto,
			filePathToUnusedDependencyFilePaths[fileDescriptorProto.GetName()],
		)
		protoFileDescriptors[i] = &descriptorv1.FileDescriptor{
			FileDescriptorProto: fileDescriptorProto,
			IsImport:            !isNotImport,
			IsSyntaxUnspecified: isSyntaxUnspecified,
			UnusedDependency:    unusedDependencyIndexes,
		}
	}
	return descriptor.FileDescriptorsForProtoFileDescriptors(protoFileDescriptors)
}

// *** PRIVATE ***

func unusedDependencyIndexesForFilePathToUnusedDependencyFilePaths(
	fileDescriptorProto *descriptorpb.FileDescriptorProto,
	unusedDependencyFilePaths map[string]struct{},
) []int32 {
	unusedDependencyIndexes := make([]int32, 0, len(unusedDependencyFilePaths))
	if len(unusedDependencyFilePaths) == 0 {
		return unusedDependencyIndexes
	}
	dependencyFilePaths := fileDescriptorProto.GetDependency()
	for i := 0; i < len(dependencyFilePaths); i++ {
		if _, ok := unusedDependencyFilePaths[dependencyFilePaths[i]]; ok {
			unusedDependencyIndexes = append(unusedDependencyIndexes, int32(i))
		}
	}
	return unusedDependencyIndexes
}

func maybeAddSyntaxUnspecified(
	syntaxUnspecifiedFilePaths map[string]struct{},
	errorWithPos reporter.ErrorWithPos,
) {
	if !errors.Is(errorWithPos, parser.ErrNoSyntax) {
		return
	}
	syntaxUnspecifiedFilePaths[errorWithPos.GetPosition().Filename] = struct{}{}
}

func maybeAddUnusedDependency(
	filePathToUnusedDependencyFilePaths map[string]map[string]struct{},
	errorWithPos reporter.ErrorWithPos,
) {
	var errorUnusedImport linker.ErrorUnusedImport
	if !errors.As(errorWithPos, &errorUnusedImport) {
		return
	}
	pos := errorWithPos.GetPosition()
	unusedDependencyFilePaths, ok := filePathToUnusedDependencyFilePaths[pos.Filename]
	if !ok {
		unusedDependencyFilePaths = make(map[string]struct{})
		filePathToUnusedDependencyFilePaths[pos.Filename] = unusedDependencyFilePaths
	}
	unusedDependencyFilePaths[errorUnusedImport.UnusedImport()] = struct{}{}
}

func fileDescriptorSetForFileDescriptors[D protoreflect.FileDescriptor](files []D) *descriptorpb.FileDescriptorSet {
	soFar := make(map[string]struct{}, len(files))
	slice := make([]*descriptorpb.FileDescriptorProto, 0, len(files))
	for _, file := range files {
		toFileDescriptorProtoSlice(file, &slice, soFar)
	}
	return &descriptorpb.FileDescriptorSet{File: slice}
}

func toFileDescriptorProtoSlice(file protoreflect.FileDescriptor, results *[]*descriptorpb.FileDescriptorProto, soFar map[string]struct{}) {
	if _, exists := soFar[file.Path()]; exists {
		return
	}
	soFar[file.Path()] = struct{}{}
	// Add dependencies first so the resulting slice is in topological order
	imports := file.Imports()
	for i, length := 0, imports.Len(); i < length; i++ {
		toFileDescriptorProtoSlice(imports.Get(i).FileDescriptor, results, soFar)
	}
	*results = append(*results, protoutil.ProtoFromFileDescriptor(file))
}

func fromSlashPaths(paths []string) []string {
	fromSlashPaths := make([]string, len(paths))
	for i, path := range paths {
		fromSlashPaths[i] = filepath.Clean(filepath.FromSlash(path))
	}
	return fromSlashPaths
}
