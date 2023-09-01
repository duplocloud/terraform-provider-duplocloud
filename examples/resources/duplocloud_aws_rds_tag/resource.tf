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
  engine_version = "12.5"
  size           = "db.t3.medium"

  master_username = "myuser"
  master_password = random_password.mypassword.result

  encrypt_storage = true
}

// Create RDS Tag for type "instance".
resource "duplocloud_aws_rds_tag" "tag" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  resource_type = "instance"
  resource_id   = duplocloud_rds_instance.mydb.identifier
  key           = "CreatedBy"
  value         = "DuploCloud"
}
