data "duplocloud_gcp_sql_database_instance" "app" {
  tenant_id = "tenant_id"
  name      = "sql-shortname"
}

output "sql_output" {
  value = {
    tenant_id        = data.duplocloud_gcp_sql_database_instance.app.tenant_id
    name             = data.duplocloud_gcp_sql_database_instance.app.name
    fullname         = data.duplocloud_gcp_sql_database_instance.app.fullname
    self_link        = data.duplocloud_gcp_sql_database_instance.app.self_link
    database_version = data.duplocloud_gcp_sql_database_instance.app.database_version
    tier             = data.duplocloud_gcp_sql_database_instance.app.tier
    disk_size        = data.duplocloud_gcp_sql_database_instance.app.disk_size
    labels           = data.duplocloud_gcp_sql_database_instance.app.labels
    wait_until_ready = data.duplocloud_gcp_sql_database_instance.app.wait_until_ready
  }
}
