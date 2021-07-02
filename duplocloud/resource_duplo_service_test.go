package duplocloud

import (
	"reflect"
	"testing"
)

func TestReorderOtherDockerConfigEnvironmentVariables(t *testing.T) {
	cases := []struct {
		given    map[string]interface{}
		expected map[string]interface{}
	}{
		// basic case
		{
			given: map[string]interface{}{
				"Env": []interface{}{
					map[string]interface{}{"Name": "foo", "Value": "bar"},
					map[string]interface{}{"Name": "bar", "Value": "foo"},
				},
			},
			expected: map[string]interface{}{
				"Env": []interface{}{
					map[string]interface{}{"Name": "bar", "Value": "foo"},
					map[string]interface{}{"Name": "foo", "Value": "bar"},
				},
			},
		},

		// user giving wrong capitalization
		{
			given: map[string]interface{}{
				"Env": []interface{}{
					map[string]interface{}{"name": "foo", "value": "bar"},
					map[string]interface{}{"name": "bar", "value": "foo"},
				},
			},
			expected: map[string]interface{}{
				"Env": []interface{}{
					map[string]interface{}{"Name": "bar", "Value": "foo"},
					map[string]interface{}{"Name": "foo", "Value": "bar"},
				},
			},
		},

		// improper env var format shouldn't crash
		{
			given: map[string]interface{}{
				"Env": []interface{}{
					map[string]interface{}{"badname": "foo", "Value": "bar"},
					map[string]interface{}{"badname": "bar", "Value": "foo"},
				},
			},
			expected: map[string]interface{}{
				"Env": []interface{}{
					map[string]interface{}{"Badname": "foo", "Value": "bar"},
					map[string]interface{}{"Badname": "bar", "Value": "foo"},
				},
			},
		},
	}

	for _, c := range cases {
		reorderOtherDockerConfigsEnvironmentVariables(c.given)
		if !reflect.DeepEqual(c.given, c.expected) {
			t.Fatalf("Error matching output and expected: %#v vs %#v", c.given, c.expected)
		}
	}
}

func TestReduceOtherDockerConfig(t *testing.T) {
	cases := []struct {
		given    map[string]interface{}
		expected map[string]interface{}
	}{
		// basic case
		{
			given: map[string]interface{}{
				"Annotations":        nil,
				"Labels":             nil,
				"PodAnnotations":     nil,
				"PodLabels":          nil,
				"ServiceAnnotations": nil,
				"ServiceLabels":      nil,
				"Command":            nil,
				"LivenessProbe":      nil,
				"ReadinessProbe": map[string]interface{}{
					"HttpGet": map[string]interface{}{
						"Path": "/",
					},
				},
				"Env": []interface{}{
					map[string]interface{}{"Name": "foo", "Value": "bar", "ValueFrom": nil},
					map[string]interface{}{"Name": "bar", "Value": "foo"},
				},
			},
			expected: map[string]interface{}{
				"HostNetwork": false,
				"ReadinessProbe": map[string]interface{}{
					"HttpGet": map[string]interface{}{
						"Path": "/",
					},
				},
				"Env": []interface{}{
					map[string]interface{}{"Name": "bar", "Value": "foo"},
					map[string]interface{}{"Name": "foo", "Value": "bar"},
				},
			},
		},

		// user giving wrong capitalization
		{
			given: map[string]interface{}{
				"annotations":        nil,
				"labels":             nil,
				"podAnnotations":     nil,
				"podLabels":          nil,
				"serviceAnnotations": nil,
				"serviceLabels":      nil,
				"command":            nil,
				"env": []interface{}{
					map[string]interface{}{"name": "foo", "value": "bar"},
					map[string]interface{}{"name": "bar", "value": "foo"},
				},
			},
			expected: map[string]interface{}{
				"HostNetwork": false,
				"Env": []interface{}{
					map[string]interface{}{"Name": "bar", "Value": "foo"},
					map[string]interface{}{"Name": "foo", "Value": "bar"},
				},
			},
		},

		// user missing HostNetwork
		{
			given: map[string]interface{}{},
			expected: map[string]interface{}{
				"HostNetwork": false,
			},
		},
		{
			given: map[string]interface{}{
				"HostNetwork": nil,
			},
			expected: map[string]interface{}{
				"HostNetwork": false,
			},
		},

		/*
			// do not crash when types are wrong.
			{
				given: map[string]interface{}{
					"Cpu":   "hi",
					"Name":  "default",
					"Image": "nginx:latest",
					"Environment": []interface{}{
						map[string]interface{}{"Name": "bar", "Value": "foo"},
						map[string]interface{}{"Name": "foo", "Value": "bar"},
					},
					"PortMappings": map[string]interface{}{
						"this": []string{"is", "wrong", "json"},
					},
				},
				expected: map[string]interface{}{
					"Cpu":       "hi",
					"Name":      "default",
					"Image":     "nginx:latest",
					"Essential": true,
					"Environment": []interface{}{
						map[string]interface{}{"Name": "bar", "Value": "foo"},
						map[string]interface{}{"Name": "foo", "Value": "bar"},
					},
					"PortMappings": map[string]interface{}{
						"this": []string{"is", "wrong", "json"},
					},
				},
			},
		*/
	}

	for _, c := range cases {
		err := reduceOtherDockerConfig(c.given)
		if err != nil {
			t.Fatalf("Unexpected error from reduceOtherDockerConfig: %s", err)
		}
		if !reflect.DeepEqual(c.given, c.expected) {
			t.Fatalf("Error matching output and expected: %#v vs %#v", c.given, c.expected)
		}
	}
}
