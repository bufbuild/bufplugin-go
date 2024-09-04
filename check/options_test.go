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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionsRoundTrip(t *testing.T) {
	t.Parallel()

	testOptionsRoundTrip(t, true)
	testOptionsRoundTrip(t, int64(1))
	testOptionsRoundTrip(t, float64(1.0))
	testOptionsRoundTrip(t, "foo")
	testOptionsRoundTrip(t, []byte("foo"))
	testOptionsRoundTrip(t, []bool{true, true})
	testOptionsRoundTrip(t, []int64{1, 2})
	testOptionsRoundTrip(t, []float64{1.0, 2.0})
	testOptionsRoundTrip(t, []string{"foo", "bar"})
	testOptionsRoundTrip(t, [][]string{{"foo", "bar"}, {"baz, bat"}})
	testOptionsRoundTripDifferentInputOutput(
		t,
		[]any{"foo", "bar"},
		[]string{"foo", "bar"},
	)
	testOptionsRoundTripDifferentInputOutput(
		t,
		[]any{[]string{"foo"}, []string{"bar"}},
		[][]string{{"foo"}, {"bar"}},
	)
}

func TestOptionsValidateValueError(t *testing.T) {
	t.Parallel()

	err := validateValue(false)
	assert.Error(t, err)
	err = validateValue(0)
	assert.Error(t, err)
	err = validateValue([]any{1, "foo"})
	assert.Error(t, err)
	err = validateValue([]any{[]string{"foo"}, "foo"})
	assert.Error(t, err)
}

func testOptionsRoundTrip(t *testing.T, value any) {
	protoValue, err := valueToProtoValue(value)
	require.NoError(t, err)
	actualValue, err := protoValueToValue(protoValue)
	require.NoError(t, err)
	assert.Equal(t, value, actualValue)
}

func testOptionsRoundTripDifferentInputOutput(t *testing.T, input any, expectedOutput any) {
	protoValue, err := valueToProtoValue(input)
	require.NoError(t, err)
	actualValue, err := protoValueToValue(protoValue)
	require.NoError(t, err)
	assert.Equal(t, expectedOutput, actualValue)
}
