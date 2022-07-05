package util

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func GetStringFromEnvWithDefault(key, defaultValue string) string {
	if str, found := os.LookupEnv(key); found {
		return str
	}
	return defaultValue
}

func GetIntFromEnvWithDefault(key string, defaultValue int) int {
	if str, found := os.LookupEnv(key); found {
		intVal, err := strconv.Atoi(str)
		if err != nil {
			fmt.Printf("Could not parse integer value from environment variable %s\n", key);
			return defaultValue
		}
		return intVal
	}
	return defaultValue
}

func GetBoolFromEnvWithDefault(key string, defaultValue bool) bool {
	if str, found := os.LookupEnv(key); found {
		b, err := strconv.ParseBool(str)
		if err != nil {
			fmt.Printf("Could not parse boolean value from environment variable %s\n", key);
			return defaultValue
		}
		return b
	}
	return defaultValue
}

func GetStringSliceFromEnvWithDefault(key string, defaultValue []string) []string {
	if str, found := os.LookupEnv(key); found {
		return strings.Split(str, ":")
	}
	return defaultValue
}
