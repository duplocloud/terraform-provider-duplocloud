---

# Resource: duplocloud_rds_instance



The `duplocloud_rds_instance` resource in DuploCloud manages the lifecycle of an RDS (Relational Database Service) instance within a cloud environment. It allows you to define, provision, and maintain database instances with customizable configurations, such as engine type, storage, and instance class, all within DuploCloud's automated infrastructure management.


## Example Usage

### Provision an RDS instance using the PostgreSQL engine named dev-db in DuploCloud platform.

```terraform
# Before creating an RDS instance, you must first set up the infrastructure and tenant. Below is the resource for creating the infrastructure.
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "dev"
  cloud             = 0 # AWS Cloud
  region            = "us-east-1"
  enable_k8_cluster = false
  address_prefix    = "10.13.0.0/16"
}

# Use the infrastructure name as the 'plan_id' from the 'duplocloud_infrastructure' resource while creating tenant.
resource "duplocloud_tenant" "tenant" {
 account_name = "dev"
 plan_id      = duplocloud_infrastructure.infra.infra_name
}

# Generate a random password for the RDS instance.
resource "random_password" "password" {
  length  = 16
  special = false
}

# Create an RDS instance.
resource "duplocloud_rds_instance" "dev-db" {
  tenant_id      = duplocloud_tenant.tenant.tenant_id
  name           = "dev-db"
  engine         = 1 # PostgreSQL DB engine
  engine_version = "15.2" # PostgreSQL DB engine version
  size           = "db.t3.medium" # RDS instance size/class

  master_username = "postgres"
  master_password = random_password.password.result

  encrypt_storage         = true
  backup_retention_period = 7
}
```

### Provision an RDS instance using the PostgreSQL engine named dev-db with deletion protection enabled and multi-az enabled.

```terraform
# Ensure the 'dev' tenant is already created before creating the RDS instance.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Generate a random password for the RDS instance.
resource "random_password" "mypassword" {
  length  = 16
  special = false
}

resource "duplocloud_rds_instance" "dev-db" {
  tenant_id      = data.duplocloud_tenant.tenant.id
  name           = "dev-db"
  engine         = 1 # PostgreSQL DB engine
  engine_version = "15.2"
  size           = "db.t3.medium"

  deletion_protection = true
  multi_az            = true

  master_username = "postgres"
  master_password = random_password.mypassword.result

  encrypt_storage         = true
  backup_retention_period = 7
}
```

### Create an RDS instance using the Aurora-PostgreSQL engine named aurora-postgres-db with instance class db.m5.large.

```terraform
# Ensure the 'dev' tenant is already created before creating the RDS instance.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Generate a random password for the RDS instance.
resource "random_password" "mypassword" {
  length  = 16
  special = false
}

resource "duplocloud_rds_instance" "aurora-postgres-db" {
  tenant_id      = data.duplocloud_tenant.tenant.id
  name           = "aurora-postgres-db"
  engine         = 9 # AuroraDB PostgreSQL engine
  engine_version = "15.2"
  size           = "db.m5.large"

  master_username = "postgres"
  master_password = random_password.mypassword.result

  encrypt_storage         = true
  backup_retention_period = 7
}
```

### Create an Aurora serverless RDS instance using the PostgreSQL engine named aurora-postgres with engine version 15.5, minimum capacity of 0.5, maximum capacity of 2, with deletion protection enabled and store the DB credentials in AWS secrets manager. Also create a read replica for this database.

```terraform
# Ensure the 'dev' tenant is already created before creating the RDS instance.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Generate a random password for the RDS instance.
resource "random_password" "mypassword" {
  length  = 16
  special = false
}

resource "duplocloud_rds_instance" "aurora-serverless" {
  tenant_id      = data.duplocloud_tenant.tenant.id
  name           = "aurora-postgres"
  engine         = 9
  engine_version = "15.5"
  size           = "db.serverless"

  master_username = "postgres"
  master_password = random_password.mypassword.result

  encrypt_storage         = true
  backup_retention_period = 7

  v2_scaling_configuration {
    min_capacity = 0.5
    max_capacity = 2
  }

  store_details_in_secret_manager = true
  deletion_protection             = true
}

resource "duplocloud_rds_read_replica" "read-replica" {
  tenant_id          = data.duplocloud_tenant.tenant.id
  name               = "aurora-postgres-read-replica"
  size               = "db.serverless"
  cluster_identifier = duplocloud_rds_instance.aurora-serverless.cluster_identifier
  depends_on         = [duplocloud_rds_instance.aurora-serverless]
}
```

