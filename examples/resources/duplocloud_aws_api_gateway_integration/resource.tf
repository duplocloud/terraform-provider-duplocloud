resource "duplocloud_tenant" "duplo-app" {
  account_name = "duplo-app"
  plan_id      = "default"
}

resource "duplocloud_aws_api_gateway_integration" "apigw-lambda" {
  tenant_id            = duplocloud_tenant.duplo-app.tenant_id
  name                 = "test-api-lambda"
  lambda_function_name = "duploservices-dev01-osrm-engine"
}
