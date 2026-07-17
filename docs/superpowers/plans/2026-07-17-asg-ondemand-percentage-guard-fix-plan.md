# ASG On-Demand Percentage Guard Fix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stop `expandAsgMixedInstancesPolicy` from sending a spurious literal `0` for `on_demand_percentage_above_base_capacity` on every ASG that doesn't explicitly configure it, matching the guard its three sibling fields already have.

**Architecture:** One-line fix: add the missing `&& v.(int) > 0` guard to the existing `if v, ok := distMap["on_demand_percentage_above_base_capacity"]; ok` check in `resource_duplo_asg_profile.go`, exactly matching `on_demand_base_capacity` and `spot_instance_pools` (same function) and `spot_max_price_percentage_over_lowest_price` (sibling function). This is the companion fix for DUPLO-43369 in the `duplo` backend repo — see that repo's `docs/superpowers/plans/2026-07-17-asg-ondemand-percentage-explicit-flag-backend-plan.md` and `docs/superpowers/specs/2026-07-17-asg-ondemand-percentage-explicit-flag-design.md`.

**Tech Stack:** Go 1.24 (module requires 1.24.0; this environment has 1.25.7 installed), `terraform-plugin-sdk/v2`, standard library `testing` + `github.com/stretchr/testify/assert`.

## Global Constraints

