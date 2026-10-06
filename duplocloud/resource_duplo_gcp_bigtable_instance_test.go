package duplocloud

import (
	"encoding/json"
	"testing"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	}, nil)

	assert.ErrorContains(t, err, "duplicate cluster_id")
}

func TestValidateBigtableClusterBlocks_NoNodesAndNoAutoscaling_Rejected(t *testing.T) {
	err := validateBigtableClusterBlocks([]interface{}{
		bigtableManualCluster("c1", "us-east1-b", 0),
	}, nil)

	assert.ErrorContains(t, err, "either 'num_nodes' (> 0) or 'autoscaling_config' must be set")
}

func TestValidateBigtableClusterBlocks_AutoscalingMaxBelowMin_Rejected(t *testing.T) {
	err := validateBigtableClusterBlocks([]interface{}{
		bigtableAutoscaledCluster("c1", "us-east1-b", 5, 2),
	}, nil)

	assert.ErrorContains(t, err, "'max_nodes' (2) must be greater than or equal to 'min_nodes' (5)")
}

func TestValidateBigtableClusterBlocks_AutoscalingEqualBounds_Accepted(t *testing.T) {
	err := validateBigtableClusterBlocks([]interface{}{
		bigtableAutoscaledCluster("c1", "us-east1-b", 3, 3),
	}, nil)

	assert.NoError(t, err)
}

func TestBigtableAutoscalingBounds_MustBePositive(t *testing.T) {
	ac := gcpBigtableClusterSchema()["autoscaling_config"].Elem.(*schema.Resource).Schema
	for _, field := range []string{"min_nodes", "max_nodes"} {
		_, errs := ac[field].ValidateFunc(0, field)
		assert.NotEmpty(t, errs, field)
		_, errs = ac[field].ValidateFunc(1, field)
		assert.Empty(t, errs, field)
	}
}

func TestValidateBigtableClusterBlocks_Valid_Accepted(t *testing.T) {
	err := validateBigtableClusterBlocks([]interface{}{
		bigtableManualCluster("c1", "us-east1-b", 3),
		bigtableAutoscaledCluster("c2", "us-east1-c", 1, 5),
	}, nil)

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

// The backend replaces the labels whenever the field is present and skips them when it
// is absent, so removing every label has to send an empty object rather than omit it.
func TestBigtableInstanceUpdateRequest_EmptyLabelsAreSent(t *testing.T) {
	empty := map[string]string{}
	body, err := json.Marshal(duplosdk.DuploBigtableInstanceUpdateRequest{DisplayName: "x", Labels: &empty})
	assert.NoError(t, err)
	assert.Contains(t, string(body), `"labels":{}`)

	body, err = json.Marshal(duplosdk.DuploBigtableInstanceUpdateRequest{DisplayName: "x"})
	assert.NoError(t, err)
	assert.NotContains(t, string(body), "labels")
}

func TestValidateBigtableClusterBlocks_AutoscalingMaxOverTenTimesMin_Rejected(t *testing.T) {
	err := validateBigtableClusterBlocks([]interface{}{
		bigtableAutoscaledCluster("c1", "us-east1-b", 1, 11),
	}, nil)

	assert.ErrorContains(t, err, "cannot be more than 10 times 'min_nodes' (1)")
	assert.NoError(t, validateBigtableClusterBlocks([]interface{}{bigtableAutoscaledCluster("c1", "us-east1-b", 1, 10)}, nil))
}

func bigtableClusterWithStorageTarget(target int) map[string]interface{} {
	cl := bigtableAutoscaledCluster("c1", "us-east1-b", 1, 5)
	cl["autoscaling_config"].([]interface{})[0].(map[string]interface{})["storage_target"] = target
	return cl
}

func TestValidateBigtableStorageTargets(t *testing.T) {
	for _, tc := range []struct {
		storageType string
		target      int
		ok          bool
	}{
		{"SSD", 0, true},
		{"SSD", 2560, true},
		{"SSD", 5120, true},
		{"SSD", 2559, false},
		{"SSD", 8192, false},
		{"HDD", 0, true},
		{"HDD", 8192, true},
		{"HDD", 16384, true},
		{"HDD", 5120, false},
		{"HDD", 16385, false},
	} {
		err := validateBigtableStorageTargets([]interface{}{bigtableClusterWithStorageTarget(tc.target)}, tc.storageType)
		if tc.ok {
			assert.NoError(t, err, "%s %d", tc.storageType, tc.target)
		} else {
			assert.ErrorContains(t, err, "'storage_target'", "%s %d", tc.storageType, tc.target)
		}
	}
	assert.NoError(t, validateBigtableStorageTargets([]interface{}{bigtableManualCluster("c1", "us-east1-b", 3)}, "SSD"), "manual cluster")
}

// bigtableRawConfig builds a raw configuration of the resource's own type holding the
// given cluster blocks, with every other attribute null.
func bigtableRawConfig(clusters ...map[string]cty.Value) cty.Value {
	ty := resourceGcpBigtableInstance().CoreConfigSchema().ImpliedType()
	clusterTy := ty.AttributeType("cluster").ElementType()
	elems := []cty.Value{}
	for _, attrs := range clusters {
		vals := map[string]cty.Value{}
		for name, at := range clusterTy.AttributeTypes() {
			if v, ok := attrs[name]; ok {
				vals[name] = v
			} else {
				vals[name] = cty.NullVal(at)
			}
		}
		elems = append(elems, cty.ObjectVal(vals))
	}
	vals := map[string]cty.Value{}
	for name, at := range ty.AttributeTypes() {
		vals[name] = cty.NullVal(at)
	}
	vals["cluster"] = cty.ListVal(elems)
	return cty.ObjectVal(vals)
}

func bigtableRawAutoscaling() cty.Value {
	acTy := resourceGcpBigtableInstance().CoreConfigSchema().ImpliedType().AttributeType("cluster").ElementType().AttributeType("autoscaling_config")
	return cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
		"min_nodes":      cty.NumberIntVal(1),
		"max_nodes":      cty.NumberIntVal(3),
		"cpu_target":     cty.NumberIntVal(50),
		"storage_target": cty.NullVal(acTy.ElementType().AttributeType("storage_target")),
	})})
}

