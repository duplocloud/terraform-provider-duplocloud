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

resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {

  tenant_id                   = duplocloud_tenant.myapp.tenant_id
  name                        = "mytable"
  read_capacity               = 80
  write_capacity              = 40
  billing_mode                = "PROVISIONED"
  is_point_in_time_recovery   = false
  deletion_protection_enabled = false

  tag {
    key   = "school"
    value = "admission"
  }
  attribute {
    name = "ForumName"
    type = "S"
  }
  attribute {
    name = "Subject"
    type = "S"
  }
  attribute {
    name = "LastPostDateTime"
    type = "S"
  }
  attribute {
    name = "PostMonth"
    type = "S"
  }

  attribute {
    name = "GamerZone"
    type = "S"
  }
  attribute {
    name = "TopScore"
    type = "N"
  }
  key_schema {
    attribute_name = "ForumName"
    key_type       = "HASH"
  }
  key_schema {
    attribute_name = "Subject"
    key_type       = "RANGE"
  }
  global_secondary_index {
    name            = "PostDate"
    hash_key        = "PostMonth"
    range_key       = "LastPostDateTime"
    write_capacity  = 2
    read_capacity   = 2
    projection_type = "KEYS_ONLY"
  }
  global_secondary_index {
    name            = "GamerZone"
    hash_key        = "GamerZone"
    range_key       = "TopScore"
    write_capacity  = 5
    read_capacity   = 5
    projection_type = "ALL"
    #delete_index = true #To remove an global secondary index from associated table
  }
  server_side_encryption {
    enabled = false
  }
  local_secondary_index { #local secondary index doesnot support updation
    hash_key        = "ForumName"
    name            = "LastPostIndex"
    range_key       = "LastPostDateTime"
    projection_type = "KEYS_ONLY"
  }

}


resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {

  tenant_id                   = duplocloud_tenant.myapp.tenant_id
  name                        = "mytable"
  read_capacity               = 80
  write_capacity              = 40
  billing_mode                = "PROVISIONED"
  is_point_in_time_recovery   = false
  deletion_protection_enabled = false

  tag {
    key   = "school"
    value = "admission"
  }
  attribute {
    name = "ForumName"
    type = "S"
  }
  attribute {
    name = "Subject"
    type = "S"
  }
  attribute {
    name = "LastPostDateTime"
    type = "S"
  }
  attribute {
    name = "PostMonth"
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

  ttl {
    attribute_name = "TimeToExist"
    enabled        = true
  }
}
