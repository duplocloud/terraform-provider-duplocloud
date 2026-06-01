resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Create the ASG that the warm pool will be attached to.
resource "duplocloud_asg_profile" "myapp" {
  tenant_id          = duplocloud_tenant.myapp.tenant_id
  friendly_name      = "myapp-asg"
  instance_count     = 1
  min_instance_count = 1
  max_instance_count = 4

  image_id       = "ami-0abcdef1234567890" # <== put the AWS duplo docker AMI ID here
  capacity       = "t3a.medium"
  agent_platform = 0 # Duplo native container agent
  zone           = 0 # Zone A
  user_account   = duplocloud_tenant.myapp.account_name

  metadata {
    key   = "OsDiskSize"
    value = "100"
  }
}

# Simple Example 1: Stopped warm pool with a single pre-initialized instance.
resource "duplocloud_aws_asg_warm_pool" "stopped" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  asg_name  = duplocloud_asg_profile.myapp.fullname

  min_size                    = 1
  max_group_prepared_capacity = 4
  pool_state                  = "Stopped" # one of "Stopped", "Running", or "Hibernated"
}

# Simple Example 2: Hibernated warm pool that returns instances to the pool on scale-in.
resource "duplocloud_aws_asg_warm_pool" "hibernated" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  asg_name  = duplocloud_asg_profile.myapp.fullname

  min_size   = 2
  pool_state = "Hibernated"

  instance_reuse_policy {
    reuse_on_scale_in = true
  }
}
