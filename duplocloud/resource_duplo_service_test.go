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

func TestOtherDockerConfigsAreEquivalent(t *testing.T) {
	cases := []struct {
		name string
		old  string
		new  string
		want bool
	}{
		{
			name: "single env entry casing",
			old:  `{"env":[{"name":"foo","value":"bar"}]}`,
			new:  `{"Env":[{"Name":"foo","Value":"bar"}]}`,
			want: true,
		},
		{
			name: "envFrom reference casing",
			old:  `{"envFrom":[{"secretRef":{"name":"secret","optional":true}},{"configMapRef":{"name":"config"}}]}`,
			new:  `{"EnvFrom":[{"SecretRef":{"Name":"secret","Optional":true}},{"ConfigMapRef":{"Name":"config"}}]}`,
			want: true,
		},
		{
			name: "startup probe casing",
			old:  `{"startupProbe":{"httpGet":{"path":"/ready","port":8080}}}`,
			new:  `{"StartupProbe":{"HttpGet":{"Path":"/ready","Port":8080}}}`,
			want: true,
		},
		{
			name: "envFrom Name value",
			old:  `{"EnvFrom":[{"SecretRef":{"Name":"first"}}]}`,
			new:  `{"EnvFrom":[{"SecretRef":{"Name":"second"}}]}`,
			want: false,
		},
		{
			name: "PodAnnotations key casing",
			old:  `{"PodAnnotations":{"example.com/key":"value"}}`,
			new:  `{"PodAnnotations":{"Example.com/key":"value"}}`,
			want: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			equal, err := otherDockerConfigsAreEquivalent(c.old, c.new)
			if err != nil {
				t.Fatalf("Unexpected error from otherDockerConfigsAreEquivalent: %s", err)
			}
			if equal != c.want {
				t.Fatalf("Expected equivalence %t, got %t", c.want, equal)
			}
		})
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
				"readinessProbe": map[string]interface{}{
					"httpGet": map[string]interface{}{
						"path": "/",
					},
				},
				"env": []interface{}{
					map[string]interface{}{"name": "foo", "value": "bar"},
					map[string]interface{}{"name": "bar", "value": "foo"},
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
