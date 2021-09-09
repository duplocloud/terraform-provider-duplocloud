// Example 1 - look up an infrastructure by tenant ID.
data "duplocloud_infrastructure" "myinfra1" {
  tenant_id = var.tenant_id
}

// Example 2 - look up an infrastructure by name.
data "duplocloud_infrastructure" "myinfra2" {
  infra_name = "myinfra"
}

// Example 3 - look up list of certificates by plan ID.
data "duplocloud_plan_certificates" "cert_list" {
   plan_id = "default"
}

// Example 3 - look up plan certificates by plan ID and certificate name.
data "duplocloud_plan_certificate" "single_cert" {
   plan_id = "default"
   name    = "poc.duplocloud.net"
}