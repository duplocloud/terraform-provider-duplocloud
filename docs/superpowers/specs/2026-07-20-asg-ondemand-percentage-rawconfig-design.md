# DUPLO-43369 (Terraform): Faithful on_demand_percentage_above_base_capacity via GetRawConfig

## Context

This supersedes the guard-only fix currently committed on branch `DUPLO-43369`
(`docs/superpowers/plans/2026-07-17-asg-ondemand-percentage-guard-fix-plan.md`, commit adding
`&& v.(int) > 0` to the expand). That guard stopped the provider from silently force-pushing a
spurious `0` for an omitted attribute — a real safety fix — but it does so by treating **every**
`0` as "unset," which means a Terraform user who *deliberately* writes
`on_demand_percentage_above_base_capacity = 0` (wanting 100% spot above base capacity) has their
intent silently discarded. This spec replaces the guard with a faithful implementation that
distinguishes "omitted" from "explicitly set" using `GetRawConfig()`.

Companion backend work (separate repo, `duplocloud-internal/duplo`, already implemented on its own
`DUPLO-43369` branch) added a nullable `OnDemandPercentageAboveBaseCapacityExplicit` flag on
`AsgInstancesDistribution` and gated the backend's `0 → 100` coercion on it: a bare `0` (no flag)
is coerced to `100` for backward compatibility; a `0` with the flag `true` is honored as a literal
`0`. This Terraform spec is the client side that sends that flag.

### The bug being fixed, precisely

1. **Silent discard of explicit `0`:** the committed guard drops any `0`, so an explicit
   `on_demand_percentage_above_base_capacity = 0` never reaches AWS as `0` — the user cannot get
   100% spot through this attribute.
2. **State drift (confirmed empirically):** `on_demand_percentage_above_base_capacity` is
   `Optional`, not `Computed`, with no diff suppression (`resource_duplo_asg_profile.go:309-313`).
   The backend converges the stored value to `100` (the coercion result) within one reconcile
   cycle (~minutes after apply). The provider's flatten unconditionally writes that `100` into
   state (`resource_duplo_asg_profile.go:1024-1025`). On the next `terraform plan`, config (omitted
   → `0`) vs state (`100`) renders a perpetual `100 -> 0` diff that never converges.

### Why `Computed` is NOT the fix (validated during design)

Adding `Computed: true` to the attribute was tested and rejected: the plan changed from
`100 -> 0` to `100 -> null` but still drifted. `Computed` on an attribute nested inside a
config-known `TypeList`/`MaxItems:1` block does not carry the prior state value forward — in a
block that has any sibling attribute set, an omitted `Computed` nested attribute resolves to
`null` rather than to prior state. (`on_demand_allocation_strategy` at line 302 already carries
`Computed` and appears clean only because the backend returns no meaningful value for it to
preserve — it is not evidence that `Computed` works here.)

## Design

### Scope

`on_demand_percentage_above_base_capacity` only. The sibling fields in the same block
(`spot_instance_pools`, `spot_max_price`, `on_demand_base_capacity`) share the same
Optional-not-Computed drift shape, but only `on_demand_percentage` carries the spot-flip safety
stakes; their (cosmetic) drift is a separately-tracked follow-up, out of scope here.

### 1. SDK model (`duplosdk/asg.go`)

Add to `DuploAsgInstancesDistribution` (currently at `duplosdk/asg.go:78-85`):

```go
OnDemandPercentageAboveBaseCapacityExplicit *bool `json:"OnDemandPercentageAboveBaseCapacityExplicit,omitempty"`
```

Mirrors the backend DTO. `*bool` + `omitempty` so it is absent from the wire unless the provider
sets it — an old provider version simply never emits it, which is what the backend's
backward-compatible coercion relies on.

### 2. Expand (`expandAsgMixedInstancesPolicy`)

Replace the current guard block (`resource_duplo_asg_profile.go:946-949`):

```go
if v, ok := distMap["on_demand_percentage_above_base_capacity"]; ok && v.(int) > 0 {
    val := v.(int)
    distribution.OnDemandPercentageAboveBaseCapacity = &val
}
```

with `GetRawConfig`-driven logic that distinguishes omitted from explicit. The decision must be
derived from the **raw config** (what the user literally wrote), not `distMap`/`d.Get` (which
cannot tell `0` from unset). Behavior:

