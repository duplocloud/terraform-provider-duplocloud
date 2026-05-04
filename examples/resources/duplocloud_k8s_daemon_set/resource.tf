resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_k8s_daemon_set" "myapp" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  metadata {
    name = "daemonsetname"
  }
  spec {
    selector {
      match_labels = {
        app = "myapp"
      }
    }
    template {
      metadata {
        labels = {
          app = "myapp"
        }
      }
      spec {
        container {
          name  = "containername"
          image = "nginx:latest"
        }
      }
    }
    update_strategy {
      type = "RollingUpdate"
      rolling_update {
        max_unavailable = "1"
        max_surge       = "0"
      }
    }
  }
}
