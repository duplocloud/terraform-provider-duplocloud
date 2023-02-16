locals {
  tenant_id = "2a80c75d-9f58-4572-83b7-157b05bce259"
}

data "duplocloud_tenant_aws_kms_key" "kms_key" {
  tenant_id = local.tenant_id
}

resource "duplocloud_aws_timestreamwrite_database" "timestreamwrite_database" {
  tenant_id  = local.tenant_id
  name       = "test"
  kms_key_id = data.duplocloud_tenant_aws_kms_key.kms_key.key_arn
}

resource "duplocloud_aws_timestreamwrite_table" "timestreamwrite_database_tbl" {
  tenant_id     = local.tenant_id
  database_name = duplocloud_aws_timestreamwrite_database.timestreamwrite_database.fullname
  name          = "example"

  retention_properties {
    magnetic_store_retention_period_in_days = 30
    memory_store_retention_period_in_hours  = 8
  }
  magnetic_store_write_properties {
    enable_magnetic_store_writes = true
    magnetic_store_rejected_data_location {
      s3_configuration {
        bucket_name       = "test"
        encryption_option = "SSE_KMS"
        kms_key_id        = data.duplocloud_tenant_aws_kms_key.kms_key.key_arn
      }
    }
  }
}
