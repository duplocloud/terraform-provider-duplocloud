// Example 1 - look up a standard SQS queue by tenant ID and name.
data "duplocloud_aws_sqs_queue" "myqueue" {
  tenant_id = var.tenant_id
  name      = "myqueue"
}

// Output the queue ARN
output "queue_arn" {
  value = data.duplocloud_aws_sqs_queue.myqueue.arn
}

// Output the queue URL
output "queue_url" {
  value = data.duplocloud_aws_sqs_queue.myqueue.url
}

// Output the full queue name
output "queue_fullname" {
  value = data.duplocloud_aws_sqs_queue.myqueue.fullname
}

// Example 2 - look up a FIFO SQS queue.
data "duplocloud_aws_sqs_queue" "myfifoqueue" {
  tenant_id = var.tenant_id
  name      = "myfifoqueue.fifo"
}

// Output FIFO queue details
output "fifo_queue_arn" {
  value = data.duplocloud_aws_sqs_queue.myfifoqueue.arn
}

output "is_fifo_queue" {
  value = data.duplocloud_aws_sqs_queue.myfifoqueue.fifo_queue
}
