resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Example 1:  EMR cluster with auto-scaling.
resource "duplocloud_emr_cluster" "test" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  name          = "emrp1"
  release_label = "emr-6.2.0"

  // custom_ami_id                  = "pravemr1"
  log_uri                           = "s3://name-of-my-bucket"
  visible_to_all_users              = true
  master_instance_type              = "m4.large"
  slave_instance_type               = "m4.large"
  instance_count                    = 3
  keep_job_flow_alive_when_no_steps = true

  applications = jsonencode([
    { "Name" : "Hadoop" },
    { "Name" : "JupyterHub" },
    { "Name" : "Spark" },
    { "Name" : "Hive" }
  ])

  managed_scaling_policy = jsonencode({
    "ComputeLimits" : {
      "UnitType" : "Instances",
      "MinimumCapacityUnits" : 2,
      "MaximumCapacityUnits" : 5,
      "MaximumOnDemandCapacityUnits" : 5,
      "MaximumCoreCapacityUnits" : 3
    }
  })

  configurations = jsonencode([
    {
      "Classification" : "hive-site",
      "Properties" : {
        "hive.metastore.client.factory.class" : "com.amazonaws.glue.catalog.metastore.AWSGlueDataCatalogHiveClientFactory",
        "spark.sql.catalog.my_catalog" : "org.apache.iceberg.spark.SparkCatalog",
        "spark.sql.catalog.my_catalog.catalog-impl" : "org.apache.iceberg.aws.glue.GlueCatalog",
        "spark.sql.catalog.my_catalog.io-impl" : "org.apache.iceberg.aws.s3.S3FileIO",
        "spark.sql.catalog.my_catalog.lock-impl" : "org.apache.iceberg.aws.glue.DynamoLockManager",
        "spark.sql.catalog.my_catalog.lock.table" : "myGlueLockTable",
        "spark.sql.catalog.sampledb.warehouse" : "s3://name-of-my-bucket/parquet5"
      }
    }
  ])
  bootstrap_actions = jsonencode([
    {
      Name : "InstallApacheIceberg",
      ScriptBootstrapAction : {
        Args : ["name", "value"],
        Path : "s3://name-of-my-bucket/bootstrap-iceberg.sh"
      }
    }
  ])
}
