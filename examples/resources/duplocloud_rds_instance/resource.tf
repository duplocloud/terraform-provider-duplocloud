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

//performance insights example
resource "duplocloud_rds_instance" "mydb" {
  tenant_id      = "5d3171c2-0fbc-4195-bb5e-05cd757ef786"
  name           = "mydb1psql"
  engine         = 1 // PostgreSQL
  engine_version = "14.11"
  size           = "db.t3.micro"

  master_username = "myuser"
  master_password = "Qaazwedd#1"

  encrypt_storage                 = true
  store_details_in_secret_manager = true
  enhanced_monitoring             = 0
  storage_type                    = "gp2"
  # parameter_group_name = "psql-group"
  performance_insights {
    enable           = false
    retention_period = 7
  }
}
