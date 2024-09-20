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
	"strconv"
	"testing"

	descriptorv1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/descriptor/v1"
	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/descriptor"
	"buf.build/go/bufplugin/internal/pkg/xslices"
	"buf.build/go/bufplugin/option"
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

// SpecTest tests your spec with check.ValidateSpec.
//
// Almost every plugin should run a test with SpecTest.
//
//	func TestSpec(t *testing.T) {
//	  t.Parallel()
//	  checktest.SpecTest(t, yourSpec)
//	}
func SpecTest(t *testing.T, spec *check.Spec) {
	require.NoError(t, check.ValidateSpec(spec))
}

// CheckTest is a single Check test to run against a Spec.
type CheckTest struct {
	// Request is the request spec to test.
	Request *RequestSpec
	// Spec is the Spec to test.
	//
	// Required.
	Spec *check.Spec
	// ExpectedAnnotations are the expected Annotations that should be returned.
	ExpectedAnnotations []ExpectedAnnotation
}

// Run runs the test.
//
// This will:
//
//   - Build the Files and AgainstFiles.
//   - Create a new Request.
//   - Create a new Client based on the Spec.
//   - Call Check on the Client.
//   - Compare the resulting Annotations with the ExpectedAnnotations, failing if there is a mismatch.
func (c CheckTest) Run(t *testing.T) {
	ctx := context.Background()

	require.NotNil(t, c.Request)
	require.NotNil(t, c.Spec)

	request, err := c.Request.ToRequest(ctx)
	require.NoError(t, err)
	client, err := check.NewClientForSpec(c.Spec)
	require.NoError(t, err)
	response, err := client.Check(ctx, request)
	require.NoError(t, err)
	AssertAnnotationsEqual(t, c.ExpectedAnnotations, response.Annotations())
}

// RequestSpec specifies request parameters to be compiled for testing.
//
// This allows a Request to be built from a directory of .proto files.
type RequestSpec struct {
	// Files specifies the input files to test against.
	//
	// Required.
	Files *ProtoFileSpec
	// AgainstFiles specifies the input against files to test against, if anoy.
	AgainstFiles *ProtoFileSpec
	// RuleIDs are the specific RuleIDs to run.
	RuleIDs []string
	// Options are any options to pass to the plugin.
	Options map[string]any
}

