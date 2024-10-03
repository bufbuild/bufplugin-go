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
	"fmt"
	"path/filepath"
	"strings"
)

func validateAndNormalizePath(path string) (string, error) {
	if path == "" {
		return "", errors.New("path was empty")
	}
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("path must be relative: %q", path)
	}
	path = filepath.Clean(path)
	if slashPath := filepath.ToSlash(path); path != slashPath {
		return "", fmt.Errorf("path must use / as the separator: %q", path)
	}
	if strings.HasPrefix(path, "../") {
		return "", fmt.Errorf("path cannot contain \"..\": %q", path)
	}
	return path, nil

}
