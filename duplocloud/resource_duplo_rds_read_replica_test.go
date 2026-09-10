package duplocloud

import (
	"strings"
	"testing"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
)

func TestRdsClusterRecordName(t *testing.T) {
	cases := []struct {
		name     string
		given    string
		expected string
	}{
		{
			name:     "aurora cluster identifier maps to its record",
			given:    "duploservices-myapp-mydb-cluster",
			expected: "duploservices-myapp-mydb",
		},
		{
			// A global database secondary is named after the secondary tenant
			// and follows the same record-plus-suffix invariant.
			name:     "global secondary cluster identifier maps to its record",
			given:    "duploservices-dr-mydb-cluster",
			expected: "duploservices-dr-mydb",
		},
		{
			name:     "only the trailing suffix is stripped",
			given:    "duploservices-myapp-mydb-cluster-cluster",
			expected: "duploservices-myapp-mydb-cluster",
		},
		{
			name:     "suffix in the middle of the name is kept",
			given:    "duploservices-myapp-cluster-db-cluster",
			expected: "duploservices-myapp-cluster-db",
		},
		{
			name:     "writer instance identifier is used as-is",
			given:    "duploservices-myapp-mydb",
			expected: "duploservices-myapp-mydb",
		},
		{
			name:     "suffix match is case-insensitive",
			given:    "duploservices-myapp-mydb-CLUSTER",
			expected: "duploservices-myapp-mydb",
		},
		{
			name:     "surrounding whitespace is trimmed",
			given:    "  duploservices-myapp-mydb-cluster  ",
			expected: "duploservices-myapp-mydb",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := rdsClusterRecordName(tc.given); got != tc.expected {
				t.Errorf("rdsClusterRecordName(%q) = %q, want %q", tc.given, got, tc.expected)
			}
		})
	}
}

func TestHasRdsClusterSuffix(t *testing.T) {
	cases := map[string]bool{
		"duploservices-myapp-mydb-cluster": true,
		"duploservices-myapp-mydb-CLUSTER": true,
		"duploservices-myapp-mydb":         false,
		"duploservices-myapp-cluster-db":   false,
		"":                                 false,
	}
	for given, expected := range cases {
		if got := hasRdsClusterSuffix(given); got != expected {
			t.Errorf("hasRdsClusterSuffix(%q) = %v, want %v", given, got, expected)
		}
	}
}

func TestValidateReadReplicaClusterTarget(t *testing.T) {
	cases := []struct {
		name    string
		cluster *duplosdk.DuploRdsInstance
		wantErr string
	}{
		{
			name: "headless global secondary is rejected with a make_headless hint",
			cluster: &duplosdk.DuploRdsInstance{
				IsGlobalClusterMember:   true,
				GlobalClusterMemberRole: "secondary",
				IsHeadlessCluster:       true,
				GlobalClusterId:         "mydb-global",
			},
			wantErr: "make_headless = false",
		},
		{
			name: "global secondary with headless off is allowed",
			cluster: &duplosdk.DuploRdsInstance{
				IsGlobalClusterMember:   true,
				GlobalClusterMemberRole: "secondary",
				IsHeadlessCluster:       false,
			},
		},
		{
			// Aurora Limitless clusters are permanently headless but not global
			// members; they keep their existing path.
			name: "headless cluster that is not a global member is allowed",
			cluster: &duplosdk.DuploRdsInstance{
				IsHeadlessCluster: true,
			},
		},
		{
			name: "headless global primary is allowed",
			cluster: &duplosdk.DuploRdsInstance{
				IsGlobalClusterMember:   true,
				GlobalClusterMemberRole: "primary",
				IsHeadlessCluster:       true,
			},
		},
		{
			name: "role comparison is case-insensitive",
			cluster: &duplosdk.DuploRdsInstance{
				IsGlobalClusterMember:   true,
				GlobalClusterMemberRole: "Secondary",
				IsHeadlessCluster:       true,
			},
			wantErr: "headless mode",
		},
		{
			name:    "regular cluster is allowed",
			cluster: &duplosdk.DuploRdsInstance{},
		},
		{
			name:    "nil cluster is a no-op",
			cluster: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateReadReplicaClusterTarget(tc.cluster, "duploservices-dr-mydb-cluster")
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected an error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}