func TestValidateBigtableScalingModes_BothSet_Rejected(t *testing.T) {
	err := validateBigtableScalingModes(bigtableRawConfig(map[string]cty.Value{
		"cluster_id":         cty.StringVal("c1"),
		"zone":               cty.StringVal("us-east1-b"),
		"num_nodes":          cty.NumberIntVal(3),
		"autoscaling_config": bigtableRawAutoscaling(),
	}))

	assert.ErrorContains(t, err, `cluster "c1": 'num_nodes' and 'autoscaling_config' are mutually exclusive`)
}

func TestValidateBigtableScalingModes_EitherAlone_Accepted(t *testing.T) {
	err := validateBigtableScalingModes(bigtableRawConfig(
		map[string]cty.Value{"cluster_id": cty.StringVal("c1"), "zone": cty.StringVal("us-east1-b"), "num_nodes": cty.NumberIntVal(3)},
		map[string]cty.Value{"cluster_id": cty.StringVal("c2"), "zone": cty.StringVal("us-east1-c"), "autoscaling_config": bigtableRawAutoscaling()},
	))

	assert.NoError(t, err)
	assert.NoError(t, validateBigtableScalingModes(cty.NullVal(cty.DynamicPseudoType)), "no config")
}

func bigtableAutoscaledRequest(minNodes, maxNodes, cpu, storage int) *duplosdk.DuploBigtableCluster {
	return &duplosdk.DuploBigtableCluster{
		ClusterConfig: &duplosdk.DuploBigtableClusterConfig{
			ClusterAutoscalingConfig: &duplosdk.DuploBigtableClusterAutoscalingConfig{
				AutoscalingLimits:  duplosdk.DuploBigtableAutoscalingLimits{MinServeNodes: minNodes, MaxServeNodes: maxNodes},
				AutoscalingTargets: duplosdk.DuploBigtableAutoscalingTargets{CpuUtilizationPercent: cpu, StorageUtilizationGibPerNode: storage},
			},
		},
	}
}

func TestBigtableClusterUpdateApplied_Manual(t *testing.T) {
	want := &duplosdk.DuploBigtableCluster{ServeNodes: 5}

	assert.True(t, bigtableClusterUpdateApplied(&duplosdk.DuploBigtableCluster{ServeNodes: 5}, want))
	assert.False(t, bigtableClusterUpdateApplied(&duplosdk.DuploBigtableCluster{ServeNodes: 3}, want))
}

