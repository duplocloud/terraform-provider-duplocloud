resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_sqs_queue" "sqs_queue" {
  tenant_id                   = duplocloud_tenant.myapp.tenant_id
  name                        = "duplo_queue"
  fifo_queue                  = true
  message_retention_seconds   = 345600
  visibility_timeout_seconds  = 30
  content_based_deduplication = true
  delay_seconds               = 10
}

# SQS queue with dead letter queue configuration
resource "duplocloud_aws_sqs_queue" "sqs_queue_with_dlq" {
  tenant_id                   = duplocloud_tenant.myapp.tenant_id
  name                        = "duplo_queue"
  fifo_queue                  = true
  message_retention_seconds   = 345600
  visibility_timeout_seconds  = 30
  content_based_deduplication = true
  delay_seconds               = 10

  dead_letter_queue_configuration {
    target_sqs_dlq_name          = duplocloud_aws_sqs_queue.sqs_queue.fullname
    max_message_receive_attempts = 5
  }
}
