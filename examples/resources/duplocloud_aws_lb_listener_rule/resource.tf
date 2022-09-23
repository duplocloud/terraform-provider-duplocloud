locals {
  tenant_id = "053a2efa-d03f-4e1d-b3b5-33c04cbed87e"
  vpc_id    = "vpc-0ba0667cfc200f060"
  cert_arn  = "arn:aws:acm:us-west-2:957282632678:certificate/2e882320-5aa5-4b8d-881f-998050178205"
}

resource "duplocloud_aws_load_balancer" "alb" {
  tenant_id            = local.tenant_id
  name                 = "tst-alb"
  is_internal          = true
  enable_access_logs   = true
  drop_invalid_headers = true
}

resource "duplocloud_aws_lb_target_group" "tg" {
  tenant_id   = local.tenant_id
  name        = "tg1"
  port        = 80
  protocol    = "HTTP"
  vpc_id      = local.vpc_id
  target_type = "instance"
}

resource "duplocloud_aws_load_balancer_listener" "alb-listener" {
  tenant_id          = local.tenant_id
  load_balancer_name = duplocloud_aws_load_balancer.alb.name
  port               = 8443
  protocol           = "HTTPS"
  target_group_arn   = duplocloud_aws_lb_target_group.tg.arn
  certificate_arn    = local.cert_arn
}

resource "duplocloud_aws_lb_listener_rule" "static" {
  tenant_id    = local.tenant_id
  listener_arn = duplocloud_aws_load_balancer_listener.alb-listener.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = duplocloud_aws_lb_target_group.tg.arn
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }

  condition {
    host_header {
      values = ["example.com"]
    }
  }
}

# Forward action

resource "duplocloud_aws_lb_listener_rule" "host_based_weighted_routing" {
  tenant_id    = local.tenant_id
  listener_arn = duplocloud_aws_load_balancer_listener.alb-listener.arn
  priority     = 99

  action {
    type             = "forward"
    target_group_arn = duplocloud_aws_lb_target_group.tg.arn
  }

  condition {
    host_header {
      values = ["my-service.*.terraform.io"]
    }
  }
}

# Redirect action

resource "duplocloud_aws_lb_listener_rule" "redirect_http_to_https" {
  tenant_id    = local.tenant_id
  listener_arn = duplocloud_aws_load_balancer_listener.alb-listener.arn
  priority     = 98
  action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }

  condition {
    http_header {
      http_header_name = "X-Forwarded-For"
      values           = ["192.168.1.*"]
    }
  }
}

# Fixed-response action

resource "duplocloud_aws_lb_listener_rule" "health_check" {
  tenant_id    = local.tenant_id
  listener_arn = duplocloud_aws_load_balancer_listener.alb-listener.arn
  priority     = 97
  action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "HEALTHY"
      status_code  = "200"
    }
  }

  condition {
    query_string {
      key   = "health"
      value = "check"
    }

    query_string {
      key   = "foo"
      value = "bar"
    }
  }
}
