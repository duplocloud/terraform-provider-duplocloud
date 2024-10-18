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
resource "duplocloud_aws_host" "eks2" {
  tenant_id     = "81f6043b-1480-4c92-a0d8-4d3d3a6ae13a"
  friendly_name = "tf-v2only"

  image_id       = "ami-0b896ca73d6b87976" # <== put the AWS EKS 1.18 AMI ID here
  capacity       = "t3.small"
  agent_platform = 7 # Duplo EKS agent
  zone           = 0 # Zone A
  user_account   = "jt-1303"
  keypair_type   = "1"
}

# Simple Example 3:  Create a host with instance metadata service
resource "duplocloud_aws_host" "host" {
  tenant_id     = "81f6043b-1480-4c92-a0d8-4d3d3a6ae13a"
  friendly_name = "tf-v2only"

  image_id       = "ami-0b896ca73d6b87976" # <== put the AWS EKS 1.18 AMI ID here
  capacity       = "t3.small"
  agent_platform = 7 # Duplo EKS agent
  zone           = 0 # Zone A
  user_account   = "jt-1303"
  keypair_type   = "1"

  metadata {
    key   = "OsDiskSize" # <== This is the size of the OS disk in GB
    value = "100"
  }

  # Create a host with instance metadata v2 only
  metadata {
    key   = "MetadataServiceOption"
    value = "enabled_v2_only"
  }

  # Create a host with instance metadata v1 and v2
  /* metadata {
    key   = "MetadataServiceOption"
    value = "enabled"
  } */

  # Create a host with instance metadata disabled
  /* metadata {
     key   = "MetadataServiceOption"
     value = "disabled"
   } */
}

# Simple Example 4:  Enabling Hibernation for ec2 host

resource "duplocloud_aws_host" "myhost" {
  capacity      = "t3a.small"
  cloud         = 0
  friendly_name = "host-5"
  image_id      = "ami-0adf7efba3226042e"
  tenant_id     = "abd5fb54-306a-46e7-a0e3-39a4de441bfd"

  metadata {
    key   = "EnableHibernation"
    value = "True"
  }

}

resource "duplocloud_aws_host" "host" {
  tenant_id     = "81f6043b-1480-4c92-a0d8-4d3d3a6ae13a"
  friendly_name = "tf-v2only"

  image_id       = "ami-0b896ca73d6b87976" # <== put the AWS EKS 1.18 AMI ID here
  capacity       = "t3.small"
  agent_platform = 7 # Duplo EKS agent
  zone           = 0 # Zone A
  user_account   = "jt-1303"
  keypair_type   = "1"

  metadata {
    key   = "OsDiskSize" # <== This is the size of the OS disk in GB
    value = "100"
  }

  # Create a host with instance metadata v2 only
  metadata {
    key   = "MetadataServiceOption"
    value = "enabled_v2_only"
  }

  custom_node_labels = {
    "key1" = "value1"
    "key2" = "value2"

  }

}

//example for taints
resource "duplocloud_aws_host" "native" {
  tenant_id     = "8cb23b3c-b7f3-4be1-8326-2a3cc4397a37"
  friendly_name = "host7"

  image_id         = "ami-0006160aad5007c19" # <== put the AWS duplo docker AMI ID here
  capacity         = "t3a.medium"
  agent_platform   = 7 # Duplo native container agent
  zone             = 1 # Zone A
  user_account     = "test13"
  is_ebs_optimized = true
  taints {
    key    = "tk1"
    value  = "tv2"
    effect = "NoSchedule"
  }
}
