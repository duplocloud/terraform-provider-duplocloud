data "duplocloud_tenant" "grantor" {
  name = "tenant1"
}

data "duplocloud_tenant" "grantee" {
  name = "tenant2"
}

resource "duplocloud_tenant_access_grant" "dynamodbGrant" {
  grantee_tenant_id = data.duplocloud_tenant.grantee.id
  grantor_tenant_id = data.duplocloud_tenant.grantor.id
  grant_area        = "dynamodb"
}