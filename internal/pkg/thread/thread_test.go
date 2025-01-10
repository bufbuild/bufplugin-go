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

package thread

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The bulk of the code relies on subtle timing that's difficult to
// reproduce, but we can test the most basic use cases.

func TestParallelizeSimple(t *testing.T) {
	t.Parallel()

	numJobs := 10
	var executed atomic.Int64
	jobs := make([]func(context.Context) error, 0, numJobs)
	for range numJobs {
		jobs = append(
			jobs,
			func(context.Context) error {
				executed.Add(1)
				return nil
			},
		)
	}
	ctx := context.Background()
	assert.NoError(t, Parallelize(ctx, jobs))
	assert.Equal(t, int64(numJobs), executed.Load())
}

func TestParallelizeImmediateCancellation(t *testing.T) {
	t.Parallel()

	numJobs := 10
	var executed atomic.Int64
	jobs := make([]func(context.Context) error, 0, numJobs)
	for range numJobs {
		jobs = append(
			jobs,
			func(context.Context) error {
				executed.Add(1)
				return nil
			},
		)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.Error(t, Parallelize(ctx, jobs))
	assert.Equal(t, int64(0), executed.Load())
}
