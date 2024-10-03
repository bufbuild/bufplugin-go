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
	"errors"
)

var errCannotReuseResponseWriter = errors.New("cannot reuse ResponseWriter")

// ResponseWriter is used by plugin implmentations to add Files to responses.
type ResponseWriter interface {
	//NewFileWriter(name string) (FileWriter, error)

	isResponseWriter()
}

// *** PRIVATE ***

type responseWriter struct {
}

func newResponseWriter() *responseWriter {
	return &responseWriter{}
}

func (*responseWriter) isResponseWriter() {}
