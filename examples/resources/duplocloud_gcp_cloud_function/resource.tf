resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

// A simple cloud function with an HTTPS trigger
resource "duplocloud_gcp_cloud_function" "myfunc" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  name = "myfunc"

  source_archive_url = "gs://my-function-code-bucket/myfunc.zip"

  entrypoint = "my_entrypoint"
  runtime    = "nodejs10"

  available_memory_mb = 256
  ingress_type        = 1 // Allow All
  timeout             = 60

  environment_variables = {
    foo = "bar"
  }

  https_trigger {}
}

//Example of a cloud function with an event trigger
// This example shows how to create a cloud function that is triggered by a Pub/Sub event
// The function will be triggered when a message is published to the specified Pub/Sub topic
resource "duplocloud_gcp_cloud_function" "example_pubsub_function" {
  tenant_id  = duplocloud_tenant.myapp.tenant_id
  name       = "example-pubsub-fn"
  runtime    = "python39"
  entrypoint = "hello_pubsub"

  event_trigger {
    event_type = "google.pubsub.topic.publish"
    resource   = "projects/projectid/topics/topicname"
  }

  environment_variables = {
    EXAMPLE_ENV = "value"
  }

  available_memory_mb = 128
  timeout             = 60
  source_archive_url  = "gs://bucketname/code.ext"
}