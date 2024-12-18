resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

data "duplocloud_aws_launch_template" "op" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "launch_template_name"
}

resource "duplocloud_aws_launch_template_default_version" "name" {
  tenant_id       = data.duplocloud_aws_launch_template.op.tenant_id
  name            = data.duplocloud_aws_launch_template.op.name
  default_version = data.duplocloud_aws_launch_template.op.latest_version
}