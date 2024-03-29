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

package fw4

import (
	"fmt"
	"testing"

	"github.com/dynatrace-oss/opentelemetry-exporter-go/core/configuration"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func traceIdOrPanic(hex string) trace.TraceID {
	result, err := trace.TraceIDFromHex(hex)
	if err != nil {
		panic(err)
	}
	return result
}

func spanIdOrPanic(hex string) trace.SpanID {
	result, err := trace.SpanIDFromHex(hex)
	if err != nil {
		panic(err)
	}
	return result
}

func setDummyTagValues(fw4 Fw4Tag) Fw4Tag {
	fw4.ClusterID = 666
	fw4.ServerID = 1
	fw4.AgentID = -332051242
	fw4.TagID = 1
	fw4.encodedLinkID = 2
	fw4.TenantID = 113948091
	fw4.PathInfo = 12345
	return fw4
}

const dummyBlob = "Hello, World!"

func TestParseFormatFw4(t *testing.T) {
	testCases := []struct {
		comment            string
		tracestate         string
		xdt                string
		parsed             Fw4Tag
		expectedTracestate string
		expectedXdt        string
	}{
		{
			"custom blob and some other extension fields",
			// aka RuxitV4LinkExtensionsTestTestCustomTagTraceContext
			"11-81@dt=fw4;1;20e;0;0;0;0;4d;69b9;1h02230045880011;2h01;3h020e;4h00;5h01",
			"FW4;129;1;526;0;0;17;77;69b9;1h02230045880011;2h01;3h020e;4h00;5h01",
			Fw4Tag{
				ClusterID:     129,
				ServerID:      1,
				AgentID:       0x20e,
				TagID:         0,
				encodedLinkID: 0,
				TenantID:      17,
				PathInfo:      77,
				CustomBlob:    "\x02\x23\x00\x45\x88\x00\x11",
				tagDepth:      1,
				entryAgentID:  0x20e,
				entryTagID:    0,
				payloadBitset: 1,
			},
			"",
			"",
		},
		{
			"no extension fields", // aka RuxitV4LinkTestRuxitV4LinkCreation
			"11-81@dt=fw4;1;2;3;4;0;0;3039",
			"FW4;129;1;2;3;4;17;12345",
			Fw4Tag{
				ClusterID:     129,
				ServerID:      1,
				AgentID:       2,
				TagID:         3,
				encodedLinkID: 4,
				TenantID:      17,
				PathInfo:      12345,
				entryTagID:    emptyEntryTagId,
				entryAgentID:  emptyEntryAgentId,
			},
			"",
			"",
		},
		{
			"negative agent ID no extensions", // aka RuxitV4LinkTestFromStringMatchesToString
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039",
			"FW4;666;1;-332051242;1;2;113948091;12345",
			setDummyTagValues(EmptyTag()),
			"",
			"",
		},
		{
			"negative agent ID with custom blob", // aka RuxitV4LinkTestFromStringWithCustomBlob
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;727c;1h48656c6c6f2c20576f726c6421",
			"FW4;666;1;-332051242;1;2;113948091;12345;727c;1h48656c6c6f2c20576f726c6421",
			setDummyTagValues(Fw4Tag{
				CustomBlob:   dummyBlob,
				entryAgentID: emptyEntryAgentId,
				entryTagID:   emptyEntryTagId,
			}),
			"",
			"",
		},
		{
			"negative agent ID with only unuspported extension ID",
			// aka RuxitV4LinkTestFromStringWithUnsupportedTagIdUsed
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;e9ae;999h48656c6c6f2c20576f726c6421",
			"FW4;666;1;-332051242;1;2;113948091;12345;e9ae;999h48656c6c6f2c20576f726c6421",
			setDummyTagValues(EmptyTag()),
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039",
			"FW4;666;1;-332051242;1;2;113948091;12345",
		},

		{
			"negative agent ID with unuspported extension ID and custom blob",
			// aka RuxitV4LinktestFromStringWithCustomBlobAndUnsupportedtagId
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;f21a;999h4865;1h48656c6c6f2c20576f726c6421",
			"FW4;666;1;-332051242;1;2;113948091;12345;f21a;999h4865;1h48656c6c6f2c20576f726c6421",
			setDummyTagValues(Fw4Tag{
				CustomBlob:   dummyBlob,
				entryAgentID: emptyEntryAgentId,
				entryTagID:   emptyEntryTagId,
			}),
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;727c;1h48656c6c6f2c20576f726c6421",
			"FW4;666;1;-332051242;1;2;113948091;12345;727c;1h48656c6c6f2c20576f726c6421",
		},
		{
			"negative agent ID with custom blob that includes NULs",
			// aka RuxitV4LinktestFromStringWithCustomBlobZeros
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;e663;1h0000ab0000",
			"FW4;666;1;-332051242;1;2;113948091;12345;e663;1h0000ab0000",
			setDummyTagValues(Fw4Tag{
				CustomBlob:   "\x00\x00\xab\x00\x00",
				entryAgentID: emptyEntryAgentId,
				entryTagID:   emptyEntryTagId,
			}),
			"",
			"",
		},
		{
			"extreme non-extension values",
			// aka RuxitV4LinktestFromStringMatchesToStringExtremeNumbers
			"6cab5bb-29a@dt=fw4;7fffffff;80000000;1;2;0;0;1e240",
			"FW4;666;2147483647;-2147483648;1;2;113948091;123456",
			Fw4Tag{
				ClusterID:     666,
				ServerID:      2147483647,
				AgentID:       -2147483648,
				TagID:         1,
				encodedLinkID: 2,
				TenantID:      113948091,
				PathInfo:      123456,
				entryAgentID:  emptyEntryAgentId,
				entryTagID:    emptyEntryTagId,
			},
			"",
			"",
		},
		{
			"custom blob and tag depth, non-canonical order",
			// aka RuxitV4LinktestFromStringAndBlobWithTagDepthAndCustomBlob
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;4095;2h0123;1h48656c6c6f2c20576f726c6421",
			"FW4;666;1;-332051242;1;2;113948091;12345;4095;2h0123;1h48656c6c6f2c20576f726c6421",
			setDummyTagValues(Fw4Tag{
				CustomBlob:   dummyBlob,
				tagDepth:     0x123,
				entryAgentID: emptyEntryAgentId,
				entryTagID:   emptyEntryTagId,
			}),
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;4589;1h48656c6c6f2c20576f726c6421;2h0123",
			"FW4;666;1;-332051242;1;2;113948091;12345;4589;1h48656c6c6f2c20576f726c6421;2h0123",
		},
		{
			"another unknown tag ID",
			// aka RuxitV4LinktestFromStringUnknownId57
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;5719;57x123;1h11223344",
			"FW4;666;1;-332051242;1;2;113948091;12345;5719;57x123;1h11223344",
			setDummyTagValues(Fw4Tag{
				CustomBlob:   "\x11\x22\x33\x44",
				entryAgentID: emptyEntryAgentId,
				entryTagID:   emptyEntryTagId,
			}),
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;f00d;1h11223344",
			"FW4;666;1;-332051242;1;2;113948091;12345;f00d;1h11223344",
		},
		{
			"unknown tag type 'g' before others",
			// aka RuxitV4LinktestFromStringUnknownTypeOnFirstPosition
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;7c1f;88g123;1h11223344",
			"FW4;666;1;-332051242;1;2;113948091;12345;7c1f;88g123;1h11223344",
			setDummyTagValues(Fw4Tag{
				CustomBlob:   "\x11\x22\x33\x44",
				entryAgentID: emptyEntryAgentId,
				entryTagID:   emptyEntryTagId,
			}),
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;f00d;1h11223344",
			"FW4;666;1;-332051242;1;2;113948091;12345;f00d;1h11223344",
		},
		{
			"unknown tag type 'g' after others",
			// aka RuxitV4LinktestFromStringUnknownTypeOnLastPosition
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;5600;1h11223344;23g123",
			"FW4;666;1;-332051242;1;2;113948091;12345;5600;1h11223344;23g123",
			setDummyTagValues(Fw4Tag{
				CustomBlob:   "\x11\x22\x33\x44",
				entryAgentID: emptyEntryAgentId,
				entryTagID:   emptyEntryTagId,
			}),
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;f00d;1h11223344",
			"FW4;666;1;-332051242;1;2;113948091;12345;f00d;1h11223344",
		},
		{
			"negative int value in extension",
			// aka TestNegativeExtensionInt
			"11-81@dt=fw4;1;fffffffe;3;4;0;0;3039;4061;3hfffffffe",
			"FW4;129;1;-2;3;4;17;12345;4061;3hfffffffe",
			Fw4Tag{
				ClusterID:     129,
				ServerID:      1,
				AgentID:       -2,
				TagID:         3,
				encodedLinkID: 4,
				TenantID:      17,
				PathInfo:      12345,
				entryAgentID:  -2,
				entryTagID:    emptyEntryTagId,
			},
			"",
			"",
		},
		{
			"with trace and span ID", // aka RuxitV4LinkTraceContextExtensionsrootPathChildPathsWithExtensions
			"11-81@dt=fw4;1;20e;0;0;0;0;457;e672;6hd728b0e6d2c9d2dbbfc086bdb6e0f6dd;7hb2c389f43fbb6576",
			"FW4;129;1;526;0;0;17;1111;e672;6hd728b0e6d2c9d2dbbfc086bdb6e0f6dd;7hb2c389f43fbb6576",
			Fw4Tag{
				ClusterID:     129,
				ServerID:      1,
				AgentID:       526,
				TagID:         0,
				encodedLinkID: 0,
				TenantID:      17,
				PathInfo:      1111,
				TraceID:       traceIdOrPanic("d728b0e6d2c9d2dbbfc086bdb6e0f6dd"),
				SpanID:        spanIdOrPanic("b2c389f43fbb6576"),
				entryAgentID:  emptyEntryAgentId,
				entryTagID:    emptyEntryTagId,
			},
			"",
			"",
		},
		{
			"just tag depth extension",
			// aka RuxitV4LinkExtensionstestFromString
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;b03b;2h0123",
			"FW4;666;1;-332051242;1;2;113948091;12345;b03b;2h0123",
			setDummyTagValues(Fw4Tag{
				tagDepth:     0x123,
				entryAgentID: emptyEntryAgentId,
				entryTagID:   emptyEntryTagId,
			}),
			"",
			"",
		},
		{
			"tag depth nearly overflows",
			// aka RuxitV4LinkExtensionstestFromStringNearlyOverflow
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;4253;2h7fffffff",
			"FW4;666;1;-332051242;1;2;113948091;12345;4253;2h7fffffff",
			setDummyTagValues(Fw4Tag{
				tagDepth:     0x7fffffff,
				entryAgentID: emptyEntryAgentId,
				entryTagID:   emptyEntryTagId,
			}),
			"",
			"",
		},
		{
			"test -1 in extension",
			// aka RuxitV4LinkExtensionstestFromStringOverflowAndZero1
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;e6ad;2hffffffff",
			"FW4;666;1;-332051242;1;2;113948091;12345;e6ad;2hffffffff",
			setDummyTagValues(Fw4Tag{
				tagDepth:     -1,
				entryAgentID: emptyEntryAgentId,
				entryTagID:   emptyEntryTagId,
			}),
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039",
			"FW4;666;1;-332051242;1;2;113948091;12345",
		},
		{
			"test max negative value in tag depth",
			// aka RuxitV4LinkExtensionstestFromStringOverflowAndZero2
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;5c62;2hf0000001",
			"FW4;666;1;-332051242;1;2;113948091;12345;5c62;2hf0000001",
			setDummyTagValues(Fw4Tag{
				tagDepth:     -0xfffffff,
				entryAgentID: emptyEntryAgentId,
				entryTagID:   emptyEntryTagId,
			}),
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039",
			"FW4;666;1;-332051242;1;2;113948091;12345",
		},
		{
			"test zero value for tag depth",
			// aka RuxitV4LinkExtensionstestFromStringOverflowAndZero3
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039;57ea;2h00",
			"FW4;666;1;-332051242;1;2;113948091;12345;57ea;2h00",
			setDummyTagValues(Fw4Tag{
				entryAgentID: emptyEntryAgentId,
				entryTagID:   emptyEntryTagId,
			}),
			"6cab5bb-29a@dt=fw4;1;ec354cd6;1;2;0;0;3039",
			"FW4;666;1;-332051242;1;2;113948091;12345",
		},
		{
			"minimal",
			// aka TestMinimal
			"11-81@dt=fw4;4;1;2;3;0;0;1",
			"FW4;129;4;1;2;3;17;1",
			Fw4Tag{
				ClusterID:     129,
				ServerID:      4,
				AgentID:       1,
				TagID:         2,
				encodedLinkID: 3,
				TenantID:      17,
				PathInfo:      1,
				entryAgentID:  emptyEntryAgentId,
				entryTagID:    emptyEntryTagId,
			},
			"",
			"",
		},
		{
			"minimal except with custom blob",
			// aka TestMinimal
			"11-81@dt=fw4;4;1;2;3;0;0;1;727c;1h48656c6c6f2c20576f726c6421",
			"FW4;129;4;1;2;3;17;1;727c;1h48656c6c6f2c20576f726c6421",
			Fw4Tag{
				ClusterID:     129,
				ServerID:      4,
				AgentID:       1,
				TagID:         2,
				encodedLinkID: 3,
				TenantID:      17,
				PathInfo:      1,
				CustomBlob:    dummyBlob,
				entryAgentID:  emptyEntryAgentId,
				entryTagID:    emptyEntryTagId,
			},
			"",
			"",
		},
		{
			"negative encoded link ID plus extensions",
			// aka IgnoredPathsGCTest
			"11-81@dt=fw4;1;20e;0;0;1;0;21df;a515;2h01;3h020e;4h00;5h01;6h1c90dcad033ff3444ba500dc717df3e6;7he0989607a1448c20",
			"FW4;129;1;526;0;-2147483648;17;8671;a515;2h01;3h020e;4h00;5h01;6h1c90dcad033ff3444ba500dc717df3e6;7he0989607a1448c20",
			Fw4Tag{
				ClusterID:     129,
				ServerID:      1,
				AgentID:       526,
				encodedLinkID: 2147483648, // 2**31 => -x == x
				TenantID:      17,
				PathInfo:      8671,
				tagDepth:      1,
				entryAgentID:  526,
				entryTagID:    0,
				payloadBitset: 1,
				TraceID:       traceIdOrPanic("1c90dcad033ff3444ba500dc717df3e6"),
				SpanID:        spanIdOrPanic("e0989607a1448c20"),
			},
			"",
			"",
		},
		{
			"negative tenant id",
			"968a934b-81@dt=fw4;1;2;3;4;0;0;3039",
			"FW4;129;1;2;3;4;-1769303221;12345",
			Fw4Tag{
				ClusterID:     129,
				ServerID:      1,
				AgentID:       2,
				TagID:         3,
				encodedLinkID: 4,
				TenantID:      -1769303221,
				PathInfo:      12345,
				entryTagID:    emptyEntryTagId,
				entryAgentID:  emptyEntryAgentId,
			},
			"",
			"",
		},
		{
			"negative cluster id",
			"9b-993f1f5b@dt=fw4;1;2;3;4;0;0;3039",
			"FW4;-1723916453;1;2;3;4;155;12345",
			Fw4Tag{
				ClusterID:     -1723916453,
				ServerID:      1,
				AgentID:       2,
				TagID:         3,
				encodedLinkID: 4,
				TenantID:      155,
				PathInfo:      12345,
				entryTagID:    emptyEntryTagId,
				entryAgentID:  emptyEntryAgentId,
			},
			"",
			"",
		},
		{
			"negative tenant and cluster id",
			"968a934b-993f1f5b@dt=fw4;1;2;3;4;0;0;3039",
			"FW4;-1723916453;1;2;3;4;-1769303221;12345",
			Fw4Tag{
				ClusterID:     -1723916453,
				ServerID:      1,
				AgentID:       2,
				TagID:         3,
				encodedLinkID: 4,
				TenantID:      -1769303221,
				PathInfo:      12345,
				entryTagID:    emptyEntryTagId,
				entryAgentID:  emptyEntryAgentId,
			},
			"",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.comment, func(t *testing.T) {
			t.Run("X-dynaTrace", func(t *testing.T) {

				// Test parsing
				fw4, err := ParseXDynatrace(tc.xdt)
				if err != nil {
					t.Errorf("Parser error for X-dynaTrace: %q: %v (shambles=%v)", tc.xdt, err, fw4)
				} else if fw4 != tc.parsed {
					t.Errorf("Expected %v got %v from %q", tc.parsed, fw4, tc.xdt)
				}

				// Test formatting
				expectedXdt := tc.expectedXdt
				if len(expectedXdt) == 0 {
					expectedXdt = tc.xdt
				}
				actualXdt := tc.parsed.ToXDynatrace()
				if expectedXdt != actualXdt {
					t.Errorf("Expected %q got %q from %v", expectedXdt, actualXdt, tc.parsed)
				}

			})
			t.Run("tracestate@dt", func(t *testing.T) {
				ts, err := trace.ParseTraceState(tc.tracestate)
				if err != nil {
					t.Fatal(err)
				}
				key := tc.parsed.TraceStateKey()
				val := ts.Get(key)
				if val == "" {
					t.Fatalf("Could not find tracestate entry, key %q", key)
				}
				fw4, err := ParseTracestateEntryValue(val)
				if err != nil {
					t.Fatalf("Parser error for X-dynaTrace: %q: %v (shambles=%v)", tc.xdt, err, fw4)
				}
				fw4.ClusterID = tc.parsed.ClusterID
				fw4.TenantID = tc.parsed.TenantID
				if fw4 != tc.parsed {
					t.Errorf("Expected %v got %v from %q", tc.parsed, fw4, tc.xdt)
				}

				// Test formatting
				expectedTracestate := tc.expectedTracestate
				if len(expectedTracestate) == 0 {
					expectedTracestate = tc.tracestate
				}
				actualTracestate := fmt.Sprintf("%s=%s", tc.parsed.TraceStateKey(), tc.parsed.ToTracestateEntryValue())
				if expectedTracestate != actualTracestate {
					t.Errorf("Expected %q got %q from %v", expectedTracestate, actualTracestate, tc.parsed)
				}
			})
		})
	}

}

func TestParseHexInt(t *testing.T) {
	result, err := parseHexIntToken("ec354cd6")
	if err != nil {
		t.Fatal(err)
	}
	if result != -332051242 {
		t.Fatalf("Unexpected result %v", result)
	}
}

func TestFW4TagWithTraceIdToTracestateEntryValue(t *testing.T) {
	fw4, err := ParseXDynatrace("FW4;129;1;526;0;0;17;12345;ce1f;2h01;6h11223344556677889900112233445566;7h663055fc5bca216f")
	require.NoError(t, err)

	require.Equal(t, fw4.ToTracestateEntryValueWithoutTraceId(), "fw4;1;20e;0;0;0;0;3039;28aa;2h01;7h663055fc5bca216f")
	require.Equal(t, fw4.ToTracestateEntryValue(), "fw4;1;20e;0;0;0;0;3039;ce1f;2h01;6h11223344556677889900112233445566;7h663055fc5bca216f")
}

func TestFW4TagWithoutTraceIdToTracestateEntryValue(t *testing.T) {
	fw4, err := ParseXDynatrace("FW4;129;1;526;0;0;17;12345;28aa;2h01;7h663055fc5bca216f")
	require.NoError(t, err)

	require.Equal(t, fw4.ToTracestateEntryValueWithoutTraceId(), "fw4;1;20e;0;0;0;0;3039;28aa;2h01;7h663055fc5bca216f")
	// FW4 tag does not contain traceId, thus there must not be traceId in tracestate at all
	require.Equal(t, fw4.ToTracestateEntryValue(), "fw4;1;20e;0;0;0;0;3039;28aa;2h01;7h663055fc5bca216f")
}

func TestGetMatchingFw4FromXDynatrace(t *testing.T) {
	var tenantId, clusterId int32 = 17, 129
	tag, err := GetMatchingFw4FromXDynatrace("FW4;129;1;526;0;0;17;12345;ce1f;2h01;6h11223344556677889900112233445566;7h663055fc5bca216f",
		configuration.QualifiedTenantId{TenantId: tenantId, ClusterId: clusterId})
	require.NoError(t, err)

	require.EqualValues(t, tag.TenantID, tenantId)
	require.EqualValues(t, tag.ClusterID, clusterId)
	require.Equal(t, tag.TraceID.String(), "11223344556677889900112233445566")
	require.Equal(t, tag.SpanID.String(), "663055fc5bca216f")
}

func TestGetMatchingFw4FromTracestate(t *testing.T) {
	var tenantId, clusterId int32 = 17, 129
	ts := trace.TraceState{}
	ts, err := ts.Insert("11-81@dt", "fw4;1;ec354cd6;1;2;0;0;3039;fa64;1h48656c6c6f2c20576f726c6421;2h0123;7h8877665544332211")
	require.NoError(t, err)

	tag, err := GetMatchingFw4FromTracestate(ts, configuration.QualifiedTenantId{TenantId: tenantId, ClusterId: clusterId})
	require.NoError(t, err)

	require.EqualValues(t, tag.TenantID, tenantId)
	require.EqualValues(t, tag.ClusterID, clusterId)
	require.False(t, tag.TraceID.IsValid())
	require.Equal(t, tag.SpanID.String(), "8877665544332211")
}
