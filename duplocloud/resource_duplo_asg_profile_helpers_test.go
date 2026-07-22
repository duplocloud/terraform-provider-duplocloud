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

func TestAsgConfiguredOnDemandPercentage_Omitted_NotExplicitButKnown(t *testing.T) {
	raw := buildAsgRawConfig(map[string]cty.Value{
		"on_demand_percentage_above_base_capacity": cty.NullVal(cty.Number),
	})

	value, explicit, known := asgConfiguredOnDemandPercentage(raw)

	assert.False(t, explicit)
	assert.True(t, known) // a known omission — safe to suppress
	assert.Equal(t, 0, value)
}

func TestAsgConfiguredOnDemandPercentage_ExplicitZero_IsExplicit(t *testing.T) {
	raw := buildAsgRawConfig(map[string]cty.Value{
		"on_demand_percentage_above_base_capacity": cty.NumberIntVal(0),
	})

	value, explicit, known := asgConfiguredOnDemandPercentage(raw)

	assert.True(t, explicit)
	assert.True(t, known)
	assert.Equal(t, 0, value)
}

func TestAsgConfiguredOnDemandPercentage_ExplicitNonZero_IsExplicit(t *testing.T) {
	raw := buildAsgRawConfig(map[string]cty.Value{
		"on_demand_percentage_above_base_capacity": cty.NumberIntVal(40),
	})

	value, explicit, known := asgConfiguredOnDemandPercentage(raw)

	assert.True(t, explicit)
	assert.True(t, known)
	assert.Equal(t, 40, value)
}

// The regression this fix targets: a value unknown at plan time (e.g. driven by a
// variable or another resource's computed output) must be reported not-explicit AND
// not-known, so the caller does NOT suppress its diff.
func TestAsgConfiguredOnDemandPercentage_UnknownValue_NotExplicitNotKnown(t *testing.T) {
	raw := buildAsgRawConfig(map[string]cty.Value{
		"on_demand_percentage_above_base_capacity": cty.UnknownVal(cty.Number),
	})

	value, explicit, known := asgConfiguredOnDemandPercentage(raw)

	assert.False(t, explicit)
	assert.False(t, known) // unknown, not omitted — must not be suppressed
	assert.Equal(t, 0, value)
}

func TestAsgConfiguredOnDemandPercentage_NoMixedInstancesPolicy_NotExplicitButKnown(t *testing.T) {
	raw := cty.ObjectVal(map[string]cty.Value{
		"mixed_instances_policy": cty.NullVal(cty.List(cty.EmptyObject)),
	})

	value, explicit, known := asgConfiguredOnDemandPercentage(raw)

	assert.False(t, explicit)
	assert.True(t, known) // block absent => attribute known-omitted
	assert.Equal(t, 0, value)
}

func TestAsgConfiguredOnDemandPercentage_UnknownMixedInstancesPolicy_NotKnown(t *testing.T) {
	raw := cty.ObjectVal(map[string]cty.Value{
		"mixed_instances_policy": cty.UnknownVal(cty.List(cty.EmptyObject)),
	})

	value, explicit, known := asgConfiguredOnDemandPercentage(raw)

	assert.False(t, explicit)
	assert.False(t, known) // whole block pending => can't decide, must not suppress
	assert.Equal(t, 0, value)
}

func TestAsgConfiguredOnDemandPercentage_NullRaw_NotExplicitButKnown(t *testing.T) {
	value, explicit, known := asgConfiguredOnDemandPercentage(cty.NullVal(cty.EmptyObject))

	assert.False(t, explicit)
	assert.True(t, known)
	assert.Equal(t, 0, value)
}
