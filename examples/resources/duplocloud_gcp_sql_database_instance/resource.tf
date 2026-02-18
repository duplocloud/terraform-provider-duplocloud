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


// Backup configuration example
resource "duplocloud_gcp_sql_database_instance" "sql" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  name             = "mysqlbckp"
  database_version = "POSTGRES_14"
  disk_size        = 10
  tier             = "db-f1-micro"

  root_password = "qwerty"
  need_backup   = true
}

resource "duplocloud_gcp_sql_database_instance" "sql_instance" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  name             = "customtier"
  database_version = "SQLSERVER_2019_STANDARD"
  tier             = "db-custom-2-7680"
  disk_size        = 19
  root_password    = "test@123Abc"
  labels = {
    managed-by = "duplocloud"
    created-by = "terraform"
  }
}

// SSL Ip configuration example for instance
resource "duplocloud_gcp_sql_database_instance" "db" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  name             = "mypsql"
  database_version = "POSTGRES_15"
  disk_size        = 20
  tier             = "db-custom-1-3840"
  database_flag {
    name  = "max_connections"
    value = "1000"
  }
  ip_configuration {
    ssl_mode    = "ENCRYPTED_ONLY"
    require_ssl = false
  }
}

// SSL Ip configuration example for server instance

resource "duplocloud_gcp_sql_database_instance" "db" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  name             = "mysqlserver"
  database_version = "SQLSERVER_2022_EXPRESS"
  disk_size        = 20
  tier             = "db-custom-1-3840"
  ip_configuration {

    ssl_mode    = "ENCRYPTED_ONLY"
    require_ssl = true
  }
  root_password = "Guide#123"
}

