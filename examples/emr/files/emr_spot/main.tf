terraform {
  required_providers {
    duplocloud = {
      version = "0.9.10" # RELEASE VERSION
      source  = "registry.terraform.io/duplocloud/duplocloud"
    }
  }
}

provider "duplocloud" {
  // duplo_host = "https://xxx.duplocloud.net"  # you can also set the duplo_host env var
  // duplo_token = ".."                         # please *ONLY* specify using a duplo_token env var (avoid checking secrets into git)
}

variable "plan_id" {
  type = string
}

variable "tenant_id" {
  type = string
}

# Tenant information 
data "duplocloud_tenant" "test" { name = "default" }
output "tenant" { value = data.duplocloud_tenant.test }

# Tenant listing
data "duplocloud_tenants" "test" {}
output "tenants" { value = data.duplocloud_tenants.test.tenants.*.name }


resource "duplocloud_emr_cluster" "test" {
  tenant_id     = var.tenant_id
  name          = "emrp1"
  release_label = "emr-6.2.0"
  // custom_ami_id                       = "pravemr1"
  log_uri                           = "s3://name-of-my-bucket"
  visible_to_all_users              = true
  keep_job_flow_alive_when_no_steps = true


  applications = jsonencode([
    {
      "Name" : "Hadoop"
    },
    {
      "Name" : "JupyterHub"
    },
    {
      "Name" : "Spark"
    },
    {
      "Name" : "Hive"
    }
    ]
  )




  configurations = jsonencode(
    [
      {
        "Classification" : "hive-site",
        "Properties" : {
          "hive.metastore.client.factory.class" : "com.amazonaws.glue.catalog.metastore.AWSGlueDataCatalogHiveClientFactory",
          "spark.sql.catalog.my_catalog" : "org.apache.iceberg.spark.SparkCatalog",
          "spark.sql.catalog.my_catalog.catalog-impl" : "org.apache.iceberg.aws.glue.GlueCatalog",
          "spark.sql.catalog.my_catalog.io-impl" : "org.apache.iceberg.aws.s3.S3FileIO",
          "spark.sql.catalog.my_catalog.lock-impl" : "org.apache.iceberg.aws.glue.DynamoLockManager",
          "spark.sql.catalog.my_catalog.lock.table" : "myGlueLockTable",
          "spark.sql.catalog.sampledb.warehouse" : "s3://name-of-my-bucket/icebergcatalog"
        }
      }
    ]
  )
  bootstrap_actions = jsonencode(
    [
      {
        Name : "InstallApacheIceberg",
        ScriptBootstrapAction : {
          Args : ["name", "value"],
          Path : "s3://name-of-my-bucket/bootstrap-iceberg.sh"
        }
      }
    ]
  )


  //https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-instance-fleet.html
  instance_fleets = jsonencode(
    [
      {
        "Name" : "Masterfleet",
        "InstanceFleetType" : "MASTER",
        "TargetSpotCapacity" : 1,
        "LaunchSpecifications" : {
          "SpotSpecification" : {
            "TimeoutDurationMinutes" : 120,
            "TimeoutAction" : "SWITCH_TO_ON_DEMAND"
          }
        },
        "InstanceTypeConfigs" : [
          {
            "InstanceType" : "m5.large",
            "BidPrice" : "0.89"
          }
        ]
      },
      {
        "Name" : "Corefleet",
        "InstanceFleetType" : "CORE",
        "TargetSpotCapacity" : 1,
        "TargetOnDemandCapacity" : 1,
        "LaunchSpecifications" : {
          "OnDemandSpecification" : {
            "AllocationStrategy" : "lowest-price",
            "CapacityReservationOptions" : {
              "UsageStrategy" : "use-capacity-reservations-first",
              "CapacityReservationResourceGroupArn" : "String"
            }
          },
          "SpotSpecification" : {
            "AllocationStrategy" : "capacity-optimized",
            "TimeoutDurationMinutes" : 120,
            "TimeoutAction" : "TERMINATE_CLUSTER"
          }
        },
        "InstanceTypeConfigs" : [
          {
            "InstanceType" : "m4.large",
            "BidPriceAsPercentageOfOnDemandPrice" : 100
          }
        ]
      },
      {
        "Name" : "Taskfleet",
        "InstanceFleetType" : "TASK",
        "TargetSpotCapacity" : 1,
        "LaunchSpecifications" : {
          "OnDemandSpecification" : {
            "AllocationStrategy" : "lowest-price",
            "CapacityReservationOptions" : {
              "CapacityReservationPreference" : "none"
            }
          },
          "SpotSpecification" : {
            "TimeoutDurationMinutes" : 120,
            "TimeoutAction" : "TERMINATE_CLUSTER"
          }
        },
        "InstanceTypeConfigs" : [
          {
            "InstanceType" : "m4.large",
            "BidPrice" : "0.89"
          }
        ]
      }
    ]
  )
}

