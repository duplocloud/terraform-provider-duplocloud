
resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

#1. Example: Basic dynamo db table. Every key definition should have attribute associated to it.
resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {

  tenant_id                   = duplocloud_tenant.myapp.tenant_id
  name                        = "mytable"
  read_capacity               = 80
  write_capacity              = 40
  billing_mode                = "PROVISIONED"
  is_point_in_time_recovery   = false
  deletion_protection_enabled = false
  attribute {
    name = "ForumName"
    type = "S"
  }
  attribute {
    name = "Subject"
    type = "S"
  }
  key_schema {
    attribute_name = "ForumName"
    key_type       = "HASH"
  }
  key_schema {
    attribute_name = "Subject"
    key_type       = "RANGE"
  }
}

#2. Example: To add tag for dynamo db table.

resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {

  tenant_id                   = duplocloud_tenant.myapp.tenant_id
  name                        = "mytable"
  read_capacity               = 80
  write_capacity              = 40
  billing_mode                = "PROVISIONED"
  is_point_in_time_recovery   = false
  deletion_protection_enabled = false
  attribute {
    name = "ForumName"
    type = "S"
  }
  attribute {
    name = "Subject"
    type = "S"
  }
  key_schema {
    attribute_name = "ForumName"
    key_type       = "HASH"
  }
  key_schema {
    attribute_name = "Subject"
    key_type       = "RANGE"
  }

  tag {
    key   = "m1"
    value = "me1"
  }

}

#3. Example: Creating Dynamo db table with local secondary index. Local secondary index can only be implemented during creation. Here attribute of LastPostDateTime is defined for range key of local secondary index.
resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {

  tenant_id                   = duplocloud_tenant.myapp.tenant_id
  name                        = "mytable"
  read_capacity               = 80
  write_capacity              = 40
  billing_mode                = "PROVISIONED"
  is_point_in_time_recovery   = false
  deletion_protection_enabled = false

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
  key_schema {
    attribute_name = "ForumName"
    key_type       = "HASH"
  }
  key_schema {
    attribute_name = "Subject"
    key_type       = "RANGE"
  }

  tag {
    key   = "m1"
    value = "me1"
  }

  local_secondary_index {
    hash_key        = "ForumName"
    name            = "LastPostIndex"
    range_key       = "LastPostDateTime"
    projection_type = "KEYS_ONLY"
  }

}

#4. Example : Creating Dynamo db table with Global Secondary Index. Here attribute of PostMonth and TopScore is defined for hash key and range key of global secondary index respectively.
resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {

  tenant_id                   = duplocloud_tenant.myapp.tenant_id
  name                        = "mytable"
  read_capacity               = 80
  write_capacity              = 40
  billing_mode                = "PROVISIONED"
  is_point_in_time_recovery   = false
  deletion_protection_enabled = false
  tag {
    key   = "m1"
    value = "me1"
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
    name            = "PostCreate"
    hash_key        = "PostMonth"
    range_key       = "TopScore"
    write_capacity  = 2
    read_capacity   = 2
    projection_type = "KEYS_ONLY"
  }
  local_secondary_index {
    hash_key        = "ForumName"
    name            = "LastPostIndex"
    range_key       = "LastPostDateTime"
    projection_type = "KEYS_ONLY"
  }

}

#5 Example : Adding/Updating global secondary index in dynamo db table

resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {

  tenant_id                   = duplocloud_tenant.myapp.tenant_id
  name                        = "mytable"
  read_capacity               = 80
  write_capacity              = 40
  billing_mode                = "PROVISIONED"
  is_point_in_time_recovery   = false
  deletion_protection_enabled = false
  tag {
    key   = "m1"
    value = "me1"
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
    name            = "PostCreate"
    hash_key        = "PostMonth"
    range_key       = "TopScore"
    write_capacity  = 2
    read_capacity   = 2
    projection_type = "KEYS_ONLY"
  }
  global_secondary_index {
    name            = "GamerZone"
    hash_key        = "GamerZone"
    range_key       = "LastPostDateTime"
    write_capacity  = 5
    read_capacity   = 5
    projection_type = "ALL"
  }

  local_secondary_index {
    hash_key        = "ForumName"
    name            = "LastPostIndex"
    range_key       = "LastPostDateTime"
    projection_type = "KEYS_ONLY"
  }

}

#6 Example : Another eample related to global secondary index

resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {

  tenant_id                   = duplocloud_tenant.myapp.tenant_id
  name                        = "mytable"
  read_capacity               = 80
  write_capacity              = 40
  billing_mode                = "PROVISIONED"
  is_point_in_time_recovery   = false
  deletion_protection_enabled = false
  tag {
    key   = "m1"
    value = "me1"
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
  attribute {
    name = "GameTitle"
    type = "S"
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
    name            = "PostCreate"
    hash_key        = "PostMonth"
    range_key       = "TopScore"
    write_capacity  = 2
    read_capacity   = 2
    projection_type = "KEYS_ONLY"
  }
  global_secondary_index {
    name            = "GamerZone"
    hash_key        = "GamerZone"
    range_key       = "LastPostDateTime"
    write_capacity  = 5
    read_capacity   = 5
    projection_type = "ALL"
  }

  local_secondary_index {
    hash_key        = "ForumName"
    name            = "LastPostIndex"
    range_key       = "LastPostDateTime"
    projection_type = "KEYS_ONLY"
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

#7 Example : TTL example for Dynamo db table.

resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {

  tenant_id                   = "f762ec1d-0433-4ae4-abd1-2034d7562321"
  name                        = "mytable"
  read_capacity               = 80
  write_capacity              = 40
  billing_mode                = "PROVISIONED"
  is_point_in_time_recovery   = false
  deletion_protection_enabled = false
  tag {
    key   = "m1"
    value = "me1"
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
    name            = "PostCreate"
    hash_key        = "PostMonth"
    range_key       = "TopScore"
    write_capacity  = 2
    read_capacity   = 2
    projection_type = "KEYS_ONLY"
  }
  global_secondary_index {
    name            = "GamerZone"
    hash_key        = "GamerZone"
    range_key       = "LastPostDateTime"
    write_capacity  = 5
    read_capacity   = 5
    projection_type = "ALL"
  }
  local_secondary_index {
    hash_key        = "ForumName"
    name            = "LastPostIndex"
    range_key       = "LastPostDateTime"
    projection_type = "KEYS_ONLY"
  }
  ttl {
    attribute_name = "TimeToExist"
    enabled        = true
  }


}

#8 Example: Server side encryption example for dynamo db

resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {

  tenant_id                   = "f762ec1d-0433-4ae4-abd1-2034d7562321"
  name                        = "mytable"
  read_capacity               = 80
  write_capacity              = 40
  billing_mode                = "PROVISIONED"
  is_point_in_time_recovery   = false
  deletion_protection_enabled = false
  tag {
    key   = "m1"
    value = "me1"
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
    name            = "PostCreate"
    hash_key        = "PostMonth"
    range_key       = "TopScore"
    write_capacity  = 2
    read_capacity   = 2
    projection_type = "KEYS_ONLY"
  }
  global_secondary_index {
    name            = "GamerZone"
    hash_key        = "GamerZone"
    range_key       = "LastPostDateTime"
    write_capacity  = 5
    read_capacity   = 5
    projection_type = "ALL"
  }
  server_side_encryption {
    enabled = true
  }
  local_secondary_index {
    hash_key        = "ForumName"
    name            = "LastPostIndex"
    range_key       = "LastPostDateTime"
    projection_type = "KEYS_ONLY"
  }
  ttl {
    attribute_name = "TimeToExist"
    enabled        = true
  }


}

#9 Example : Delete protection and point in time recovery example for dynamo db

resource "duplocloud_aws_dynamodb_table_v2" "tst-dynamodb-table" {

  tenant_id                   = "f762ec1d-0433-4ae4-abd1-2034d7562321"
  name                        = "mytable"
  read_capacity               = 80
  write_capacity              = 40
  billing_mode                = "PROVISIONED"
  is_point_in_time_recovery   = true
  deletion_protection_enabled = true
  tag {
    key   = "m1"
    value = "me1"
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
    name            = "PostCreate"
    hash_key        = "PostMonth"
    range_key       = "TopScore"
    write_capacity  = 2
    read_capacity   = 2
    projection_type = "KEYS_ONLY"
  }
  global_secondary_index {
    name            = "GamerZone"
    hash_key        = "GamerZone"
    range_key       = "LastPostDateTime"
    write_capacity  = 5
    read_capacity   = 5
    projection_type = "ALL"
  }
  server_side_encryption {
    enabled = true
  }
  local_secondary_index {
    hash_key        = "ForumName"
    name            = "LastPostIndex"
    range_key       = "LastPostDateTime"
    projection_type = "KEYS_ONLY"
  }
  ttl {
    attribute_name = "TimeToExist"
    enabled        = true
  }


}