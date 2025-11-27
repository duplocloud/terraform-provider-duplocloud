data "duplocloud_native_host_image" "img" {
  tenant_id   = "f4bf01f0-5077-489e-aa51-95fb77049608"
  name        = "EKS-Oregon-1.32"
  os          = "AmazonLinux2023"
  k8s_version = "1.32"
  arch        = "amd64"
}

