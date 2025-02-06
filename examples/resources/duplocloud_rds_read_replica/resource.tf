resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_rds_instance" "rds" {
  tenant_id       = duplocloud_tenant.myapp.tenant_id
  enable_logging  = false
  encrypt_storage = true
  engine          = 8
  engine_version  = "8.0.mysql_aurora.3.04.0"
  master_password = "test!!1234"
  master_username = "masteruser"
  multi_az        = false
  name            = "mysqltest"
  size            = "db.t2.small"
}

resource "duplocloud_rds_read_replica" "replica" {
  tenant_id          = duplocloud_rds_instance.rds.tenant_id
  name               = "read-replica"
  size               = "db.t2.small"
  cluster_identifier = duplocloud_rds_instance.rds.cluster_identifier
}

//Performance insight example for document db read replica
resource "duplocloud_rds_instance" "rds" {
  tenant_id      = duplocloud_tenant.myapp.tenant_id
  name           = "docDb"
  engine         = 13
  engine_version = "3.6.0"
  size           = "db.t3.medium"

  master_username = "myuser"
  master_password = "Qaazwedd#1"

  encrypt_storage                 = true
  store_details_in_secret_manager = true
  enhanced_monitoring             = 0

}

resource "duplocloud_rds_read_replica" "replica" {
  tenant_id          = duplocloud_rds_instance.rds.tenant_id
  name               = "read-replica"
  size               = "db.t3.medium"
  cluster_identifier = duplocloud_rds_instance.rds.cluster_identifier
  performance_insights {
    enabled          = true
    retention_period = 7
  }
}


//Performance insight example for cluster db read replica
//Performance insight for aurora cluster db is applied at cluster level
resource "duplocloud_rds_instance" "rds" {
  tenant_id      = duplocloud_tenant.myapp.tenant_id
  name           = "clust"
  engine         = 8
  engine_version = "8.0.mysql_aurora.3.07.1"
  size           = "db.r5.large"

  master_username                 = "myuser"
  master_password                 = "Qaazwedd#1"
  encrypt_storage                 = true
  store_details_in_secret_manager = true
  enhanced_monitoring             = 0
  performance_insights {
    enabled          = false
    retention_period = 31
    kms_key_id       = "arn:aws:kms:us-west-2:182680712604:key/6b8dc967-92bf-43de-a850-ee7b945260f8"
  }
}
//referencing performance insights block from writer or primary resource is must, 
//to maintain tfstate to be in sync for performance insights block.
resource "duplocloud_rds_read_replica" "replica" {
  tenant_id          = duplocloud_rds_instance.rds.tenant_id
  name               = "read-replica"
  size               = "db.r5.large"
  cluster_identifier = duplocloud_rds_instance.rds.cluster_identifier
  performance_insights {
    enabled          = duplocloud_rds_instance.rds.performance_insights.0.enabled
    retention_period = duplocloud_rds_instance.rds.performance_insights.0.retention_period
    kms_key_id       = duplocloud_rds_instance.rds.performance_insights.0.kms_key_id
  }
}

resource "duplocloud_rds_read_replica" "replica2" {
  tenant_id          = duplocloud_rds_instance.rds.tenant_id
  name               = "read-replica2"
  size               = "db.r5.large"
  cluster_identifier = duplocloud_rds_instance.rds.cluster_identifier
  performance_insights {
    enabled          = duplocloud_rds_instance.rds.performance_insights.0.enabled
    retention_period = duplocloud_rds_instance.rds.performance_insights.0.retention_period
    kms_key_id       = duplocloud_rds_instance.rds.performance_insights.0.kms_key_id
  }

}


//Performance insight example for instance db read replica
resource "duplocloud_rds_instance" "mydb" {
  tenant_id      = duplocloud_tenant.myapp.tenant_id
  name           = "mydbpsql"
  engine         = 1 // PostgreSQL
  engine_version = "14.11"
  size           = "db.t3.medium"

  master_username = "myuser"
  master_password = "Qaazwedd#1"

  encrypt_storage                 = false
  store_details_in_secret_manager = false
  enhanced_monitoring             = 0
  storage_type                    = "gp2"
  performance_insights {
    enabled          = true
    retention_period = 7
  }
}

resource "duplocloud_rds_read_replica" "replica" {
  tenant_id          = duplocloud_rds_instance.mydb.tenant_id
  name               = "inst-replica"
  size               = "db.t3.medium"
  cluster_identifier = duplocloud_rds_instance.mydb.cluster_identifier
  performance_insights {
    enabled          = true
    retention_period = 7
  }
}

//Example to create serverless read replica for a provisioned writer

resource "duplocloud_rds_instance" "provisioned3" {
  tenant_id      = duplocloud_rds_instance.mydb.tenant_id
  name           = "provisioned1"
  engine         = 8
  engine_version = "8.0.mysql_aurora.3.05.2"
  size           = "db.t3.medium"

  master_username = "duploadmin"
  master_password = "duploadmin"

  encrypt_storage         = true
  backup_retention_period = 1

}



resource "duplocloud_rds_read_replica" "serverless_replica2" {
  tenant_id          = duplocloud_rds_instance.mydb.tenant_id
  name               = "provisioned1-serverless-replica"
  size               = "db.serverless"
  cluster_identifier = duplocloud_rds_instance.provisioned3.cluster_identifier
  v2_scaling_configuration {
    max_capacity = 2
    min_capacity = 1
  }
}
