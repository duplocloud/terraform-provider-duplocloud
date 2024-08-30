resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

// Generate a random password.
resource "random_password" "mypassword" {
  length  = 16
  special = false
}

// Create an RDS instance.
resource "duplocloud_rds_instance" "mydb" {
  tenant_id      = duplocloud_tenant.myapp.tenant_id
  name           = "mydb"
  engine         = 1 // PostgreSQL
  engine_version = "15.2"
  size           = "db.t3.medium"

  master_username = "myuser"
  master_password = random_password.mypassword.result

  encrypt_storage         = true
  backup_retention_period = 1
  availability_zone       = "us-west-2a"
}


// Create an RDS instance.
resource "duplocloud_rds_instance" "aurora-mydb" {
  tenant_id      = duplocloud_tenant.myapp.tenant_id
  name           = "aurora-mydb"
  engine         = 9 // AuroraDB
  engine_version = "15.2"
  size           = "db.t3.medium"

  master_username              = "myuser"
  master_password              = random_password.mypassword.result
  cluster_parameter_group_name = "default.cluster-groupname"

  encrypt_storage         = true
  backup_retention_period = 1
}
