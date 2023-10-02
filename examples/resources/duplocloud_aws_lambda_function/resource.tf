resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_lambda_function" "myfunction" {

  tenant_id   = duplocloud_tenant.this.tenant_id
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

  image_config {
    command = ["app.handler5"]
    entry_point = [
      "/usr/local/bin/python",
      "awslambdaruntimeclient"
    ]
    working_directory = "/var/task1"
  }

  tracing_config {
    mode = "PassThrough"
  }

  timeout     = 60
  memory_size = 512
}
