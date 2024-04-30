resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_gcp_firestore" "app" {
  tenant_id                     = duplocloud_tenant.myapp.tenant_id
  name                          = "firestore-tf-2"
  type                          = "FIRESTORE_NATIVE"
  location_id                   = "us-west2"
  enable_delete_protection      = false
  enable_point_in_time_recovery = false
}

resource "duplocloud_gcp_firestore" "firestore-app" {

}