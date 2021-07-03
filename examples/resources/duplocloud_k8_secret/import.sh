# Example: Importing an existing kubernetes secret
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the config map name
#
terraform import duplocloud_k8_secret.myapp v2/subscriptions/*TENANT_ID*/K8SecretApiV2/*NAME*
