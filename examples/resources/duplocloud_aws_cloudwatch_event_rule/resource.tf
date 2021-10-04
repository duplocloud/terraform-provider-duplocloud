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
