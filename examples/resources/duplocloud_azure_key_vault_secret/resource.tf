resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_azure_key_vault_secret" "myapp" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "base01-test"
  value     = "tst"
  type      = "duplo_container_env"
}
