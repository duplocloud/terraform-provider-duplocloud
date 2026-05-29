resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_launch_template" "lt" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  name                = "launch-template-name"
  instance_type       = "t3a.medium"
  version             = "1"
  version_description = "launch template description"
  ami                 = "ami-123test"
}

resource "duplocloud_aws_launch_template" "name" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  instance_type       = "t3a.small"
  name                = "launch-template-name"
  version_description = "launch template block device mapping"
  version             = "2"
  block_device_mapping {
    device_name = "/dev/xvda"
    ebs {
      volume_size           = 30
      volume_type           = "gp3"
      delete_on_termination = true
      encrypted             = false
      iops                  = 3000
    }
    virtual_name = "ephemeral1"
  }

}

resource "duplocloud_aws_launch_template" "template" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  name                = "launch-template-name"
  version_description = "launch template block device mapping"
  version             = "3"
  instance_requirements {
    allowed_instance_types = ["t3a.*", "c5.*"]
    vcpu_count {
      min = 0
      max = 2
    }
    memory_mib {
      min = 4096
      max = 5120
    }
  }
}

resource "duplocloud_aws_launch_template" "mixed_instance_test" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  name                = duplocloud_asg_profile.mixed-instances-asg.fullname
  version_description = "launch template with extended instance requirements"

  instance_requirements {
    allowed_instance_types = ["m5.*", "m5a.*", "c5.*"]
    vcpu_count {
      min = 2
      max = 8
    }
    memory_mib {
      min = 4096
      max = 32768
    }
    cpu_manufacturers                           = ["intel", "amd"]
    instance_generations                        = ["current"]
    spot_max_price_percentage_over_lowest_price = 100
  }
}
