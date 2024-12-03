resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_datafactory" "df" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  name          = "tf-dft2"
  public_access = true
}