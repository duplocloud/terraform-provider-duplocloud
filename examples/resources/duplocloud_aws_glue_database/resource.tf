data "duplocloud_tenant" "tenant" {
  name = "dev"
}

resource "duplocloud_aws_glue_database" "analytics" {
  tenant_id = data.duplocloud_tenant.tenant.id
  name      = "analytics"
  body_json = jsonencode({
    Description = "Analytics tables for the data team"
    Parameters = {
      classification = "csv"
    }
  })
}
