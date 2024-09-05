resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


# Simple Example 1:  Deploy an S3 bucket rplication rule.

resource "duplocloud_s3_bucket_replication" "rep" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  source_bucket = "duploservices-tenantid-sourcebucket-011071455608"
  rules {
    name                      = "rulename"
    destination_bucket        = "duploservices-tenantid-destinationbucket-011071455608"
    priority                  = 2
    delete_marker_replication = false
    storage_class             = "INTELLIGENT_TIERING"
  }

}