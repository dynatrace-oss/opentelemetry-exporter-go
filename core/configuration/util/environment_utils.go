package util

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetStringFromEnvWithDefault(key, defaultValue string) string {
	str, found := os.LookupEnv(key)
	if found {
		return str
	}
	return defaultValue
}

func GetIntFromEnvWithDefault(key string, defaultValue int) int {
	str, found := os.LookupEnv(key)
	if found {
		intVal, err := strconv.Atoi(str)
		if err != nil {
			panic(fmt.Sprintf("Could not parse integer value from environment variable %s", key))
		}
		return intVal
	}
	return defaultValue
}

func GetBoolFromEnvWithDefault(key string, defaultValue bool) bool {
	str, found := os.LookupEnv(key)
	if found {
		return toBool(str)
	}
	return defaultValue
}

func toBool(value string) bool {
	lowercaseValue := strings.ToLower(value)
	return lowercaseValue != "false" && lowercaseValue != "0"
}

func GetStringSliceFromEnvWithDefault(key string, defaultValue []string) []string {
	str, found := os.LookupEnv(key)
	if found {
		return strings.Split(str, ":")
	}
	return defaultValue
}
