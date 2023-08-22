resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Simple Example 1:  Deploy a host to be used with Duplo's native container agent
resource "duplocloud_aws_host" "native" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  friendly_name = "host1"

  image_id       = "ami-abcd1234" # <== put the AWS duplo docker AMI ID here
  capacity       = "t3a.medium"
  agent_platform = 0 # Duplo native container agent
  zone           = 0 # Zone A
  user_account   = duplocloud_tenant.myapp.account_name

  metadata {
    key   = "OsDiskSize" # <== This is the size of the OS disk in GB
    value = "100"
  }
}

# Simple Example 2:  Deploy a host to be used with EKS
resource "duplocloud_aws_host" "eks" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  friendly_name = "host2"

  image_id       = "ami-12345678" # <== put the AWS EKS 1.18 AMI ID here
  capacity       = "t3a.medium"
  agent_platform = 7 # Duplo EKS agent
  zone           = 0 # Zone A
  user_account   = duplocloud_tenant.myapp.account_name

  metadata {
    key   = "OsDiskSize" # <== This is the size of the OS disk in GB
    value = "100"
  }
}
