resource "duplocloud_infrastructure_setting" "settings" {
  infra_name            = duplocloud_infrastructure.myinfra
  setting = [
    {
      key   = "exampleKey1"
      value = "exampleValue1"
    },
    {
      key   = "exampleKey2"
      value = "exampleValue2"
    }
  ]

  delete_unspecified_settings = false
}
