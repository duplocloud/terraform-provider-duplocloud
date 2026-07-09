package duplocloud

import (
	"reflect"
	"testing"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
)

func TestExpandV2ScalingConfiguration(t *testing.T) {
	cases := []struct {
		name     string
		given    []interface{}
		expected *duplosdk.V2ScalingConfiguration
	}{
		{
			// Auto-pause: min_capacity 0 is valid and must be preserved (not
			// treated as "unset"), along with seconds_until_auto_pause.
			name: "auto-pause with min_capacity 0",
			given: []interface{}{
				map[string]interface{}{
					"min_capacity":             float64(0),
					"max_capacity":             float64(4),
					"seconds_until_auto_pause": 3600,
				},
			},
			expected: &duplosdk.V2ScalingConfiguration{
				MinCapacity:           0,
				MaxCapacity:           4,
				SecondsUntilAutoPause: 3600,
			},
		},
		{
			// Standard config: positive min capacity, no auto-pause.
			name: "standard min/max without auto-pause",
			given: []interface{}{
				map[string]interface{}{
					"min_capacity":             float64(0.5),
					"max_capacity":             float64(2),
					"seconds_until_auto_pause": 0,
				},
			},
			expected: &duplosdk.V2ScalingConfiguration{
				MinCapacity: 0.5,
				MaxCapacity: 2,
			},
		},
		{
			// Empty block -> nil.
			name:     "empty config",
			given:    []interface{}{},
			expected: nil,
		},
		{
			// max_capacity 0 means the block was not populated -> nil.
			name: "unpopulated block (max_capacity 0)",
			given: []interface{}{
				map[string]interface{}{
					"min_capacity": float64(0),
					"max_capacity": float64(0),
				},
			},
			expected: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := expandV2ScalingConfiguration(c.given)
			if !reflect.DeepEqual(got, c.expected) {
				t.Fatalf("Error matching output and expected: %#v vs %#v", got, c.expected)
			}
		})
	}
}
