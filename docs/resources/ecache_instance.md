---

# Resource: duplocloud_ecache_instance



`duplocloud_ecache_instance` used to manage Amazon ElastiCache instances within a DuploCloud-managed environment. <p>This resource allows you to define and manage Redis or Memcached instances on AWS through Terraform, with DuploCloud handling the underlying infrastructure and integration aspects.</p>


## Example Usage

### Create an Amazon ElastiCache cluster of type Redis.

```terraform
# Before creating a ElastiCache cluster, you must first set up the infrastructure and tenant. Below is the resource for creating the infrastructure.
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0 # AWS Cloud
  region            = "us-west-2"
  enable_k8_cluster = false
  address_prefix    = "10.11.0.0/16"
}

# Use the infrastructure name as the 'plan_id' from the 'duplocloud_infrastructure' resource while creating tenant.
resource "duplocloud_tenant" "tenant" {
  account_name = "prod"
  plan_id      = duplocloud_infrastructure.infra.infra_name
}

# Use the tenant_id from the duplocloud_tenant, which will be populated after the tenant resource is created, when setting up the Redis ElastiCache cluster.
resource "duplocloud_ecache_instance" "redis_cache" {
  tenant_id           = duplocloud_tenant.tenant.tenant_id
  name                = "mycache"
  cache_type          = 0 # 0: Redis, 1: Memcache
  replicas            = 1
  size                = "cache.t2.small"
  enable_cluster_mode = true # applicable only for Redis
  number_of_shards    = 1    # applicable only for Redis
}
```

### Create an Amazon ElastiCache cluster of type Redis with 2 replicas of type cache.t2.small in dev tenant.

```terraform
# Assuming the 'dev' tenant is already created, use a data source to fetch the tenant ID.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Use the tenant_id from the duplocloud_tenant, which will be populated after the tenant resource is created, when setting up the Redis ElastiCache cluster.
resource "duplocloud_ecache_instance" "redis_cache" {
  tenant_id           = data.duplocloud_tenant.tenant.id
  name                = "mycache"
  cache_type          = 0 # 0: Redis, 1: Memcache
  replicas            = 2
  size                = "cache.t2.small"
  enable_cluster_mode = true # applicable only for Redis
  number_of_shards    = 1    # applicable only for Redis
}
```

### Create an Amazon ElastiCache of type Redis with log delivery configuration and automatic failover enabled in dev tenant.


```terraform
# Assuming the 'dev' tenant is already created, use a data source to fetch the tenant ID.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Use the tenant_id from the duplocloud_tenant, which will be populated after the tenant resource is created, when setting up the Redis ElastiCache.
resource "duplocloud_ecache_instance" "redis_cache" {
  tenant_id           = data.duplocloud_tenant.tenant.id
  name                = "mycache"
  cache_type          = 0 # 0: Redis, 1: Memcache
  replicas            = 2
  size                = "cache.t2.small"
  automatic_failover_enabled = true # minimum 2 replicas or enable_cluster_mode is true

  log_delivery_configuration {
    log_group        = "/elasticache/redis" # cloudwatch log group
    destination_type = "cloudwatch-logs"
    log_format       = "text" # log_format in text format.
    log_type         = "slow-log"  # log_type of type low-log
  }

  log_delivery_configuration {
    log_group        = "/elasticache/redis" # cloudwatch log group
    destination_type = "cloudwatch-logs"
    log_format       = "json" # log_format in json format.
    log_type         = "engine-log" # log_type of type low-log
  }
}

```

### Set up an ElastiCache Redis cluster with 2 shards and 2 cache.t2.small replicas in the dev tenant.

```terraform
# Assuming the 'dev' tenant is already created, use a data source to fetch the tenant ID.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Use the tenant_id from the duplocloud_tenant, which will be populated after the tenant resource is created, when setting up the Redis ElastiCache cluster.
resource "duplocloud_ecache_instance" "redis_cache" {
  tenant_id           = data.duplocloud_tenant.tenant.id
  name                = "mycache"
  cache_type          = 0 # 0: Redis, 1: Memcache
  replicas            = 2
  size                = "cache.t2.small"
  enable_cluster_mode = true # applicable only for Redis
  number_of_shards    = 2    # applicable only for Redis
}
```

### Create an Amazon ElastiCache cluster of type Memcached.

```terraform

# Assuming the 'dev' tenant is already created, use a data source to fetch the tenant ID.
data "duplocloud_tenant" "tenant" {
  name = "dev"
}

# Use the tenant_id from the duplocloud_tenant data source, which will be populated after the tenant data source is created, when setting up the Memcached ElastiCache cluster.
resource "duplocloud_ecache_instance" "mem_cache" {
  tenant_id           = data.duplocloud_tenant.tenant.id
  name                = "mycache"
  cache_type          = 1 # 0: Redis, 1: Memcache
  replicas            = 1
  size                = "cache.t2.small"
}
```

