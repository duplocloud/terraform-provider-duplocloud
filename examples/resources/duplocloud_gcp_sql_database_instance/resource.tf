resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_gcp_sql_database_instance" "sql_instance" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  name             = "sqlinstances01"
  database_version = "MYSQL_8_0"
  tier             = "db-n1-standard-1"
  disk_size        = 17
  labels = {
    managed-by = "duplocloud"
    created-by = "terraform"
  }
}

