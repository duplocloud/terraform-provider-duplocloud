resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_kafka_cluster" "mycluster" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  name          = "mycluster"
  kafka_version = "3.5.1"
  instance_type = "kafka.m5.large"
  storage_size  = 20
}


resource "duplocloud_aws_kafka_cluster" "serverless" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  name          = "serverlesscluster"
  subnets       = ["subnet-0c9a95f287b4fc38f", "subnet-0f629e07cf54d3ca0"]
  is_serverless = true
}
