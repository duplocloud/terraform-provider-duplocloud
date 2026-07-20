package duplocloud

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
)

// buildAsgRawConfig wraps an instances_distribution attribute map into the
// nested cty shape GetRawConfig() produces:
// { mixed_instances_policy = [ { instances_distribution = [ { <distAttrs> } ] } ] }
func buildAsgRawConfig(distAttrs map[string]cty.Value) cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"mixed_instances_policy": cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"instances_distribution": cty.ListVal([]cty.Value{
					cty.ObjectVal(distAttrs),
				}),
			}),
		}),
	})
}

func TestAsgConfiguredOnDemandPercentage_Omitted_NotExplicit(t *testing.T) {
	raw := buildAsgRawConfig(map[string]cty.Value{
		"on_demand_percentage_above_base_capacity": cty.NullVal(cty.Number),
	})

	value, explicit := asgConfiguredOnDemandPercentage(raw)

	assert.False(t, explicit)
	assert.Equal(t, 0, value)
}

func TestAsgConfiguredOnDemandPercentage_ExplicitZero_IsExplicit(t *testing.T) {
	raw := buildAsgRawConfig(map[string]cty.Value{
		"on_demand_percentage_above_base_capacity": cty.NumberIntVal(0),
	})

	value, explicit := asgConfiguredOnDemandPercentage(raw)

	assert.True(t, explicit)
	assert.Equal(t, 0, value)
}

func TestAsgConfiguredOnDemandPercentage_ExplicitNonZero_IsExplicit(t *testing.T) {
	raw := buildAsgRawConfig(map[string]cty.Value{
		"on_demand_percentage_above_base_capacity": cty.NumberIntVal(40),
	})

	value, explicit := asgConfiguredOnDemandPercentage(raw)

	assert.True(t, explicit)
	assert.Equal(t, 40, value)
}

func TestAsgConfiguredOnDemandPercentage_NoMixedInstancesPolicy_NotExplicit(t *testing.T) {
	raw := cty.ObjectVal(map[string]cty.Value{
		"mixed_instances_policy": cty.NullVal(cty.List(cty.EmptyObject)),
	})

	value, explicit := asgConfiguredOnDemandPercentage(raw)

	assert.False(t, explicit)
	assert.Equal(t, 0, value)
}

func TestAsgConfiguredOnDemandPercentage_NullRaw_NotExplicit(t *testing.T) {
	value, explicit := asgConfiguredOnDemandPercentage(cty.NullVal(cty.EmptyObject))

	assert.False(t, explicit)
	assert.Equal(t, 0, value)
}
