resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


resource "duplocloud_gcp_node_pool" "myNodePool" {
  tenant_id              = duplocloud_tenant.myapp.tenant_id
  name                   = "myNodePool"
  is_autoscaling_enabled = false
  accelerator {
    accelerator_count  = "2"
    accelerator_type   = "nvidia-tesla-p100"
    gpu_partition_size = ""
    gpu_sharing_config {
      gpu_sharing_strategy       = "GPU_SHARING_STRATEGY_UNSPECIFIED"
      max_shared_clients_per_gpu = "2"
    }
    gpu_driver_installation_config {
      gpu_driver_version = "DEFAULT"
    }
  }
  upgrade_settings {
    strategy        = "NODE_POOL_UPDATE_STRATEGY_UNSPECIFIED"
    max_surge       = 4
    max_unavailable = 2
  }
  zones           = ["us-east1-c"]
  location_policy = "BALANCED"
  auto_upgrade    = true
  image_type      = "cos_containerd"
  machine_type    = "n2-highcpu-32"
  disc_type       = "pd-standard"
  disc_size_gb    = 100
}

resource "duplocloud_gcp_node_pool" "node_pool" {

}

