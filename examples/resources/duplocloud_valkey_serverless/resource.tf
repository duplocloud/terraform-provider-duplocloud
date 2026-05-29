resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_valkey_serverless" "valkey" {
  tenant_id                = duplocloud_tenant.myapp.tenant_id
  name                     = "valkey-sl"
  description              = "my serverless valkey"
  snapshot_retention_limit = 1
  engine_version           = "7"
  kms_key_id               = "kms-keyid-arn"
}