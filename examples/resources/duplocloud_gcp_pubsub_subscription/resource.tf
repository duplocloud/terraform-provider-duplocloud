resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

//example for push config
resource "duplocloud_gcp_pubsub_subscription" "sub"{
  tenant_id=duplocloud_tenant.myapp.tenant_id
  name="subtest6"
  topic="test-topic1"
   ack_deadline_seconds = 20

  labels = {
    foo = "bar"
  }

  push_config {
    push_endpoint = "https://example.com/push"

    attributes = {
      x-goog-version = "v1"
    }
  }
}


//example for pull
resource "duplocloud_gcp_pubsub_subscription" "pullsub" {
  tenant_id=duplocloud_tenant.myapp.tenant_id
  name  = "subtest8"
  topic = "test-topic1"

  labels = {
    foo = "bar"
  }

  # 20 minutes
  message_retention_duration = "1200s"
  retain_acked_messages      = true

  ack_deadline_seconds = 20

  expiration_policy {
    ttl = "300000.5s"
  }
  retry_policy {
    minimum_backoff = "10s"
  }

  enable_message_ordering    = false
}

//example Deadletter policy
resource "duplocloud_gcp_pubsub_subscription" "pullsub" {
  tenant_id=duplocloud_tenant.myapp.tenant_id
  name  = "subtest9"
  topic = "test-topic1"

  dead_letter_policy {
    dead_letter_topic = "projects/<project-identifier>/topics/<topic-name>"
    max_delivery_attempts = 10
  }
}