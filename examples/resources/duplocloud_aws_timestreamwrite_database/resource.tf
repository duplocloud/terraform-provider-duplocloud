locals {
  tenant_id = "d186700c-ad18-4525-9593-aad467c843ff"
}

data "duplocloud_tenant_aws_kms_key" "kms_key" {
  tenant_id = local.tenant_id
}

resource "duplocloud_aws_timestreamwrite_database" "timestreamwrite_database" {
  tenant_id  = local.tenant_id
  name       = "test"
  kms_key_id = data.duplocloud_tenant_aws_kms_key.kms_key.key_arn
}
