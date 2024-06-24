resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_apigateway_event" "apigateway_event" {
  tenant_id          = duplocloud_tenant.myapp.tenant_id
  api_gateway_id     = "t84tb3skz0"
  method             = "POST"
  path               = "/v2/docs"
  cors               = true
  authorization_type = "COGNITO_USER_POOLS"
  authorizer_id      = "gto03x"

  integration {
    type    = "AWS_PROXY"
    uri     = "arn:aws:apigateway:us-west-2:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:366133256645:function:duploservices-dev-valuation-manheimAspects/invocations"
    timeout = 29000
  }
}