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
