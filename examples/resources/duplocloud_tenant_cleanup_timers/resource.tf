resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "non-default-tenant"
}

# Create or Update a tenant clean up timer
resource "duplocloud_tenant_cleanup_timers" "mytimers" {
  tenant_id   = duplocloud_tenant.myapp.tenant_id
  expiry_time = "2024-10-10T00:00:00Z"
  pause_time  = "2024-10-11T00:00:00Z"
}

# Clear expiry timer
resource "duplocloud_tenant_cleanup_timers" "mytimers" {
  tenant_id          = duplocloud_tenant.myapp.tenant_id
  remove_expiry_time = true
}

# Clear pause timer
resource "duplocloud_tenant_cleanup_timers" "mytimers" {
  tenant_id         = duplocloud_tenant.myapp.tenant_id
  remove_pause_time = true
}