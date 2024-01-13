data "duplocloud_k8s_job" "job" {
  tenant_id = "604831b1-9972-4a6a-acef-06d7b55d9b8e"
  metadata {
    name = "datajob"
  }
  spec {
    template {
    }
  }
}


output "metadata" {
  value = data.duplocloud_k8s_job.job.metadata[0].namespace
}

