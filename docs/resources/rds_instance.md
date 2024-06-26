---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_rds_instance Resource - terraform-provider-duplocloud"
subcategory: ""
description: |-
  duplocloud_rds_instance manages an AWS RDS instance in Duplo.
---

# duplocloud_rds_instance (Resource)

`duplocloud_rds_instance` manages an AWS RDS instance in Duplo.

## Example Usage

```terraform
resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

// Generate a random password.
resource "random_password" "mypassword" {
  length  = 16
  special = false
}

// Create an RDS instance.
resource "duplocloud_rds_instance" "mydb" {
  tenant_id      = duplocloud_tenant.myapp.tenant_id
  name           = "mydb"
  engine         = 1 // PostgreSQL
  engine_version = "15.2"
  size           = "db.t3.medium"

  master_username = "myuser"
  master_password = random_password.mypassword.result

  encrypt_storage         = true
  backup_retention_period = 1
}


// Create an RDS instance.
resource "duplocloud_rds_instance" "aurora-mydb" {
  tenant_id      = duplocloud_tenant.myapp.tenant_id
  name           = "aurora-mydb"
  engine         = 9 // AuroraDB
  engine_version = "15.2"
  size           = "db.t3.medium"

  master_username              = "myuser"
  master_password              = random_password.mypassword.result
  cluster_parameter_group_name = "default.cluster-groupname"

  encrypt_storage         = true
  backup_retention_period = 1
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `engine` (Number) The numerical index of database engine to use the for the RDS instance.
Should be one of:

   - `0` : MySQL
   - `1` : PostgreSQL
   - `2` : MsftSQL-Express
   - `3` : MsftSQL-Standard
   - `8` : Aurora-MySQL
   - `9` : Aurora-PostgreSQL
   - `10` : MsftSQL-Web
   - `11` : Aurora-Serverless-MySql
   - `12` : Aurora-Serverless-PostgreSql
   - `13` : DocumentDB
   - `14` : MariaDB
   - `16` : Aurora
- `name` (String) The short name of the RDS instance.  Duplo will add a prefix to the name.  You can retrieve the full name from the `identifier` attribute.
- `size` (String) The instance type of the RDS instance.
See AWS documentation for the [available instance types](https://aws.amazon.com/rds/instance-types/).
- `tenant_id` (String) The GUID of the tenant that the RDS instance will be created in.

### Optional

- `allocated_storage` (Number) (Required unless a `snapshot_id` is provided) The allocated storage in gigabytes.
- `backup_retention_period` (Number) Specifies backup retention period between 1 and 35 day(s). Default backup retention period is 1 day. Defaults to `1`.
- `cluster_parameter_group_name` (String) Parameter group associated with this instance's DB Cluster.
- `db_name` (String) The name of the database to create when the DB instance is created. This is not applicable for update.
- `db_subnet_group_name` (String) Name of DB subnet group. DB instance will be created in the VPC associated with the DB subnet group.
- `deletion_protection` (Boolean) If the DB instance should have deletion protection enabled.The database can't be deleted when this value is set to `true`. This setting is not applicable for document db cluster instance. Defaults to `false`.
- `enable_iam_auth` (Boolean) Whether or not to enable the RDS IAM authentication.
- `enable_logging` (Boolean) Whether or not to enable the RDS instance logging. This setting is not applicable for document db cluster instance.
- `encrypt_storage` (Boolean) Whether or not to encrypt the RDS instance storage.
- `engine_version` (String) The database engine version to use the for the RDS instance.
If you don't know the available engine versions for your RDS instance, you can use the [AWS CLI](https://docs.aws.amazon.com/cli/latest/reference/rds/describe-db-engine-versions.html) to retrieve a list.
- `enhanced_monitoring` (Number) Interval to capture metrics in real time for the operating system (OS) that your Amazon RDS DB instance runs on.
- `iops` (Number) The IOPS (Input/Output Operations Per Second) value. Should be specified only if `storage_type` is either io1 or gp3.
- `kms_key_id` (String) The globally unique identifier for the key.
- `master_password` (String, Sensitive) The master password of the RDS instance.
- `master_username` (String) The master username of the RDS instance.
- `multi_az` (Boolean) Specifies if the RDS instance is multi-AZ.
- `parameter_group_name` (String) A RDS parameter group name to apply to the RDS instance.
- `skip_final_snapshot` (Boolean) If the final snapshot should be taken. When set to true, the final snapshot will not be taken when the resource is deleted. Defaults to `false`.
- `snapshot_id` (String) A database snapshot to initialize the RDS instance from, at launch.
- `storage_type` (String) Valid values: gp2 | gp3 | io1 | standard. Storage type to be used for RDS instance storage.
- `store_details_in_secret_manager` (Boolean) Whether or not to store RDS details in the AWS secrets manager.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `v2_scaling_configuration` (Block List, Max: 1) Serverless v2_scaling_configuration min and max scalling capacity. (see [below for nested schema](#nestedblock--v2_scaling_configuration))

### Read-Only

- `arn` (String) The ARN of the RDS instance.
- `cluster_identifier` (String) The RDS Cluster Identifier
- `endpoint` (String) The endpoint of the RDS instance.
- `host` (String) The DNS hostname of the RDS instance.
- `id` (String) The ID of this resource.
- `identifier` (String) The full name of the RDS instance.
- `instance_status` (String) The current status of the RDS instance.
- `port` (Number) The listening port of the RDS instance.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)


<a id="nestedblock--v2_scaling_configuration"></a>
### Nested Schema for `v2_scaling_configuration`

Required:

- `max_capacity` (Number) Specifies max scalling capacity.
- `min_capacity` (Number) Specifies min scalling capacity.

## Import

Import is supported using the following syntax:

```shell
# Example: Importing an existing RDS instance
#  - *TENANT_ID* is the tenant GUID
#  - *SHORTNAME* is the short name of the database (without the duplo prefix)
#
terraform import duplocloud_rds_instance.mydb v2/subscriptions/*TENANT_ID*/RDSDBInstance/*SHORTNAME*
```
