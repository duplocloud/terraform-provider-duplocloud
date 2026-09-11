package duplocloud

import (
	"reflect"
	"testing"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
)

func Test_expandTrustedKeyGroups(t *testing.T) {
	cases := []struct {
		name     string
		given    []interface{}
		expected *duplosdk.DuploCFDTrustedKeyGroups
	}{
		{
			name:  "empty list disables trusted key groups",
			given: []interface{}{},
			expected: &duplosdk.DuploCFDTrustedKeyGroups{
				Enabled:  false,
				Quantity: 0,
			},
		},
		{
			name:  "non-empty list enables trusted key groups",
			given: []interface{}{"key-group-1", "key-group-2"},
			expected: &duplosdk.DuploCFDTrustedKeyGroups{
				Enabled:  true,
				Quantity: 2,
				Items:    []string{"key-group-1", "key-group-2"},
			},
		},
		{
			name:  "quantity matches filtered items when list contains empty strings",
			given: []interface{}{"key-group-1", "", "key-group-2"},
			expected: &duplosdk.DuploCFDTrustedKeyGroups{
				Enabled:  true,
				Quantity: 2,
				Items:    []string{"key-group-1", "key-group-2"},
			},
		},
		{
			name:  "list of only empty strings disables trusted key groups",
			given: []interface{}{""},
			expected: &duplosdk.DuploCFDTrustedKeyGroups{
				Enabled:  false,
				Quantity: 0,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := expandTrustedKeyGroups(c.given)
			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("expected %+v, got %+v", c.expected, actual)
			}
		})
	}
}

func Test_flattenTrustedKeyGroups(t *testing.T) {
	cases := []struct {
		name     string
		given    *duplosdk.DuploCFDTrustedKeyGroups
		expected []interface{}
	}{
		{
			name:     "nil items flattens to empty list",
			given:    &duplosdk.DuploCFDTrustedKeyGroups{Enabled: false, Quantity: 0},
			expected: []interface{}{},
		},
		{
			name: "items flatten to string list",
			given: &duplosdk.DuploCFDTrustedKeyGroups{
				Enabled:  true,
				Quantity: 2,
				Items:    []string{"key-group-1", "key-group-2"},
			},
			expected: []interface{}{"key-group-1", "key-group-2"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := flattenTrustedKeyGroups(c.given)
			if !reflect.DeepEqual(actual, c.expected) {
				t.Errorf("expected %+v, got %+v", c.expected, actual)
			}
		})
	}
}
