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

// Package checktest provides testing helpers when writing lint and breaking change plugins.
//
// The easiest entry point is TestCase. This allows you to set up a test and run it extremely
// easily. Other functions provide lower-level primitives if TestCase doesn't meet your needs.
package checktest

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	checkv1beta1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1beta1"
	"github.com/bufbuild/bufplugin-go/check"
	"github.com/bufbuild/bufplugin-go/internal/pkg/xslices"
	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/protoutil"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/bufbuild/protocompile/wellknownimports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// TestCase is a single test to run against a set of RuleSpecs.
type TestCase struct {
	// Spec is the Spec to test.
	//
	// Required.
	Spec *check.Spec
	// Files specifies the input files to test against.
	//
	// Required.
	Files *ProtoFileSpec
	// AgainstFiles specifies the input against files to test against, if anoy.
	AgainstFiles *ProtoFileSpec
	// Options are any options to pass to the plugin.
	Options map[string][]byte
	// ExpectedAnnotations are the expected Annotations that should be returned.
	ExpectedAnnotations []ExpectedAnnotation
}

// Run runs the TestCase.
//
// This will:
//
//   - Build the Files and AgainstFiles.
//   - Create a new Request.
//   - Create a new Client based on the RuleSpecs.
//   - Call Check on the Client.
//   - Compare the resulting Annotations with the ExpectedAnnotations, failing if there is a mismatch.
func (c TestCase) Run(t *testing.T) {
	ctx := context.Background()

	require.NotNil(t, c.Spec)
	require.NotNil(t, c.Files)

	againstFiles, err := c.AgainstFiles.Compile(ctx)
	require.NoError(t, err)
	requestOptions := []check.RequestOption{
		check.WithAgainstFiles(againstFiles),
	}
	for key, value := range c.Options {
		requestOptions = append(
			requestOptions,
			check.WithOption(key, value),
		)
	}

	files, err := c.Files.Compile(ctx)
	require.NoError(t, err)
	request, err := check.NewRequest(files, requestOptions...)
	require.NoError(t, err)
	client, err := check.NewClientForSpec(c.Spec)
	require.NoError(t, err)

	response, err := client.Check(ctx, request)
	require.NoError(t, err)
	AssertAnnotationsEqual(t, c.ExpectedAnnotations, response.Annotations())
}

// ProtoFileSpec specifies files to be compiled for testing.
//
// This allows tests to effectively point at a directory, and get back a
// *descriptorpb.FileDesriptorSet, or more to the point, check.Files
// that can be passed on a Request.
type ProtoFileSpec struct {
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
	// Paths should be relative to DirPaths.
	//
	// This corresponds to arguments passed to protoc.
	FilePaths []string
}

// Compile compiles the files into check.Files.
//
// If p is nil, this returns an empty slice.
func (p *ProtoFileSpec) Compile(ctx context.Context) ([]check.File, error) {
	if p == nil {
		return nil, nil
	}
	if err := validateProtoFileSpec(p); err != nil {
		return nil, err
	}
	return compile(ctx, p.DirPaths, p.FilePaths)
}

// ExpectedAnnotation contains the values expected from an Annotation.
type ExpectedAnnotation struct {
	// ID is the ID of the Rule.
	//
	// Required.
	ID string
	// Location is the location of the failure.
	Location *ExpectedLocation
	// AgainstLocation is the against location of the failure.
	AgainstLocation *ExpectedLocation
}

// ExpectedLocation contains the values expected from a Location.
type ExpectedLocation struct {
	// FileName is the name of the file.
	FileName string
	// StartLine is the zero-indexed start line.
	StartLine int
	// StartColumn is the zero-indexed start column.
	StartColumn int
	// EndLine is the zero-indexed end line.
	EndLine int
	// EndColumn is the zero-indexed end column.
	EndColumn int
}

