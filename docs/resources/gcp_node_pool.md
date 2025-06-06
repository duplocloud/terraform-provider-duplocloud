---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_gcp_node_pool Resource - terraform-provider-duplocloud"
subcategory: ""
description: |-
  duplocloud_gcp_node_pool manages a GCP Node Pool in Duplo.
---

# duplocloud_gcp_node_pool (Resource)

`duplocloud_gcp_node_pool` manages a GCP Node Pool in Duplo.

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `image_type` (String) The image type to use for this node. Note that for a given image type, the latest version of it will be used
- `machine_type` (String) The name of a Google Compute Engine machine type.
				If unspecified, the default machine type is e2-medium.
- `name` (String) The short name of the node pool.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.
- `tenant_id` (String) The GUID of the tenant that the node pool will be created in.
- `zones` (List of String) The list of Google Compute Engine zones in which the NodePool's nodes should be located.

### Optional

- `accelerator` (Block List) (see [below for nested schema](#nestedblock--accelerator))
- `allocation_tags` (String) Allocation tag to give to the nodes 
			if specified it would be added as a label and that can be used while creating services
- `auto_repair` (Boolean) Whether the nodes will be automatically repaired.
- `auto_upgrade` (Boolean) Whether the nodes will be automatically upgraded.
- `disc_size_gb` (Number) Size of the disk attached to each node, specified in GB. The smallest allowed disk size is 10GB.
				If unspecified, the default disk size is 100GB.
- `disc_type` (String) Type of the disk attached to each node
				If unspecified, the default disk type is 'pd-standard'
- `initial_node_count` (Number) The initial node count for the pool Defaults to `1`.
- `is_autoscaling_enabled` (Boolean) Is autoscaling enabled for this node pool.
- `labels` (Map of String) The map of Kubernetes labels (key/value pairs) to be applied to each node.
- `linux_node_config` (Block List, Max: 1) Parameters that can be configured on Linux nodes (see [below for nested schema](#nestedblock--linux_node_config))
- `location_policy` (String) Update strategy of the node pool. Defaults to `BALANCED`.
- `max_node_count` (Number) Maximum number of nodes for one location in the NodePool. Must be >= minNodeCount.
- `metadata` (Map of String) The metadata key/value pairs assigned to instances in the cluster.
- `min_node_count` (Number) Minimum number of nodes for one location in the NodePool. Must be >= 1 and <= maxNodeCount.
- `node_pool_logging_config` (Block List, Max: 1) Logging configuration. (see [below for nested schema](#nestedblock--node_pool_logging_config))
- `oauth_scopes` (List of String) The set of Google API scopes to be made available on all of the node VMs under the default service account.
- `resource_labels` (Map of String) Resource labels associated to node pool
- `spot` (Boolean) Spot flag for enabling Spot VM
- `tags` (List of String) The list of instance tags applied to all nodes.
				Tags are used to identify valid sources or targets for network firewalls and are specified by the client during cluster or node pool creation.
				Each tag within the list must comply with RFC1035.
- `taints` (Block List) (see [below for nested schema](#nestedblock--taints))
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `total_max_node_count` (Number) Maximum number of nodes for one location in the NodePool. Must be >= minNodeCount.
- `total_min_node_count` (Number) Minimum number of nodes for one location in the NodePool. Must be >= 1 and <= maxNodeCount.
- `upgrade_settings` (Block List) Upgrade settings control disruption and speed of the upgrade. (see [below for nested schema](#nestedblock--upgrade_settings))

### Read-Only

- `fullname` (String) The short name of the node pool.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.
- `id` (String) The ID of this resource.

<a id="nestedblock--accelerator"></a>
### Nested Schema for `accelerator`

Optional:

- `accelerator_count` (String) The number of the accelerator cards exposed to an instance.
- `accelerator_type` (String) The accelerator type resource name.
- `gpu_driver_installation_config` (Block List) (see [below for nested schema](#nestedblock--accelerator--gpu_driver_installation_config))
- `gpu_partition_size` (String) Size of partitions to create on the GPU
- `gpu_sharing_config` (Block List) (see [below for nested schema](#nestedblock--accelerator--gpu_sharing_config))
- `max_time_shared_clients_per_gpu` (String) The number of time-shared GPU resources to expose for each physical GPU.

<a id="nestedblock--accelerator--gpu_driver_installation_config"></a>
### Nested Schema for `accelerator.gpu_driver_installation_config`

Optional:

- `gpu_driver_version` (String)


<a id="nestedblock--accelerator--gpu_sharing_config"></a>
### Nested Schema for `accelerator.gpu_sharing_config`

Optional:

- `gpu_sharing_strategy` (String) The configuration for GPU sharing options.
- `max_shared_clients_per_gpu` (String) The max number of containers that can share a physical GPU.



<a id="nestedblock--linux_node_config"></a>
### Nested Schema for `linux_node_config`

Optional:

- `cgroup_mode` (String) cgroupMode specifies the cgroup mode to be used on the node.
- `sysctls` (Map of String) The Linux kernel parameters to be applied to the nodes and all pods running on the nodes.


<a id="nestedblock--node_pool_logging_config"></a>
### Nested Schema for `node_pool_logging_config`

Optional:

- `variant_config` (Map of String)


<a id="nestedblock--taints"></a>
### Nested Schema for `taints`

Optional:

- `effect` (String) Update strategy of the node pool. Supported effect's are : 
	- EFFECT_UNSPECIFIED 
	- NO_SCHEDULE 
	- PREFER_NO_SCHEDULE
	- NO_EXECUTE
- `key` (String)
- `value` (String)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)


<a id="nestedblock--upgrade_settings"></a>
### Nested Schema for `upgrade_settings`

Optional:

- `blue_green_settings` (Block List) (see [below for nested schema](#nestedblock--upgrade_settings--blue_green_settings))
- `max_surge` (Number) The maximum number of nodes that can be created beyond the current size of the node pool during the upgrade process.
- `max_unavailable` (Number) The maximum number of nodes that can be simultaneously unavailable during the upgrade process. A node is considered available if its status is Ready
- `strategy` (String) Update strategy of the node pool.

<a id="nestedblock--upgrade_settings--blue_green_settings"></a>
### Nested Schema for `upgrade_settings.blue_green_settings`

Optional:

- `node_pool_soak_duration` (String) Note: The node_pool_soak_duration should not be used along with standard_rollout_policy
- `standard_rollout_policy` (Block List) Note: The standard_rollout_policy should not be used along with node_pool_soak_duration (see [below for nested schema](#nestedblock--upgrade_settings--blue_green_settings--standard_rollout_policy))

<a id="nestedblock--upgrade_settings--blue_green_settings--standard_rollout_policy"></a>
### Nested Schema for `upgrade_settings.blue_green_settings.standard_rollout_policy`

Optional:

- `batch_node_count` (Number) Note: The batch_node_count should not be used along with batch_percentage
- `batch_percentage` (Number) Note: The batch_percentage should not be used along with batch_node_count
- `batch_soak_duration` (String)

## Import

Import is supported using the following syntax:

```shell
# Example: Importing an existing GCP Node Pool
#  - *TENANT_ID* is the tenant GUID
#  - *FULLNAME* is the  name of the Node Pool
#

terraform import duplocloud_gcp_node_pool.node_pool *TENANT_ID*/*FULLNAME*
```
