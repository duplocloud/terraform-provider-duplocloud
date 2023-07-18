package duplocloud

import (
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
