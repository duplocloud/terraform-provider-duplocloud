data "duplocloud_gke_credentials" "credential" {
  plan_id = "non-prod"
}

output "credential_value" {
  value = {
    ca_certificate_data = data.duplocloud_gke_credentials.credential.ca_certificate_data
    endpoint            = data.duplocloud_gke_credentials.credential.endpoint
    name                = data.duplocloud_gke_credentials.credential.name
    version             = data.duplocloud_gke_credentials.credential.version
    region              = data.duplocloud_gke_credentials.credential.region
    token               = data.duplocloud_gke_credentials.credential.token
  }
}
