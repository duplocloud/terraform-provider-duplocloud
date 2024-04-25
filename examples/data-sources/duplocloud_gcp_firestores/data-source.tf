data "duplocloud_gcp_firestores" "app" {
  tenant_id = "tenant_id"
}

output "out" {
  value = {
    firestores = data.duplocloud_gcp_firestores.app.firestores
  }
}