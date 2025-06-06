---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_gcp_sql_database_instance Resource - terraform-provider-duplocloud"
subcategory: ""
description: |-
  duplocloud_gcp_sql_database_instance manages a GCP SQL Database Instance in Duplo.
---

# duplocloud_gcp_sql_database_instance (Resource)

`duplocloud_gcp_sql_database_instance` manages a GCP SQL Database Instance in Duplo.

## Example Usage

```terraform
resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_gcp_sql_database_instance" "sql_instance" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  name             = "sqlinstances01"
  database_version = "MYSQL_8_0"
  tier             = "db-n1-standard-1"
  disk_size        = 17
  labels = {
    managed-by = "duplocloud"
    created-by = "terraform"
  }
}


// Backup configuration example
resource "duplocloud_gcp_sql_database_instance" "sql" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  name             = "mysqlbckp"
  database_version = "POSTGRES_14"
  disk_size        = 10
  tier             = "db-f1-micro"

  root_password = "qwerty"
  need_backup   = true
}

resource "duplocloud_gcp_sql_database_instance" "sql_instance" {
  tenant_id        = duplocloud_tenant.myapp.tenant_id
  name             = "customtier"
  database_version = "SQLSERVER_2019_STANDARD"
  tier             = "db-custom-2-7680"
  disk_size        = 19
  root_password    = "test@123Abc"
  labels = {
    managed-by = "duplocloud"
    created-by = "terraform"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `database_version` (String) The MySQL, PostgreSQL or SQL Server version to use.Supported values include `MYSQL_5_6`,`MYSQL_5_7`, `MYSQL_8_0`, `POSTGRES_9_6`,`POSTGRES_10`,`POSTGRES_11`,`POSTGRES_12`, `POSTGRES_13`, `POSTGRES_14`, `POSTGRES_15`,`POSTGRES_16`,`POSTGRES_17`, `SQLSERVER_2017_STANDARD`,`SQLSERVER_2017_ENTERPRISE`,`SQLSERVER_2017_EXPRESS`, `SQLSERVER_2017_WEB`,`SQLSERVER_2019_STANDARD`, `SQLSERVER_2019_ENTERPRISE`, `SQLSERVER_2019_EXPRESS`,`SQLSERVER_2019_WEB`,`SQLSERVER_2022_WEB`,`SQLSERVER_2022_EXPRESS`,`SQLSERVER_2022_ENTERPRISE`,`SQLSERVER_2022_STANDARD`.[Database Version Policies](https://cloud.google.com/sql/docs/db-versions) includes an up-to-date reference of supported versions.
- `name` (String) The short name of the sql database.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.
- `tenant_id` (String) The GUID of the tenant that the sql database will be created in.
- `tier` (String) The machine type to use. See tiers for more details and supported versions. Postgres supports only shared-core machine types, and custom machine types, format for custom machine type db-custom-{vCPU}-{memory-in-MB} example `db-custom-2-13312`.See the [Machine Type Documentation](https://cloud.google.com/compute/docs/machine-resource) to learn more about machine types.

### Optional

- `disk_size` (Number) The size of data disk, in GB. Size of a running instance cannot be reduced but can be increased. The minimum value is 10GB.
- `labels` (Map of String) Map of string keys and values that can be used to organize and categorize this resource.
- `need_backup` (Boolean) Flag to enable backup process on delete of database Defaults to `true`.
- `root_password` (String) Provide root password for specific database versions.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `wait_until_ready` (Boolean) Whether or not to wait until sql database instance to be ready, after creation. Defaults to `true`.

### Read-Only

- `connection_name` (String) Connection name of the database.
- `fullname` (String) The full name of the sql database.
- `id` (String) The ID of this resource.
- `ip_address` (List of String) List of IP addresses of the database.
- `self_link` (String) The SelfLink of the sql database.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)

## Import

Import is supported using the following syntax:

```shell
# Example: Importing an existing GCP SQL database instance
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the GCP SQL database instance
#
terraform import duplocloud_gcp_sql_database_instance.sql_instance *TENANT_ID*/*SHORT_NAME*
```
