resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {

  tenant_id    = duplocloud_tenant.myapp.tenant_id
  name         = "tst-dynamodb-table"
  billing_mode = "PAY_PER_REQUEST"
  tag {
    key   = "CreatedBy"
    value = "Duplo"
  }

  tag {
    key   = "CreatedFrom"
    value = "Duplo"
  }

  attribute {
    name = "testattr"
    type = "S"
  }

  attribute {
    name = "annotherattr"
    type = "N"
  }

  key_schema {
    attribute_name = "testattr"
    key_type       = "HASH"
  }

  key_schema {
    attribute_name = "annotherattr"
    key_type       = "RANGE"
  }

}
