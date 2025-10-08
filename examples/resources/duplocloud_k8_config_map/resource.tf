resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_k8_config_map" "myapp" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  name = "myconfigmap"

  data = jsonencode({ foo = "bar2" })
  labels = {
    ke1 = "val1"
    ke2 = "val3"
  }
}


