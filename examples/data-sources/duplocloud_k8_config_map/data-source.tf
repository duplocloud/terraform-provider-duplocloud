data "duplocloud_k8_config_map" "test" {
  tenant_id = var.tenant_id
  name      = "myconfigmap"
}

output "config_map" {
  value = jsondecode(data.duplocloud_k8_config_map.test.data)
}