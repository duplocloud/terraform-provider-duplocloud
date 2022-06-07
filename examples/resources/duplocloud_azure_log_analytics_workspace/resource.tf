resource "duplocloud_azure_log_analytics_workspace" "log_analytics_workspace" {
  infra_name          = "demo"
  resource_group_name = "duploinfra-demo"
  name                = "test"
}
