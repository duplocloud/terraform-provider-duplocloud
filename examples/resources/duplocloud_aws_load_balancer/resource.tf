resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_load_balancer" "myapp" {
  tenant_id            = duplocloud_tenant.myapp.tenant_id
  name                 = "myapp"
  is_internal          = true
  enable_access_logs   = true
  drop_invalid_headers = true
}
