resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_mq_broker" "example" {
  tenant_id                  = duplocloud_tenant.myapp.tenant_id
  engine_type                = "ACTIVEMQ"
  deployment_mode            = "SINGLE_INSTANCE"
  broker_storage_type        = "EBS"
  broker_name                = "example5"
  host_instance_type         = "mq.m5.large"
  engine_version             = "5.18"
  authentication_strategy    = "SIMPLE"
  auto_minor_version_upgrade = true
  data_replication_mode      = "NONE"
  publicly_accessible        = false
  security_groups            = ["sg-0aead878f4e404689"]
  subnet_ids                 = ["subnet-0c9a95f287b4fc38f"]
  logs {
    general = true
    audit   = false
  }
  user {
    user_name = "admin"
    password  = "q2eqsaxasd121312sQsdfweqw"
    groups    = ["admins"]
  }
  encryption_options {
    kms_key_id        = "arn:aws:kms:us-west-2:123456789012:key/abcd-1234"
    use_aws_owned_key = false
  }
  maintenance_window {
    time_of_day = "02:00"
    time_zone   = "UTC"
    day_of_week = "SUNDAY"
  }
  tags = {
    Environment = "dev"
    Owner       = "team"
  }
}



resource "duplocloud_aws_mq_broker" "example" {
  tenant_id                  = duplocloud_tenant.myapp.tenant_id
  engine_type                = "ACTIVEMQ"
  deployment_mode            = "SINGLE_INSTANCE"
  broker_storage_type        = "EBS"
  broker_name                = "example5"
  host_instance_type         = "mq.m5.large"
  engine_version             = "5.18"
  authentication_strategy    = "LDAP"
  auto_minor_version_upgrade = true
  data_replication_mode      = "NONE"
  publicly_accessible        = false
  security_groups            = ["sg-0aead878f4e404689"]
  subnet_ids                 = ["subnet-0c9a95f287b4fc38f"] //, "subnet-09252308e1a093bda"]
  logs {
    general = true
    audit   = false
  }
  user {
    user_name = "admin"
    password  = "q2eqsaxasd121312sQsdfweqw"
  }
  ldap_server_metadata {
    hosts                    = ["ldap.example.com"]
    role_base                = "ou=roles,dc=example,dc=com"
    role_name                = "role"
    role_search_matching     = "(member={0})"
    role_search_subtree      = true
    service_account_password = "ldap-password#"
    service_account_username = "cn=admin,dc=example,dc=com"
    user_base                = "ou=users,dc=example,dc=com"
    user_role_name           = "role"
    user_search_matching     = "(uid={0})"
    user_search_subtree      = true
  }
}


resource "duplocloud_aws_mq_broker" "example" {
  tenant_id                  = duplocloud_tenant.myapp.tenant_id
  engine_type                = "RABBITMQ"
  deployment_mode            = "SINGLE_INSTANCE"
  broker_storage_type        = "EBS"
  broker_name                = "example5"
  host_instance_type         = "mq.t3.micro"
  engine_version             = "3.13"
  authentication_strategy    = "SIMPLE"
  auto_minor_version_upgrade = true
  data_replication_mode      = "NONE"
  publicly_accessible        = false
  security_groups            = ["sg-0aead878f4e404689"]
  logs {
    general = false
  }
  user {
    user_name = "admin"
    password  = "q2eqsaxasd121312sQsdfweqw"
    groups    = ["admins"]
  }
}


resource "duplocloud_aws_mq_broker" "example" {
  tenant_id                  = duplocloud_tenant.myapp.tenant_id
  engine_type                = "RABBITMQ"
  deployment_mode            = "CLUSTER_MULTI_AZ"
  broker_storage_type        = "EBS"
  broker_name                = "example5"
  host_instance_type         = "mq.m7g.medium"
  engine_version             = "3.13"
  authentication_strategy    = "SIMPLE"
  auto_minor_version_upgrade = true
  data_replication_mode      = "NONE"
  publicly_accessible        = false
  security_groups            = ["sg-0aead878f4e404689"]
  logs {
    general = false
  }
  user {
    user_name = "admin"
    password  = "q2eqsaxasd121312sQsdfweqw"
  }
}
