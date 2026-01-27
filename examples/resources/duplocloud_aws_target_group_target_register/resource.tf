resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Using target group ARN
resource "duplocloud_aws_target_group_target_register" "trgt" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  target_group_arn = duplocloud_aws_lb_target_group.tg.arn

  targets {
    id = "<instnace type id>"
  }
  targets {
    id = "<instnace type id>"
  }

}

