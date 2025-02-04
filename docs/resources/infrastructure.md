---

# Resource: duplocloud_infrastructure



`duplocloud_infrastructure` manages an infrastructure in Duplo.<p>**DuploCloud infrastructure** refers to the cloud resources and configurations managed within the DuploCloud platform.It includes the setup, organization, and management of cloud services like networks, compute instances, databases, and other cloud-native services within a specific environment or tenant.</p>


## Example Usage

### Create a DuploCloud infrastructure named nonprod

```terraform
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "nonprod"
  cloud             = 0             # 0-AWS, 1- Oracle, 2- Azure, 3-Google; Defaults to 0
  region            = "us-west-2"
  enable_k8_cluster = true
  address_prefix    = "10.11.0.0/16"
}
```

### Create a DuploCloud infrastructure named nonprod with cidr 10.34.0.0/16 in us-west-2 region

```terraform
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "nonprod"
  cloud             = 0             # 0-AWS, 1- Oracle, 2- Azure, 3-Google
  region            = "us-west-2"
  enable_k8_cluster = true
  address_prefix    = "10.34.0.0/16"
}
```

### Create a DuploCloud infrastructure named nonprod with cidr 10.30.0.0/16 in us-east-1 region with EKS cluster

```terraform
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "nonprod"
  cloud             = 0             # 0-AWS, 1- Oracle, 2- Azure, 3-Google
  region            = "us-east-1"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = true
  address_prefix    = "10.30.0.0/16"
}
```

### Create a DuploCloud infrastructure named 'prod' in the us-east-2 region, with a VPC CIDR of 10.30.0.0/16, a subnet mask of 24, and EKS cluster enabled

```terraform
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0             # 0-AWS, 1- Oracle, 2- Azure, 3-Google
  region            = "us-east-2"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = true
  address_prefix    = "10.30.0.0/16"
  subnet_cidr       = 24
}
```

### Create a DuploCloud infrastructure named 'prod' in the us-east-2 region, with a VPC CIDR of 10.30.0.0/16, a subnet mask of 24, and an EKS cluster disabled with an ingress controller

```terraform
resource "duplocloud_infrastructure" "prod_infra" {
  infra_name        = "prod"
  cloud             = 0             # 0-AWS, 1- Oracle, 2- Azure, 3-Google
  region            = "us-east-2"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = false
  address_prefix    = "10.30.0.0/16"
  subnet_cidr       = 24
}

resource "duplocloud_infrastructure_setting" "settings" {
  infra_name = duplocloud_infrastructure.prod_infra.infra_name

  setting {
    key   = "EnableAwsAlbIngress"
    value = "true"
  }
}

```

### Create a DuploCloud infrastructure named 'nonprod' in the us-west-2 region, with a VPC CIDR of 10.60.0.0/16, a subnet mask of 24, and an EKS cluster enabled with an autoscaler

```terraform
resource "duplocloud_infrastructure" "nonprod_infra" {
  infra_name        = "nonprod"
  cloud             = 0             # 0-AWS, 1- Oracle, 2- Azure, 3-Google
  region            = "us-west-2"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = true
  address_prefix    = "10.60.0.0/16"
  subnet_cidr       = 24
}

resource "duplocloud_infrastructure_setting" "nonprod_settings" {
  infra_name = duplocloud_infrastructure.nonprod_infra.infra_name

  setting {
    key   = "EnableClusterAutoscaler"
    value = "true"
  }
}

```

### Create a DuploCloud infrastructure named 'nonprod' in the us-west-2 region, with a VPC CIDR of 10.60.0.0/16, a subnet mask of 24, and an EKS cluster enabled with an autoscaler, ingress controller, and Secrets Store CSI Driver

```terraform
resource "duplocloud_infrastructure" "nonprod_infra" {
  infra_name        = "nonprod"
  cloud             = 0             # 0-AWS, 1- Oracle, 2- Azure, 3-Google
  region            = "us-west-2"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = true
  address_prefix    = "10.60.0.0/16"
  subnet_cidr       = 24
}

resource "duplocloud_infrastructure_setting" "nonprod_settings" {
  infra_name = duplocloud_infrastructure.nonprod_infra.infra_name

  setting {
    key   = "EnableClusterAutoscaler"
    value = "true"
  }
  setting {
    key   = "EnableAwsAlbIngress"
    value = "true"
  }
  setting {
    key   = "EnableSecretCsiDriver"
    value = "true"
  }
}

```

