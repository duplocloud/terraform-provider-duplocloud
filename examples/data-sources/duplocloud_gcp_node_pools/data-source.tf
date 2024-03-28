data "duplocloud_gcp_node_pools" "pool" {
  tenant_id = "tenantid"
}

output "nodepool_output" {
  value = {
    node_pools = data.duplocloud_gcp_node_pools.pool.node_pools
  }
}
 