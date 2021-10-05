resource "duplocloud_tenant" "duplo-app" {
  account_name = "duplo-app"
  plan_id      = "default"
}

resource "duplocloud_aws_cloudwatch_event_rule" "cw_erule" {
  tenant_id           = duplocloud_tenant.duplo-app.tenant_id
  name                = "cw_erule"
  description         = "this is a test cloudwatch event rule."
  schedule_expression = "rate(10 minutes)"
  state               = "DISABLED"

  tag {
    key   = "CreatedBy"
    value = "Duplo"
  }

  tag {
    key   = "CreatedFrom"
    value = "Duplo"
  }
}

resource "duplocloud_aws_cloudwatch_event_target" "cw_etarget1" {
  tenant_id  = duplocloud_tenant.duplo-app.tenant_id
  rule_name  = duplocloud_aws_cloudwatch_event_rule.cw_erule.fullname
  target_arn = "arn:aws:lambda:us-west-2:294468937448:function:orphan-resource-tag"
  target_id  = "lamda-tst1"
}

resource "duplocloud_aws_cloudwatch_event_target" "cw_etarget2" {
  tenant_id  = duplocloud_tenant.duplo-app.tenant_id
  rule_name  = duplocloud_aws_cloudwatch_event_rule.cw_erule.fullname
  target_arn = "arn:aws:lambda:us-west-2:294468937448:function:orphan-resource-tag"
  target_id  = "lamda-tst2"
}