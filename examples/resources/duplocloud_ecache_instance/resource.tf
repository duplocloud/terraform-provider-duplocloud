resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_ecache_instance" "mycache" {
  tenant_id  = duplocloud_tenant.myapp.tenant_id
  name       = "mycache"
  cache_type = 0 // Redis
  replicas   = 1
  size       = "cache.t2.small"
}
