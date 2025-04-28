resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_tag" "custom" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  arn       = "resource-arn"
  key       = "custom1"
  value     = "customvalue1"
}