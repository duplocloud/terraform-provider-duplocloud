resource "duplocloud_infrastructure" "infra" {
  infra_name            = "tst-0206"
  account_id            = "143ffc59-9394-4ec6-8f5a-c408a238be62" // Subscription Id
  cloud                 = 2
  azcount               = 2
  region                = "West US 2"
  enable_k8_cluster     = false
  address_prefix        = "10.50.0.0/16"
  subnet_cidr           = 0
  subnet_name           = "sub01"
  subnet_address_prefix = "10.50.1.0/24"
}

resource "duplocloud_azure_k8s_cluster" "cluster" {
  infra_name = duplocloud_infrastructure.infra.infra_name
}

