# Example: Importing an existing Kafka cluster
#  - *TENANT_ID* is the tenant GUID
#  - *SHORTNAME* is the short name of the cluster (without the duploservices prefix)
#
terraform import duplocloud_kafka_cluster.mycluster *TENANT_ID*/*SHORTNAME*
