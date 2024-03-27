data "duplocloud_gcp_node_pools" "pool" {
  tenant_id = "tenant_id"
}
#
output "nodepool_output" {
  value = {
    node_pools = data.duplocloud_gcp_node_pools.pool.node_pools

  }
}
