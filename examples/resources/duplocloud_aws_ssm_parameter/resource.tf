resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_ssm_parameter" "ssm_param" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "ssm_param"
  type      = "String"
  value     = "ssm_param_value"
}
