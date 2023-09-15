package duplocloud

import (
	"reflect"
	"terraform-provider-duplocloud/duplosdk"
	"testing"
)

func TestSortCommaDelimitedString(t *testing.T) {
	cases := []struct {
		given    string
		expected string
	}{
		// basic case
		{
			given:    "z-1.foo,z-3.foo,z-2.foo,z-2.bar",
			expected: "z-1.foo,z-2.bar,z-2.foo,z-3.foo",
		},

		// empty string
		{
			given:    "",
			expected: "",
		},

		// single value
		{
			given:    "foo",
			expected: "foo",
		},
	}

	for _, c := range cases {
		actual := sortCommaDelimitedString(c.given)
		if actual != c.expected {
			t.Errorf("Expected %s, got %s", c.expected, actual)
		}
	}
}

func Test_keyValueToState(t *testing.T) {
	// Test case with non-nil input
	input := []duplosdk.DuploKeyStringValue{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
		{Key: "key3", Value: "value3"},
	}
	expectedOutput := []interface{}{
		map[string]interface{}{"key": "key1", "value": "value1"},
		map[string]interface{}{"key": "key2", "value": "value2"},
		map[string]interface{}{"key": "key3", "value": "value3"},
	}
	type args struct {
		fieldName    string
		duploObjects *[]duplosdk.DuploKeyStringValue
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "non-nil input",
			args: args{
				fieldName:    "testFieldName",
				duploObjects: &input,
			},
			want: expectedOutput,
		},
		{
			name: "nil input",
			args: args{
				fieldName:    "testFieldName",
				duploObjects: nil,
			},
			want: []interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keyValueToState(tt.args.fieldName, tt.args.duploObjects); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("keyValueToState() = %v, want %v", got, tt.want)
			}
		})
	}
}
