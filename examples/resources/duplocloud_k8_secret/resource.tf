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
#
# Import the secret before the first apply:
#
#   terraform import duplocloud_k8_secret.env_var_global \
#     v2/subscriptions/*TENANT_ID*/K8SecretApiV2/env-var-os-global
#
# Importing is not strictly required - a create carries any existing data forward - but
# it keeps Terraform from treating a secret that already exists as one it is making, and
# it is the only way to catch a wrong secret_type before the plan proposes a replacement.
# Note that the import itself writes the secret's values to state once, so rotate the
# secret afterwards if that matters.
resource "duplocloud_k8_secret" "env_var_global" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  secret_name = "env-var-os-global"
  secret_type = "Opaque"

  # secret_data is left out on purpose: Terraform manages only the existence of this
  # secret, and its values are masked in state.
  manage_secret_data = false
}
