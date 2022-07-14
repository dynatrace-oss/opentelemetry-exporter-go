package fw4

import (
	"fmt"
	"testing"
)

func TestLatin1MurmurHash2_64A(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{";1h02230045880011;2h01;3h020e;4h00;5h01", 0x4d2a12b9bd7669b9},
		{"", -0x64051f5b19ec03c4},
		{"1", -0x5d87b4edc81d812d},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%q", tc.input), func(t *testing.T) {
			h := latin1MurmurHash2_64A(tc.input)
			if h != tc.expected {
				t.Fatalf("Unexpected hash: %x for %q", h, tc.input)
			}
		})
	}
}
