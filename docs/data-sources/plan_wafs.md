---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_plan_wafs Data Source - terraform-provider-duplocloud"
subcategory: ""
description: |-
  duplocloud_plans retrieves a list of plans from Duplo.
---

# duplocloud_plan_wafs (Data Source)

`duplocloud_plans` retrieves a list of plans from Duplo.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `plan_id` (String) The ID of the plan for waf.

### Read-Only

- `data` (List of Object) (see [below for nested schema](#nestedatt--data))
- `id` (String) The ID of this resource.

<a id="nestedatt--data"></a>
### Nested Schema for `data`

Read-Only:

- `dashboard_url` (String)
- `waf_arn` (String)
- `waf_name` (String)