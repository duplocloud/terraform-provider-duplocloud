---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_tenant Data Source - terraform-provider-duplocloud"
subcategory: ""
description: |-
  
---

# duplocloud_tenant (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `name` (String)

### Read-Only

- `id` (String) The ID of this resource.
- `infra_owner` (String)
- `plan_id` (String)
- `policy` (List of Object) (see [below for nested schema](#nestedatt--policy))
- `tags` (List of Object) (see [below for nested schema](#nestedatt--tags))

<a id="nestedatt--policy"></a>
### Nested Schema for `policy`

Read-Only:

- `allow_volume_mapping` (Boolean)
- `block_external_ep` (Boolean)


<a id="nestedatt--tags"></a>
### Nested Schema for `tags`

Read-Only:

- `key` (String)
- `value` (String)
