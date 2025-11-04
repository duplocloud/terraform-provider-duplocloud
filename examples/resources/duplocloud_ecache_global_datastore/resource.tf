resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


resource "duplocloud_ecache_instance" "re" {
  tenant_id                  = duplocloud_tenant.myapp.tenant_id
  name                       = "redp11"
  cache_type                 = 0
  replicas                   = 2
  size                       = "cache.r6g.large"
  engine_version             = "6.0"
  enable_cluster_mode        = true
  number_of_shards           = 2
  automatic_failover_enabled = true
}

resource "duplocloud_ecache_global_datastore" "gds" {
  tenant_id                     = duplocloud_tenant.myapp.tenant_id
  primary_instance_name         = duplocloud_ecache_instance.re.identifier
  global_replication_group_name = "ggrp11"
  description                   = "rgb global datastore"
}
