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
  capacity       = "t2.micro"
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
  capacity       = "t2.micro"
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
  capacity       = "t2.micro"
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
  capacity       = "t2.micro"
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
  capacity       = "t2.micro"
  agent_platform = 0      # Duplo native container agent
  zones          = [0, 1] # [Zone A, Zone B]
  user_account   = duplocloud_tenant.duplo-app.account_name

  metadata {
    key   = "OsDiskSize" # <== This is the size of the OS disk in GB
    value = "100"
  }

}


#Example to create ASG profile and updating it with launch template , setting default version and instance refresh
resource "duplocloud_asg_profile" "asgtest" {
  tenant_id           = duplocloud_tenant.duplo-app.tenant_id
  friendly_name       = "myasg"
  instance_count      = 1
  min_instance_count  = 1
  max_instance_count  = 1
  image_id            = "<ami-id>"
  capacity            = "t3a.medium"
  agent_platform      = 7
  zones               = [0]
  is_minion           = true
  is_ebs_optimized    = false
  encrypt_disk        = false
  allocated_public_ip = false
  cloud               = 0
  keypair_type        = 0
  use_spot_instances  = false
  custom_node_labels = {
    "new" = "data"
  }
  metadata {
    key   = "OsDiskSize"
    value = "50"
  }

}

resource "duplocloud_aws_launch_template" "name" {
  tenant_id           = duplocloud_tenant.duplo-app.tenant_id
  instance_type       = "t3a.micro"
  name                = duplocloud_asg_profile.asgtest.fullname
  version_description = "launch template allowed instance types"
  version             = "1"
  block_device_mapping {
    device_name = "/dev/xvda"

    ebs {
      volume_type           = "gp3"
      volume_size           = 150
      throughput            = 1000
      iops                  = 4000
      delete_on_termination = true
      encrypted             = true
    }
  }

}
resource "duplocloud_aws_launch_template_default_version" "set" {
  tenant_id       = duplocloud_tenant.duplo-app.tenant_id
  name            = duplocloud_asg_profile.asgtest.fullname
  default_version = duplocloud_aws_launch_template.name.latest_version
}


resource "duplocloud_asg_instance_refresh" "name" {
  asg_name                       = duplocloud_asg_profile.asgtest.fullname
  instance_warmup                = 300
  max_healthy_percentage         = 100
  min_healthy_percentage         = 90
  refresh_identifier             = "1"
  tenant_id                      = duplocloud_aws_launch_template_default_version.set.tenant_id
  update_launch_template_version = duplocloud_aws_launch_template_default_version.set.default_version
}