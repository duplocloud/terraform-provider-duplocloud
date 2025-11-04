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

resource "duplocloud_azure_k8s_cluster" "ac" {
  infra_name                              = duplocloud_infrastructure.infra.infra_name
  private_cluster_enabled                 = true
  enable_workload_identity                = true
  enable_blob_csi_driver                  = true
  disable_run_command                     = true
  add_critical_taint_to_system_agent_pool = true
  enable_image_cleaner                    = true
  image_cleaner_interval_in_days          = 7
  pricing_tier                            = "Free"
  linux_admin_username                    = "kubuser"
  linux_ssh_public_key                    = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC666PWPnhOI3oc+4t4CmW6HtTKfns3uOa3ZW6EN57qti20Ln4SvoBT8mwMvwnZq6Z413Kp5MFbSdkVv1t+5ZXQ0E0NdJKM59O6bTtUriekkQoeoBgu2AU2Gmk20SbMZ/7lRJDhHYg0JM3HWup7RoL3tGEJDKmv0fZ1WYYsqGkX6Dc/XP1DfmUVwd2I41yVjDWpXY/FG9/t2tKoG4DONGOJY974C6P1cxhptWyt/yqzEU7VyOB3L/kdbhTe4Z64TEYSR57jW7GsnYBbmvX8lLTAhkIFbqENXNJHl26OcwCj4M8+HU2Y4oba7vTUxb7rcgQ0vDsYgjlK6zLzPs5mcbIzjTW4VMcXBC3bciiXlurXe+ByoEUSKiXAzgszg2aD6LlMWfS6jQwGDpnfC962RxeDv/EY8ggL7xBVTe9B8H3khbeLTQpFvDYtY1GwYq0+/911LHvdRJycP7GuEWghhSDGNmh1/MhG/Qgmqh49NYhKn1RNZkYn7ePxNkTUA7h9lyU= noname"
  network_plugin                          = "kubenet"
  active_directory_config {
    ad_tenant_id           = "<ad-tenant-id>"
    admin_group_object_ids = ["<admin-group-object-id>"]
    enable_ad              = true
    enable_rbac            = true
  }
  kubernetes_version = "1.31.5"
}
