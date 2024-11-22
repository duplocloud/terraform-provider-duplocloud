### Create a DuploCloud onpremise infrastructure named onprem with eks vendor

#### Solution:

resource "duplocloud_infrastructure_onprem" "infra" {
  infra_name                         = "onprem"
  cluster_name                       = "onprem"
  region                             = "us-west-2"
  azcount                            = 2
  enable_k8_cluster                  = true
  vendor                             = 2
  cluster_endpoint                   = "https://BB3C2589BAE34AD680060B5FDBA12BA1.gr7.us-west-2.eks.amazonaws.com"
  api_token                          = "<api-token>"
  cluster_certificate_authority_data = "<certificate-authority-data>"
  data_center                        = "us"
  eks_config {
    private_subnets            = ["subnet-06c1b3a338ace60ce", "subnet-09252308e1a093bda"]
    public_subnets             = ["subnet-0d5b3c3a3ae9d129f", "subnet-065ab3e894092dd1c"]
    ingress_security_group_ids = ["sg-0331e348b886ed796"]
    vpc_id                     = "vpc-0961fc6b0903ad63f"
  }
  custom_data {
    key   = "K8sVersion"
    value = "1.31"
  }
}
