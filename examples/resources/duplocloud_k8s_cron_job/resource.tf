resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_k8s_cron_job" "myapp" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  metadata {
    name = "jobname"
  }
  spec {
    job_template {
      spec {
        template {
          spec {
            container {
              name  = "containername"
              image = "nginx:latest"
            }
          }
        }
      }
    }
    schedule = "*/1 * * * *"
  }
}
