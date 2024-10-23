resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Simple Example 1:  Grant access of tenant to user
resource "duplocloud_user_tenant_access" "acc" {
  username    = "user@domain"
  tenant_id   = duplocloud_tenant.myapp.tenant_id
  is_readonly = false
}
