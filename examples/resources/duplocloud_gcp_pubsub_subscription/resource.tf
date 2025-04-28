resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

//example for push config
resource "duplocloud_gcp_pubsub_subscription" "sub" {
  tenant_id            = duplocloud_tenant.myapp.tenant_id
  name                 = "subtest6"
  topic                = "{topic-name}"
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
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "subtest8"
  topic     = "{topic-name}"

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

  enable_message_ordering = false
}

//example Deadletter policy
resource "duplocloud_gcp_pubsub_subscription" "pullsub" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "subtest9"
  topic     = "{topic-name}"

  dead_letter_policy {
    dead_letter_topic     = "projects/{project-identifier}/topics/{topic-name}"
    max_delivery_attempts = 10
  }
}


//example BigQuery
resource "duplocloud_gcp_pubsub_subscription" "sub" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "subtest12"
  topic     = "{topic-name}"
  big_query {
    table                 = "{project}.{dataset}.{table}"
    service_account_email = "abc@xxyz.com"
    use_table_schema      = true
  }
}

//example cloud storage
resource "duplocloud_gcp_pubsub_subscription" "pullsub" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "subtestcloud"
  topic     = "{topic-name}"

  cloud_storage_config {
    bucket = "{cloudstorage-bucketname}"

    filename_prefix          = "pre-"
    filename_suffix          = "10"
    filename_datetime_format = "YYYY-MM-DD/hh_mm_ssZ"

    max_bytes    = 1000
    max_duration = "300s"
    max_messages = 1000
  }
}


resource "duplocloud_gcp_pubsub_subscription" "sub" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "subtest2"
  topic     = "{topic-name}"
  big_query {
    table                 = "gcp-test10-431717.pbdataset.pbtable"
    service_account_email = "abc@xxyz.com"
    use_table_schema      = true
    drop_unknown_fields   = false
  }
}