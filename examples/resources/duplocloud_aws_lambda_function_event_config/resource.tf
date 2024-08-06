resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "nondefault"
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

resource "duplocloud_aws_sqs_queue" "failure_queue" {
  tenant_id                   = duplocloud_tenant.myapp.tenant_id
  name                        = "failure_queue"
  fifo_queue                  = false
  message_retention_seconds   = 345600
  visibility_timeout_seconds  = 30
  content_based_deduplication = false
  delay_seconds               = 10
}

resource "duplocloud_aws_sqs_queue" "success_queue" {
  tenant_id                   = duplocloud_tenant.myapp.tenant_id
  name                        = "success_queue"
  fifo_queue                  = false
  message_retention_seconds   = 345600
  visibility_timeout_seconds  = 30
  content_based_deduplication = false
  delay_seconds               = 10
}

resource "duplocloud_aws_lambda_function_event_config" "event-invoke-config" {
  tenant_id                = duplocloud_tenant.myapp.tenant_id
  function_name            = duplocloud_aws_lambda_function.myfunction.fullname
  max_retry_attempts       = 1
  max_event_age_in_seconds = 100


  destination_config {
    on_failure {
      destination = duplocloud_aws_sqs_queue.failure_queue.arn
    }

    on_success {
      destination = duplocloud_aws_sqs_queue.success_queue.arn
    }
  }
}

