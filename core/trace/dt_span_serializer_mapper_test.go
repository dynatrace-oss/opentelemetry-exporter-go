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

package trace

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	protoCommon "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/common/v1"
	protoTrace "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/trace/v1"
)

func TestGetProtoCustomTag(t *testing.T) {
	require.NotNil(t, getProtoCustomTag("customTag"))
	require.Nil(t, getProtoCustomTag(""))
}

func TestCreateProtoAttribute(t *testing.T) {
	testCases := []struct {
		attribute                       attribute.KeyValue
		valueGetter                     func(*protoCommon.AttributeKeyValue) interface{}
		expectedProtoAttributeValueType protoCommon.AttributeKeyValue_ValueType
	}{
		{
			attribute: attribute.String("string_attr", "value"),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetStringValue()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_STRING,
		},
		{
			attribute: attribute.Int("int_attr", 123),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetIntValue()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_INT,
		},
		{
			attribute: attribute.Float64("double_attr", 123.45),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetDoubleValue()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_DOUBLE,
		},
		{
			attribute: attribute.Bool("bool_attr", true),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetBoolValue()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_BOOL,
		},
		{
			attribute: attribute.StringSlice("string_array_attr", []string{"foo", "bar"}),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetStringValues()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_STRING_ARRAY,
		},
		{
			attribute: attribute.Int64Slice("int_array_attr", []int64{1, 2, 3}),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetIntValues()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_INT_ARRAY,
		},
		{
			attribute: attribute.Float64Slice("double_array_attr", []float64{1.1, 2.2, 3.3}),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetDoubleValues()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_DOUBLE_ARRAY,
		},
		{
			attribute: attribute.BoolSlice("bool_array_attr", []bool{true, false, true}),
			valueGetter: func(kv *protoCommon.AttributeKeyValue) interface{} {
				return kv.GetBoolValues()
			},
			expectedProtoAttributeValueType: protoCommon.AttributeKeyValue_BOOL_ARRAY,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Test proto attribute creation for attribute %s", tc.attribute.Key), func(t *testing.T) {
			protoAttribute, err := createProtoAttribute(tc.attribute)
			require.NoError(t, err)
			require.Equal(t, string(tc.attribute.Key), protoAttribute.GetKey())
			require.Equal(t, tc.expectedProtoAttributeValueType, protoAttribute.GetType())
			require.Equal(t, tc.valueGetter(protoAttribute), tc.attribute.Value.AsInterface())
		})
	}
}

func TestGetProtoAttributes(t *testing.T) {
	attributes := []attribute.KeyValue{
		attribute.String("string_attr", "value"),
		attribute.Int("int_attr", 123),
		attribute.Float64("double_attr", 123.45),
		attribute.Bool("bool_attr", true),
		attribute.StringSlice("string_array_attr", []string{"foo", "bar"}),
		attribute.Int64Slice("int_array_attr", []int64{1, 2, 3}),
		attribute.Float64Slice("double_array_attr", []float64{1.1, 2.2, 3.3}),
		attribute.BoolSlice("bool_array_attr", []bool{true, false, true}),
	}

	protoAttributes, err := getProtoAttributes(attributes)
	require.NoError(t, err)
	require.Len(t, protoAttributes, len(attributes))
}

func TestGetProtoStatus(t *testing.T) {
	status := sdktrace.Status{
		Code:        codes.Ok,
		Description: "description",
	}
	protoStatus, err := getProtoStatus(status)
	require.NoError(t, err)
	require.Equal(t, protoStatus.GetCode(), protoTrace.Status_Ok)
	require.Equal(t, protoStatus.GetMessage(), status.Description)
}

func TestGetProtoStatusCode(t *testing.T) {
	codeToProtoStatusCodeMap := map[codes.Code]protoTrace.Status_StatusCode{
		codes.Ok:    protoTrace.Status_Ok,
		codes.Unset: protoTrace.Status_Ok,
		codes.Error: protoTrace.Status_UnknownError,
	}

	for code, expectedProtoStatusCode := range codeToProtoStatusCodeMap {
		protoStatusCode, err := getProtoStatusCode(code)
		require.NoError(t, err)
		require.Equal(t, expectedProtoStatusCode, protoStatusCode)
	}
}

func TestGetProtoStatusCode_InvalidCode(t *testing.T) {
	_, err := getProtoStatusCode(codes.Ok + 1)
	require.Error(t, err)
}
