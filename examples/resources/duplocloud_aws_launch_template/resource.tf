resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_launch_template" "lt" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  name                = "launch-template-name"
  instance_type       = "t3a.medium"
  version             = "1"
  version_description = "launch template description"
  ami                 = "ami-123test"
}