### Create a DuploCloud infrastructure named 'prod' in the us-east-2 region, with a VPC CIDR of 10.50.0.0/16, a subnet mask of 22, and ECS cluster enabled

```terraform
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0             # 0-AWS, 1- Oracle, 2- Azure, 3-Google
  region            = "us-east-2"
  azcount           = 2             # The number of availability zones.
  enable_ecs_cluster = true
  address_prefix    = "10.50.0.0/16"
  subnet_cidr       = 22
}
```

### Create a DuploCloud infrastructure named 'prod' in the us-east-2 region, with a VPC CIDR of 10.49.0.0/16, a subnet mask of 24, and EKS, ECS cluster enabled

```terraform
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0             # 0-AWS, 1- Oracle, 2- Azure, 3-Google
  region            = "us-east-2"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = true
  enable_ecs_cluster= true
  address_prefix    = "10.49.0.0/16"
  subnet_cidr       = 24
}
```

### Set up a DuploCloud infrastructure named 'nonprod' in the us-west-2 region, with a VPC CIDR of 10.60.0.0/16, a subnet mask of 24, and an EKS cluster configured with an autoscaler, ingress controller, and Secrets Store CSI Driver using variables and dynamic blocks.

```terraform
variable "infra_settings" {
  type = list(object({
    key   = string
    value = string
  }))

  default = [{
    key   = "EnableAwsAlbIngress"
    value = "true"
    },
    {
      key   = "EnableClusterAutoscaler"
      value = true
    },
    {
      key   = "EnableSecretCsiDriver"
      value = true
    }
  ]
}

resource "duplocloud_infrastructure" "nonprod_infra" {
  infra_name        = "nonprod"
  cloud             = 0             # 0-AWS, 1- Oracle, 2- Azure, 3-Google
  region            = "us-west-2"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = true
  address_prefix    = "10.60.0.0/16"
  subnet_cidr       = 22
}

resource "duplocloud_infrastructure_setting" "nonprod_settings" {
  infra_name = duplocloud_infrastructure.nonprod_infra.infra_name

  dynamic "setting" {
    for_each = var.infra_settings
    content {
      key   = setting.value["key"]
      value = setting.value["value"]
    }
  }
}

```

### Provision a DuploCloud infrastructure named 'prod' in the us-east-2 region, with a VPC CIDR of 10.49.0.0/16, a subnet mask of 24, and enable EKS and ECS clusters using all attribute values from variables.

