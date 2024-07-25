resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_ecr_repository" "test-ecr" {
  tenant_id                 = duplocloud_tenant.myapp.tenant_id
  name                      = "test-ecr"
  enable_scan_image_on_push = true
  enable_tag_immutability   = true
  force_delete              = false
}
