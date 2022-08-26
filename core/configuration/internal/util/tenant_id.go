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
	"crypto/md5"
	"strings"
	"unicode"
)

func CalculateTenantId(tenant string) int32 {
	sanitized := sanitizeTenantUuid(tenant)
	md5Hash := md5.Sum([]byte(sanitized))

	hash := int32(0)
	for i := 0; i < 16; i++ {
		shift := (3 - (i % 4)) * 8
		hash ^= int32(md5Hash[i]) << shift
	}
	return hash
}

func sanitizeTenantUuid(tenant string) string {
	// Replace non-ASCII characters with '?'
	return strings.Map(func(r rune) rune {
		if r > unicode.MaxASCII {
			return '?'
		}
		return r
	}, tenant)
}
