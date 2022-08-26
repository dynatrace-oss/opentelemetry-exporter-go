// Copyright 2022 Dynatrace LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const ENV_KEY = "TEST_KEY"

func TestGetStringFromEnvWithDefault_UseEnvIfSet(t *testing.T) {
	defer os.Unsetenv(ENV_KEY)
	os.Setenv(ENV_KEY, "foo")
	assert.Equal(t, GetStringFromEnvWithDefault(ENV_KEY, "bar"), "foo")

	os.Setenv(ENV_KEY, "")
	assert.Equal(t, GetStringFromEnvWithDefault(ENV_KEY, "bar"), "")
}

func TestGetStringFromEnvWithDefault_UseDefault(t *testing.T) {
	assert.Equal(t, GetStringFromEnvWithDefault(ENV_KEY, "bar"), "bar")
	assert.Equal(t, GetStringFromEnvWithDefault(ENV_KEY, ""), "")
}

func TestGetIntFromEnvWithDefault_UseEnvIfSet(t *testing.T) {
	defer os.Unsetenv(ENV_KEY)
	os.Setenv(ENV_KEY, "123")
	assert.Equal(t, GetIntFromEnvWithDefault(ENV_KEY, 456), 123)

	os.Setenv(ENV_KEY, "0")
	assert.Equal(t, GetIntFromEnvWithDefault(ENV_KEY, 456), 0)
}

func TestGetIntFromEnvWithDefault_UseDefaultForInvalidValue(t *testing.T) {
	defer os.Unsetenv(ENV_KEY)
	os.Setenv(ENV_KEY, "foo")
	assert.Equal(t, GetIntFromEnvWithDefault(ENV_KEY, 456), 456)

	os.Setenv(ENV_KEY, "")
	assert.Equal(t, GetIntFromEnvWithDefault(ENV_KEY, 456), 456)
}

func TestGetIntFromEnvWithDefault_UseDefault(t *testing.T) {
	assert.Equal(t, GetIntFromEnvWithDefault(ENV_KEY, 456), 456)
	assert.Equal(t, GetIntFromEnvWithDefault(ENV_KEY, 0), 0)
	assert.Equal(t, GetIntFromEnvWithDefault(ENV_KEY, -1), -1)
}

func TestGetBoolFromEnvWithDefault_UseEnvIfSet(t *testing.T) {
	defer os.Unsetenv(ENV_KEY)
	os.Setenv(ENV_KEY, "false")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, true), false)

	os.Setenv(ENV_KEY, "0")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, true), false)

	os.Setenv(ENV_KEY, "FALSE")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, true), false)

	os.Setenv(ENV_KEY, "False")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, true), false)

	os.Setenv(ENV_KEY, "true")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, false), true)

	os.Setenv(ENV_KEY, "1")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, false), true)
}

func TestGetBoolFromEnvWithDefault_UseDefaultIfInvalidValue(t *testing.T) {
	defer os.Unsetenv(ENV_KEY)
	os.Setenv(ENV_KEY, "foo")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, true), true)
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, false), false)

	os.Setenv(ENV_KEY, "fAlSe")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, true), true)
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, false), false)

	os.Setenv(ENV_KEY, "tRuE")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, false), false)
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, true), true)
}

func TestGetBoolFromEnvWithDefault_UseDefault(t *testing.T) {
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, true), true)
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, false), false)
}

func TestGetStringSliceFromEnvWithDefault_UseEnvIfSet(t *testing.T) {
	defer os.Unsetenv(ENV_KEY)
	os.Setenv(ENV_KEY, "foo:bar:baz")
	assert.Equal(t, GetStringSliceFromEnvWithDefault(ENV_KEY, []string{"a", "b", "c"}), []string{"foo", "bar", "baz"})

	os.Setenv(ENV_KEY, "")
	assert.Equal(t, GetStringSliceFromEnvWithDefault(ENV_KEY, []string{"a", "b", "c"}), []string{""})

	os.Setenv(ENV_KEY, "foo,bar,baz")
	assert.Equal(t, GetStringSliceFromEnvWithDefault(ENV_KEY, []string{"a", "b", "c"}), []string{"foo,bar,baz"})

	os.Setenv(ENV_KEY, "foo;bar;baz")
	assert.Equal(t, GetStringSliceFromEnvWithDefault(ENV_KEY, []string{"a", "b", "c"}), []string{"foo;bar;baz"})
}

func TestGetStringSliceFromEnvWithDefault_UseDefault(t *testing.T) {
	assert.Equal(t, GetStringSliceFromEnvWithDefault(ENV_KEY, []string{"a", "b", "c"}), []string{"a", "b", "c"})
	assert.Equal(t, GetStringSliceFromEnvWithDefault(ENV_KEY, []string{}), []string{})
}
