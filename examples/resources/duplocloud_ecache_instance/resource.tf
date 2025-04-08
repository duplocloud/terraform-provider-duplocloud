resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}
variable "engine_version" {
  type = string
}

resource "duplocloud_ecache_instance" "mycache" {
  tenant_id                  = duplocloud_tenant.myapp.tenant_id
  name                       = "mycache"
  cache_type                 = 0 // Redis
  replicas                   = 1
  size                       = "cache.t2.small"
  enable_cluster_mode        = true  // applicable only for redis
  number_of_shards           = 1     // applicable only for redis
  automatic_failover_enabled = false // enable auto failover, set replicas to 2 or more
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


resource "duplocloud_ecache_instance" "mycaches" {
  tenant_id                = duplocloud_tenant.myapp.tenant_id
  name                     = "myvalkey"
  cache_type               = 2 //valkey
  size                     = "cache.t3.medium"
  engine_version           = "7.2"
  snapshot_window          = "19:50-20:51"
  snapshot_retention_limit = 12
}

resource "duplocloud_ecache_instance" "mycachesnap" {
  tenant_id      = duplocloud_tenant.myapp.tenant_id
  name           = "fromsnap"
  cache_type     = 2
  size           = "cache.t3.medium"
  engine_version = var.engine_version //"7.2"
  snapshot_name  = "duploservices-march13-mysnap-snapshot"

}
// Example: cluster mode example
resource "duplocloud_ecache_instance" "mycaches" {
  tenant_id                  = duplocloud_tenant.myapp.tenant_id
  name                       = "tf-clust1"
  cache_type                 = 2
  size                       = "cache.t3.medium"
  engine_version             = var.engine_version //"7.2"
  snapshot_window            = "19:50-20:51"
  snapshot_retention_limit   = 12
  enable_cluster_mode        = true
  number_of_shards           = 3
  automatic_failover_enabled = true
  replicas                   = 2
}
