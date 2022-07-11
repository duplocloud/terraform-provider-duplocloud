resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {

  tenant_id      = duplocloud_tenant.myapp.tenant_id
  name           = "tst-dynamodb-table"
  read_capacity  = 10
  write_capacity = 10
  #billing_mode = "PAY_PER_REQUEST"
  tag {
    key   = "CreatedBy"
    value = "Duplo"
  }

  tag {
    key   = "CreatedFrom"
    value = "Duplo"
  }

  attribute {
    name = "UserId"
    type = "S"
  }

  attribute {
    name = "GameTitle"
    type = "S"
  }

  attribute {
    name = "TopScore"
    type = "N"
  }

  key_schema {
    attribute_name = "UserId"
    key_type       = "HASH"
  }

  key_schema {
    attribute_name = "GameTitle"
    key_type       = "RANGE"
  }

  global_secondary_index {
    name               = "GameTitleIndex"
    hash_key           = "GameTitle"
    range_key          = "TopScore"
    write_capacity     = 10
    read_capacity      = 10
    projection_type    = "INCLUDE"
    non_key_attributes = ["UserId"]
  }

}
