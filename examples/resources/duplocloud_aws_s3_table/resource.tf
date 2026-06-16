resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


resource "duplocloud_aws_s3_table" "table" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "table"
}


