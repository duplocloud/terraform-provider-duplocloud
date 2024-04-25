data "duplocloud_gcp_node_pools" "app" {
  tenant_id = "tenantid"
}

output "nodepool_output" {
  value = {
    node_pools = data.duplocloud_gcp_node_pools.app.node_pools
  }
}
 