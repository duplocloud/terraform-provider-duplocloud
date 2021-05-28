resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_kafka_cluster" "mycluster" {
  tenant_id     = duplocloud_tenant.this.tenant_id
  name          = "mycluster"
  kafka_version = "2.4.1.1"
  instance_type = "kafka.m5.large"
  storage_size  = 20
}
