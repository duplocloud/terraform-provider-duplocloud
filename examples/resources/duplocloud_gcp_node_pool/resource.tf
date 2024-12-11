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


resource "duplocloud_gcp_node_pool" "core" {
  tenant_id              = "afe2dfab-d47b-4abb-85af-13da6f1423cf"
  name                   = "core1"
  is_autoscaling_enabled = true
  min_node_count         = 2
  initial_node_count     = 2
  max_node_count         = 5
  zones                  = ["us-west2-c"]
  location_policy        = "BALANCED"
  auto_upgrade           = true
  auto_repair            = true
  image_type             = "cos_containerd"
  machine_type           = "e2-standard-4"
  disc_type              = "pd-standard"
  disc_size_gb           = 200
  oauth_scopes           = local.node_scopes
  labels = {
    galileo-node-type = "galileo-core"
  }

  node_pool_logging_config {
    variant_config = {
      variant = "MAX_THROUGHPUT"
    }
  }
  tags = ["avxcsd"]
  resource_labels = {
    test1 = "a"
    test2 = "b"
    test3 = "c"
  }
  upgrade_settings {
    strategy  = "SURGE"
    max_surge = 1
  }
}


locals {
  node_scopes = [
    "https://www.googleapis.com/auth/devstorage.read_write",
    "https://www.googleapis.com/auth/logging.write",
    "https://www.googleapis.com/auth/monitoring",
    "https://www.googleapis.com/auth/servicecontrol",
    "https://www.googleapis.com/auth/service.management.readonly",
    "https://www.googleapis.com/auth/trace.append",
  ]
}
