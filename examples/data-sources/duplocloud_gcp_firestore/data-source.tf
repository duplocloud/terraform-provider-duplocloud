data "duplocloud_gcp_firestore" "app" {
  tenant_id = "tenant_id"
  name      = "name"
}

output "out" {
  value = {
    name                          = data.duplocloud_gcp_firestore.app.name
    type                          = data.duplocloud_gcp_firestore.app.type
    location_id                   = data.duplocloud_gcp_firestore.app.location_id
    enable_delete_protection      = data.duplocloud_gcp_firestore.app.enable_delete_protection
    enable_point_in_time_recovery = data.duplocloud_gcp_firestore.app.enable_point_in_time_recovery
    etag                          = data.duplocloud_gcp_firestore.app.etag
    uid                           = data.duplocloud_gcp_firestore.app.uid
    version_retention_period      = data.duplocloud_gcp_firestore.app.version_retention_period
    earliest_version_time         = data.duplocloud_gcp_firestore.app.earliest_version_time
    concurrency_mode              = data.duplocloud_gcp_firestore.app.concurrency_mode
    app_engine_integration_mode   = data.duplocloud_gcp_firestore.app.app_engine_integration_mode
  }
}
