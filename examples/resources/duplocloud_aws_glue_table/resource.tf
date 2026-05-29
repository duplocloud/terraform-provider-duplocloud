data "duplocloud_tenant" "tenant" {
  name = "dev"
}

resource "duplocloud_aws_glue_database" "analytics" {
  tenant_id = data.duplocloud_tenant.tenant.id
  name      = "analytics"
}

# Parquet-backed external table under the analytics database.
# database_name is the short name; the provider applies the tenant prefix.
resource "duplocloud_aws_glue_table" "events" {
  tenant_id     = data.duplocloud_tenant.tenant.id
  database_name = duplocloud_aws_glue_database.analytics.name
  name          = "events"
  body_json = jsonencode({
    Description = "Click events"
    TableType   = "EXTERNAL_TABLE"
    Parameters = {
      classification = "parquet"
    }
    StorageDescriptor = {
      Location     = "s3://my-bucket/events/"
      InputFormat  = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"
      OutputFormat = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"
      SerdeInfo = {
        SerializationLibrary = "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe"
      }
      Columns = [
        { Name = "id", Type = "string" },
        { Name = "ts", Type = "timestamp" },
        { Name = "url", Type = "string" },
      ]
    }
  })
}
