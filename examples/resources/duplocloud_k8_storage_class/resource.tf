locals {
  tenant_id = "3a0b2ea5-7403-4765-ad6e-8771ca8fa0fd"
}

resource "duplocloud_k8_storage_class" "sc" {
  tenant_id              = local.tenant_id
  name                   = "sc"
  storage_provisioner    = "efs.csi.aws.com"
  reclaim_policy         = "Delete"
  volume_binding_mode    = "Immediate"
  allow_volume_expansion = false
  parameters = {
    fileSystemId     = "fs-0d2f79aca4712c6e8",
    basePath         = "/dynamic_provisioning",
    directoryPerms   = "700",
    gidRangeStart    = "1000",
    gidRangeEnd      = "2000",
    provisioningMode = "efs-ap",
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
