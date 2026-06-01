data "duplocloud_tenant" "tenant" {
  name = "dev"
}

resource "duplocloud_aws_glue_connection" "warehouse" {
  tenant_id = data.duplocloud_tenant.tenant.id
  name      = "warehouse"
  body_json = jsonencode({
    Description    = "Connection to the Postgres warehouse"
    ConnectionType = "JDBC"
    ConnectionProperties = {
      JDBC_CONNECTION_URL = "jdbc:postgresql://warehouse.example.com:5432/analytics"
      USERNAME            = "glue_user"
      PASSWORD            = "REDACTED"
    }
  })
}
