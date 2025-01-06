// Copyright 2024-2025 Buf Technologies, Inc.
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

package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var count int
	singleton := NewSingleton(
		func(context.Context) (int, error) {
			count++
			return count, nil
		},
	)
	value, err := singleton.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, value)
	value, err = singleton.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, value)

	count = 0
	singleton = NewSingleton(
		func(context.Context) (int, error) {
			count++
			return 0, fmt.Errorf("%d", count)
		},
	)
	_, err = singleton.Get(ctx)
	require.Error(t, err)
	require.Equal(t, "1", err.Error())
	_, err = singleton.Get(ctx)
	require.Error(t, err)
	require.Equal(t, "1", err.Error())
}
