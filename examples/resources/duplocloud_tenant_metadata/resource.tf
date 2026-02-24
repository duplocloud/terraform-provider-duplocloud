
locals {

  tenant_id = "913a4498-db09-42c0-95b1-88ed26d87b83"

  tenant_metadata = {
    meta1 = {
      key   = "env"
      value = "dev"
      type  = "text"
    }
    meta2 = {
      key   = "team"
      value = "platform"
      type  = "aws_console"
    }
    meta3 = {
      key   = "cost-center"
      value = "12345"
      type  = "url"
    }
  }
}

resource "duplocloud_tenant_metadata" "meta" {
  for_each = local.tenant_metadata

  tenant_id = local.tenant_id
  key       = each.value.key
  value     = each.value.value
  type      = each.value.type
}
