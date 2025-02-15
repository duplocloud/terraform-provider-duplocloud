resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}



resource "duplocloud_k8_helm_repository" "repo" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "repo-name"
  interval  = "06m00s"
  url       = "https://helm.github.com"
}
