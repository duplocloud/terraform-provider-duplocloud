resource "duplocloud_tenant" "duplo-app" {
  account_name = "duplo-app"
  plan_id      = "default"
}

resource "duplocloud_aws_cloudwatch_metric_alarm" "mAlarm" {
  tenant_id           = duplocloud_tenant.duplo-app.tenant_id
  metric_name         = "CPUUtilization"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  namespace           = "AWS/EC2"
  period              = 300
  threshold           = 80
  statistic           = "Average"

  dimension {
    key   = "InstanceId"
    value = "i-1234567abcdefghj"
  }
}
