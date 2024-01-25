data "duplocloud_k8s_cron_job" "job" {
  tenant_id = var.tenant_id
  metadata {
    name = "datajob"
  }
}

output "metadata" {
  value = data.duplocloud_k8s_cron_job.job.metadata[0].namespace
}
