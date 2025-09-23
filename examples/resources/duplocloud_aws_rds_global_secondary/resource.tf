resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

// Generate a random password.
resource "duplocloud_rds_instance" "mydb" {
  tenant_id                       = duplocloud_tenant.myapp.tenant_id
  name                            = "primarydb"
  engine                          = 9 // PostgreSQL
  engine_version                  = "17.5"
  size                            = "db.r7g.large"
  kms_key_id                      = "arn:aws:kms:us-west-2:000000000000:key/69569059-6055-4238-9eb3-3e5235e1e262"
  master_username                 = "myuser"
  master_password                 = "$$$QWdqwe23dq"
  storage_type                    = "aurora"
  encrypt_storage                 = true
  store_details_in_secret_manager = true
  enhanced_monitoring             = 0
  enable_iam_auth                 = false
}

resource "duplocloud_aws_rds_global_secondary" "gs" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  cluster_identifier  = duplocloud_rds_instance.mydb.cluster_identifier
  secondary_tenant_id = "a54598b1-0d8f-4a7b-ba7e-4a20f890a57d"
  region              = "us-east-2"
}