### Provision an RDS instance using the MySQL engine named dev-db, with username mysql_user1 in DuploCloud platform.

```terraform
# Ensure the 'dev' tenant is already created before creating the RDS instance.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Generate a random password for the RDS instance.
resource "random_password" "mypassword" {
  length  = 16
  special = false
}

# Create an RDS instance.
resource "duplocloud_rds_instance" "dev-db" {
  tenant_id      = data.duplocloud_tenant.tenant.id
  name           = "dev-db"
  engine         = 0              # MySQL DB engine
  engine_version = "8.0.32"       # MySQL DB engine version
  size           = "db.t3.medium" # RDS instance size/class

  master_username = "mysql_user1"
  master_password = random_password.mypassword.result

  encrypt_storage         = true
  backup_retention_period = 7
}
```

### Provision an RDS instance using the MySQL engine named dev-db with engine version 5.7, allocated storage 50 GB and enable IAM auth and logging for this DB.

```terraform
# Ensure the 'dev' tenant is already created before creating the RDS instance.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Generate a random password for the RDS instance.
resource "random_password" "mypassword" {
  length  = 16
  special = false
}

# Create an RDS instance.
resource "duplocloud_rds_instance" "dev-db" {
  tenant_id      = data.duplocloud_tenant.tenant.id
  name           = "dev-db"
  engine         = 0              # MySQL DB engine
  engine_version = "5.7.44"       # MySQL DB engine version
  size           = "db.t3.medium" # RDS instance size/class

  master_username = "mysql_user1"
  master_password = random_password.mypassword.result

  encrypt_storage         = true
  backup_retention_period = 7
  allocated_storage       = 50
  enable_iam_auth         = true
  enable_logging          = true
}
```

### Create an RDS instance using the Aurora MySQL engine named mysql-db with engine version 5.7, allocated storage 100 GB and storage type io1 with number of iops 6000. It should skip the final snapshot and store the credentials in secrets manager.

```terraform
# Ensure the 'dev' tenant is already created before creating the RDS instance.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Generate a random password for the RDS instance.
resource "random_password" "mypassword" {
  length  = 16
  special = false
}

# Create an RDS instance.
resource "duplocloud_rds_instance" "mysql-db" {
  tenant_id      = data.duplocloud_tenant.tenant.id
  name           = "mysql-db"
  engine         = 8                         # Aurora MySQL DB engine
  engine_version = "5.7.mysql_aurora.2.11.6" # MySQL DB engine version
  size           = "db.t3.medium"            # RDS instance size/class

  master_username = "mysql_user1"
  master_password = random_password.mypassword.result

  encrypt_storage         = true
  backup_retention_period = 7
  allocated_storage       = 100
  storage_type            = "io1"
  iops                    = 6000
  skip_final_snapshot     = true

  store_details_in_secret_manager = true
}
```


# Example for enabling performance insights in a RDS instance

```
resource "duplocloud_rds_instance" "mydb" {
  tenant_id      = data.duplocloud_tenant.tenant.id
  name           = "mydb1psql"
  engine         = 1 // PostgreSQL
  engine_version = "14.11"
  size           = "db.t3.micro"

  master_username = "myuser"
  master_password = "Qaazwedd#1"

  encrypt_storage                 = true
  store_details_in_secret_manager = true
  enhanced_monitoring             = 0
  storage_type                    = "gp2"
  performance_insights {
    enabled          = true
    retention_period = 7
  }
}
```

# Example for enabling performance insights in a RDS cluster

```
resource "duplocloud_rds_instance" "mydb" {
  tenant_id      = data.duplocloud_tenant.tenant.id
  name           = "clust"
  engine         = 8 // PostgreSQL
  engine_version = "8.0.mysql_aurora.3.07.1"
  size           = "db.r5.large"

  master_username = "myuser"
  master_password = "Qaazwedd#1"

  encrypt_storage                 = true
  store_details_in_secret_manager = true
  enhanced_monitoring             = 0
  performance_insights {
    enabled          = true
    retention_period = 7
  }
}
```

# Example for enabling performance insights in a Document DB
```
resource "duplocloud_rds_instance" "mydb" {
  tenant_id      = data.duplocloud_tenant.tenant.id
  name           = "doc"
  engine         = 13
  engine_version = "3.6.0"
  size           = "db.t3.medium"

  master_username = "myuser"
  master_password = "Qaazwedd#1"

  encrypt_storage                 = true
  store_details_in_secret_manager = true
  enhanced_monitoring             = 0
  performance_insights {
    enabled          = true
    retention_period = 7
  }
}
```