```terraform
variable "infra_config" {
  type = object({
    name           = string
    cloud          = number
    region         = string
    azcount        = number
    address_prefix = string
    subnet_cidr    = number
  })

  default = {
    name           = "prod"
    cloud          = 0
    region         = "us-east-2"
    azcount        = 2
    address_prefix = "10.49.0.0/16"
    subnet_cidr    = 24
  }
}

resource "duplocloud_infrastructure" "infra" {
  infra_name         = var.infra_config["name"]
  cloud              = var.infra_config["cloud"]
  region             = var.infra_config["region"]
  azcount            = var.infra_config["azcount"]
  enable_k8_cluster  = true
  enable_ecs_cluster = true
  address_prefix     = var.infra_config["address_prefix"]
  subnet_cidr        = var.infra_config["subnet_cidr"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `address_prefix` (String) The CIDR to use for the VPC or VNet.
- `enable_k8_cluster` (Boolean) Whether or not to provision a kubernetes cluster.
- `infra_name` (String) The name of the infrastructure.  Infrastructure names are globally unique and less than 13 characters.
- `region` (String) The cloud provider region.  The Duplo portal must have already been configured to support this region.

### Optional

- `account_id` (String) The cloud account ID.
- `azcount` (Number) The number of availability zones.  Must be one of: `2`, `3`, or `4`. This is applicable only for AWS.
- `cloud` (Number) The numerical index of cloud provider to use for the infrastructure.
Should be one of:

   - `0` : AWS
   - `2` : Azure
   - `3` : Google
 Defaults to `0`.
- `cluster_ip_cidr` (String) cluster IP CIDR defines a private IP address range used for internal Kubernetes services.
- `custom_data` (Block List, Deprecated) A list of configuration settings to apply on creation, expressed as key / value pairs. The custom_data argument is only applied on creation, and is deprecated in favor of the settings argument. (see [below for nested schema](#nestedblock--custom_data))
- `delete_unspecified_settings` (Boolean) Whether or not this resource should delete any settings not specified by this resource. **WARNING:**  It is not recommended to change the default value of `false`. Defaults to `false`.
- `disable_public_subnet` (Boolean) This flag is used to disable public subnet, applicable only for AWS cloud infra provisioning.
- `enable_container_insights` (Boolean) Whether or not to enable container insights for an ECS cluster.
- `enable_ecs_cluster` (Boolean) Whether or not to provision an ECS cluster.
- `is_serverless_kubernetes` (Boolean) Whether or not to make GKE with autopilot.
- `setting` (Block List) A list of configuration settings to manage, expressed as key / value pairs. (see [below for nested schema](#nestedblock--setting))
- `subnet_address_prefix` (String) The address prefixe to use for the subnet. This is applicable only for Azure
- `subnet_cidr` (Number) The CIDR subnet size (in bits) for the automatically created subnets. This is applicable only for AWS.
- `subnet_name` (String) The name of the subnet. This is applicable only for Azure.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `wait_until_deleted` (Boolean) Whether or not to wait until Duplo has destroyed the infrastructure. Defaults to `false`.

### Read-Only

- `all_settings` (List of Object) A complete list of configuration settings for this infrastructure, even ones not being managed by this resource. (see [below for nested schema](#nestedatt--all_settings))
- `id` (String) The ID of this resource.
- `private_subnets` (Set of Object) The private subnets for the VPC or VNet. (see [below for nested schema](#nestedatt--private_subnets))
- `public_subnets` (Set of Object) The public subnets for the VPC or VNet. (see [below for nested schema](#nestedatt--public_subnets))
- `security_groups` (Set of Object) The security groups for the VPC or VNet. (see [below for nested schema](#nestedatt--security_groups))
- `specified_settings` (List of String) A list of configuration setting key being managed by this resource.
- `status` (String) The status of the infrastructure.
- `subnet_fullname` (String) The full name of the subnet. This is applicable only for Azure.
- `vpc_id` (String) The VPC or VNet ID.
- `vpc_name` (String) The VPC or VNet name.

<a id="nestedblock--custom_data"></a>
### Nested Schema for `custom_data`

Required:

- `key` (String)
- `value` (String)


<a id="nestedblock--setting"></a>
### Nested Schema for `setting`

Required:

- `key` (String)
- `value` (String)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)


<a id="nestedatt--all_settings"></a>
### Nested Schema for `all_settings`

Read-Only:

- `key` (String)
- `value` (String)


<a id="nestedatt--private_subnets"></a>
### Nested Schema for `private_subnets`

Read-Only:

- `cidr_block` (String)
- `id` (String)
- `name` (String)
- `tags` (List of Object) (see [below for nested schema](#nestedobjatt--private_subnets--tags))
- `type` (String)
- `zone` (String)

<a id="nestedobjatt--private_subnets--tags"></a>
### Nested Schema for `private_subnets.tags`

Read-Only:

- `key` (String)
- `value` (String)



<a id="nestedatt--public_subnets"></a>
### Nested Schema for `public_subnets`

Read-Only:

- `cidr_block` (String)
- `id` (String)
- `name` (String)
- `tags` (List of Object) (see [below for nested schema](#nestedobjatt--public_subnets--tags))
- `type` (String)
- `zone` (String)

<a id="nestedobjatt--public_subnets--tags"></a>
### Nested Schema for `public_subnets.tags`

Read-Only:

- `key` (String)
- `value` (String)



<a id="nestedatt--security_groups"></a>
### Nested Schema for `security_groups`

Read-Only:

- `id` (String)
- `name` (String)
- `read_only` (Boolean)
- `rules` (Set of Object) (see [below for nested schema](#nestedobjatt--security_groups--rules))
- `type` (String)

<a id="nestedobjatt--security_groups--rules"></a>
### Nested Schema for `security_groups.rules`

Read-Only:

- `action` (String)
- `destination_rule_type` (Number)
- `direction` (String)
- `priority` (Number)
- `protocol` (String)
- `source_address_prefix` (String)
- `source_port_range` (String)
- `source_rule_type` (Number)

## Import

Import is supported using the following syntax:

```
# Example: Importing an existing infrastructure
#  - *NAME* is the infrastructure name
#
terraform import duplocloud_infrastructure.myinfra v2/admin/InfrastructureV2/*NAME*
```