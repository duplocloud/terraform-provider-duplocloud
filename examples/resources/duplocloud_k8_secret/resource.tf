resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_k8_secret" "myapp" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  secret_name = "mysecret"
  secret_type = "Opaque"
  secret_data = jsonencode({ foo = "bar2" })
  secret_label = {
    KeyA = "ValueA"
    KeyB = "ValueB"
  }
}
