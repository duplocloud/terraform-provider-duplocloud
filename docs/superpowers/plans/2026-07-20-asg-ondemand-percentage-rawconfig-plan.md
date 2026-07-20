# ASG On-Demand Percentage GetRawConfig Fix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the Terraform provider honor an explicit `on_demand_percentage_above_base_capacity = 0` (100% spot) and stop drifting on an omitted value, by deriving intent from `GetRawConfig()` and sending the backend's `Explicit` flag — replacing the crude `> 0` guard that silently discards an explicit `0`.

**Architecture:** A pure helper reads the nested attribute from the raw config `cty.Value` and returns `(value, explicit)`, distinguishing "user wrote it (incl 0)" from "omitted." The expand uses it to send `{value, Explicit=true}` on an explicit value and `nil` on omission; a `DiffSuppressFunc` uses it to suppress the state-vs-omitted-config drift while still letting an explicit change apply. Keeps the attribute `Optional` (not `Computed`).

**Tech Stack:** Go 1.24 (module floor; sandbox has 1.25.7), `terraform-plugin-sdk/v2 v2.26.1`, `github.com/hashicorp/go-cty/cty`, `github.com/stretchr/testify/assert`.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-07-20-asg-ondemand-percentage-rawconfig-design.md` — read first.
- Scope: `on_demand_percentage_above_base_capacity` ONLY. Do not touch sibling fields (`spot_instance_pools`, `spot_max_price`, `on_demand_base_capacity`) — their drift is a separate follow-up.
- Do NOT add `Computed` to the attribute — it was tested and rejected (does not carry state forward in a config-known nested block). Keep it `Optional`, add a `DiffSuppressFunc`.
- Branch: work on `DUPLO-43369` (already checked out, guard fix committed at `01f9ef4e`). Confirm with `git branch --show-current` before starting.
- `cty` import path is `github.com/hashicorp/go-cty/cty` (matches `resource_duplo_s3_bucket.go:13`).
- This environment CAN build and run this repo's Go unit tests (`go build ./duplocloud/...`, `go test ./duplocloud/... -run <name> -v`, `go vet ./duplocloud/...`) — run them for real. Acceptance tests (`TF_ACC`) and the live-tenant validation gate need Duplo credentials the sandbox lacks; those are run by a human/CI, not here.
- Known testing gotcha: `schema.TestResourceDataRaw` does NOT populate `GetRawConfig`, so expand/suppress logic that reads raw config CANNOT be exercised through it. Unit coverage lives in the pure helper's tests (Task 1); full round-trip coverage is the live gate (Task 2 verification).

---

## File Structure

| File | Change |
|---|---|
| `duplosdk/asg.go` | Add `OnDemandPercentageAboveBaseCapacityExplicit *bool` to `DuploAsgInstancesDistribution` |
| `duplocloud/resource_duplo_asg_profile.go` | Add pure helper `asgConfiguredOnDemandPercentage`; rewrite expand block to use it; add `DiffSuppressFunc` `suppressOmittedOnDemandPercentage` + wire it into the schema attribute |
| `duplocloud/resource_duplo_asg_profile_helpers_test.go` | New — unit tests for the pure helper (constructed `cty.Value` inputs) |
| `duplocloud/resource_duplo_asg_profile_test.go` | Remove the 3 stale `TestResourceDataRaw`-based expand tests (they cannot exercise `GetRawConfig` and the non-zero one would now fail) |

---

### Task 1: SDK flag field + pure raw-config helper (fully unit-tested)

**Files:**
- Modify: `duplosdk/asg.go:78-85` (`DuploAsgInstancesDistribution`)
- Modify: `duplocloud/resource_duplo_asg_profile.go` (add helper near the other ASG helpers, e.g. after `expandAsgMixedInstancesPolicy`)
- Create: `duplocloud/resource_duplo_asg_profile_helpers_test.go`

**Interfaces:**
- Produces: `asgConfiguredOnDemandPercentage(raw cty.Value) (value int, explicit bool)` — `explicit` is false when the attribute is absent/null/unknown in raw config; true when the user wrote any value (including `0`), with `value` the written int. Task 2's expand and DiffSuppressFunc both consume this.
- Produces: `DuploAsgInstancesDistribution.OnDemandPercentageAboveBaseCapacityExplicit *bool` (JSON `OnDemandPercentageAboveBaseCapacityExplicit,omitempty`).

- [ ] **Step 1: Add the SDK flag field**

In `duplosdk/asg.go`, `DuploAsgInstancesDistribution` currently is:

```go
type DuploAsgInstancesDistribution struct {
	OnDemandAllocationStrategy          string `json:"OnDemandAllocationStrategy,omitempty"`
	OnDemandBaseCapacity                *int   `json:"OnDemandBaseCapacity,omitempty"`
	OnDemandPercentageAboveBaseCapacity *int   `json:"OnDemandPercentageAboveBaseCapacity,omitempty"`
	SpotAllocationStrategy              string `json:"SpotAllocationStrategy,omitempty"`
	SpotInstancePools                   *int   `json:"SpotInstancePools,omitempty"`
	SpotMaxPrice                        string `json:"SpotMaxPrice,omitempty"`
}
```

Add the flag directly after `OnDemandPercentageAboveBaseCapacity`:

```go
type DuploAsgInstancesDistribution struct {
	OnDemandAllocationStrategy          string `json:"OnDemandAllocationStrategy,omitempty"`
	OnDemandBaseCapacity                *int   `json:"OnDemandBaseCapacity,omitempty"`
	OnDemandPercentageAboveBaseCapacity *int   `json:"OnDemandPercentageAboveBaseCapacity,omitempty"`
	OnDemandPercentageAboveBaseCapacityExplicit *bool `json:"OnDemandPercentageAboveBaseCapacityExplicit,omitempty"`
	SpotAllocationStrategy              string `json:"SpotAllocationStrategy,omitempty"`
	SpotInstancePools                   *int   `json:"SpotInstancePools,omitempty"`
	SpotMaxPrice                        string `json:"SpotMaxPrice,omitempty"`
}
```

- [ ] **Step 2: Write the failing helper tests**

Create `duplocloud/resource_duplo_asg_profile_helpers_test.go`:

```go
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
```

- [ ] **Step 3: Run the helper tests to confirm they fail**

```bash
go test ./duplocloud/... -run TestAsgConfiguredOnDemandPercentage -v
```
Expected: compile failure — `asgConfiguredOnDemandPercentage` is undefined — until Step 4.

- [ ] **Step 4: Implement the pure helper**

In `duplocloud/resource_duplo_asg_profile.go`, add the following. This file does **not** currently import `cty`, so add `"github.com/hashicorp/go-cty/cty"` to its import block (alongside the existing `github.com/hashicorp/terraform-plugin-sdk/v2/...` imports):

```go
// asgConfiguredOnDemandPercentage extracts
// mixed_instances_policy[0].instances_distribution[0].on_demand_percentage_above_base_capacity
// from the raw config (via GetRawConfig; works with both *schema.ResourceData and
// *schema.ResourceDiff). explicit is false when the attribute is absent/null/unknown,
// which is how we tell "user omitted it" (leave the backend value alone) from "user
// wrote a value, including 0" (send it with the Explicit flag so the backend honors it
// instead of coercing 0 -> 100). Mirrors the s3ConfiguredKmsKeyId pattern.
func asgConfiguredOnDemandPercentage(raw cty.Value) (value int, explicit bool) {
	if raw.IsNull() || !raw.IsKnown() {
		return 0, false
	}
	mip := raw.GetAttr("mixed_instances_policy")
	if mip.IsNull() || !mip.IsKnown() || mip.LengthInt() == 0 {
		return 0, false
	}
	mipBlock := mip.AsValueSlice()[0]
	if mipBlock.IsNull() || !mipBlock.IsKnown() {
		return 0, false
	}
	dist := mipBlock.GetAttr("instances_distribution")
	if dist.IsNull() || !dist.IsKnown() || dist.LengthInt() == 0 {
		return 0, false
	}
	distBlock := dist.AsValueSlice()[0]
	if distBlock.IsNull() || !distBlock.IsKnown() {
		return 0, false
	}
	pct := distBlock.GetAttr("on_demand_percentage_above_base_capacity")
	if pct.IsNull() || !pct.IsKnown() {
		return 0, false
	}
	i64, _ := pct.AsBigFloat().Int64()
	return int(i64), true
}
```

- [ ] **Step 5: Run the helper tests to confirm they pass**

```bash
go test ./duplocloud/... -run TestAsgConfiguredOnDemandPercentage -v
go build ./duplocloud/... ./duplosdk/...
```
Expected: all 5 helper tests PASS; build clean (no output). Hand-check each: omitted/null-attr → `(0,false)`; explicit `0` → `(0,true)`; explicit `40` → `(40,true)`; no `mixed_instances_policy` → `(0,false)`; null raw → `(0,false)`.

- [ ] **Step 6: Commit**

```bash
git add duplosdk/asg.go duplocloud/resource_duplo_asg_profile.go duplocloud/resource_duplo_asg_profile_helpers_test.go
git commit -m "DUPLO-43369: Add explicit-percentage SDK flag and raw-config helper"
```

---

### Task 2: Wire helper into expand + DiffSuppressFunc; remove stale expand tests

**Files:**
- Modify: `duplocloud/resource_duplo_asg_profile.go` — the expand block (currently lines 946-949), the schema attribute (lines 309-313), and add `suppressOmittedOnDemandPercentage`
- Modify: `duplocloud/resource_duplo_asg_profile_test.go` — remove the 3 stale expand tests

**Interfaces:**
- Consumes: `asgConfiguredOnDemandPercentage(raw cty.Value) (int, bool)` and `DuploAsgInstancesDistribution.OnDemandPercentageAboveBaseCapacityExplicit *bool` from Task 1.
- Consumes: the expand function `expandAsgMixedInstancesPolicy(d *schema.ResourceData)` reads `d.GetRawConfig()`.

- [ ] **Step 1: Rewrite the expand block to derive intent from raw config**

In `duplocloud/resource_duplo_asg_profile.go`, replace the guard block (currently lines 946-949):

```go
		if v, ok := distMap["on_demand_percentage_above_base_capacity"]; ok && v.(int) > 0 {
			val := v.(int)
			distribution.OnDemandPercentageAboveBaseCapacity = &val
		}
