data "duplocloud_gcp_sql_database_instances" "app" {
  tenant_id = "tenant_id"
}

output "sql_output" {
  value = {
    databases = data.duplocloud_gcp_sql_database_instances.app.databases
  }
}

