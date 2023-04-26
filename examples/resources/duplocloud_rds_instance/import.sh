# Example: Importing an existing RDS instance
#  - *TENANT_ID* is the tenant GUID
#  - *SHORTNAME* is the short name of the database (without the duplo prefix)
#
terraform import duplocloud_rds_instance.mydb v3/subscriptions/*TENANT_ID*/RDSDBInstance/*SHORTNAME*
