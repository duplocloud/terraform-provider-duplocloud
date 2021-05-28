package duplocloud

import (
	"reflect"
	"testing"
)

func TestReorderEcsEnvironmentVariables(t *testing.T) {
	cases := []struct {
		given    map[string]interface{}
		expected map[string]interface{}
	}{
		// basic case
		{
			given: map[string]interface{}{
				"Environment": []interface{}{
					map[string]interface{}{"Name": "foo", "Value": "bar"},
					map[string]interface{}{"Name": "bar", "Value": "foo"},
				},
			},
			expected: map[string]interface{}{
				"Environment": []interface{}{
					map[string]interface{}{"Name": "bar", "Value": "foo"},
					map[string]interface{}{"Name": "foo", "Value": "bar"},
				},
			},
		},

		// user giving wrong capitalization
		{
			given: map[string]interface{}{
				"Environment": []interface{}{
					map[string]interface{}{"name": "foo", "value": "bar"},
					map[string]interface{}{"name": "bar", "value": "foo"},
				},
			},
			expected: map[string]interface{}{
				"Environment": []interface{}{
					map[string]interface{}{"Name": "bar", "Value": "foo"},
					map[string]interface{}{"Name": "foo", "Value": "bar"},
				},
			},
		},

		// improper env var format shouldn't crash
		{
			given: map[string]interface{}{
				"Environment": []interface{}{
					map[string]interface{}{"badname": "foo", "Value": "bar"},
					map[string]interface{}{"badname": "bar", "Value": "foo"},
				},
			},
			expected: map[string]interface{}{
				"Environment": []interface{}{
					map[string]interface{}{"Badname": "foo", "Value": "bar"},
					map[string]interface{}{"Badname": "bar", "Value": "foo"},
				},
			},
		},
	}

	for _, c := range cases {
		reorderEcsEnvironmentVariables(c.given)
		if !reflect.DeepEqual(c.given, c.expected) {
			t.Fatalf("Error matching output and expected: %#v vs %#v", c.given, c.expected)
		}
	}
}

func TestReduceContaineDefinition(t *testing.T) {
	cases := []struct {
		given    map[string]interface{}
		isAWSVPC bool
		expected map[string]interface{}
	}{
		// basic case
		{
			given: map[string]interface{}{
				"Name":  "default",
				"Image": "nginx:latest",
				"Environment": []interface{}{
					map[string]interface{}{"Name": "foo", "Value": "bar"},
					map[string]interface{}{"Name": "bar", "Value": "foo"},
				},
			},
			expected: map[string]interface{}{
				"Name":      "default",
				"Image":     "nginx:latest",
				"Essential": true,
				"Environment": []interface{}{
					map[string]interface{}{"Name": "bar", "Value": "foo"},
					map[string]interface{}{"Name": "foo", "Value": "bar"},
				},
			},
		},

		// user giving wrong capitalization
		{
			given: map[string]interface{}{
				"name":      "default",
				"image":     "nginx:latest",
				"essential": false,
				"environment": []interface{}{
					map[string]interface{}{"name": "foo", "value": "bar"},
					map[string]interface{}{"name": "bar", "value": "foo"},
				},
			},
			expected: map[string]interface{}{
				"Name":      "default",
				"Image":     "nginx:latest",
				"Essential": false,
				"Environment": []interface{}{
					map[string]interface{}{"Name": "bar", "Value": "foo"},
					map[string]interface{}{"Name": "foo", "Value": "bar"},
				},
			},
		},

		// user missing HostPort
		{
			given: map[string]interface{}{
				"Name":      "default",
				"Image":     "nginx:latest",
				"Essential": true,
				"PortMappings": []interface{}{
					map[string]interface{}{
						"Protocol": "tcp",
						"HostPort": 0,
					},
				},
			},
			isAWSVPC: true,
			expected: map[string]interface{}{
				"Name":      "default",
				"Image":     "nginx:latest",
				"Essential": true,
				"PortMappings": []interface{}{
					map[string]interface{}{
						"Protocol": nil,
						"HostPort": nil,
					},
				},
			},
		},
		{
			given: map[string]interface{}{
				"Name":  "default",
				"Image": "nginx:latest",
				"PortMappings": []interface{}{
					map[string]interface{}{
						"Protocol":      "tcp",
						"ContainerPort": "80",
					},
				},
			},
			isAWSVPC: true,
			expected: map[string]interface{}{
				"Name":      "default",
				"Image":     "nginx:latest",
				"Essential": true,
				"PortMappings": []interface{}{
					map[string]interface{}{
						"Protocol":      nil,
						"HostPort":      "80",
						"ContainerPort": "80",
					},
				},
			},
		},

		// user giving ports as strings
		{
			given: map[string]interface{}{
				"Name":      "default",
				"Image":     "nginx:latest",
				"Essential": true,
				"PortMappings": []interface{}{
					map[string]interface{}{
						"Protocol": "tcp",
						"HostPort": "80",
					},
				},
			},
			isAWSVPC: true,
			expected: map[string]interface{}{
				"Name":      "default",
				"Image":     "nginx:latest",
				"Essential": true,
				"PortMappings": []interface{}{
					map[string]interface{}{
						"Protocol": nil,
						"HostPort": "80",
					},
				},
			},
		},
		{
			given: map[string]interface{}{
				"Name":      "default",
				"Image":     "nginx:latest",
				"Essential": true,
				"PortMappings": []interface{}{
					map[string]interface{}{
						"Protocol": "tcp",
						"HostPort": "0",
					},
				},
			},
			isAWSVPC: true,
			expected: map[string]interface{}{
				"Name":      "default",
				"Image":     "nginx:latest",
				"Essential": true,
				"PortMappings": []interface{}{
					map[string]interface{}{
						"Protocol": nil,
						"HostPort": nil,
					},
				},
			},
		},

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
	}

	for _, c := range cases {
		err := reduceContainerDefinition(c.given, c.isAWSVPC)
		if err != nil {
			t.Fatalf("Unexpected error from reduceContainerDefinition: %s", err)
		}
		if !reflect.DeepEqual(c.given, c.expected) {
			t.Fatalf("Error matching output and expected: %#v vs %#v", c.given, c.expected)
		}
	}
}
