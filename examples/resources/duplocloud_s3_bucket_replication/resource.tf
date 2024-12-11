resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


# Simple Example 1:  Deploy an S3 bucket rplication rule.

resource "duplocloud_s3_bucket_replication" "rep" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  source_bucket = "duploservices-tenantname-sourcebucket-011071455608"
  rules {
    name                      = "rulename"
    destination_bucket        = "duploservices-tenantname-destinationbucket-011071455608"
    priority                  = 2
    delete_marker_replication = false
    storage_class             = "INTELLIGENT_TIERING"
  }

}

# Simple Example 2: Deploy multiple S3 bucket replication rule for a source bucket

resource "duplocloud_s3_bucket_replication" "rep" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  source_bucket = "duploservices-tenantname-src-182680712604"
  rules {
    name                      = "tfruleA"
    destination_bucket        = "duploservices-tenantname-dest1-182680712604"
    priority                  = 2
    delete_marker_replication = false
    storage_class             = "STANDARD"
  }

}

resource "duplocloud_s3_bucket_replication" "rep1" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  source_bucket = "duploservices-tenantname-src-182680712604"
  rules {
    name                      = "tfruleB"
    destination_bucket        = "duploservices-tenantname-dest2-182680712604"
    priority                  = 1 #priority should not conflict
    delete_marker_replication = false
    storage_class             = "STANDARD"
  }
}