### Create an Amazon ElastiCache with snapshot window
```
resource "duplocloud_ecache_instance" "mycaches" {
  tenant_id  = data.duplocloud_tenant.tenant.id
   name                  = "mycache"
  cache_type            = 0
  replicas              = 2
  size                  = "cache.t3.small"
  engine_version    = "7.1"
  snapshot_window="04:00-13:00"
  
}
```

### Creating an Amazon Valkey with snapshot window
```terraform

resource "duplocloud_ecache_instance" "mycaches" {
  tenant_id                = data.duplocloud_tenant.tenant.id
  name                     = "myvalkey"
  cache_type               = 2 
  size                     = "cache.t3.medium"
  engine_version           = "7.2"
  snapshot_window          = "19:50-20:51"
  snapshot_retention_limit = 12
}

```

### Creating an Amazon Valkey with cluster mode enabled
```terraform
resource "duplocloud_ecache_instance" "mycaches" {
  tenant_id                  = data.duplocloud_tenant.tenant.id
  name                       = "tf-clust1"
  cache_type                 = 2
  size                       = "cache.t3.medium"
  engine_version             = "7.2"
  snapshot_window            = "19:50-20:51"
  snapshot_retention_limit   = 12
  enable_cluster_mode        = true
  number_of_shards           = 3
  automatic_failover_enabled = true
  replicas                   = 2
}
```
<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The short name of the elasticache instance.  Duplo will add a prefix to the name.  You can retrieve the full name from the `identifier` attribute.
- `size` (String) The instance type of the elasticache instance.
See AWS documentation for the [available instance types](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/CacheNodes.SupportedTypes.html).
- `tenant_id` (String) The GUID of the tenant that the elasticache instance will be created in.

### Optional

- `auth_token` (String) Set a password for authenticating to the ElastiCache instance.  Only supported if `encryption_in_transit` is to to `true`.

See AWS documentation for the [required format](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/auth.html) of this field.
- `automatic_failover_enabled` (Boolean) Enables automatic failover. Defaults to `false`.
- `cache_type` (Number) The numerical index of elasticache instance type.
Should be one of:

   - `0` : Redis
   - `1` : Memcache

 Defaults to `0`.
- `enable_cluster_mode` (Boolean) Flag to enable/disable redis/valkey cluster mode.
- `encryption_at_rest` (Boolean) Enables encryption-at-rest. Defaults to `false`.
- `encryption_in_transit` (Boolean) Enables encryption-in-transit. Defaults to `false`.
- `engine_version` (String) The engine version of the elastic instance.
See AWS documentation for the [available Redis and Valkey instance types](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/supported-engine-versions.html) or the [available Memcached instance types](https://docs.aws.amazon.com/AmazonElastiCache/latest/mem-ug/supported-engine-versions-mc.html).
- `kms_key_id` (String) The globally unique identifier for the key.
- `log_delivery_configuration` (Block Set, Max: 2) (see [below for nested schema](#nestedblock--log_delivery_configuration))
- `number_of_shards` (Number) The number of shards to create. Applicable only if enable_cluster_mode is set to true
- `parameter_group_name` (String) The REDIS/Valkey parameter group to supply.
- `replicas` (Number) The number of replicas to create. Supported number of replicas is 1 to 6 Defaults to `1`.
- `snapshot_arns` (List of String) Specify the ARN of a Redis/Valkey RDB snapshot file stored in Amazon S3. User should have the access to export snapshot to s3 bucket. One can find steps to provide access to export snapshot to s3 on following link https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/backups-exporting.html
- `snapshot_name` (String) Select the snapshot/backup you want to use for creating redis/valkey.
- `snapshot_retention_limit` (Number) Specify retention limit in days. Accepted values - 1-35.
- `snapshot_window` (String) Specify snapshot window limit The daily time range (in UTC) during which ElastiCache begins taking a daily snapshot of your node group (shard). Example: 05:00-09:00. If you do not specify this parameter, ElastiCache automatically chooses an appropriate time range.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `actual_engine_version` (String)
- `arn` (String) The ARN of the elasticache instance.
- `endpoint` (String) The endpoint of the elasticache instance.
- `host` (String) The DNS hostname of the elasticache instance.
- `id` (String) The ID of this resource.
- `identifier` (String) The full name of the elasticache instance.
- `instance_status` (String) The status of the elasticache instance.
- `port` (Number) The listening port of the elasticache instance.

<a id="nestedblock--log_delivery_configuration"></a>
### Nested Schema for `log_delivery_configuration`

Required:

- `destination_type` (String) destination type : must be cloudwatch-logs.
- `log_format` (String) log_format: Value must be one of the ['json', 'text']
- `log_type` (String) log_type: Value must be one of the ['slow-log', 'engine-log']

Optional:

- `log_group` (String) cloudwatch log_group


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)

## Import

Import is supported using the following syntax:

```
# Example: Importing an existing AWS ElastiCache cluster
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the AWS ElastiCache cluster
#
terraform import duplocloud_ecache_instance.mycluster v2/subscriptions/*TENANT_ID*/ECacheDBInstance/*SHORT_NAME*
```