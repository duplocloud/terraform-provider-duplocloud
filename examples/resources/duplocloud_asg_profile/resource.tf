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


#Asg example to enable metrics
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
  enabled_metrics = ["GroupDesiredCapacity", "GroupInServiceInstances", "GroupPendingInstances", "GroupStandbyInstances"] #, "GroupTerminatingInstances", "GroupTotalInstances"]

}


#custom node label example
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

  custom_node_labels = {
    "key1" = "value1"
    "key2" = "value2"
  }

}


//example for taints

resource "duplocloud_asg_profile" "duplo-test-asg" {
  tenant_id          = duplocloud_tenant.duplo-app.tenant_id
  friendly_name      = "asgtaint"
  instance_count     = 1
  min_instance_count = 1
  max_instance_count = 2

  image_id       = "ami-id" # <== put the AWS duplo docker AMI ID here
  capacity       = "t2.small"
  agent_platform = 7 # Duplo native container agent
  zone           = 1 # Zone A
  user_account   = "oct15"

  taints {
    key    = "tk1"
    value  = "tv2"
    effect = "NoSchedule"
  }
}


#secondary volume example
resource "duplocloud_asg_profile" "duplo-test-asg" {
  tenant_id             = duplocloud_tenant.duplo-app.tenant_id
  friendly_name         = "duplo-test-asg"
  instance_count        = 1
  min_instance_count    = 1
  max_instance_count    = 1
  image_id              = "ami-077e0e0754c311a47" # <== put the AWS duplo docker AMI ID here
  capacity              = "t3a.small"
  agent_platform        = 7 # Duplo native container agent
  zone                  = 0 # Zone A
  user_account          = duplocloud_tenant.duplo-app.account_name
  keypair_type          = 2
  prepend_user_data     = false
  use_spot_instances    = true
  can_scale_from_zero   = false
  is_cluster_autoscaled = false

  metadata {
    key   = "OsDiskSize" # <== This is the size of the OS disk in GB
    value = "100"
  }

  minion_tags {
    key   = "AllocationTags"
    value = "test"
  }

  volume {
    name        = "/dev/sda2"
    volume_type = "gp3"
    size        = 100
  }
}


resource "duplocloud_asg_profile" "duplo-test-asg" {
  tenant_id          = duplocloud_tenant.duplo-app.tenant_id
  friendly_name      = "duplo-test-asg"
  instance_count     = 1
  min_instance_count = 1
  max_instance_count = 2

  image_id       = "ami-077e0e0754c311a47" # <== put the AWS duplo docker AMI ID here
  capacity       = "t2.small"
  agent_platform = 0      # Duplo native container agent
  zones          = [1, 2] # [Zone A, Zone B]
  user_account   = duplocloud_tenant.duplo-app.account_name

  metadata {
    key   = "OsDiskSize" # <== This is the size of the OS disk in GB
    value = "100"
  }

}
