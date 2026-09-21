resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_k8_secret" "myapp" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  secret_name = "mysecret"
  secret_type = "Opaque"
  secret_data = jsonencode({ foo = "bar2" })
  secret_labels = {
    KeyA                          = "ValueA"
    KeyB                          = "ValueB"
    "app.duplocloud.net/app-name" = "<appname>"
  }
  secret_annotations = {
    annotA = "ValueA"
    annotB = "ValueB"
  }
}



# Track a secret whose values are owned outside of Terraform - created by the Duplo
# installer, rotated out of band - so that services can depend on it without its
# contents ever being written to Terraform state.
resource "duplocloud_k8_secret" "env_var_global" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  secret_name = "env-var-os-global"
  secret_type = "Opaque"

  # secret_data is left out on purpose: Terraform manages only the existence of this
  # secret, and its values are masked in state.
  manage_secret_data = false
}
