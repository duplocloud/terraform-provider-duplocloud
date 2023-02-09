locals {
  tenant_id = "d186700c-ad18-4525-9593-aad467c843ff"
}

resource "duplocloud_aws_batch_scheduling_policy" "bsp" {
  tenant_id = local.tenant_id
  name      = "rtt"
  fair_share_policy {
    compute_reservation = 2
    share_decay_seconds = 3600

    share_distribution {
      share_identifier = "A1*"
      weight_factor    = 0.1
    }

    share_distribution {
      share_identifier = "A2"
      weight_factor    = 0.2
    }
  }

  tags = {
    "Name" = "Example Batch Scheduling Policy"
  }
}

resource "duplocloud_aws_batch_compute_environment" "bce" {
  tenant_id = local.tenant_id
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

resource "duplocloud_aws_batch_job_queue" "jq" {
  tenant_id = local.tenant_id
  name      = "tf-test-batch-job-queue"

  scheduling_policy_arn = duplocloud_aws_batch_scheduling_policy.bsp.arn
  state                 = "ENABLED"
  priority              = 1

  compute_environments = [
    duplocloud_aws_batch_compute_environment.bce.arn,
  ]
}
