data "duplocloud_tenant" "tenant" {
  name = "dev"
}

resource "duplocloud_aws_glue_job" "report" {
  tenant_id = data.duplocloud_tenant.tenant.id
  name      = "daily-report"
  role      = "arn:aws:iam::123456789012:role/duploservices-dev-glue"
  body_json = jsonencode({
    Command = {
      Name           = "pythonshell"
      ScriptLocation = "s3://my-scripts/daily_report.py"
      PythonVersion  = "3.9"
    }
    GlueVersion = "3.0"
    MaxCapacity = 0.0625
  })
}

# Cross-resource references inside body_json use `.fullname` (tenant-prefixed),
# not `.name`. AWS does not recognize the short user-facing name.
resource "duplocloud_aws_glue_trigger" "nightly" {
  tenant_id = data.duplocloud_tenant.tenant.id
  name      = "nightly"
  type      = "SCHEDULED"
  body_json = jsonencode({
    Schedule        = "cron(0 7 * * ? *)"
    StartOnCreation = true
    Actions = [
      { JobName = duplocloud_aws_glue_job.report.fullname }
    ]
  })
}
