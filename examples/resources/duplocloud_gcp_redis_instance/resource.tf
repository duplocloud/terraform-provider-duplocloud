resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_gcp_redis_instance" "redis-demo" {
  tenant_id      = duplocloud_tenant.myapp.tenant_id
  name           = "redis-demo"
  display_name   = "redis-demo"
  tier           = "BASIC"
  redis_version  = "REDIS_4_0"
  memory_size_gb = 1
}