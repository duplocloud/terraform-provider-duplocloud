resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_rds_instance" "rds" {
  tenant_id       = duplocloud_tenant.myapp.tenant_id
  enable_logging  = false
  encrypt_storage = true
  engine          = 8
  engine_version  = "8.0.mysql_aurora.3.04.0"
  master_password = "test!!1234"
  master_username = "masteruser"
  multi_az        = false
  name            = "mysqltest"
  size            = "db.t2.small"
}

resource "duplocloud_rds_read_replica" "replica" {
  tenant_id          = duplocloud_rds_instance.rds.tenant_id
  name               = "read-replica"
  size               = "db.t2.small"
  cluster_identifier = duplocloud_rds_instance.rds.cluster_identifier
}
