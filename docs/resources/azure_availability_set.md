---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_azure_availability_set Resource - terraform-provider-duplocloud"
subcategory: ""
description: |-
  duplocloud_azure_availability_set manages logical groupings of VMs that enhance reliability by placing VMs in different fault domains to minimize correlated failures, offering improved VM-to-VM latency and high availability, with no extra cost beyond the VM instances themselves, though they may still be affected by shared infrastructure failures.
---

# duplocloud_azure_availability_set (Resource)

`duplocloud_azure_availability_set` manages logical groupings of VMs that enhance reliability by placing VMs in different fault domains to minimize correlated failures, offering improved VM-to-VM latency and high availability, with no extra cost beyond the VM instances themselves, though they may still be affected by shared infrastructure failures.

## Example Usage

```terraform
resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_availability_set" "st" {
  tenant_id                    = duplocloud_tenant.myapp.tenant_id
  name                         = "availset"
  use_managed_disk             = "Aligned"
  platform_update_domain_count = 5
  platform_fault_domain_count  = 2
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name for availability set
- `tenant_id` (String) The GUID of the tenant that the host will be created in.

### Optional

- `platform_fault_domain_count` (Number) Specify platform fault domain count betweem 1-3, for availability set. Virtual machines in the same fault domain share a common power source and physical network switch. Defaults to `2`.
- `platform_update_domain_count` (Number) Specify platform update domain count between 1-20, for availability set. Virtual machines in the same update domain will be restarted together during planned maintenance. Azure never restarts more than one update domain at a time. Defaults to `5`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `use_managed_disk` (String) Set this to `Aligned` if you plan to create virtual machines in this availability set with managed disks. Defaults to `Classic`.

### Read-Only

- `availability_set_id` (String)
- `id` (String) The ID of this resource.
- `location` (String)
- `tags` (Map of String)
- `type` (String)
- `virtual_machines` (List of Object) (see [below for nested schema](#nestedatt--virtual_machines))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)


<a id="nestedatt--virtual_machines"></a>
### Nested Schema for `virtual_machines`

Read-Only:

- `id` (String)

## Import

Import is supported using the following syntax:

```shell
# Example: Importing an existing Azure Availablitu set
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the  name of the Azure Availability set
#
terraform import duplocloud_azure_availability_set.this *TENANT_ID*/availability-set/*NAME*
```