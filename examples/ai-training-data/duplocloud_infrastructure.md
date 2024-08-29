### Create a DuploCloud infrastructure named nonprod with cidr 10.34.0.0/16 in us-west-2 region

```
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "nonprod"
  cloud             = 0             # AWS Cloud
  region            = "us-west-2"
  enable_k8_cluster = false
  address_prefix    = "10.34.0.0/16"
}
```

### Create a DuploCloud infrastructure named nonprod with cidr 10.30.0.0/16 in us-east-1 region with EKS cluster

```
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "nonprod"
  cloud             = 0             # AWS Cloud
  region            = "us-east-1"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = true
  address_prefix    = "10.30.0.0/16"
}
```

### Create a DuploCloud infrastructure named 'preprod' in the us-east-2 region, with a VPC CIDR of 10.30.0.0/16, a subnet mask of 24, and EKS cluster enabled

```
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "preprod"
  cloud             = 0             # AWS Cloud
  region            = "us-east-2"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = true
  address_prefix    = "10.30.0.0/16"
  subnet_cidr       = 24
}
```

### Create a DuploCloud infrastructure named 'prod' in the us-east-2 region, with a VPC CIDR of 10.30.0.0/16, a subnet mask of 24, and EKS cluster enabled

```
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0             # AWS Cloud
  region            = "us-east-2"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = true
  address_prefix    = "10.30.0.0/16"
  subnet_cidr       = 24
}
```

### Create a DuploCloud infrastructure named 'prod' in the us-east-2 region, with a VPC CIDR of 10.30.0.0/16, a subnet mask of 24, and an EKS cluster enabled with an ingress controller

```
resource "duplocloud_infrastructure" "prod_infra" {
  infra_name        = "prod"
  cloud             = 0             # AWS Cloud
  region            = "us-east-2"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = true
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

```
resource "duplocloud_infrastructure" "nonprod_infra" {
  infra_name        = "nonprod"
  cloud             = 0             # AWS Cloud
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

```
resource "duplocloud_infrastructure" "nonprod_infra" {
  infra_name        = "nonprod"
  cloud             = 0             # AWS Cloud
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

```
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0             # AWS Cloud
  region            = "us-east-2"
  azcount           = 2             # The number of availability zones.
  enable_ecs_cluster = true
  address_prefix    = "10.50.0.0/16"
  subnet_cidr       = 22
}
```

### Create a DuploCloud infrastructure named 'prod' in the us-east-2 region, with a VPC CIDR of 10.49.0.0/16, a subnet mask of 24, and EKS, ECS cluster enabled

```
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0             # AWS Cloud
  region            = "us-east-2"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = true
  enable_ecs_cluster= true
  address_prefix    = "10.49.0.0/16"
  subnet_cidr       = 24
}
```

### Set up a DuploCloud infrastructure named 'nonprod' in the us-west-2 region, with a VPC CIDR of 10.60.0.0/16, a subnet mask of 24, and an EKS cluster configured with an autoscaler, ingress controller, and Secrets Store CSI Driver.

```
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
    }
    {
      key   = "EnableSecretCsiDriver"
      value = true
    }
  ]
}

resource "duplocloud_infrastructure" "nonprod_infra" {
  infra_name        = "nonprod"
  cloud             = 0             # AWS Cloud
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

### Provision a DuploCloud infrastructure named 'prod' in the us-east-2 region, with a VPC CIDR of 10.49.0.0/16, a subnet mask of 24, and enable EKS and ECS clusters.

```
variable "infra_config" {
  type = object({
    name     = string
    cloud    = number
    region   = string
    azcount  = number
    address_prefix = string
    subnet_cidr = number
  })

  default = {
    name     = "prod"
    cloud    = 0
    region   = "us-east-2"
    azcount  = 2
    address_prefix = "10.49.0.0/16"
    subnet_cidr = 24
  }
}

resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0             # AWS Cloud
  region            = "us-east-2"
  azcount           = 2             # The number of availability zones.
  enable_k8_cluster = true
  enable_ecs_cluster= true
  address_prefix    = "10.49.0.0/16"
  subnet_cidr       = 24
}
```