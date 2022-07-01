package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const ENV_KEY = "TEST_KEY"

func TestGetStringFromEnvWithDefault_UseEnvIfSet(t *testing.T) {
	t.Setenv(ENV_KEY, "foo")
	assert.Equal(t, GetStringFromEnvWithDefault(ENV_KEY, "bar"), "foo")

	t.Setenv(ENV_KEY, "")
	assert.Equal(t, GetStringFromEnvWithDefault(ENV_KEY, "bar"), "")
}

func TestGetStringFromEnvWithDefault_UseDefault(t *testing.T) {
	assert.Equal(t, GetStringFromEnvWithDefault(ENV_KEY, "bar"), "bar")
	assert.Equal(t, GetStringFromEnvWithDefault(ENV_KEY, ""), "")
}

func TestGetIntFromEnvWithDefault_UseEnvIfSet(t *testing.T) {
	t.Setenv(ENV_KEY, "123")
	assert.Equal(t, GetIntFromEnvWithDefault(ENV_KEY, 456), 123)

	t.Setenv(ENV_KEY, "0")
	assert.Equal(t, GetIntFromEnvWithDefault(ENV_KEY, 456), 0)
}

func TestGetIntFromEnvWithDefault_PanicForInvalidValue(t *testing.T) {
	t.Setenv(ENV_KEY, "foo")
	assert.Panics(t, func() { GetIntFromEnvWithDefault(ENV_KEY, 456) })

	t.Setenv(ENV_KEY, "")
	assert.Panics(t, func() { GetIntFromEnvWithDefault(ENV_KEY, 456) })
}

func TestGetIntFromEnvWithDefault_UseDefault(t *testing.T) {
	assert.Equal(t, GetIntFromEnvWithDefault(ENV_KEY, 456), 456)
	assert.Equal(t, GetIntFromEnvWithDefault(ENV_KEY, 0), 0)
	assert.Equal(t, GetIntFromEnvWithDefault(ENV_KEY, -1), -1)
}

func TestGetBoolFromEnvWithDefault_UseEnvIfSet(t *testing.T) {
	t.Setenv(ENV_KEY, "false")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, true), false)

	t.Setenv(ENV_KEY, "0")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, true), false)

	t.Setenv(ENV_KEY, "FALSE")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, true), false)

	t.Setenv(ENV_KEY, "fAlSe")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, true), false)

	t.Setenv(ENV_KEY, "true")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, false), true)

	t.Setenv(ENV_KEY, "1")
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, false), true)
}

func TestGetBoolFromEnvWithDefault_UseDefault(t *testing.T) {
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, true), true)
	assert.Equal(t, GetBoolFromEnvWithDefault(ENV_KEY, false), false)
}

func TestGetStringSliceFromEnvWithDefault_UseEnvIfSet(t *testing.T) {
	t.Setenv(ENV_KEY, "foo:bar:baz")
	assert.Equal(t, GetStringSliceFromEnvWithDefault(ENV_KEY, []string{"a", "b", "c"}), []string{"foo", "bar", "baz"})

	t.Setenv(ENV_KEY, "")
	assert.Equal(t, GetStringSliceFromEnvWithDefault(ENV_KEY, []string{"a", "b", "c"}), []string{""})

	t.Setenv(ENV_KEY, "foo,bar,baz")
	assert.Equal(t, GetStringSliceFromEnvWithDefault(ENV_KEY, []string{"a", "b", "c"}), []string{"foo,bar,baz"})

	t.Setenv(ENV_KEY, "foo;bar;baz")
	assert.Equal(t, GetStringSliceFromEnvWithDefault(ENV_KEY, []string{"a", "b", "c"}), []string{"foo;bar;baz"})
}

func TestGetStringSliceFromEnvWithDefault_UseDefault(t *testing.T) {
	assert.Equal(t, GetStringSliceFromEnvWithDefault(ENV_KEY, []string{"a", "b", "c"}), []string{"a", "b", "c"})
	assert.Equal(t, GetStringSliceFromEnvWithDefault(ENV_KEY, []string{}), []string{})
}
