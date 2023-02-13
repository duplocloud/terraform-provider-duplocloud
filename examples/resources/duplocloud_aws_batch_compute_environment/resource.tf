resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_batch_compute_environment" "bce" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "sample"

  compute_resources {

    ec2_configuration {
      image_type = "ECS_AL2"
    }

    instance_type = [
      "optimal",
    ]

    allocation_strategy = "BEST_FIT"

    max_vcpus     = 8
    min_vcpus     = 1
    desired_vcpus = 2

    bid_percentage = 100

    type = "EC2"
  }

  type = "MANAGED"
}
