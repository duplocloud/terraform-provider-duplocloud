data "duplocloud_tenant" "tenant" {
  name = "dev"
}

resource "duplocloud_aws_glue_workflow" "pipeline" {
  tenant_id = data.duplocloud_tenant.tenant.id
  name      = "pipeline"
  body_json = jsonencode({
    Description = "End-to-end data pipeline"
    DefaultRunProperties = {
      env = "dev"
    }
    MaxConcurrentRuns = 1
  })
}

# Triggers reference the workflow by its tenant-prefixed `fullname`, not `name`.
# See the aws_glue_trigger example.
