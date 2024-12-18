resource "duplocloud_tenant" "tenant" {
  account_name = "test"
  plan_id      = "test"
}

resource "duplocloud_azure_tenant_key_vault" "kv" {
  tenant_id                  = duplocloud_tenant.tenant.tenant_id
  name                       = "tst-kv001"
  sku_name                   = "standard"
  purge_protection_enabled   = true
  soft_delete_retention_days = 90
}
