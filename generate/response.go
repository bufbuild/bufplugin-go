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
	"bytes"
	"io"
	"io/fs"
	"slices"

	filev1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/file/v1"
	generatev1 "buf.build/gen/go/bufbuild/bufplugin/protocolbuffers/go/buf/plugin/generate/v1"
	"buf.build/go/bufplugin/internal/pkg/xslices"
)

// Response is a response from a plugin for a generate call.
type Response interface {
	// Paths gets all file paths on the response.
	//
	// The paths will be relative, use "/" as the path separator, and not contain
	// any "." or ".." components.
	//
	// The paths will be sorted and unique.
	Paths() []string
	// Get gets the Reader for the file of the given path.
	//
	// The paths must be relative, use "/" as the path separator, and not contain
	// any "." or ".." components.
	//
	// If the path does not exist, an error satisfying fs.ErrNotExist is returned.
	Get(path string) (io.Reader, error)

	toProto() *generatev1.GenerateResponse

	isResponse()
}

// *** PRIVATE ***

type response struct {
	pathToBuffer map[string]*bytes.Buffer
	sortedPaths  []string
}

func newResponse(pathToBuffer map[string]*bytes.Buffer) (*response, error) {
	return &response{
		pathToBuffer: pathToBuffer,
		sortedPaths:  xslices.MapKeysToSortedSlice(pathToBuffer),
	}, nil
}

func (r *response) Paths() []string {
	return slices.Clone(r.sortedPaths)
}

func (r *response) Get(path string) (io.Reader, error) {
	path, err := validateAndNormalizePath(path)
	if err != nil {
		return nil, err
	}
	buffer, ok := r.pathToBuffer[path]
	if !ok {
		return nil, &fs.PathError{Op: "read", Path: path, Err: fs.ErrNotExist}
	}
	return buffer, nil
}

func (r *response) toProto() *generatev1.GenerateResponse {
	protoFiles := make([]*filev1.File, len(r.sortedPaths))
	for i, path := range r.sortedPaths {
		protoFiles[i] = &filev1.File{
			Path: path,
			// We know the key exists because of how we created the response.
			Content: r.pathToBuffer[path].Bytes(),
		}
	}
	return &generatev1.GenerateResponse{
		Files: protoFiles,
	}
}

func (*response) isResponse() {}
