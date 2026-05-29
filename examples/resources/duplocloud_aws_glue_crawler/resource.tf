data "duplocloud_tenant" "tenant" {
  name = "dev"
}

resource "duplocloud_aws_glue_database" "analytics" {
  tenant_id = data.duplocloud_tenant.tenant.id
  name      = "analytics"
}

# The IAM role must trust glue.amazonaws.com and have read access to the
# S3 target plus write access to the Glue catalog and CloudWatch Logs.
resource "duplocloud_aws_glue_crawler" "events" {
  tenant_id = data.duplocloud_tenant.tenant.id
  name      = "events"
  role      = "arn:aws:iam::123456789012:role/duploservices-dev-glue"
  body_json = jsonencode({
    Description  = "Crawl click events"
    DatabaseName = duplocloud_aws_glue_database.analytics.fullname
    Targets = {
      S3Targets = [
        { Path = "s3://my-bucket/events/" }
      ]
    }
    SchemaChangePolicy = {
      UpdateBehavior = "UPDATE_IN_DATABASE"
      DeleteBehavior = "LOG"
    }
  })
}