```

with logic driven by the raw config (not `distMap`, which cannot tell `0` from unset):

```go
		if pct, explicit := asgConfiguredOnDemandPercentage(d.GetRawConfig()); explicit {
			val := pct
			flag := true
			distribution.OnDemandPercentageAboveBaseCapacity = &val
			distribution.OnDemandPercentageAboveBaseCapacityExplicit = &flag
		}
```

Note: `expandAsgMixedInstancesPolicy`'s signature already takes `d *schema.ResourceData`, so `d.GetRawConfig()` is available here with no signature change. Leave the sibling guards (`on_demand_base_capacity`, `spot_instance_pools`) exactly as they are.

- [ ] **Step 2: Add the DiffSuppressFunc**

In the same file, add:

```go
// suppressOmittedOnDemandPercentage suppresses the diff for
// on_demand_percentage_above_base_capacity only when the user omitted it from config.
// The backend converges the stored value to 100 (its coercion default), which flatten
// writes into state; without this, an omitted config (0) perpetually diffs against
// state (100). An explicitly-written value (including 0) is NOT suppressed, so real
// changes still apply. Reads GetRawConfig so it sees what the user wrote, not state.
func suppressOmittedOnDemandPercentage(k, old, new string, d *schema.ResourceData) bool {
	_, explicit := asgConfiguredOnDemandPercentage(d.GetRawConfig())
	return !explicit
}
```

- [ ] **Step 3: Wire the DiffSuppressFunc into the schema attribute**

In `duplocloud/resource_duplo_asg_profile.go`, the attribute (lines 309-313) is:

```go
							"on_demand_percentage_above_base_capacity": {
								Description: "Percentage of On-Demand instances above the base capacity (0-100).",
								Type:        schema.TypeInt,
								Optional:    true,
							},
