---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_tenant_secrets Data Source - terraform-provider-duplocloud"
subcategory: ""
description: |-
  
---

# duplocloud_tenant_secrets (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `tenant_id` (String)

### Read-Only

- `id` (String) The ID of this resource.
- `secrets` (List of Object) (see [below for nested schema](#nestedatt--secrets))

<a id="nestedatt--secrets"></a>
### Nested Schema for `secrets`

Read-Only:

- `arn` (String)
- `name` (String)
- `name_suffix` (String)
- `rotation_enabled` (Boolean)
- `tags` (List of Object) (see [below for nested schema](#nestedobjatt--secrets--tags))
- `tenant_id` (String)

<a id="nestedobjatt--secrets--tags"></a>
### Nested Schema for `secrets.tags`

Read-Only:

- `key` (String)
- `value` (String)
