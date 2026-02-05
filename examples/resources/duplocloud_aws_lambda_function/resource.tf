resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_lambda_function" "myfunction" {

  tenant_id   = duplocloud_tenant.myapp.tenant_id
  name        = "myfunction"
  description = "A description of my function"

  runtime   = "java11"
  handler   = "com.example.MyFunction::handleRequest"
  s3_bucket = "my-bucket-name"
  s3_key    = "my-function.zip"

  environment {
    variables = {
      "foo" = "bar"
    }
  }

  timeout     = 60
  memory_size = 512
}

resource "duplocloud_aws_lambda_function" "thisfunction" {

  tenant_id   = duplocloud_tenant.myapp.tenant_id
  name        = "thisfunction"
  description = "A description of my function"

  package_type = "Image"
  image_uri    = "dkr.ecr.us-west-2.amazonaws.com/myimage:latest"

  image_config {
    command           = ["echo", "hello world"]
    entry_point       = ["echo hello workd"]
    working_directory = "/tmp3"
  }

  tracing_config {
    mode = "PassThrough"
  }

  timeout     = 60
  memory_size = 512
}

resource "duplocloud_aws_lambda_function" "edgefunction" {
  tenant_id   = duplocloud_tenant.myapp.tenant_id
  name        = "edgefunction"
  description = "An example edge function"

  package_type = "Image"
  image_uri    = "dkr.ecr.us-east-1.amazonaws.com/myimage:1.0"

  image_config {
    command           = ["echo", "hello world"]
    entry_point       = ["echo hello workd"]
    working_directory = "/tmp3"
  }

  tags = {
    IsEdgeDeploy = true
  }

  timeout     = 5
  memory_size = 128
}


#Example for the usage of invoke_arn attribute


resource "duplocloud_aws_lambda_function" "myfunction" {

  tenant_id   = duplocloud_tenant.myapp.tenant_id
  name        = "mylambda"
  description = "A description of my function"

  runtime     = "python3.14"
  handler     = "main.lambda_handler"
  s3_bucket   = "<s3-bucket-name>"
  s3_key      = "main.py.zip"
  timeout     = 3
  memory_size = 128
}

resource "duplocloud_aws_lambda_permission" "apigw_lambda" {
  action        = "lambda:InvokeFunction"
  function_name = duplocloud_aws_lambda_function.myfunction.fullname
  principal     = "apigateway.amazonaws.com"
  source_arn    = duplocloud_aws_lambda_function.myfunction.arn
  statement_id  = "AllowExecutionFromAPIGateway"
  tenant_id     = duplocloud_tenant.myapp.tenant_id
}

resource "duplocloud_aws_api_gateway_integration" "api" {
  tenant_id            = duplocloud_tenant.myapp.tenant_id
  name                 = "apigatewayname"
  lambda_function_name = duplocloud_aws_lambda_function.myfunction.fullname
}


resource "duplocloud_aws_apigateway_event" "apigateway_event" {
  tenant_id          = duplocloud_tenant.myapp.tenant_id
  api_gateway_id     = duplocloud_aws_api_gateway_integration.api.metadata
  method             = "GET"
  path               = "/"
  cors               = true
  authorization_type = "NONE"

  integration {
    type    = "AWS_PROXY"
    uri     = duplocloud_aws_lambda_function.myfunction.invoke_arn
    timeout = 29000
  }
}