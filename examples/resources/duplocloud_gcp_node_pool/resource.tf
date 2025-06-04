resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

#Example for upgrade strategy NODE_POOL_UPDATE_STRATEGY_UNSPECIFIED
resource "duplocloud_gcp_node_pool" "myNodePool" {
  tenant_id              = duplocloud_tenant.myapp.tenant_id
  name                   = "mynodepool"
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
  zones           = ["us-east1-c"]
  location_policy = "BALANCED"
  auto_upgrade    = true
  image_type      = "cos_containerd"
  machine_type    = "n2-highcpu-32"
  disc_type       = "pd-standard"
  disc_size_gb    = 100
}


resource "duplocloud_gcp_node_pool" "core" {
  tenant_id              = duplocloud_tenant.myapp.tenant_id
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


resource "duplocloud_gcp_node_pool" "core" {
  tenant_id              = duplocloud_tenant.myapp.tenant_id
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
  resource_labels = {
    test1 = "a"
    test2 = "b"
    test3 = "c"
  }
  upgrade_settings {
    strategy  = "SURGE"
    max_surge = 1
  }
  upgrade_settings {
    strategy        = "BLUE_GREEN"
    max_surge       = 2
    max_unavailable = 1
    blue_green_settings {
      standard_rollout_policy {
        batch_percentage    = 0.1
        batch_soak_duration = "10s"

      }
    }
  }
  taints {
    key    = "taint-key"
    value  = "taint-value"
    effect = "NO_EXECUTE"
  }
  tags = [
    "environment",
    "team",
  ]


  linux_node_config {
    cgroup_mode = "CGROUP_MODE_UNSPECIFIED"
  }
  metadata = {
    "abc"  = "xyz"
    "dabc" = "dxyz"
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
