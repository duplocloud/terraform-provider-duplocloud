resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# A production instance with a single, manually-scaled cluster.
resource "duplocloud_gcp_bigtable_instance" "bigtable-demo" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  name          = "bigtable-demo"
  display_name  = "bigtable-demo"
  instance_type = "PRODUCTION"
  storage_type  = "SSD"

  cluster {
    cluster_id = "bigtable-demo-c1"
    zone       = "us-east1-b"
    num_nodes  = 1
  }
}

# An instance with an autoscaled cluster and a second replicated cluster.
resource "duplocloud_gcp_bigtable_instance" "bigtable-autoscale" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  name          = "bigtable-autoscale"
  display_name  = "bigtable-autoscale"
  instance_type = "PRODUCTION"
  storage_type  = "SSD"

  cluster {
    cluster_id = "bigtable-autoscale-c1"
    zone       = "us-east1-b"

    autoscaling_config {
      min_nodes  = 1
      max_nodes  = 3
      cpu_target = 60
    }
  }

  cluster {
    cluster_id = "bigtable-autoscale-c2"
    zone       = "us-east1-c"
    num_nodes  = 1
  }
}
