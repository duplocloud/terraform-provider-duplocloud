# Example: Importing an existing RDS Global Database.
#  - *TENANT_ID* is the tenant GUID.
#  - *GLOBAL_DATABASE* Name of the global database.
#  - *PRIMARY_REGION* The region where primary cluster has been created.
#  - *SECONDARY_REGION* The region where secondary cluster has been created.
#
terraform import duplocloud_aws_rds_global_secondary.gs *TENANT_ID*/*GLOBAL_DATABASE*/*PRIMARY_REGION*/*SECONDARY_REGION*