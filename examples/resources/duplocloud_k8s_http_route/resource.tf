resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_k8s_gateway" "gateway" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "my-gateway"
  is_public = false

  listener {
    name     = "http"
    port     = 80
    protocol = "HTTP"
  }
}

// Example 1: Route all traffic for a hostname to a single backend service.
resource "duplocloud_k8s_http_route" "basic" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "basic-route"

  parent_ref {
    name = duplocloud_k8s_gateway.gateway.name
  }

  hostnames = ["app.example.com"]

  rule {
    match {
      path {
        type  = "PathPrefix"
        value = "/"
      }
    }
    backend_ref {
      name = "my-service"
      port = 8080
    }
  }
}

// Example 2: Path-based routing with weighted backends and a redirect filter.
resource "duplocloud_k8s_http_route" "advanced" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "advanced-route"

  parent_ref {
    name         = duplocloud_k8s_gateway.gateway.name
    section_name = "http"
  }

  rule {
    match {
      path {
        type  = "PathPrefix"
        value = "/api"
      }
    }
    backend_ref {
      name   = "api-v1"
      port   = 8080
      weight = 90
    }
    backend_ref {
      name   = "api-v2"
      port   = 8080
      weight = 10
    }
  }

  rule {
    match {
      path {
        type  = "PathPrefix"
        value = "/old"
      }
    }
    backend_ref {
      name = "my-service"
      port = 8080
    }
    filters = jsonencode([
      {
        type = "RequestRedirect"
        requestRedirect = {
          path = {
            type               = "ReplacePrefixMatch"
            replacePrefixMatch = "/new"
          }
          statusCode = 301
        }
      }
    ])
  }
}
