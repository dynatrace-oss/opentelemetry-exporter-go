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
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	protoCommon "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/common/v1"
	protoTrace "github.com/dynatrace-oss/opentelemetry-exporter-go/core/internal/odin-proto/trace/v1"
)

type testProtoAttributeAccessor struct {
	keyFuncName   string
	protoAttrName string
}

var testProtoAttributeMapping = map[protoCommon.AttributeKeyValue_ValueType]testProtoAttributeAccessor{
	protoCommon.AttributeKeyValue_BOOL:         {"Bool", "BoolValue"},
	protoCommon.AttributeKeyValue_BOOL_ARRAY:   {"BoolSlice", "BoolValues"},
	protoCommon.AttributeKeyValue_INT:          {"Int64", "IntValue"},
	protoCommon.AttributeKeyValue_INT_ARRAY:    {"Int64Slice", "IntValues"},
	protoCommon.AttributeKeyValue_DOUBLE:       {"Float64", "DoubleValue"},
	protoCommon.AttributeKeyValue_DOUBLE_ARRAY: {"Float64Slice", "DoubleValues"},
	protoCommon.AttributeKeyValue_STRING:       {"String", "StringValue"},
	protoCommon.AttributeKeyValue_STRING_ARRAY: {"StringSlice", "StringValues"},
}

func protoAttributesToMap(t *testing.T, protoAttributes []*protoCommon.AttributeKeyValue) map[attribute.Key]attribute.KeyValue {
	attributeMap := make(map[attribute.Key]attribute.KeyValue, len(protoAttributes))
	for _, protoAttr := range protoAttributes {
		key := attribute.Key(protoAttr.Key)
		require.NotContainsf(t, attributeMap, key, "Duplicate attribute '%s'", key)

		accessor, ok := testProtoAttributeMapping[protoAttr.Type]
		require.Truef(t, ok, "Unknown type '%v' for attribute '%s'", protoAttr.Type, key)

		protoValue := reflect.ValueOf(protoAttr).Elem().FieldByName(accessor.protoAttrName).Interface()
		attrValue := reflect.ValueOf(key).MethodByName(accessor.keyFuncName).Call([]reflect.Value{reflect.ValueOf(protoValue)})
		attributeMap[key] = attrValue[0].Interface().(attribute.KeyValue)
	}
	return attributeMap
}

func attributesToMap(t *testing.T, attributes []attribute.KeyValue) map[attribute.Key]attribute.KeyValue {
	mappedAttributes := make(map[attribute.Key]attribute.KeyValue)
	for _, attr := range attributes {
		require.NotContainsf(t, mappedAttributes, attr.Key, "Duplicate attribute '%s'", attr.Key)
		mappedAttributes[attr.Key] = attr
	}
	return mappedAttributes
}

func assertProtoAttributes(t *testing.T, protoAttributes []*protoCommon.AttributeKeyValue, expected []attribute.KeyValue) {
	protoAttrMap := protoAttributesToMap(t, protoAttributes)
	expectedAttrMap := attributesToMap(t, expected)

	assertAttributeEquals(t, expectedAttrMap, protoAttrMap)
}

func assertAttributeEquals(t *testing.T, expected, other map[attribute.Key]attribute.KeyValue) {
	require.Equal(t, len(expected), len(other))
	for key, expectedAttr := range expected {
		otherAttr, ok := other[key]
		require.Truef(t, ok, "Proto attribute '%s' missing", key)
		require.Equal(t, expectedAttr.Value.AsInterface(), otherAttr.Value.AsInterface())
	}
}

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
	assertProtoAttributes(t, protoAttributes, attributes)
}

func TestGetProtoSpanAttributes(t *testing.T) {
	testCases := []struct {
		propagatedAttributes map[attribute.Key]attribute.KeyValue
	}{
		{nil},
		{make(map[attribute.Key]attribute.KeyValue)},
	}
	attributes := []attribute.KeyValue{
		attribute.String("string_attr", "value"),
		attribute.Int("int_attr", 123),
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			protoAttributes, err := getProtoSpanAttributes(attributes, tc.propagatedAttributes)
			require.NoError(t, err)
			require.Len(t, protoAttributes, len(attributes))
		})
	}
}

