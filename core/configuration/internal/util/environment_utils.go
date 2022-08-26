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
			fmt.Printf("Could not parse integer value from environment variable %s\n", key)
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
			fmt.Printf("Could not parse boolean value from environment variable %s\n", key)
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
