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
	"pluginrpc.com/pluginrpc"
)

// Main is the main entrypoint for a plugin that implements the given Spec.
//
// A plugin just needs to provide a Spec, and then call this function within main.
//
//		func main() {
//			generate.Main(
//				&generate.Spec {
//					Handler: generate.HandlerFunc(
//	               func(ctx context.Context, responseWriter generate.ResponseWriter, request generate.Request) error {
//	                 return errors.New("implement your generator here")
//	               },
//	             ),
//				},
//			)
//		}
func Main(spec *Spec, _ ...MainOption) {
	pluginrpc.Main(
		func() (pluginrpc.Server, error) {
			return NewServer(spec)
		},
	)
}

// MainOption is an option for Main.
type MainOption func(*mainOptions)

// *** PRIVATE ***

type mainOptions struct{}
