data "duplocloud_k8_config_maps" "test" {
  tenant_id = var.tenant_id
}

output "config_maps" {
  value = data.duplocloud_k8_config_maps.test.config_maps
}
