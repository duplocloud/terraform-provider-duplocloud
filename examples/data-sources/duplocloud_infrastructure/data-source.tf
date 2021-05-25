// Example 1 - look up an infrastructure by tenant ID.
data "duplocloud_infrastructure" "myinfra1" {
  tenant_id = var.tenant_id
}

// Example 2 - look up an infrastructure by name.
data "duplocloud_infrastructure" "myinfra2" {
  infra_name = "myinfra"
}
