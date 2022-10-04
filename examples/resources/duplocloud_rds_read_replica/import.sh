# Example: Importing an existing RDS read replica
#  - *TENANT_ID* is the tenant GUID
#  - *SHORTNAME* is the short name of the database read replica (without the duplo prefix)
#
terraform import duplocloud_rds_read_replica.read_replica v2/subscriptions/*TENANT_ID*/RDSDBInstance/*SHORTNAME*
