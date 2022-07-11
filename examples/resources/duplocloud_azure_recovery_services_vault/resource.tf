resource "duplocloud_azure_recovery_services_vault" "recovery_services_vault" {
  infra_name          = "demo"
  resource_group_name = "duploinfra-demo"
  name                = "test"
}
