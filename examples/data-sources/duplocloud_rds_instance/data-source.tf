
data "duplocloud_rds_instance" "reader" {
  tenant_id = var.tenant_id
  name      = "duploread-replica"
}

output "rds_reader_host" {
  value = data.duplocloud_rds_instance.reader.host
}


data "duplocloud_rds_instance" "cluster" {
  tenant_id = var.tenant_id
  name      = "duplocluster"
}

output "rds_cluster_host" {
  value = data.duplocloud_rds_instance.cluster.host
}

