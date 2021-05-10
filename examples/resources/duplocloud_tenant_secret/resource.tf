resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Example with plaintext data.
resource "duplocloud_tenant_secret" "mysecret1" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  # The full name will be:  duploservices-myapp-mytext
  name_suffix = "mytext"

  data = "hi"
}

# Example with JSON data.
resource "duplocloud_tenant_secret" "mysecret2" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  # The full name will be:  duploservices-myapp-myjson
  name_suffix = "myjson"

  data = jsonencode({ foo = "bar" })
}
