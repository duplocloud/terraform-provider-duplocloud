resource "duplocloud_infrastructure" "myinfra" {
  infra_name        = "myinfra"
  cloud             = 0 // AWS
  region            = "us-east-1"
  azcount           = 2
  enable_k8_cluster = true
  address_prefix    = "10.34.0.0/16"
  subnet_cidr       = 24
}

resource "duplocloud_infrastructure_setting" "settings" {
  infra_name = duplocloud_infrastructure.myinfra.name

  setting {
    key   = "EnableSecretCsiDriver"
    value = "true"
  }
  setting {
    key   = "EnableAWSEfsVolumes"
    value = "true"
  }
  setting {
    key   = "EnableAwsAlbIngress"
    value = "true"
  }
}

