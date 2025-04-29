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


resource "duplocloud_k8_helm_repository" "repo" {
  name             = "repo-name"
  interval         = "06m00s"
  url              = "oci://registry-1.docker.io/bitnamicharts"
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  insecure         = true
  pass_credentials = true
  helm_provider    = "generic"
  suspend          = true
  type             = "oci"

}