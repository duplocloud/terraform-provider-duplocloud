---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_tenant_internal_subnets Data Source - terraform-provider-duplocloud"
subcategory: ""
description: |-
  The duplocloud_tenant_internal_subnets data source retrieves a list of tenant's internal subnet IDs.
---

# duplocloud_tenant_internal_subnets (Data Source)

The `duplocloud_tenant_internal_subnets` data source retrieves a list of tenant's internal subnet IDs.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `tenant_id` (String) The GUID of the tenant.

### Read-Only

- `id` (String) The ID of this resource.
- `subnet_ids` (List of String) The list of subnet IDs.
