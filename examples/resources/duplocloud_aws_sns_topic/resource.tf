resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Without KMS Key running as fifo
resource "duplocloud_aws_sns_topic" "sns_topic" {
  tenant_id  = duplocloud_tenant.myapp.tenant_id
  name       = "duplo_topic.fifo" # AWS requires the ".fifo" extension for fifo sns topics
  fifo_topic = true
}

# With Tenant KMS Key
data "duplocloud_tenant_aws_kms_key" "tenant_kms_key" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
}

resource "duplocloud_aws_sns_topic" "sns_topic" {
  tenant_id  = duplocloud_tenant.myapp.tenant_id
  name       = "duplo_topic"
  kms_key_id = data.duplocloud_tenant_aws_kms_key.tenant_kms_key.key_arn
}