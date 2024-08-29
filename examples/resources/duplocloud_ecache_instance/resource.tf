resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_ecache_instance" "mycache" {
  tenant_id                  = duplocloud_tenant.myapp.tenant_id
  name                       = "mycache"
  cache_type                 = 0 // Redis
  replicas                   = 1
  size                       = "cache.t2.small"
  enable_cluster_mode        = true  // applicable only for redis
  number_of_shards           = 1     // applicable only for redis
  automatic_failover_enabled = false // enable auto failover
  engine_version             = var.engine_version

  log_delivery_configuration {
    log_group        = "/elasticache/redis"
    destination_type = "cloudwatch-logs"
    log_format       = "text"
    log_type         = "slow-log"
  }

  log_delivery_configuration {
    log_group        = "/elasticache/redis"
    destination_type = "cloudwatch-logs"
    log_format       = "json"
    log_type         = "engine-log"
  }

}
