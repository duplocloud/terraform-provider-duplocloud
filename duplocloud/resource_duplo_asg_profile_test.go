package duplocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacityOmitted_NotSent(t *testing.T) {
	d := schema.TestResourceDataRaw(t, autoscalingGroupSchema(), map[string]interface{}{
		"mixed_instances_policy": []interface{}{
			map[string]interface{}{
				"instances_distribution": []interface{}{
					map[string]interface{}{
						"on_demand_allocation_strategy": "lowest-price",
					},
				},
			},
		},
	})

	policy := expandAsgMixedInstancesPolicy(d)

	assert.Nil(t, policy.InstancesDistribution.OnDemandPercentageAboveBaseCapacity)
}

func TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacityExplicitZero_NotSent(t *testing.T) {
	// Terraform's schema always populates an unset nested-block int with its zero
	// value, so this is indistinguishable from the omitted case above without the
	// deferred GetRawConfig()-based fix. Confirms this field is now *consistent*
	// with its siblings (on_demand_base_capacity, spot_instance_pools), not that a
	// Terraform user can express literal 0 through this attribute today.
	d := schema.TestResourceDataRaw(t, autoscalingGroupSchema(), map[string]interface{}{
		"mixed_instances_policy": []interface{}{
			map[string]interface{}{
				"instances_distribution": []interface{}{
					map[string]interface{}{
						"on_demand_percentage_above_base_capacity": 0,
					},
				},
			},
		},
	})

	policy := expandAsgMixedInstancesPolicy(d)

	assert.Nil(t, policy.InstancesDistribution.OnDemandPercentageAboveBaseCapacity)
}

func TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacityNonZero_Sent(t *testing.T) {
	d := schema.TestResourceDataRaw(t, autoscalingGroupSchema(), map[string]interface{}{
		"mixed_instances_policy": []interface{}{
			map[string]interface{}{
				"instances_distribution": []interface{}{
					map[string]interface{}{
						"on_demand_percentage_above_base_capacity": 40,
					},
				},
			},
		},
	})

	policy := expandAsgMixedInstancesPolicy(d)

	assert.NotNil(t, policy.InstancesDistribution.OnDemandPercentageAboveBaseCapacity)
	assert.Equal(t, 40, *policy.InstancesDistribution.OnDemandPercentageAboveBaseCapacity)
}
