resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


resource "duplocloud_k8_helm_release" "release" {
  tenant_id    = duplocloud_tenant.myapp.tenant_id
  name         = "helm-release-name"
  interval     = "05m00s"
  release_name = "helm-release-1"
  chart {
    name               = "chart-name"
    version            = "v1"
    reconcile_strategy = "ChartVersion"
    source_type        = "HelmRepository"
    source_name        = duplocloud_k8_helm_repository.repo.name
  }
  values = jsonencode({
    "replicaCount" : 2,
    "serviceAccount" : {
      "create" : false
    }
    }
  )
}