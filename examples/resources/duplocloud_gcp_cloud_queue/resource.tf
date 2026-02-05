resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_gcp_cloud_queue" "queue" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "queue"
  #location="us-west2"
}