- Scope: this one field only, this one guard. Do **not** attempt the `GetRawConfig()`-based full parity fix (letting Terraform users genuinely express `0`) — that was explicitly deferred to a separate, larger effort during design.
- Do **not** send any new flag/attribute alongside this fix. The backend's new `OnDemandPercentageAboveBaseCapacityExplicit` field (see the `duplo` repo plan) is not touched by this provider fix at all.
- This environment CAN build and run this repo's Go tests directly (`go build ./duplocloud/...` and `go test ./duplocloud/... -run <name> -v` both work here, confirmed) — unlike the backend's `duplo` repo, do not skip real test execution for this plan.
- Branch `DUPLO-43369` already exists in this repo (created off `develop` at `a6eaf0da`, per this session's established branch-before-changes convention) — work on it, don't create a second branch of the same name.

---

## File Structure

| File | Change |
|---|---|
| `duplocloud/resource_duplo_asg_profile.go` | Add `&& v.(int) > 0` to the `on_demand_percentage_above_base_capacity` guard at (currently) line 946 |
| `duplocloud/resource_duplo_asg_profile_test.go` | New file — unit tests for `expandAsgMixedInstancesPolicy`'s handling of this field |

No other files change. `duplosdk.DuploAsgInstancesDistribution.OnDemandPercentageAboveBaseCapacity` is already `*int` with `json:"OnDemandPercentageAboveBaseCapacity,omitempty"` (`duplosdk/asg.go:81`) — the SDK layer already supports this correctly; only the provider's expand guard needs fixing.

---

### Task 1: Add the missing guard and verify with real tests

**Files:**
- Create branch: `DUPLO-43369` (from `develop`)
- Modify: `duplocloud/resource_duplo_asg_profile.go` (currently line 946, inside `expandAsgMixedInstancesPolicy`)
- Create: `duplocloud/resource_duplo_asg_profile_test.go`

**Interfaces:**
- Consumes: `expandAsgMixedInstancesPolicy(d *schema.ResourceData) *duplosdk.DuploAsgMixedInstancesPolicy` (existing, unchanged signature), `autoscalingGroupSchema() map[string]*schema.Schema` (existing, used to build test `ResourceData` via `schema.TestResourceDataRaw`).
- Produces: nothing new consumed elsewhere — this is a self-contained bugfix with no new public surface.

- [ ] **Step 1: Confirm the branch**

The `DUPLO-43369` branch was already created off `develop` (`a6eaf0da`) while writing this plan. Confirm you're on it before proceeding:

```bash
cd ~/Desktop/duplocloud/terraform-provider-duplocloud
git branch --show-current
```
Expected: `DUPLO-43369`. If it's not, run `git checkout DUPLO-43369` (do not create a second branch of the same name).

- [ ] **Step 2: Write the failing tests**

Create `duplocloud/resource_duplo_asg_profile_test.go`:

```go
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
```

- [ ] **Step 3: Run tests to verify the omitted/explicit-zero cases fail**

```bash
go test ./duplocloud/... -run TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacity -v
```
Expected: `TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacityOmitted_NotSent` and `..._ExplicitZero_NotSent` FAIL — `assert.Nil` fails because today's code (missing guard) sets `OnDemandPercentageAboveBaseCapacity` to a non-nil `*int` pointing to `0` in both cases (confirmed empirically while designing this plan: both cases currently produce a non-nil pointer with value `0`). `..._NonZero_Sent` PASSES already (unaffected by the bug).

- [ ] **Step 4: Add the guard**

In `duplocloud/resource_duplo_asg_profile.go`, inside `expandAsgMixedInstancesPolicy`, change:

```go
		if v, ok := distMap["on_demand_percentage_above_base_capacity"]; ok {
			val := v.(int)
			distribution.OnDemandPercentageAboveBaseCapacity = &val
		}
```
to:
```go
		if v, ok := distMap["on_demand_percentage_above_base_capacity"]; ok && v.(int) > 0 {
			val := v.(int)
			distribution.OnDemandPercentageAboveBaseCapacity = &val
		}
```

This now matches the two adjacent guards in the same function (`on_demand_base_capacity` immediately above it, `spot_instance_pools` immediately below it) exactly.

- [ ] **Step 5: Run tests to verify they all pass**

```bash
go test ./duplocloud/... -run TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacity -v
```
Expected:
```
=== RUN   TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacityOmitted_NotSent
--- PASS: TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacityOmitted_NotSent
=== RUN   TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacityExplicitZero_NotSent
--- PASS: TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacityExplicitZero_NotSent
=== RUN   TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacityNonZero_Sent
--- PASS: TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacityNonZero_Sent
PASS
```

- [ ] **Step 6: Run the full package test suite once to confirm no regressions**

```bash
go build ./duplocloud/...
go test ./duplocloud/... -short -v 2>&1 | tail -100
```
Expected: `go build` produces no output (success). For `go test`, expect the new tests plus any other short/unit-style tests to pass; acceptance tests gated behind `TF_ACC` (the SDK v2 convention — check for `resource.Test(` calls guarding on `os.Getenv("TF_ACC")` if any failures reference missing credentials) will skip rather than fail without real cloud credentials, which is expected and fine — do not attempt to provide credentials or force these to run.

- [ ] **Step 7: Commit**

```bash
git add duplocloud/resource_duplo_asg_profile.go duplocloud/resource_duplo_asg_profile_test.go
git commit -m "DUPLO-43369: Guard on_demand_percentage_above_base_capacity like its siblings

Matches the > 0 guard already used for on_demand_base_capacity and
spot_instance_pools in the same function, and
spot_max_price_percentage_over_lowest_price in the sibling function.
Without this, every ASG using instances_distribution without
explicitly setting this attribute sends a literal 0, which the Duplo
backend used to silently coerce to 100 -- masking this bug. Once the
backend stops coercing (see duplocloud-internal/duplo DUPLO-43369),
this would otherwise flip such ASGs from on-demand to 100% spot
capacity on the next terraform apply."
```

---

## Post-Plan Notes

- This branch is independent of and should merge on its own timeline from the `duplo` backend PR — the backend's new `OnDemandPercentageAboveBaseCapacityExplicit` flag defaults to legacy-safe behavior regardless of whether this provider fix has shipped yet, so there's no strict ordering requirement between the two, but shipping this is what actually closes the resiliency risk for TF-managed ASGs.
- Deferred, not part of this plan: using `d.GetRawConfig()` so Terraform users can genuinely express `on_demand_percentage_above_base_capacity = 0`, and sending the backend's new `OnDemandPercentageAboveBaseCapacityExplicit` flag. Track as a separate follow-up if/when that's prioritized.
