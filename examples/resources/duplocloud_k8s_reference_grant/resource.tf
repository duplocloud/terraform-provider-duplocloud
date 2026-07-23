resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

// Allow HTTPRoutes in the "frontend" namespace to reference Services in this
// tenant's namespace.
resource "duplocloud_k8s_reference_grant" "grant" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "allow-frontend-routes"

  from {
    group     = "gateway.networking.k8s.io"
    kind      = "HTTPRoute"
    namespace = "frontend"
  }

  to {
    kind = "Service"
  }
}
