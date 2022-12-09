locals {
  tenant_id = "3a0b2ea5-7403-4765-ad6e-8771ca8fa0fd"
}

resource "duplocloud_k8_persistent_volume_claim" "pvc" {
  tenant_id = local.tenant_id
  name      = "pvc"
  spec {
    access_modes       = ["ReadWriteMany"]
    volume_mode        = "Filesystem"
    storage_class_name = "duploservices-dev02-sc"
    resources {
      limits = {
        storage = "20Gi"
      }
      requests = {
        storage = "10Gi"
      }
    }
  }
  annotations = {
    "a1" = "v1"
    "a2" = "v2"
  }
  labels = {
    "l1" = "v1"
    "l2" = "v2"
  }
}