```

Change it to add the suppressor and document the `0` semantics (keep `Optional`, do NOT add `Computed`):

```go
							"on_demand_percentage_above_base_capacity": {
								Description:      "Percentage of On-Demand instances above the base capacity (0-100). Set explicitly to `0` for 100% Spot above base capacity. Omit to let DuploCloud manage it (defaults to 100% On-Demand).",
								Type:             schema.TypeInt,
								Optional:         true,
								DiffSuppressFunc: suppressOmittedOnDemandPercentage,
							},
```

- [ ] **Step 4: Remove the stale expand tests**

In `duplocloud/resource_duplo_asg_profile_test.go`, delete all three functions:
`TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacityOmitted_NotSent`,
`TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacityExplicitZero_NotSent`, and
`TestExpandAsgMixedInstancesPolicy_OnDemandPercentageAboveBaseCapacityNonZero_Sent`.

They drive `expandAsgMixedInstancesPolicy` through `schema.TestResourceDataRaw`, which does not populate `GetRawConfig`. After Step 1, expand derives everything from `d.GetRawConfig()` (null under that helper), so all three would exercise the null-config path: `..._Omitted_` and `..._ExplicitZero_` would pass only coincidentally, and `..._NonZero_Sent` would FAIL (expects `40`, gets `nil`). The real coverage moved to Task 1's helper tests (branching logic) and the live gate below (full round-trip). If removing all three empties the file of imports, drop the now-unused `import` lines too so it compiles.

- [ ] **Step 5: Build, vet, and run the unit suite**

```bash
go build ./duplocloud/... ./duplosdk/...
go vet ./duplocloud/...
go test ./duplocloud/... -run 'TestAsgConfiguredOnDemandPercentage|TestExpandAsgMixedInstancesPolicy' -v
```
Expected: build + vet clean; the 5 helper tests PASS; the 3 removed tests no longer exist (no failures, no references). A repo-wide `go test ./duplocloud/...` will still show the pre-existing `TF_ACC` acceptance tests failing for missing credentials (`Duplocloud Unable to create ... client`) — that is unrelated and expected in this environment; do not treat it as a regression (confirm the same failures exist on a clean `develop` if in doubt).

- [ ] **Step 6: Commit**

```bash
git add duplocloud/resource_duplo_asg_profile.go duplocloud/resource_duplo_asg_profile_test.go
git commit -m "DUPLO-43369: Honor explicit on-demand % 0 and suppress omitted-value drift"
```

- [ ] **Step 7: Live validation gate (REQUIRED before merge — run by a human/CI with tenant creds, NOT in the sandbox)**

Document these results on the PR. This is a hard gate: the nested-block diff behavior has already proven it lies to static reasoning, so a green build is not sufficient evidence.

1. `apply` an ASG config that OMITS `on_demand_percentage_above_base_capacity`; wait ~5 min for the backend reconcile cycle; `terraform plan` must be **clean** (no `100 -> 0` / `-> null`).
2. `apply` a config that sets `on_demand_percentage_above_base_capacity = 0`; verify in AWS the ASG is actually 100% Spot; `terraform plan` again must be **clean**.
3. `apply` a config that sets it to a non-zero value (e.g. `40`); verify AWS shows 40% on-demand; `plan` clean.
4. If step 1's `plan` is NOT clean, the `DiffSuppressFunc` did not hold on the nested path — fall back to the `CustomizeDiff` + `diff.SetNew(...)` mechanism (per the spec) using the same `asgConfiguredOnDemandPercentage` helper against `diff.GetRawConfig()`, and re-run this gate.

---

## Post-Plan Notes

- Per repo convention, run `superpowers:requesting-code-review` before opening/updating the PR.
- The `DiffSuppressFunc`-vs-`CustomizeDiff` choice is settled as DiffSuppressFunc-first only pending the live gate (Step 7.4). Everything else is unit-verifiable here.
- Companion backend work (the `Explicit` flag + gated coercion) is already implemented on the backend repo's `DUPLO-43369` branch; this provider change is what makes a TF user's explicit `0` actually reach AWS.
- Out of scope, tracked separately: sibling-field drift (`spot_instance_pools`, etc.); the Angular UI change.
