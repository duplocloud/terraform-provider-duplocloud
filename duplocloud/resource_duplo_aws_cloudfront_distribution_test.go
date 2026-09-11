package duplocloud

import (
	"reflect"
	"testing"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/go-cty/cty"
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

func Test_cacheBehaviorConfiguresTrustedKeyGroups(t *testing.T) {
	behaviorWithValues := cty.ObjectVal(map[string]cty.Value{
		"trusted_key_groups": cty.ListVal([]cty.Value{cty.StringVal("kg-1")}),
	})
	behaviorWithEmptyList := cty.ObjectVal(map[string]cty.Value{
		"trusted_key_groups": cty.ListValEmpty(cty.String),
	})
	behaviorWithoutAttr := cty.ObjectVal(map[string]cty.Value{
		"trusted_key_groups": cty.NullVal(cty.List(cty.String)),
	})
	behaviorObjType := cty.Object(map[string]cty.Type{"trusted_key_groups": cty.List(cty.String)})

	cases := []struct {
		name     string
		block    cty.Value
		index    int
		expected bool
	}{
		{
			name:     "explicit values present",
			block:    cty.ListVal([]cty.Value{behaviorWithValues}),
			index:    0,
			expected: true,
		},
		{
			name:     "explicit empty list is still configured",
			block:    cty.ListVal([]cty.Value{behaviorWithEmptyList}),
			index:    0,
			expected: true,
		},
		{
			name:     "attribute omitted from config",
			block:    cty.ListVal([]cty.Value{behaviorWithoutAttr}),
			index:    0,
			expected: false,
		},
		{
			name:     "block itself absent",
			block:    cty.NullVal(cty.List(behaviorObjType)),
			index:    0,
			expected: false,
		},
		{
			name:     "index out of range",
			block:    cty.ListVal([]cty.Value{behaviorWithValues}),
			index:    1,
			expected: false,
		},
		{
			name:     "empty block list",
			block:    cty.ListValEmpty(behaviorObjType),
			index:    0,
			expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw := cty.ObjectVal(map[string]cty.Value{"default_cache_behavior": c.block})
			actual := cacheBehaviorConfiguresTrustedKeyGroups(raw, "default_cache_behavior", c.index)
			if actual != c.expected {
				t.Errorf("expected %v, got %v", c.expected, actual)
			}
		})
	}
}
