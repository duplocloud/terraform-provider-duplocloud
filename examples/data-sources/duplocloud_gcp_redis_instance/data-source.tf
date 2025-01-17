data "duplocloud_gcp_redis_instance" "app" {
  tenant_id = "tenant_id"
  name      = "name"
}

output "out" {
  value = {
    name                       = data.duplocloud_gcp_redis_instance.app.name
    display_name               = data.duplocloud_gcp_redis_instance.app.display_name
    memory_size_gb             = data.duplocloud_gcp_redis_instance.app.memory_size_gb
    read_replicas_enabled      = data.duplocloud_gcp_redis_instance.app.read_replicas_enabled
    redis_configs              = data.duplocloud_gcp_redis_instance.app.redis_configs
    redis_version              = data.duplocloud_gcp_redis_instance.app.redis_version
    replica_count              = data.duplocloud_gcp_redis_instance.app.replica_count
    auth_enabled               = data.duplocloud_gcp_redis_instance.app.auth_enabled
    transit_encryption_enabled = data.duplocloud_gcp_redis_instance.app.transit_encryption_enabled
    tier                       = data.duplocloud_gcp_redis_instance.app.tier
    labels                     = data.duplocloud_gcp_redis_instance.app.labels
  }
}

