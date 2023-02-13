locals {
  tenant_id = "d186700c-ad18-4525-9593-aad467c843ff"
}

resource "duplocloud_aws_batch_job_definition" "jd" {
  tenant_id = local.tenant_id
  name      = "tf_test_batch_job_definition"
  type      = "container"

  platform_capabilities = ["EC2"]
  retry_strategy {
    attempts = 2
    evaluate_on_exit {
      action           = "EXIT"
      on_exit_code     = "1*"
      on_reason        = "reason*"
      on_status_reason = "status"
    }
  }

  timeout {
    attempt_duration_seconds = 60
  }

  container_properties = <<CONTAINER_PROPERTIES
  {
        "Command": [
            "sleep",
            "5"
        ],
        "Image": "amazonlinux",
        "ResourceRequirements": [
            {
                "Type": { "Value": "MEMORY" },
                "Value": "2048"
            },
            {
                "Type":  { "Value": "VCPU" },
                "Value": "1"
            },
            {
                "Type":  { "Value": "GPU" },
                "Value": "2"
            }
        ]
  }
CONTAINER_PROPERTIES
}
