resource "duplocloud_infrastructure_setting" "settings" {
  infra_name = duplocloud_infrastructure.myinfra

  dynamic "setting" {
    for_each = var.infra_settings
    content {
      key   = setting.value["key"]
      value = setting.value["value"]
    }
  }

  lifecycle {
    ignore_changes = [
      setting
    ]
  }
}

variable "infra_settings" {
  type = list(object({
    key   = string
    value = string
  }))

  default = [{
    key   = "EnableSecretCsiDriver"
    value = "true"
    },
    {
      key   = "EnableAWSEfsVolumes"
      value = "true"
  }]
}
