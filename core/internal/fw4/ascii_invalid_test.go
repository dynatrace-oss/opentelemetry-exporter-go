package fw4

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

type unparsableInputTestCase struct {
	tag    string
	errmsg string
}

func runInvalidParsingTests(t *testing.T, parser func(string) (Fw4Tag, error), testCases []unparsableInputTestCase) {
	for _, tc := range testCases {
		t.Run(tc.tag, func(t *testing.T) {
			parsed, err := parser(tc.tag)
			if err == nil {
				t.Fatalf("Unexpectedly succeeded parsing %q (result=%v)", tc.tag, parsed)
			} else if !strings.Contains(err.Error(), tc.errmsg) {
				t.Fatalf(" %q unparsable, but error %q doesn't contain %q", tc.tag, err, tc.errmsg)
			}
		})
	}
}

func TestParseInvalidXDynatrace(t *testing.T) {
	runInvalidParsingTests(t, ParseXDynatrace, []unparsableInputTestCase{
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
	runInvalidParsingTests(t, ParseTracestateEntryValue, []unparsableInputTestCase{
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

type inputMismatchingXDynatraceTestCase struct {
	tag       string
	tenantId  int32
	clusterId int32
	errmsg    string
}

func runMismatchingXDynatraceTests(t *testing.T, parser func(string, int32, int32) (Fw4Tag, error), testCases []inputMismatchingXDynatraceTestCase) {
	for _, tc := range testCases {
		t.Run(tc.tag, func(t *testing.T) {
			parsed, err := parser(tc.tag, tc.tenantId, tc.clusterId)
			if err == nil {
				t.Fatalf("Unexpectedly succeeded parsing %q tag with expected tenantId: %d, clusterId: %d, (result=%v)", tc.tag,
					tc.tenantId, tc.clusterId, parsed)
			} else if !strings.Contains(err.Error(), tc.errmsg) {
				t.Fatalf(" %q tag with expected tenantId: %d, clusterId: %d not matching, but error %q doesn't contain %q", tc.tag,
					tc.tenantId, tc.clusterId, err, tc.errmsg)
			}
		})
	}
}

func TestGetMismatchingXDynatrace(t *testing.T) {
	runMismatchingXDynatraceTests(t, GetMatchingFw4FromXDynatrace, []inputMismatchingXDynatraceTestCase{
		{"FW4;129;1;526;0;0;17;12345;ce1f;2h01;6h11223344556677889900112233445566;7h663055fc5bca216f", 0, 129, "FW4 TenantId mismatch"},
		{"FW4;129;1;526;0;0;17;12345;ce1f;2h01;6h11223344556677889900112233445566;7h663055fc5bca216f", 17, 0, "FW4 ClusterId mismatch"},
		{"FW4;129;1;526;0;0;17;12345;28aa;2h01;7h663055fc5bca216f", 17, 129, "FW4 tag does not contain TraceId"},
		{"FW4;129;1;526;0;0;17;12345;d126;2h01;6h11223344556677889900112233445566", 17, 129, "FW4 tag does not contain SpanId"},
		{"FW4;129;1;526;0;0;17;12345;0f61;2h01;6h00000000000000000000000000000000;7h663055fc5bca216f", 17, 0, "unparsable extension value for ID 6: trace-id can't be all zero"},
		{"FW4;129;1;526;0;0;17;12345;47a6;2h01;6h11223344556677889900112233445566;7h0000000000000000", 17, 0, "unparsable extension value for ID 7: span-id can't be all zero"},
	})
}

func TestGetFw4TagFromTracestateWithMismatchingTenantId(t *testing.T) {
	ts := trace.TraceState{}
	ts, err := ts.Insert("11-81@dt", "fw4;1;ec354cd6;1;2;0;0;3039;4589;1h48656c6c6f2c20576f726c6421;2h0123")
	require.NoError(t, err)

	tag, err := GetMatchingFw4FromTracestate(ts, 0, 129)
	require.EqualError(t, err, "can not find @dt entry in given tracestate with key 0-81@dt")
	require.Equal(t, tag, EmptyTag())
}

func TestGetFw4TagFromTracestateWithMismatchingClustertId(t *testing.T) {
	ts := trace.TraceState{}
	ts, err := ts.Insert("11-81@dt", "fw4;1;ec354cd6;1;2;0;0;3039;4589;1h48656c6c6f2c20576f726c6421;2h0123")
	require.NoError(t, err)

	tag, err := GetMatchingFw4FromTracestate(ts, 17, 0)
	require.EqualError(t, err, "can not find @dt entry in given tracestate with key 11-0@dt")
	require.Equal(t, tag, EmptyTag())
}
