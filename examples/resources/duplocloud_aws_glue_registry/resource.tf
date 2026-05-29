data "duplocloud_tenant" "tenant" {
  name = "dev"
}

resource "duplocloud_aws_glue_registry" "events" {
  tenant_id = data.duplocloud_tenant.tenant.id
  name      = "events"
  body_json = jsonencode({
    Description = "Event schemas shared across services"
  })
}
