data "duplocloud_gcp_node_pool" "pool" {
  tenant_id = "tenantid"
  name      = "nodepool-name"
}

output "nodepool_output" {
  value = {
    name             = data.duplocloud_gcp_node_pool.pool.name
    machine_type     = data.duplocloud_gcp_node_pool.pool.machine_type
    zones            = data.duplocloud_gcp_node_pool.pool.zones
    disc_size_gb     = data.duplocloud_gcp_node_pool.pool.disc_size_gb
    disc_type        = data.duplocloud_gcp_node_pool.pool.disc_type
    upgrade_settings = data.duplocloud_gcp_node_pool.pool.upgrade_settings

  }
}
 