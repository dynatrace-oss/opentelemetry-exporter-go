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
