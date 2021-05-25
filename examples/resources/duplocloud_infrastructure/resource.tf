resource "duplocloud_infrastructure" "myinfra" {
  infra_name        = "myinfra"
  cloud             = 0 // AWS
  region            = "us-east-1"
  azcount           = 2
  enable_k8_cluster = true
  address_prefix    = "10.34.0.0/16"
  subnet_cidr       = 24
}