func TestBigtableClusterUpdateApplied_Autoscaling(t *testing.T) {
	want := bigtableAutoscaledRequest(1, 5, 50, 0)

	assert.True(t, bigtableClusterUpdateApplied(bigtableAutoscaledRequest(1, 5, 50, 2560), want), "default storage target")
	assert.False(t, bigtableClusterUpdateApplied(bigtableAutoscaledRequest(1, 3, 50, 2560), want), "old limits")
	assert.False(t, bigtableClusterUpdateApplied(bigtableAutoscaledRequest(1, 5, 60, 2560), want), "old cpu target")
	assert.False(t, bigtableClusterUpdateApplied(&duplosdk.DuploBigtableCluster{ServeNodes: 5}, want), "not yet autoscaled")

	want = bigtableAutoscaledRequest(1, 5, 50, 4000)
	assert.True(t, bigtableClusterUpdateApplied(bigtableAutoscaledRequest(1, 5, 50, 4000), want), "explicit storage target")
	assert.False(t, bigtableClusterUpdateApplied(bigtableAutoscaledRequest(1, 5, 50, 2560), want), "old storage target")
}

func TestBigtableDisplayName(t *testing.T) {
	assert.Equal(t, "My Instance", bigtableDisplayName("My Instance", "my-instance"), "configured")
	assert.Equal(t, "my-instance", bigtableDisplayName("", "my-instance"), "derived from name")
	assert.Equal(t, "abcdefghij-abcdefghij-abcdefgh", bigtableDisplayName("", "abcdefghij-abcdefghij-abcdefghij-a"), "cut to 30")
}

func TestBigtableDisplayName_LengthValidated(t *testing.T) {
	validate := gcpBigtableInstanceSchema()["display_name"].ValidateFunc
	for _, v := range []string{"abc", "abcdefghij-abcdefghij-abcdefghi"} {
		_, errs := validate(v, "display_name")
		assert.NotEmpty(t, errs, v)
	}
	for _, v := range []string{"abcd", "abcdefghij-abcdefghij-abcdefgh"} {
		_, errs := validate(v, "display_name")
		assert.Empty(t, errs, v)
	}
}

// A num_nodes taken from a value known only at apply reads as 0 in the flattened diff,
// so the node-count check waits for it rather than rejecting the plan.
func TestValidateBigtableClusterBlocks_UnknownNumNodes_Deferred(t *testing.T) {
	clusters := []interface{}{bigtableManualCluster("c1", "us-east1-b", 0)}

	assert.NoError(t, validateBigtableClusterBlocks(clusters, map[int]bool{0: true}))
	assert.Error(t, validateBigtableClusterBlocks(clusters, nil))
}

func TestBigtableUnknownNumNodes(t *testing.T) {
	unknown := bigtableUnknownNumNodes(bigtableRawConfig(
		map[string]cty.Value{"cluster_id": cty.StringVal("c1"), "zone": cty.StringVal("us-east1-b"), "num_nodes": cty.NumberIntVal(3)},
		map[string]cty.Value{"cluster_id": cty.StringVal("c2"), "zone": cty.StringVal("us-east1-c"), "num_nodes": cty.UnknownVal(cty.Number)},
		map[string]cty.Value{"cluster_id": cty.StringVal("c3"), "zone": cty.StringVal("us-east1-d")},
	))

	assert.Equal(t, map[int]bool{1: true}, unknown)
}

func TestBigtableInstanceName_Validated(t *testing.T) {
	validate := gcpBigtableInstanceSchema()["name"].ValidateFunc
	for _, v := range []string{"abc", "abcde", "1abcdef", "Abcdef", "abcdef-", "abc_def", "abcdefghij-abcdefghij-abcdefghij-a"} {
		_, errs := validate(v, "name")
		assert.NotEmpty(t, errs, v)
	}
	for _, v := range []string{"abcdef", "my-bigtable-1", "abcdefghij-abcdefghij-abcdefghij"} {
		_, errs := validate(v, "name")
		assert.Empty(t, errs, v)
	}
}

func TestBigtableInstanceTypeChangeNeedsRecreate(t *testing.T) {
	assert.True(t, bigtableInstanceTypeChangeNeedsRecreate("PRODUCTION", "DEVELOPMENT"), "downgrade")
	assert.False(t, bigtableInstanceTypeChangeNeedsRecreate("DEVELOPMENT", "PRODUCTION"), "upgrade")
}

func TestBigtableClusterID_Validated(t *testing.T) {
	validate := gcpBigtableClusterSchema()["cluster_id"].ValidateFunc
	for _, v := range []string{"abcde", "1abcdef", "abcdef-", "abc/def", "abc?def", "abcdefghij-abcdefghij-abcdefghi"} {
		_, errs := validate(v, "cluster_id")
		assert.NotEmpty(t, errs, v)
	}
	for _, v := range []string{"abcdef", "my-cluster-c1", "abcdefghij-abcdefghij-abcdefgh"} {
		_, errs := validate(v, "cluster_id")
		assert.Empty(t, errs, v)
	}
}
