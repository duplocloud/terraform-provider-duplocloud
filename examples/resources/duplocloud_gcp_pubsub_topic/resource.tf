resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_gcp_pubsub_topic" "mytopic" {

  tenant_id = duplocloud_tenant.this.tenant_id
  name      = "mytopic"
}
