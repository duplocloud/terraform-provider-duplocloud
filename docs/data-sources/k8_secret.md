---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_k8_secret Data Source - terraform-provider-duplocloud"
subcategory: ""
description: |-
  
---

# duplocloud_k8_secret (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- **filter** (Block Set) (see [below for nested schema](#nestedblock--filter))
- **tenant_id** (String)

### Read-Only

- **data** (List of Object) (see [below for nested schema](#nestedatt--data))
- **id** (String) The ID of this resource.

<a id="nestedblock--filter"></a>
### Nested Schema for `filter`

Required:

- **name** (String)
- **value** (String)


<a id="nestedatt--data"></a>
### Nested Schema for `data`

Read-Only:

- **client_secret_version** (String)
- **secret_data** (String)
- **secret_name** (String)
- **secret_type** (String)
- **secret_version** (String)
- **tenant_id** (String)


