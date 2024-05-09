resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

// A simple scheduler job with an HTTPS target, running at 9 am daily.
resource "duplocloud_gcp_scheduler_job" "myjob" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  name = "myjob"

  schedule = "* 9 * * *"
  timezone = "America/New_York"

  http_target {
    method = "GET"
    uri    = "https://www.google.com"
  }
}