- attribute **absent / null** in raw config → leave `OnDemandPercentageAboveBaseCapacity` nil and
  do **not** set the `Explicit` flag → backend leaves the value alone.
- attribute **explicitly set** (any value, including `0`) → set
  `OnDemandPercentageAboveBaseCapacity = &val` **and** `OnDemandPercentageAboveBaseCapacityExplicit
  = &true` → backend honors the literal value.

Follow the in-repo `GetRawConfig` idiom (`resource_duplo_s3_bucket.go`,
`resource_duplo_ecache_instance.go`): read the `cty.Value`, guard with `IsKnown()` and `IsNull()`.
The attribute is nested (`mixed_instances_policy[0].instances_distribution[0]
.on_demand_percentage_above_base_capacity`); navigate the cty structure (list → index 0 → nested
attr), mirroring the nested-block navigation `resource_duplo_s3_bucket.go` already does.

### 3. Drift suppression: `DiffSuppressFunc` (not `Computed`)

Attach a `DiffSuppressFunc` to `on_demand_percentage_above_base_capacity` that consults
`d.GetRawConfig()` and returns `true` (suppress) **only when the attribute is absent from config**.
This suppresses the `100`-in-state vs omitted-config diff, while still letting an explicit change
(e.g. `40 -> 60`) or an explicit `0` produce a real diff and apply. Keep the attribute `Optional`
and **not** `Computed`.

Fallback if `DiffSuppressFunc` misbehaves on the nested path (a known SDKv2 finicky area): a
`CustomizeDiff` using `diff.GetRawConfig()` + `diff.SetNew(...)` to reset the planned value to
prior state on omission (the mechanism `resource_duplo_ecache_instance.go:938-1004` uses). Prefer
`DiffSuppressFunc` for its attribute-scoped simplicity; escalate to `CustomizeDiff` only if live
validation shows the suppress func does not hold.

### 4. Coexistence (unchanged — why the backend flag exists)

| Wire | Sender | Backend behavior |
|---|---|---|
| `0`, no `Explicit` flag | old provider (any config) | coerce → `100` (safe) |
| value absent / `null` | new provider, attribute omitted | leave alone |
| `0` + `Explicit=true` | new provider, attribute explicitly `0` | honor → literal `0` (spot) |

Omission never sets `Explicit=true`, so it can never trigger the on-demand→spot flip. The flip is
reachable only by a deliberate explicit `0`, which is visible in `terraform plan`.

## Testing

### Unit tests — design around the `GetRawConfig` gotcha

`schema.TestResourceDataRaw` does **not** populate `GetRawConfig`, so expand/suppress logic that
reads raw config cannot be exercised through it directly. Mitigation (following the
`s3ConfiguredKmsKeyId(cty.Value) (string, bool)` precedent in `resource_duplo_s3_bucket.go`):

- Factor the decision — "is this attribute explicitly set in config, and what is its value" — into
  a **pure helper that takes a `cty.Value`** and returns e.g. `(value int, explicit bool)`.
- Unit-test that helper directly with hand-constructed `cty.Value` inputs: absent, explicit `0`,
  explicit non-zero. This gives real coverage of the branching logic without needing
  `ResourceData`.
- Keep the `ResourceData`/expand glue that calls the helper thin.

### Live validation gate (REQUIRED before merge)

This is a hard gate, not optional — the nested-block diff behavior has already proven it lies to
static reasoning. Must be run against a live Duplo tenant (needs credentials; run by a human or
CI, not achievable in the sandbox). Sequence:

1. `apply` an ASG config that **omits** `on_demand_percentage_above_base_capacity`.
2. Wait for the backend reconcile cycle (~minutes).
3. `terraform plan` → **must be clean** (no `100 -> 0`/`-> null` drift).
4. `apply` a config that sets `on_demand_percentage_above_base_capacity = 0` explicitly.
5. Verify in AWS the ASG is actually 100% spot (0% on-demand above base).
6. `terraform plan` again → **must be clean**.
7. Confirm an existing ASG on an **old** provider version is unaffected (still coerced to `100`, no
   flip) when the backend fix is deployed.

## Out of Scope / Deferred

- Sibling-field drift (`spot_instance_pools`, `spot_max_price`, `on_demand_base_capacity`) — same
  drift shape, cosmetic only, separate follow-up.
- Angular UI change to send the `Explicit` flag — separate repo, separate ticket.
- Backend flag + coercion gating — already implemented on the backend's `DUPLO-43369` branch.
