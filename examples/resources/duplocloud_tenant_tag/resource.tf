locals {
  tenant_id = "913a4498-db09-42c0-95b1-88ed26d87b83"
  tags      = { "key1" : "value1", "key2" : "value2" }
}

# Register tags key1;key2 in the system settings.
resource "duplocloud_admin_system_setting" "key_setting" {
  key   = "DUPLO_CUSTOM_TAGS"
  value = join(";", keys(local.tags))
  type  = "AppConfig"
}

resource "duplocloud_tenant_tag" "tags" {
  for_each = local.tags
  depends_on = [
    duplocloud_admin_system_setting.key_setting
  ]
  tenant_id = local.tenant_id
  key       = each.key
  value     = each.value
}