// ExpectedAnnotationsForAnnotations returns ExpectedAnnotations for the given Annotations.
func ExpectedAnnotationsForAnnotations(annotations []check.Annotation) []ExpectedAnnotation {
	return xslices.Map(annotations, ExpectedAnnotationForAnnotation)
}

// ExpectedAnnotationForAnnotation returns an ExpectedAnnotation for the given Annotation.
func ExpectedAnnotationForAnnotation(annotation check.Annotation) ExpectedAnnotation {
	expectedAnnotation := ExpectedAnnotation{
		ID: annotation.ID(),
	}
	if location := annotation.Location(); location != nil {
		expectedAnnotation.Location = &ExpectedLocation{
			FileName:    location.FileName(),
			StartLine:   location.StartLine(),
			StartColumn: location.StartColumn(),
			EndLine:     location.EndLine(),
			EndColumn:   location.EndColumn(),
		}
	}
	if againstLocation := annotation.AgainstLocation(); againstLocation != nil {
		expectedAnnotation.AgainstLocation = &ExpectedLocation{
			FileName:    againstLocation.FileName(),
			StartLine:   againstLocation.StartLine(),
			StartColumn: againstLocation.StartColumn(),
			EndLine:     againstLocation.EndLine(),
			EndColumn:   againstLocation.EndColumn(),
		}
	}
	return expectedAnnotation
}

// AssertAnnotationsEqual asserts that the Annotations equal the expected Annotations.
func AssertAnnotationsEqual(t *testing.T, expectedAnnotations []ExpectedAnnotation, actualAnnotations []check.Annotation) {
	if len(expectedAnnotations) == 0 {
		expectedAnnotations = nil
	}
	if len(actualAnnotations) == 0 {
		actualAnnotations = nil
	}
	assert.Equal(t, expectedAnnotations, ExpectedAnnotationsForAnnotations(actualAnnotations))
}

// RequireAnnotationsEqual requires that the Annotations equal the expected Annotations.
func RequireAnnotationsEqual(t *testing.T, expectedAnnotations []ExpectedAnnotation, actualAnnotations []check.Annotation) {
	if len(expectedAnnotations) == 0 {
		expectedAnnotations = nil
	}
	if len(actualAnnotations) == 0 {
		actualAnnotations = nil
	}
	require.Equal(t, expectedAnnotations, ExpectedAnnotationsForAnnotations(actualAnnotations))
}

// *** PRIVATE ***

func validateProtoFileSpec(protoFileSpec *ProtoFileSpec) error {
	if len(protoFileSpec.DirPaths) == 0 {
		return errors.New("no DirPaths specified on ProtoFileSpec")
	}
	if len(protoFileSpec.FilePaths) == 0 {
		return errors.New("no FilePaths specified on ProtoFileSpec")
	}
	return nil
}

func compile(ctx context.Context, dirPaths []string, filePaths []string) ([]check.File, error) {
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
			func(errorWithPos reporter.ErrorWithPos) error {
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

	protoFiles := make([]*checkv1beta1.File, len(fileDescriptorSet.GetFile()))
	for i, fileDescriptorProto := range fileDescriptorSet.GetFile() {
		_, isNotImport := toSlashFilePathMap[fileDescriptorProto.GetName()]
		_, isSyntaxUnspecified := syntaxUnspecifiedFilePaths[fileDescriptorProto.GetName()]
		unusedDependencyIndexes := unusedDependencyIndexesForFilePathToUnusedDependencyFilePaths(
			fileDescriptorProto,
			filePathToUnusedDependencyFilePaths[fileDescriptorProto.GetName()],
		)
		protoFiles[i] = &checkv1beta1.File{
			FileDescriptorProto: fileDescriptorProto,
			IsImport:            !isNotImport,
			IsSyntaxUnspecified: isSyntaxUnspecified,
			UnusedDependency:    unusedDependencyIndexes,
		}
	}
	return check.FilesForProtoFiles(protoFiles)
}

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
