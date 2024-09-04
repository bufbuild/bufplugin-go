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
	"slices"

	checkv1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/check/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Location is a reference to a File or to a location within a File.
//
// A Location always has a file name.
type Location interface {
	// File is the File associated with the Location.
	//
	// Always present.
	File() File
	// SourcePath returns the path within the FileDescriptorProto of the Location.
	SourcePath() protoreflect.SourcePath

	// StartLine returns the zero-indexed start line, if known.
	StartLine() int
	// StartColumn returns the zero-indexed start column, if known.
	StartColumn() int
	// EndLine returns the zero-indexed end line, if known.
	EndLine() int
	// EndColumn returns the zero-indexed end column, if known.
	EndColumn() int
	// LeadingComments returns any leading comments, if known.
	LeadingComments() string
	// TrailingComments returns any trailing comments, if known.
	TrailingComments() string
	// LeadingDetachedComments returns any leading detached comments, if known.
	LeadingDetachedComments() []string

	unclonedSourcePath() protoreflect.SourcePath
	unclonedLeadingDetachedComments() []string
	toProto() *checkv1.Location

	isLocation()
}

// *** PRIVATE ***

type location struct {
	file           File
	sourceLocation protoreflect.SourceLocation
}

func newLocation(
	file File,
	sourceLocation protoreflect.SourceLocation,
) *location {
	return &location{
		file:           file,
		sourceLocation: sourceLocation,
	}
}

func (l *location) File() File {
	return l.file
}

func (l *location) SourcePath() protoreflect.SourcePath {
	return slices.Clone(l.sourceLocation.Path)
}

func (l *location) StartLine() int {
	return l.sourceLocation.StartLine
}

func (l *location) StartColumn() int {
	return l.sourceLocation.StartColumn
}

func (l *location) EndLine() int {
	return l.sourceLocation.EndLine
}

func (l *location) EndColumn() int {
	return l.sourceLocation.EndColumn
}

func (l *location) LeadingComments() string {
	return l.sourceLocation.LeadingComments
}

func (l *location) TrailingComments() string {
	return l.sourceLocation.TrailingComments
}

func (l *location) LeadingDetachedComments() []string {
	return slices.Clone(l.sourceLocation.LeadingDetachedComments)
}

func (l *location) unclonedSourcePath() protoreflect.SourcePath {
	return l.sourceLocation.Path
}

func (l *location) unclonedLeadingDetachedComments() []string {
	return l.sourceLocation.LeadingDetachedComments
}

func (l *location) toProto() *checkv1.Location {
	if l == nil {
		return nil
	}
	return &checkv1.Location{
		FileName:   l.file.FileDescriptor().Path(),
		SourcePath: l.sourceLocation.Path,
	}
}

func (*location) isLocation() {}
