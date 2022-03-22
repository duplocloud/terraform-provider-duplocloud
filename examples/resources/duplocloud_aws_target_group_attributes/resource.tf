resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Using target group ARN
resource "duplocloud_aws_target_group_attributes" "tg_attrs" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  target_group_arn = element(concat(flatten(duplocloud_ecs_service.nginx.*.target_group_arns), list("")), 0)

  attribute {
    key   = "deregistration_delay.timeout_seconds"
    value = "60"
  }
  attribute {
    key   = "slow_start.duration_seconds"
    value = "0"
  }
  attribute {
    key   = "stickiness.app_cookie.duration_seconds"
    value = "86400"
  }
  attribute {
    key   = "stickiness.enabled"
    value = "false"
  }
  attribute {
    key   = "stickiness.lb_cookie.duration_seconds"
    value = "86400"
  }
  attribute {
    key   = "stickiness.type"
    value = "lb_cookie"
  }
}


# For ECS - Using name and container port
resource "duplocloud_aws_target_group_attributes" "tg_attrs" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  role_name = "nginx"
  port      = 80
  is_ecs_lb = true

  attribute {
    key   = "deregistration_delay.timeout_seconds"
    value = "60"
  }
  attribute {
    key   = "slow_start.duration_seconds"
    value = "0"
  }
  attribute {
    key   = "stickiness.app_cookie.duration_seconds"
    value = "86400"
  }
  attribute {
    key   = "stickiness.enabled"
    value = "false"
  }
  attribute {
    key   = "stickiness.lb_cookie.duration_seconds"
    value = "86400"
  }
  attribute {
    key   = "stickiness.type"
    value = "lb_cookie"
  }
}
