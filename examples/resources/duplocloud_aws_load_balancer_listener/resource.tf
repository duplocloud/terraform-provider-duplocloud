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
  protocol           = "https"
  target_group_arn   = "arn:aws:elasticloadbalancing:us-west-2:1234567890:targetgroup/duplo2-stage-antcmw-http4000/fc6f818e85fa737a"
}
