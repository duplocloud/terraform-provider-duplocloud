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
