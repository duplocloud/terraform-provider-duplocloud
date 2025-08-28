# Example: Importing an existing AWS ElastiCache Secondary cluster in a global datastore
#  - *TENANT_ID* is the tenant GUID
#  - *SECONDARY_TENANT_ID* tenant id where secondary cluster resides
#  - *GLOBAL_DATASTORE* global datastore name
#  - *SECONDARY_CLUSTER* secondary cluster name
terraform import duplocloud_ecache_associate_global_secondary_cluster.sc *TENANT_ID*/ecacheReplicationGroup/*SECONDARY_TENANT_ID*/*GLOBAL_DATASTORE*/*SECONDARY_CLUSTER*