```
# Example for creating db using snapshot

resource "duplocloud_rds_instance" "mydb" {
  tenant_id      = data.duplocloud_tenant.tenant.id
  name           = "mydbpsql"
  engine         = 0 //mysql
  engine_version = "5.7.44"
  size           = "db.t3.medium"
  snapshot_id = "rds:duplotest-snapdb-2024-12-17-07-00" //snapshot id is of previously created mysql db of version 5.7.44
}
```

# Example to showcase use of parameter group in writer and read replica for aurora cluster instance

```
resource "random_password" "mypassword" {
  length  = 16
  special = false
}

resource "duplocloud_rds_instance" "app" {
  tenant_id      = data.duplocloud_tenant.tenant.id
  name           = "writer1-sqlnew"
  engine         = 8 
  engine_version = "5.7.mysql_aurora.2.11.5"
  size           = "db.r5.large"
  master_username              = "myuser"
  master_password              = random_password.mypassword.result
  encrypt_storage         = true
  backup_retention_period = 10
  db_name         =  "auroradb"
  skip_final_snapshot = true
  store_details_in_secret_manager = false
  enhanced_monitoring = 0
  availability_zone = "us-west-2b"
  storage_type                    = "aurora"
  cluster_parameter_group_name = "c-aurora-mysql"
  parameter_group_name = "aurora-mysql-dbparam"
}

resource "duplocloud_rds_read_replica" "replica1" {
  tenant_id          = duplocloud_rds_instance.app.tenant_id
  name               = "aurora-replica-new"
  size               = "db.r5.large"
  cluster_identifier = duplocloud_rds_instance.app.cluster_identifier
  availability_zone = "us-west-2a"
  parameter_group_name = "aurora-mysql-dbparam"
  engine_type=duplocloud_rds_instance.app.engine
}
```

# Example to showcase use of parameter group in writer and read replica for standalone instance