func TestGetProtoSpanAttributes_WithPropagatedAttributes(t *testing.T) {
	propagatingAttributes := []attribute.KeyValue{
		attribute.Key("str.value").String("my.value"),
		attribute.Key("int.value").Int(42),
		attribute.Key("bool.value").Bool(true),
		attribute.Key("float.value").Float64(1.7),
		attribute.StringSlice("slice.value", []string{"hello", "prop", "attr"}),
	}
	attributes := []attribute.KeyValue{}

	protoAttributes, err := getProtoSpanAttributes(attributes, attributesToMap(t, propagatingAttributes))
	require.NoError(t, err)
	assertProtoAttributes(t, protoAttributes, propagatingAttributes)
}

func TestGetProtoSpanAttributes_OverwrittenPropagatedAttributes(t *testing.T) {
	propagatingAttributes := attributesToMap(t, []attribute.KeyValue{
		attribute.Key("str.value").String("prop.value"),
		attribute.Key("str.same.value").String("same.value"),
		attribute.Key("int.value").Int(17),
		attribute.Key("int.same.value").Int(42),
		attribute.Key("bool.value").Bool(true),
		attribute.Key("bool.same.value").Bool(true),
		attribute.Key("float.value").Float64(1.7),
		attribute.Key("float.same.value").Float64(4.2),
		attribute.Key("slice.value").StringSlice([]string{"span", "world"}),
		attribute.Key("slice.same.value").StringSlice([]string{"hello", "world"}),
	})
	attributes := []attribute.KeyValue{
		attribute.Key("str.value").String("span.value"),
		attribute.Key("str.same.value").String("same.value"),
		attribute.Key("int.value").Int(37),
		attribute.Key("int.same.value").Int(42),
		attribute.Key("bool.value").Bool(false),
		attribute.Key("bool.same.value").Bool(true),
		attribute.Key("float.value").Float64(3.7),
		attribute.Key("float.same.value").Float64(4.2),
		attribute.Key("slice.value").StringSlice([]string{"world", "span"}),
		attribute.Key("slice.same.value").StringSlice([]string{"hello", "world"}),
	}

	protoAttributes, err := getProtoSpanAttributes(attributes, propagatingAttributes)
	require.NoError(t, err)
	assertProtoAttributes(t, protoAttributes, []attribute.KeyValue{
		attribute.Key("str.value").String("prop.value"),
		attribute.Key("str.same.value").String("same.value"),
		attribute.Key("overwritten1.str.value").String("span.value"),
		attribute.Key("int.value").Int(17),
		attribute.Key("int.same.value").Int(42),
		attribute.Key("overwritten1.int.value").Int(37),
		attribute.Key("bool.value").Bool(true),
		attribute.Key("bool.same.value").Bool(true),
		attribute.Key("overwritten1.bool.value").Bool(false),
		attribute.Key("float.value").Float64(1.7),
		attribute.Key("float.same.value").Float64(4.2),
		attribute.Key("overwritten1.float.value").Float64(3.7),
		attribute.Key("slice.value").StringSlice([]string{"span", "world"}),
		attribute.Key("slice.same.value").StringSlice([]string{"hello", "world"}),
		attribute.Key("overwritten1.slice.value").StringSlice([]string{"world", "span"}),
	})
}

func TestGetProtoSpanAttributes_OverwrittenPropagatedAttributesWithDifferentType(t *testing.T) {
	propagatingAttributes := attributesToMap(t, []attribute.KeyValue{
		attribute.Key("str.value").String("my.value"),
		attribute.Key("int.value").Int64(17),
		attribute.Key("bool.value").Bool(true),
		attribute.Key("float.value").Float64(3.7),
		attribute.Key("slice.value").StringSlice([]string{"hello", "world"}),
	})
	attributes := []attribute.KeyValue{
		attribute.Key("str.value").Int64(99),
		attribute.Key("int.value").Int64Slice([]int64{1, 7}),
		attribute.Key("bool.value").String("hello bool"),
		attribute.Key("float.value").Bool(false),
		attribute.Key("slice.value").Float64(3.7),
	}

	protoAttributes, err := getProtoSpanAttributes(attributes, propagatingAttributes)
	require.NoError(t, err)
	assertProtoAttributes(t, protoAttributes, []attribute.KeyValue{
		attribute.Key("str.value").String("my.value"),
		attribute.Key("overwritten1.str.value").Int64(99),
		attribute.Key("int.value").Int64(17),
		attribute.Key("overwritten1.int.value").Int64Slice([]int64{1, 7}),
		attribute.Key("bool.value").Bool(true),
		attribute.Key("overwritten1.bool.value").String("hello bool"),
		attribute.Key("float.value").Float64(3.7),
		attribute.Key("overwritten1.float.value").Bool(false),
		attribute.Key("slice.value").StringSlice([]string{"hello", "world"}),
		attribute.Key("overwritten1.slice.value").Float64(3.7),
	})
}

