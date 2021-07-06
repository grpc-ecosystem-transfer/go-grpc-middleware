package kv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKey(t *testing.T) {
	testCases := []struct {
		name          string
		keyValue      KeyValue
		expectedValue interface{}
	}{
		{
			name:          "true",
			keyValue:      Key("bool").Bool(true),
			expectedValue: true,
		},
		{
			name:          "false",
			keyValue:      Key("bool").Bool(false),
			expectedValue: false,
		},
		{
			name:          "int64",
			keyValue:      Key("int64").Int64(43),
			expectedValue: int64(43),
		},
		{
			name:          "float64",
			keyValue:      Key("float64").Float64(43),
			expectedValue: float64(43),
		},
		{
			name:          "string",
			keyValue:      Key("string").String("foo"),
			expectedValue: "foo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			switch tc.keyValue.Value.Type() {
			case BOOL:
				assert.Equal(t, tc.expectedValue, tc.keyValue.Value.AsBool())
			case INT64:
				assert.Equal(t, tc.expectedValue, tc.keyValue.Value.AsInt64())
			case FLOAT64:
				assert.Equal(t, tc.expectedValue, tc.keyValue.Value.AsFloat64())
			case STRING:
				assert.Equal(t, tc.expectedValue, tc.keyValue.Value.AsString())
			}
		})
	}
}
