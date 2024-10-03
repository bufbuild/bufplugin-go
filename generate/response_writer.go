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
	"errors"
	"fmt"
	"io"
	"sync"
)

var errCannotReuseResponseWriter = errors.New("cannot reuse ResponseWriter")

// ResponseWriter is used by plugin implmentations to add Files to responses.
type ResponseWriter interface {
	// Put opens a new file for Writing.
	//
	// The path must be relative, use "/" as the path separator, and not contain
	// any "." or ".." components.
	//
	// Returns error if the path is not valid or has already been written.
	Put(path string) (io.Writer, error)

	isResponseWriter()
}

// *** PRIVATE ***

type responseWriter struct {
	pathToBuffer map[string]*bytes.Buffer

	written bool
	lock    sync.RWMutex
}

func newResponseWriter() *responseWriter {
	return &responseWriter{}
}

func (r *responseWriter) Put(path string) (io.Writer, error) {
	path, err := validateAndNormalizePath(path)
	if err != nil {
		return nil, err
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	if r.written {
		return nil, errCannotReuseResponseWriter
	}
	if _, ok := r.pathToBuffer[path]; ok {
		return nil, fmt.Errorf("duplicate path: %q", path)
	}
	buffer := bytes.NewBuffer(nil)
	r.pathToBuffer[path] = buffer
	return buffer, nil
}

func (r *responseWriter) toResponse() (Response, error) {
	if r.written {
		return nil, errCannotReuseResponseWriter
	}
	r.written = true

	return newResponse(r.pathToBuffer)
}

func (*responseWriter) isResponseWriter() {}
