// Copyright 2023 Dynatrace LLC
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

package trace

func boolSliceEquals(first []bool, second []bool) bool {
	if len(first) != len(second) {
		return false
	}

	for idx, value := range first {
		if value != second[idx] {
			return false
		}
	}

	return true
}

func int64SliceEquals(first []int64, second []int64) bool {
	if len(first) != len(second) {
		return false
	}

	for idx, value := range first {
		if value != second[idx] {
			return false
		}
	}

	return true
}

func float64SliceEquals(first []float64, second []float64) bool {
	if len(first) != len(second) {
		return false
	}

	for idx, value := range first {
		if value != second[idx] {
			return false
		}
	}

	return true
}

func stringSliceEquals(first []string, second []string) bool {
	if len(first) != len(second) {
		return false
	}

	for idx, value := range first {
		if value != second[idx] {
			return false
		}
	}

	return true
}
