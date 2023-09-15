resource "duplocloud_tenant" "duplo-app" {
  account_name = "duplo-app"
  plan_id      = "default"
}

# With Schedule Expression
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

# With Event Pattern
resource "duplocloud_aws_cloudwatch_event_rule" "cw_erule2" {
  tenant_id   = duplocloud_tenant.duplo-app.tenant_id
  name        = "cw_erule2"
  description = "capture-aws-sign-in."
  event_pattern = jsonencode({
    detail-type = [
      "AWS Console Sign In via CloudTrail"
    ]
  })
}
