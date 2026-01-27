resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}



resource "duplocloud_k8_oci_repository" "oci" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "myrepo"
  spec {
    interval = "6m0s"
    url      = "oci://registry-1.docker.io/bitnamicharts"
    ref {
      tag = "latest"
    }
  }
}

resource "duplocloud_k8_oci_repository" "oci" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "mybitnamicharts"
  spec {
    interval                = "6m0s"
    url                     = "oci://registry-1.docker.io/bitnamicharts"
    media_type              = "application/vnd.oci.image.config.v1+json"
    operation               = "copy"
    private_registry_secret = "dockerconfig"
  }
}

