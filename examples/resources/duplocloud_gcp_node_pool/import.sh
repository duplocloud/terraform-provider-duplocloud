# Example: Importing an existing GCP Node Pool
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the  name of the Node Pool
#

terraform import duplocloud_gcp_node_pools.node_pool v3/subscriptions/*TENANT_ID*/google/nodePools/*NAME* 