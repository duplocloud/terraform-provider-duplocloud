package duplocloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// bigtableManualCluster builds the flattened shape of a cluster block that uses
// a fixed node count.
func bigtableManualCluster(id, zone string, numNodes int) map[string]interface{} {
	return map[string]interface{}{
		"cluster_id":         id,
		"zone":               zone,
		"num_nodes":          numNodes,
		"autoscaling_config": []interface{}{},
	}
}

// bigtableAutoscaledCluster builds the flattened shape of a cluster block that
// carries an autoscaling_config.
func bigtableAutoscaledCluster(id, zone string, minNodes, maxNodes int) map[string]interface{} {
	return map[string]interface{}{
		"cluster_id": id,
		"zone":       zone,
		"num_nodes":  minNodes,
		"autoscaling_config": []interface{}{
			map[string]interface{}{
				"min_nodes":      minNodes,
				"max_nodes":      maxNodes,
				"cpu_target":     60,
				"storage_target": 0,
			},
		},
	}
}

func TestValidateBigtableClusterBlocks_DuplicateClusterID_Rejected(t *testing.T) {
	err := validateBigtableClusterBlocks([]interface{}{
		bigtableManualCluster("c1", "us-east1-b", 3),
		bigtableManualCluster("c1", "us-east1-c", 3),
	})

	assert.ErrorContains(t, err, "duplicate cluster_id")
}

func TestValidateBigtableClusterBlocks_NoNodesAndNoAutoscaling_Rejected(t *testing.T) {
	err := validateBigtableClusterBlocks([]interface{}{
		bigtableManualCluster("c1", "us-east1-b", 0),
	})

	assert.ErrorContains(t, err, "either 'num_nodes' (> 0) or 'autoscaling_config' must be set")
}

func TestValidateBigtableClusterBlocks_Valid_Accepted(t *testing.T) {
	err := validateBigtableClusterBlocks([]interface{}{
		bigtableManualCluster("c1", "us-east1-b", 3),
		bigtableAutoscaledCluster("c2", "us-east1-c", 1, 5),
	})

	assert.NoError(t, err)
}

func TestValidateBigtableClusterTransitions_ZoneChanged_Rejected(t *testing.T) {
	err := validateBigtableClusterTransitions(
		[]interface{}{bigtableManualCluster("c1", "us-east1-b", 3)},
		[]interface{}{bigtableManualCluster("c1", "us-east1-c", 3)},
	)

	assert.ErrorContains(t, err, "zone is immutable")
}

func TestValidateBigtableClusterTransitions_NodeCountChanged_Accepted(t *testing.T) {
	err := validateBigtableClusterTransitions(
		[]interface{}{bigtableManualCluster("c1", "us-east1-b", 3)},
		[]interface{}{bigtableManualCluster("c1", "us-east1-b", 5)},
	)

	assert.NoError(t, err)
}

func TestValidateBigtableClusterTransitions_AutoscalingRemoved_Rejected(t *testing.T) {
	err := validateBigtableClusterTransitions(
		[]interface{}{bigtableAutoscaledCluster("c1", "us-east1-b", 1, 5)},
		[]interface{}{bigtableManualCluster("c1", "us-east1-b", 3)},
	)

	assert.ErrorContains(t, err, "removing 'autoscaling_config' is not supported")
}

func TestValidateBigtableClusterTransitions_AutoscalingAdded_Accepted(t *testing.T) {
	err := validateBigtableClusterTransitions(
		[]interface{}{bigtableManualCluster("c1", "us-east1-b", 3)},
		[]interface{}{bigtableAutoscaledCluster("c1", "us-east1-b", 1, 5)},
	)

	assert.NoError(t, err)
}

func TestValidateBigtableClusterTransitions_AutoscalingBoundsChanged_Accepted(t *testing.T) {
	err := validateBigtableClusterTransitions(
		[]interface{}{bigtableAutoscaledCluster("c1", "us-east1-b", 1, 5)},
		[]interface{}{bigtableAutoscaledCluster("c1", "us-east1-b", 2, 8)},
	)

	assert.NoError(t, err)
}

// Removing the cluster outright and adding a different one is how an operator
// relocates a cluster or drops its autoscaling, so neither guard applies.
func TestValidateBigtableClusterTransitions_ClusterReplaced_Accepted(t *testing.T) {
	err := validateBigtableClusterTransitions(
		[]interface{}{bigtableAutoscaledCluster("c1", "us-east1-b", 1, 5)},
		[]interface{}{bigtableManualCluster("c2", "us-east1-c", 3)},
	)

	assert.NoError(t, err)
}

func TestValidateBigtableClusterTransitions_NewResource_Accepted(t *testing.T) {
	err := validateBigtableClusterTransitions(
		[]interface{}{},
		[]interface{}{bigtableAutoscaledCluster("c1", "us-east1-b", 1, 5)},
	)

	assert.NoError(t, err)
}
