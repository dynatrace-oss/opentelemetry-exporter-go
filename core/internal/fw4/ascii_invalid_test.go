package fw4

import (
	"strings"
	"testing"
)

type invalidInputTestCase struct {
	input  string
	errmsg string
}

func runInvalidInputTests(t *testing.T, parser func(string) (Fw4Tag, error), testCases []invalidInputTestCase) {
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			parsed, err := parser(tc.input)
			if err == nil {
				t.Fatalf("Unexpectedly succeeded parsing %q (result=%v)", tc.input, parsed)
			} else if !strings.Contains(err.Error(), tc.errmsg) {
				t.Fatalf(" %q unparsable, but error %q doesn't contain %q", tc.input, err, tc.errmsg)
			}
		})
	}
}

func TestParseInvalidXDynatrace(t *testing.T) {
	runInvalidInputTests(t, ParseXDynatrace, []invalidInputTestCase{
		{"", "input does not start with"},
		{"FW4;666;1;-332051242;1;2;113948091;12345;c72d;1h48656c6c6f2c20576f726c642", "odd length hex"},
		{"FW4;666;1;343434343434332051242;1;2;113948091;12345", "invalid value for agentId"},
		{"FW4;666;1;-332051242;1;2;113948091;12345;", "checksum must be exactly 4 ASCII bytes"},
		{"FW4;", "does not contain enough semicolon-separated components"},
		{"FW3;666;1;-332051242;1;2;113948091;12345", "input does not start with \"FW4;\""},
		{"FW4;2;X;4;5;6;7;8", "invalid value for ServerId"},
		{"FW4;2;3;-1;6;7", "does not contain enough semicolon-separated components"},
		{"FW4;2;3;4;6;6", "does not contain enough semicolon-separated components"},
		{"FW4;22222;2;3;-1;5;11111;1h0000ab0000", "invalid value for PathInfo"},
		{"FW4;22222;2;3;-1;5;11111;1;1234;991h000ab0000", "checksum mismatch"},
		{"FW4;22222;2;3;-1;5;11111;1;1234", "checksum mismatch"},
	})
}

func TestParseInvalidTracestateEntryValue(t *testing.T) {
	runInvalidInputTests(t, ParseTracestateEntryValue, []invalidInputTestCase{
		{"", "input does not start with"},
		{"fw4;1;2;3;4;0;0;3039;c72d;1h48656c6c6f2c20576f726c642", "odd length hex"},
		{"fw3;1;2;3;4;0;0;3039", "input does not start with \"fw4;\""},
		{"fw4;1;g;3;4;0;0;3039", "invalid value for agentId"},
		{"fw4;-1;2;3;4;0;0;3039", "invalid value for ServerId"},
		{"fw4;1;2;3;4;0;0", "does not contain enough semicolon-separated components"},
		{"fw4;1;2;3;4;0;0;", "invalid value for PathInfo"},
		{"fw4;1;2;3;4;0;0;;", "invalid value for PathInfo"},
		{"fw4;1;ffffffffffffffffffff;3;4;0;0;3039", "invalid value for agentId"},
		{"fw4;1;2;3;4;0;0;3039;11111;1h0000ab0000", "checksum must be exactly 4 ASCII bytes"},
		{"fw4;1;2;3;4;0;0;3039;1111;1h0000ab0000", "checksum mismatch"},
	})
}
