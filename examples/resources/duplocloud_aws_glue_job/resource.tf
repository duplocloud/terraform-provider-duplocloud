data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Glue ETL (Spark) job. The IAM role must trust glue.amazonaws.com.
resource "duplocloud_aws_glue_job" "etl" {
  tenant_id = data.duplocloud_tenant.tenant.id
  name      = "events-etl"
  role      = "arn:aws:iam::123456789012:role/duploservices-dev-glue"
  body_json = jsonencode({
    Description = "Events ETL"
    Command = {
      Name           = "glueetl"
      ScriptLocation = "s3://my-scripts/events_etl.py"
      PythonVersion  = "3"
    }
    GlueVersion     = "4.0"
    WorkerType      = "G.1X"
    NumberOfWorkers = 5
    Timeout         = 120
  })
}

# Lightweight Python Shell job for ad-hoc scripts.
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
    Timeout     = 60
  })
}