func TestGetProtoSpanAttributes_PropagatingAttributesWithDefaultValues(t *testing.T) {
	propagatingAttributes := attributesToMap(t, []attribute.KeyValue{
		attribute.Key("str.default").String(""),
		attribute.Key("int.default").Int64(0),
		attribute.Key("bool.default").Bool(false),
		attribute.Key("float.default").Float64(0.0),
		attribute.Key("slice.default").StringSlice(nil),
	})
	attributes := []attribute.KeyValue{
		attribute.Key("str.default").String("my.value"),
		attribute.Key("int.default").Int64(17),
		attribute.Key("bool.default").Bool(true),
		attribute.Key("float.default").Float64(1.7),
		attribute.Key("slice.default").StringSlice([]string{"hello"}),
	}

	protoAttributes, err := getProtoSpanAttributes(attributes, propagatingAttributes)
	require.NoError(t, err)
	assertProtoAttributes(t, protoAttributes, []attribute.KeyValue{
		attribute.Key("str.default").String(""),
		attribute.Key("overwritten1.str.default").String("my.value"),
		attribute.Key("int.default").Int64(0),
		attribute.Key("overwritten1.int.default").Int64(17),
		attribute.Key("bool.default").Bool(false),
		attribute.Key("overwritten1.bool.default").Bool(true),
		attribute.Key("float.default").Float64(0.0),
		attribute.Key("overwritten1.float.default").Float64(1.7),
		attribute.Key("slice.default").StringSlice(nil),
		attribute.Key("overwritten1.slice.default").StringSlice([]string{"hello"}),
	})
}

func TestGetProtSpanAttributes_PropagatedAttributesSliceValues(t *testing.T) {
	propagatingAttributes := attributesToMap(t, []attribute.KeyValue{
		attribute.Key("slice.same.value").StringSlice([]string{"same", "same"}),
		attribute.Key("slice.different.order").StringSlice([]string{"first", "second", "third"}),
		attribute.Key("slice.different.elem.type").StringSlice([]string{"1", "2"}),
		attribute.Key("slice.different.size").StringSlice([]string{"one", "two", "three"}),
		attribute.Key("slice.different.type").Int64Slice([]int64{17}),
	})
	attributes := []attribute.KeyValue{
		attribute.Key("slice.same.value").StringSlice([]string{"same", "same"}),
		attribute.Key("slice.different.order").StringSlice([]string{"first", "third", "second"}),
		attribute.Key("slice.different.elem.type").Int64Slice([]int64{1, 2}),
		attribute.Key("slice.different.size").StringSlice([]string{"one", "two"}),
		attribute.Key("slice.different.type").Int64(17),
	}

	protoAttributes, err := getProtoSpanAttributes(attributes, propagatingAttributes)
	require.NoError(t, err)
	assertProtoAttributes(t, protoAttributes, []attribute.KeyValue{
		attribute.Key("slice.same.value").StringSlice([]string{"same", "same"}),
		attribute.Key("slice.different.order").StringSlice([]string{"first", "second", "third"}),
		attribute.Key("overwritten1.slice.different.order").StringSlice([]string{"first", "third", "second"}),
		attribute.Key("slice.different.elem.type").StringSlice([]string{"1", "2"}),
		attribute.Key("overwritten1.slice.different.elem.type").Int64Slice([]int64{1, 2}),
		attribute.Key("slice.different.size").StringSlice([]string{"one", "two", "three"}),
		attribute.Key("overwritten1.slice.different.size").StringSlice([]string{"one", "two"}),
		attribute.Key("slice.different.type").Int64Slice([]int64{17}),
		attribute.Key("overwritten1.slice.different.type").Int64(17),
	})
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
