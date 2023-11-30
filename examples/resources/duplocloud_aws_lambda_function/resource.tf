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

  timeout     = 60
  memory_size = 512
}

resource "duplocloud_aws_lambda_function" "thisfunction" {

  tenant_id   = duplocloud_tenant.this.tenant_id
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
  tenant_id   = "c7163b39-43ca-4d44-81ce-9a323087039b"
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