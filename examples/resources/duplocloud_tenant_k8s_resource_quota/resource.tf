locals {
  tenant_id = "913a4498-db09-42c0-95b1-88ed26d87b83"
}


resource "duplocloud_tenant_k8s_resource_quota" "quota" {
  tenant_id = local.tenant_id
  name      = "kubequota"
  resource_quota = jsonencode({
    "requests.memory" : "4Gi",
    "pods" : "2"
    }
  )
  scope_selector = jsonencode({
    "matchExpressions" : [
      {
        "operator" : "In",
        "scopeName" : "PriorityClass",
        "values" : [
          "middle"
        ]
      }
    ]
  })
}
