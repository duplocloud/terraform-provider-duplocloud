resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Minimal example
resource "duplocloud_aws_elasticsearch" "sample" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  name      = "sample"
}

# Example with hardened settings
resource "duplocloud_aws_elasticsearch" "es-doc" {
  tenant_id                      = duplocloud_tenant.myapp.tenant_id
  name                           = "es-doc"
  enable_node_to_node_encryption = true
  require_ssl                    = true
  use_latest_tls_cipher          = true
}


resource "duplocloud_aws_elasticsearch" "essample" {
  tenant_id             = duplocloud_tenant.myapp.tenant_id
  name                  = "essamp"
  selected_zone         = 1
  elasticsearch_version = "OpenSearch_2.3"
  ebs_options {
    ebs_enabled = true
    volume_type = "gp2"
    volume_size = 10
  }
}

resource "duplocloud_aws_elasticsearch" "essample2" {
  tenant_id             = duplocloud_tenant.myapp.tenant_id
  name                  = "essamp2"
  selected_zone         = 1
  elasticsearch_version = "OpenSearch_3.3"
  cluster_config {
    instance_type          = "or2.medium.search"
    dedicated_master_type  = "or2.medium.search"
    dedicated_master_count = 3
  }
  ebs_options {
    ebs_enabled = true
    volume_type = "gp3"
    volume_size = 100
    iops        = 3000
  }
  encrypt_at_rest {
    kms_key_id = "<kms-arn>"
  }

}
