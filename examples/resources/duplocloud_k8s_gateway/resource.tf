resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

// Example 1: An internal (VPN-only) gateway. Duplo auto-selects the internal
// GatewayClass (e.g. gke-l7-rilb on GKE).
resource "duplocloud_k8s_gateway" "internal" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "internal-gateway"
  is_public = false

  listener {
    name     = "http"
    port     = 80
    protocol = "HTTP"
  }
}

// Example 2: A public gateway terminating TLS with a kubernetes secret.
resource "duplocloud_k8s_gateway" "public" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "public-gateway"
  is_public = true

  listener {
    name     = "https"
    port     = 443
    protocol = "HTTPS"
    hostname = "*.example.com"

    tls {
      mode = "Terminate"
      certificate_ref {
        name = "tls-secret"
      }
    }
  }
}

// Example 3: A gateway with an explicit GatewayClass, a static IP, and a
// Google Certificate Manager certificate map.
resource "duplocloud_k8s_gateway" "certmap" {
  tenant_id          = duplocloud_tenant.myapp.tenant_id
  name               = "certmap-gateway"
  gateway_class_name = "gke-l7-global-external-managed"

  address {
    type  = "IPAddress"
    value = "203.0.113.10"
  }

  listener {
    name     = "https"
    port     = 443
    protocol = "HTTPS"

    tls {
      mode            = "Terminate"
      certificate_map = "my-cert-map"
    }
  }
}
