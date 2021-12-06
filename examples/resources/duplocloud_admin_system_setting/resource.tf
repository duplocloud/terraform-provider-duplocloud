resource "duplocloud_admin_system_setting" "test-setting" {
  key   = "EnableVPN"
  value = "true"
  type  = "Flags"
}
