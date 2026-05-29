data "duplocloud_tenant" "tenant" {
  name = "dev"
}

resource "duplocloud_aws_glue_registry" "events" {
  tenant_id = data.duplocloud_tenant.tenant.id
  name      = "events"
}

# Avro schema under the events registry. registry_name is the short name;
# the provider adds the tenant prefix when calling the backend.
resource "duplocloud_aws_glue_schema" "click" {
  tenant_id     = data.duplocloud_tenant.tenant.id
  registry_name = duplocloud_aws_glue_registry.events.name
  name          = "click"
  data_format   = "AVRO"
  compatibility = "BACKWARD"
  body_json = jsonencode({
    Description = "Click event schema"
    SchemaDefinition = jsonencode({
      type = "record"
      name = "Click"
      fields = [
        { name = "id", type = "string" },
        { name = "ts", type = "long" },
      ]
    })
  })
}
