resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_redis_cache" "myCache" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  name                = "cache-test"
  capacity            = 1
  family              = "p"
  sku_name            = "Premium"
  subnet_id           = "/subscriptions/0c84b91e-95f5-409e-9cff-6c2e60affbb3/resourceGroups/duploinfra-demo/providers/Microsoft.Network/virtualNetworks/demo/subnets/duploinfra-default"
  enable_non_ssl_port = false
  shard_count         = 1
}
