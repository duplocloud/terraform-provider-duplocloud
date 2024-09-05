---

# Resource: duplocloud_duplo_service



A Duplo service is a microservice managed by the DuploCloud platform, which automates cloud infrastructure management. It abstracts complexities, allowing users to deploy, scale, and monitor cloud-native applications with minimal manual effort.

NOTE: For Amazon ECS services, see the `duplocloud_ecs_service` resource.


## Example Usage

### Deploy NGINX service using DuploCloud Platform's native container agent.

```terraform
# Before creating an NGINX service, you must first set up the infrastructure and tenant. Below is the resource for creating the infrastructure.
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "dev"
  cloud             = 0 # AWS Cloud
  region            = "us-east-1"
  enable_k8_cluster = false # for native container agent
  address_prefix    = "10.13.0.0/16"
}

# Use the infrastructure name as the 'plan_id' from the 'duplocloud_infrastructure' resource while creating tenant.
resource "duplocloud_tenant" "tenant" {
 account_name = "dev"
 plan_id      = duplocloud_infrastructure.infra.infra_name
}

# You will need a DuploCloud host to launch the Duplo service, so create a host using following resource configuration.
data "duplocloud_native_host_image" "image" {
  tenant_id     = duplocloud_tenant.tenant.tenant_id
  is_kubernetes = false # for native container agent
}

resource "duplocloud_aws_host" "host" {
  tenant_id     = duplocloud_tenant.tenant.tenant_id
  friendly_name = "host01"

  image_id       = data.duplocloud_native_host_image.image.image_id # get the image_id from the data source
  capacity       = "t3a.small"
  agent_platform = 0 # Duplo native container agent
  zone           = 0 # Zone A
  user_account   = duplocloud_tenant.tenant.account_name
  keypair_type   = 1
}

resource "duplocloud_duplo_service" "myservice" {
  tenant_id = duplocloud_tenant.tenant.tenant_id

  name           = "myservice"
  agent_platform = 0 # Duplo native container agent
  docker_image   = "nginx:latest"
  replicas       = 1
}
```

### Deploy NGINX service inside the 'nonprod' tenant using DuploCloud Platform's native container agent with host networking and the environment variables - NGINX_HOST and NGINX_PORT

```terraform
# Ensure the 'nonprod' tenant is already created before deploying the Nginx duplo service.
data "duplocloud_tenant" "tenant" {
  name = "nonprod"
}

# You will need a DuploCloud host to launch the Duplo service, so create a host

# Create a data source to retrieve the Machine Image ID to be used by the host
data "duplocloud_native_host_image" "image" {
  tenant_id     = data.duplocloud_tenant.tenant.id
  is_kubernetes = false # for native container agent
}

resource "duplocloud_aws_host" "host" {
  tenant_id     = data.duplocloud_tenant.tenant.id
  friendly_name = "host01"

  image_id       = data.duplocloud_native_host_image.image.image_id # get the image_id from the data source
  capacity       = "t3a.small"
  agent_platform = 0 # Duplo native container agent
  zone           = 0 # Zone A
  user_account   = data.duplocloud_tenant.tenant.name
  keypair_type   = 1
}

# Create the DuploCloud service
resource "duplocloud_duplo_service" "myservice" {
  tenant_id = data.duplocloud_tenant.tenant.id

  name           = "myservice"
  agent_platform = 0 # Duplo native container agent
  docker_image   = "nginx:latest"
  replicas       = 1

  extra_config = jsonencode({
    "NGINX_HOST" = "foo",
    "NGINX_PORT" = "8080"
  })

  # Enables host networking, and listening on ports < 1000
  other_docker_host_config = jsonencode({
    NetworkMode = "host",
    CapAdd      = ["NET_ADMIN"]
  })
}
```

### Deploy NGINX service named nginx inside the 'dev' tenant and set the resource requests and limits. Set cpu requests and limits to 200m and 300m respectively and set memory requests and limits to 100Mi and 300Mi respectively

```terraform
# Ensure the 'dev' tenant is already created before deploying the Nginx duplo service.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Assuming that a host already exists in the tenant, create a service
resource "duplocloud_duplo_service" "nginx" {
  tenant_id = data.duplocloud_tenant.tenant.id

  name           = "nginx"
  agent_platform = 7 # Duplo EKS container agent
  docker_image   = "nginx:latest"
  replicas       = 1

  other_docker_config = jsonencode({
    Resources = {
      requests = {
        cpu    = "200m"
        memory = "100Mi"
      },
      limits = {
        cpu    = "300m"
        memory = "300Mi"
      }
    }
  })
}
```

### Deploy an Nginx service named nginx inside the prod tenant and mount these environment variables from the kubernetes secrets - 1. FOO: bar 2. PING: pong

```terraform
# Ensure the 'prod' tenant is already created before deploying the Nginx duplo service.
data "duplocloud_tenant" "tenant" {
  name = "prod"
}

# Create a secret with the env vars values 1. FOO: bar 2. PING: pong if it does not exist

resource "duplocloud_k8_secret" "nginx" {
  tenant_id = data.duplocloud_tenant.tenant.id

  secret_name = "nginx-secret"
  secret_type = "Opaque"
  secret_data = jsonencode({
    FOO = "bar",
    PING = "pong"
  })
}

# Assuming that a host exists in the tenant.
resource "duplocloud_duplo_service" "nginx" {
  tenant_id = data.duplocloud_tenant.tenant.id

  name           = "nginx"
  agent_platform = 7 # Duplo EKS container agent
  docker_image   = "nginx:latest"
  replicas       = 1

  other_docker_config = jsonencode({
    EnvFrom = [
      {
        secretRef = {
          name = duplocloud_k8_secret.nginx.secret_name,
        }
      }
    ]
  })
}
```