// ToRequest converts the spec into a check.Request.
//
// If r is nil, this returns nil.
func (r *RequestSpec) ToRequest(ctx context.Context) (check.Request, error) {
	if r == nil {
		return nil, nil
	}

	if r.Files == nil {
		return nil, errors.New("RequestSpec.Files not set")
	}

	againstFileDescriptors, err := r.AgainstFiles.ToFileDescriptors(ctx)
	if err != nil {
		return nil, err
	}
	options, err := option.NewOptions(r.Options)
	if err != nil {
		return nil, err
	}
	requestOptions := []check.RequestOption{
		check.WithAgainstFileDescriptors(againstFileDescriptors),
		check.WithOptions(options),
		check.WithRuleIDs(r.RuleIDs...),
	}

	fileDescriptors, err := r.Files.ToFileDescriptors(ctx)
	if err != nil {
		return nil, err
	}
	return check.NewRequest(fileDescriptors, requestOptions...)
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

// ToFileDescriptors compiles the files into descriptor.FileDescriptors.
//
// If p is nil, this returns an empty slice.
func (p *ProtoFileSpec) ToFileDescriptors(ctx context.Context) ([]descriptor.FileDescriptor, error) {
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
	// RuleID is the ID of the Rule.
	//
	// Required.
	RuleID string
	// Message is the message returned from the annoation.
	//
	// If Message is not set on ExpectedAnnotation, this field will *not* be compared
	// against the value in Annotation. That is, it is valid to have an Annotation return
	// a message but to not set it on ExpectedAnnotation.
	Message string
	// FileLocation is the location of the failure.
	FileLocation *ExpectedFileLocation
	// AgainstFileLocation is the against location of the failure.
	AgainstFileLocation *ExpectedFileLocation
}

// String implements fmt.Stringer.
func (ea ExpectedAnnotation) String() string {
	return "ruleID=\"" + ea.RuleID + "\"" +
		" message=\"" + ea.Message + "\"" +
		" location=\"" + ea.FileLocation.String() + "\"" +
		" againstLocation=\"" + ea.AgainstFileLocation.String() + "\""
}

// ExpectedFileLocation contains the values expected from a Location.
type ExpectedFileLocation struct {
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

// String implements fmt.Stringer.
func (el *ExpectedFileLocation) String() string {
	if el == nil {
		return "nil"
	}
	return el.FileName +
		" startLine=" + strconv.Itoa(el.StartLine) +
		" startColumn=" + strconv.Itoa(el.StartColumn) +
		" endLine=" + strconv.Itoa(el.EndLine) +
		" endColumn=" + strconv.Itoa(el.EndColumn)
}

// AssertAnnotationsEqual asserts that the Annotations equal the expected Annotations.
func AssertAnnotationsEqual(t *testing.T, expectedAnnotations []ExpectedAnnotation, actualAnnotations []check.Annotation) {
	if len(expectedAnnotations) == 0 {
		expectedAnnotations = nil
	}
	if len(actualAnnotations) == 0 {
		actualAnnotations = nil
	}
	actualExpectedAnnotations := expectedAnnotationsForAnnotations(actualAnnotations)
	msgAndArgs := []any{"expected:\n%v\nactual:\n%v", expectedAnnotations, actualExpectedAnnotations}
	require.Equal(t, len(expectedAnnotations), len(actualExpectedAnnotations), msgAndArgs...)
	for i, expectedAnnotation := range expectedAnnotations {
		if expectedAnnotation.Message == "" {
			actualExpectedAnnotations[i].Message = ""
		}
	}
	assert.Equal(t, expectedAnnotations, actualExpectedAnnotations, msgAndArgs...)
}

// RequireAnnotationsEqual requires that the Annotations equal the expected Annotations.
func RequireAnnotationsEqual(t *testing.T, expectedAnnotations []ExpectedAnnotation, actualAnnotations []check.Annotation) {
	if len(expectedAnnotations) == 0 {
		expectedAnnotations = nil
	}
	if len(actualAnnotations) == 0 {
		actualAnnotations = nil
	}
	actualExpectedAnnotations := expectedAnnotationsForAnnotations(actualAnnotations)
	msgAndArgs := []any{"expected:\n%v\nactual:\n%v", expectedAnnotations, actualExpectedAnnotations}
	require.Equal(t, len(expectedAnnotations), len(actualExpectedAnnotations), msgAndArgs...)
	for i, expectedAnnotation := range expectedAnnotations {
		if expectedAnnotation.Message == "" {
			actualExpectedAnnotations[i].Message = ""
		}
	}
	require.Equal(t, expectedAnnotations, actualExpectedAnnotations, msgAndArgs...)
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

// expectedAnnotationsForAnnotations returns ExpectedAnnotations for the given Annotations.
//
// Callers will need to filter out the Messages from the returned ExpectedAnnotations to conform
// to the ExpectedAnnotations that are being compared against. See the note on ExpectedAnnotation.Message.
func expectedAnnotationsForAnnotations(annotations []check.Annotation) []ExpectedAnnotation {
	return xslices.Map(annotations, expectedAnnotationForAnnotation)
}

// expectedAnnotationForAnnotation returns an ExpectedAnnotation for the given Annotation.
//
// Callers will need to filter out the Messages from the returned ExpectedAnnotations to conform
// to the ExpectedAnnotations that are being compared against. See the note on ExpectedAnnotation.Message.
func expectedAnnotationForAnnotation(annotation check.Annotation) ExpectedAnnotation {
	expectedAnnotation := ExpectedAnnotation{
		RuleID:  annotation.RuleID(),
		Message: annotation.Message(),
	}
	if fileLocation := annotation.FileLocation(); fileLocation != nil {
		expectedAnnotation.FileLocation = &ExpectedFileLocation{
			FileName:    fileLocation.FileDescriptor().ProtoreflectFileDescriptor().Path(),
			StartLine:   fileLocation.StartLine(),
			StartColumn: fileLocation.StartColumn(),
			EndLine:     fileLocation.EndLine(),
			EndColumn:   fileLocation.EndColumn(),
		}
	}
	if againstFileLocation := annotation.AgainstFileLocation(); againstFileLocation != nil {
		expectedAnnotation.AgainstFileLocation = &ExpectedFileLocation{
			FileName:    againstFileLocation.FileDescriptor().ProtoreflectFileDescriptor().Path(),
			StartLine:   againstFileLocation.StartLine(),
			StartColumn: againstFileLocation.StartColumn(),
			EndLine:     againstFileLocation.EndLine(),
			EndColumn:   againstFileLocation.EndColumn(),
		}
	}
	return expectedAnnotation
}

func compile(ctx context.Context, dirPaths []string, filePaths []string) ([]descriptor.FileDescriptor, error) {
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
