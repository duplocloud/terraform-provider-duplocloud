resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Simple Example 1:  Deploy NGINX using Duplo's native container agent
resource "duplocloud_duplo_service" "myservice" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  name           = "myservice"
  agent_platform = 0 # Duplo native container agent
  docker_image   = "nginx:latest"
  replicas       = 1
}

# Simple Example 2:  Deploy NGINX using Duplo's EKS agent
resource "duplocloud_duplo_service" "myservice" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  name           = "myservice"
  agent_platform = 7 # Duplo EKS agent
  docker_image   = "nginx:latest"
  replicas       = 1
}

# Example 3:  Deploy NGINX with host networking and env vars, using Duplo's native container agent
resource "duplocloud_duplo_service" "myservice" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  name           = "myservice"
  agent_platform = 0 # Duplo native container agent
  docker_image   = "nginx:latest"
  replicas       = 1

  extra_config = jsonencode({
    "NGINX_HOST" = "foo",
    "NGINX_PORT" = "8080"
  })

  // Enables host networking, and listening on ports < 1000
  other_docker_host_config = jsonencode({
    NetworkMode = "host",
    CapAdd      = ["NET_ADMIN"]
  })
}

# Example 4:  Deploy NGINX with host networking and env vars, using Duplo's EKS agent
resource "duplocloud_duplo_service" "myservice" {
  tenant_id = duplocloud_tenant.myapp.tenant_id

  name           = "myservice"
  agent_platform = 7 # Duplo EKS agent
  docker_image   = "nginx:latest"
  replicas       = 1

  other_docker_config = jsonencode({
    HostNetwork = true,
    Env = [
      { Name = "NGINX_HOST", Value = "foo" },
    ]
  })
}

# Example 5:  Deploy NGINX with resources and horizontal pod autoscaling, using Duplo's EKS agent

resource "duplocloud_duplo_service" "hpa" {
  tenant_id      = duplocloud_tenant.myapp.tenant_id
  name           = "hpa"
  replicas       = 1
  agent_platform = 7
  other_docker_config = jsonencode({
    "resources" : {
      "limits" : {
        "cpu" : "500m",
        "memory" : "1Gi"
      },
      "requests" : {
        "cpu" : "500m",
        "memory" : "1Gi"
      }
    }
    }
  )
  docker_image = "nginx:latest"
  hpa_specs = jsonencode({
    "maxReplicas" : 3,
    "metrics" : [
      {
        "resource" : {
          "name" : "cpu",
          "target" : {
            "averageUtilization" : 10,
            "type" : "Utilization"
          }
        },
        "type" : "Resource"
      }
    ],
    "minReplicas" : 1
    }
  )
}

