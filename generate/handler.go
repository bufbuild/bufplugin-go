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
	"context"
)

var nopHandler = HandlerFunc(func(context.Context, ResponseWriter, Request) error { return nil })

// Handler implements the generate logic.
type Handler interface {
	Handle(ctx context.Context, responseWriter ResponseWriter, request Request) error
}

// HandlerFunc is a function that implements Handler.
type HandlerFunc func(context.Context, ResponseWriter, Request) error

// Handle implements Handler.
func (h HandlerFunc) Handle(ctx context.Context, responseWriter ResponseWriter, request Request) error {
	return h(ctx, responseWriter, request)
}
