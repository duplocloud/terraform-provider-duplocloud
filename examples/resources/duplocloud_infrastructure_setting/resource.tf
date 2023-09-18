resource "duplocloud_infrastructure_setting" "settings" {
  infra_name            = duplocloud_infrastructure.myinfra
  EnableSecretCsiDriver = "true"
  EnableAWSEfsVolumes   = "true"
}
