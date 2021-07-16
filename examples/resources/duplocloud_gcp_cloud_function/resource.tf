resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

// A simple cloud function with an HTTPS trigger
resource "duplocloud_gcp_cloud_function" "myfunc" {
  tenant_id = local.tenant_id

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