### Deploy an Nginx service named nginx inside the dev tenant, and mount these environment variables from the kubernetes configmap - 1. FOO: bar 2. PING: pong

```terraform
# Ensure the 'dev' tenant is already created before deploying the Nginx duplo service.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Create a configmap with the env vars values 1. FOO: bar 2. PING: pong if it does not exists

resource "duplocloud_k8_config_map" "nginx" {
  tenant_id = data.duplocloud_tenant.tenant.id

  name = "nginx-cm"
  data = jsonencode({
    FOO  = "bar",
    PING = "pong"
  })
}

# Ensure that the host is also created in the tenant.
resource "duplocloud_duplo_service" "nginx" {
  tenant_id = data.duplocloud_tenant.tenant.id

  name           = "nginx"
  agent_platform = 7 # Duplo EKS container agent
  docker_image   = "nginx:latest"
  replicas       = 1

  other_docker_config = jsonencode({
    EnvFrom = [
      {
        configMapRef = {
          name = duplocloud_k8_config_map.nginx.name,
        }
      }
    ]
  })
}
```

### Deploy an Nginx service named nginx inside the dev tenant, and set the replica count to 5

```terraform
# Ensure the 'dev' tenant is already created before deploying the Nginx duplo service.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Ensure that the host is also created in the tenant.
resource "duplocloud_duplo_service" "nginx" {
  tenant_id = data.duplocloud_tenant.tenant.id

  name           = "nginx"
  agent_platform = 7 # Duplo EKS container agent
  docker_image   = "nginx:latest"
  replicas       = 5
}
```

### Deploy an Nginx service named nginx with liveliness probe. Create it inside the dev tenant which already exists.

```terraform
# Ensure the 'dev' tenant is already created before deploying the Nginx duplo service.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Assuming a host already exists in the tenant, create the duplo service
resource "duplocloud_duplo_service" "nginx" {
  tenant_id      = data.duplocloud_tenant.tenant.id
  name           = "nginx"
  agent_platform = 7 # Duplo EKS container agent
  docker_image   = "nginx:latest"
  replicas       = 1

  other_docker_config = jsonencode({
    # Liveness probe to ensure container is alive
    "LivenessProbe" : {
      "initialDelaySeconds" : 10,
      "periodSeconds" : 30,
      "successThreshold" : 1,
      "httpGet" : {
        "path" : "/health",
        "port" : 80
      }
    }
  })
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `docker_image` (String) The docker image to use for the launched container(s).
- `name` (String) The name of the service to create.
- `tenant_id` (String) The GUID of the tenant that the service will be created in.

### Optional

- `agent_platform` (Number) The numeric ID of the container agent to use for deployment.
Should be one of:

   - `0` : Duplo Native container agent
   - `7` : EKS linux container agent
 Defaults to `0`.
- `allocation_tags` (String)
- `any_host_allowed` (Boolean) Whether or not the service can run on hosts in other tenants (within the the same plan as the current tenant). Defaults to `false`.
- `cloud` (Number) The numeric ID of the cloud provider to launch the service in.
Should be one of:

   - `0` : AWS (Default)
   - `1` : Oracle
   - `2` : Azure
   - `3` : Google
   - `4` : Byoh
   - `5` : Unknown
   - `6` : DigitalOcean
   - `10` : OnPrem
 Defaults to `0`.
- `cloud_creds_from_k8s_service_account` (Boolean) Whether or not the service gets it's cloud credentials from Kubernetes service account. Defaults to `false`.
- `commands` (String)
- `extra_config` (String)
- `force_recreate_on_volumes_change` (Boolean) if 'force_recreate_on_volumes_change=true' and any changing to Volumes, will results in forceNew and hence recreating the resource. Defaults to `false`.
- `force_stateful_set` (Boolean) Whether or not to force a StatefulSet to be created. Defaults to `false`.
- `hpa_specs` (String)
- `is_daemonset` (Boolean) Whether or not to enable DaemonSet. Defaults to `false`.
- `is_unique_k8s_node_required` (Boolean) Whether or not the replicas must be scheduled on separate Kubernetes nodes.  Only supported on Kubernetes. Defaults to `false`.
- `lb_synced_deployment` (Boolean) Defaults to `false`.
- `other_docker_config` (String)
- `other_docker_host_config` (String)
- `replica_collocation_allowed` (Boolean) Allow replica collocation for the service. If this is set then 2 replicas can be on the same host.
- `replicas` (Number) The number of container replicas to deploy. Defaults to `1`.
- `replicas_matching_asg_name` (String)
- `should_spread_across_zones` (Boolean) Whether or not the replicas must be spread across availability zones.  Only supported on Kubernetes. Defaults to `false`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `volumes` (String) Volumes to be attached to pod.

### Read-Only

- `domain` (String) The service domain (whichever fqdn_ex or fqdn which is non empty)
- `fqdn` (String) The fully qualified domain associated with the service
- `fqdn_ex` (String) External fully qualified domain associated with the service
- `id` (String) The ID of this resource.
- `index` (Number) The index of the service.
- `parent_domain` (String) The service's parent domain
- `tags` (List of Object) (see [below for nested schema](#nestedatt--tags))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)


<a id="nestedatt--tags"></a>
### Nested Schema for `tags`

Read-Only:

- `key` (String)
- `value` (String)

## Import

Import is supported using the following syntax:

```shell
# Example: Importing an existing service
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the name of the service
#
terraform import duplocloud_duplo_service.myservice v2/subscriptions/*TENANT_ID*/ReplicationControllerApiV2/*NAME*
```
