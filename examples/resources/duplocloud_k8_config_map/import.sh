# Example: Importing an existing kubernetes config map
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the config map name
#
terraform import duplocloud_k8_config_map.myapp v2/subscriptions/*TENANT_ID*/K8ConfigMapApiV2/*NAME*
