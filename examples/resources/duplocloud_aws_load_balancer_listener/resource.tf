resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_load_balancer" "myapp" {
  tenant_id            = duplocloud_tenant.myapp.tenant_id
  name                 = "myapp"
  is_internal          = true
  enable_access_logs   = true
  drop_invalid_headers = true
}



resource "duplocloud_aws_load_balancer_listener" "myapp-listener" {
  tenant_id          = duplocloud_tenant.myapp.tenant_id
  load_balancer_name = duplocloud_aws_load_balancer.myapp.name
  port               = 8443
  protocol           = "TCP"
  default_actions {
    forward {
      target_group_arn = "arn:aws:elasticloadbalancing:us-west-2:100000004:targetgroup/tg1/cc3e8ee8256682dd"
    }
  }
}



resource "duplocloud_aws_load_balancer_listener" "myapp-listener" {
  tenant_id          = duplocloud_tenant.myapp.tenant_id
  load_balancer_name = duplocloud_aws_load_balancer.myapp.name
  port               = 80
  protocol           = "HTTP"
  default_actions {
    fixed_response {
      content_type = "text/plain"
      message_body = "Hello, World!"
      status_code  = "200"
    }
  }
}

resource "duplocloud_aws_load_balancer_listener" "myapp-listener1" {
  tenant_id          = duplocloud_tenant.myapp.tenant_id
  load_balancer_name = duplocloud_aws_load_balancer.myapp.name
  port               = 5580
  protocol           = "HTTP"
  default_actions {
    redirect {
      path        = "/api"
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}