```
resource "duplocloud_rds_instance" "mydb" {
  tenant_id      = data.duplocloud_tenant.tenant.id
  name           = "tf-postgresql1"
  engine         = 1// PostgreSQL
  engine_version = "13.11"
  size           = "db.t3.medium"
  master_username = "myuser"
  master_password = "Qaazwedd#1"
  parameter_group_name = "psql13dbparam"
  encrypt_storage                 = false
  store_details_in_secret_manager = false
  enhanced_monitoring             = 0
  storage_type                    = "gp2"
}

resource "duplocloud_rds_read_replica" "replica" {
  tenant_id          = duplocloud_rds_instance.mydb.tenant_id
  name               = "postgresql-rep1"
  size               = "db.t3.medium"
  cluster_identifier = duplocloud_rds_instance.mydb.cluster_identifier
  #availability_zone = "us-east-1b"
  performance_insights {
    enabled          = true
    retention_period = 31
  }
  engine_type=duplocloud_rds_instance.mydb.engine
  parameter_group_name=duplocloud_rds_instance.mydb.parameter_group_name
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
   - `13` : DocumentDB
   - `14` : MariaDB
   - `16` : Aurora
- `name` (String) The short name of the RDS instance.  Duplo will add a prefix to the name.  You can retrieve the full name from the `identifier` attribute.
- `size` (String) The instance type of the RDS instance.
See AWS documentation for the [available instance types](https://aws.amazon.com/rds/instance-types/).Size should be set as db.serverless if rds instamce is created as serverless
- `tenant_id` (String) The GUID of the tenant that the RDS instance will be created in.

### Optional

- `allocated_storage` (Number) (Required unless a `snapshot_id` is provided) The allocated storage in gigabytes.
- `auto_minor_version_upgrade` (Boolean) Enable or disable auto minor version upgrade
- `availability_zone` (String) Specify a valid Availability Zone for the RDS primary instance (when Multi-AZ is disabled) or for the Aurora writer instance. e.g. us-west-2a
- `backup_retention_period` (Number) Specifies backup retention period between 1 and 35 day(s). Default backup retention period is 1 day. Defaults to `1`.
- `cluster_parameter_group_name` (String) Parameter group associated with this instance's DB Cluster.
- `db_name` (String) The name of the database to create when the DB instance is created. This is not applicable for update.
- `db_subnet_group_name` (String) Name of DB subnet group. DB instance will be created in the VPC associated with the DB subnet group.
- `deletion_protection` (Boolean) If the DB instance should have deletion protection enabled.The database can't be deleted when this value is set to `true`. This setting is not applicable for document db cluster instance. Defaults to `false`.
- `enable_iam_auth` (Boolean) Whether or not to enable the RDS IAM authentication. It can only be set during instance creation.
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
- `performance_insights` (Block List, Max: 1) Amazon RDS Performance Insights is a database performance tuning and monitoring feature that helps you quickly assess the load on your database, and determine when and where to take action. Perfomance Insights get apply when enable is set to true. (see [below for nested schema](#nestedblock--performance_insights))
- `skip_final_snapshot` (Boolean) If the final snapshot should be taken. When set to true, the final snapshot will not be taken when the resource is deleted. Defaults to `false`.
- `snapshot_id` (String) A database snapshot to initialize the RDS instance from, at launch.
- `storage_type` (String) Storage type to be used for RDS instance storage.

			|Storage Type  | Performance                        | Throughput            | Descritpion                                                                                                                                                                                                               |
			|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
			| gp2          | 3 IOPS/GB, up to 16K IOPS          | Up to 250 MB/s	    | General-purpose databases, small to medium workloads. 'gp2' provides SSD-based storage with burstable IOPS                                                                                                                |
			| gp3          | 3K to 16K IOPS                     | Up to 1,000 MB/s      | Cost-effective, customizable performance for a wide range of workloads. gp3 offers a more advanced and cost-effective version of gp2. You can provision IOPS and throughput independently of storage size.                |
			| io1          | Up to 256K IOPS                    | Up to 1,000 MB/s      | Mission-critical applications with high IOPS requirements. io1 provides provisioned IOPS, meaning you can define and guarantee IOPS performance levels independently of storage capacity.                                 |
			| standard     | Variable, low IOPS                 | Low and unpredictable | Low-cost, infrequent access, small databases, or test environments. Magnetic storage is the oldest and least performant storage option. It is mainly used for low-cost applications with low performance demands.         |
			| aurora       | Automatic scaling, up to 200K IOPS | Varies                | High-performance, fault-tolerant, distributed storage for Amazon Aurora databases. Aurora uses a unique distributed, fault-tolerant storage system that automatically replicates data across multiple Availability Zones. |
			| aurora-iopt1 | Provisioned IOPS, similar to io1   | Varies                | Aurora databases needing guaranteed, high-performance IOPS. Aurora I/O-Optimized storage offers provisioned IOPS for Aurora clusters that require consistently high performance for critical workloads.                   |
- `store_details_in_secret_manager` (Boolean) Whether or not to store RDS details in the AWS secrets manager.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `v2_scaling_configuration` (Block List, Max: 1) Serverless v2_scaling_configuration min and max scaling capacity. This configuration is only applicable for serverless instances (see [below for nested schema](#nestedblock--v2_scaling_configuration))

### Read-Only

- `arn` (String) The ARN of the RDS instance.
- `cluster_identifier` (String) The RDS Cluster Identifier
- `endpoint` (String) The endpoint of the RDS instance.
- `host` (String) The DNS hostname of the RDS instance.
- `id` (String) The ID of this resource.
- `identifier` (String) The full name of the RDS instance.
- `instance_status` (String) The current status of the RDS instance.
- `port` (Number) The listening port of the RDS instance.

<a id="nestedblock--performance_insights"></a>
### Nested Schema for `performance_insights`

Optional:

- `enabled` (Boolean) Turn on or off Performance Insights Defaults to `false`.
- `kms_key_id` (String) Specify ARN for the KMS key to encrypt Performance Insights data.
- `retention_period` (Number) Specify retention period in Days. Valid values are 7, 731 (2 years) or a multiple of 31. For Document DB retention period is 7 Defaults to `7`.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)


<a id="nestedblock--v2_scaling_configuration"></a>
### Nested Schema for `v2_scaling_configuration`

Required:

- `max_capacity` (Number) Specifies max scaling capacity.
- `min_capacity` (Number) Specifies min scaling capacity.

## Import

Import is supported using the following syntax:

```shell
# Example: Importing an existing RDS instance
#  - *TENANT_ID* is the tenant GUID
#  - *SHORTNAME* is the short name of the database (without the duplo prefix)
#
terraform import duplocloud_rds_instance.mydb v2/subscriptions/*TENANT_ID*/RDSDBInstance/*SHORTNAME*
```

