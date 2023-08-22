resource "duplocloud_tenant" "duplo-app" {
  account_name = "duplo-app"
  plan_id      = "default"
}

#Deploy a ASG to be used with Duplo's native container agent
resource "duplocloud_asg_profile" "duplo-test-asg" {
  tenant_id          = duplocloud_tenant.duplo-app.tenant_id
  friendly_name      = "duplo-test-asg"
  instance_count     = 1
  min_instance_count = 1
  max_instance_count = 2

  image_id       = "ami-077e0e0754c311a47" # <== put the AWS duplo docker AMI ID here
  capacity       = "t2.small"
  agent_platform = 0 # Duplo native container agent
  zone           = 0 # Zone A
  user_account   = duplocloud_tenant.duplo-app.account_name

  metadata {
    key   = "OsDiskSize" # <== This is the size of the OS disk in GB
    value = "100"
  }
}

