locals {
  tenant_id = "053a2efa-d03f-4e1d-b3b5-33c121342adae"
  vpc_id    = "vpc-0ba0667cfc20111122233"
}

resource "duplocloud_aws_lb_target_group" "tg" {
  tenant_id   = local.tenant_id
  name        = "tg1"
  port        = 80
  protocol    = "HTTP"
  vpc_id      = local.vpc_id
  target_type = "instance"

  health_check {
    healthy_threshold   = 8
    interval            = 300
    path                = "/health"
    port                = "9000"
    protocol            = "HTTP"
    timeout             = 60
    unhealthy_threshold = 6
  }
}
