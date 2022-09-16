# Example: Importing an existing kubernetes secret provider class
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the secret provider class name
#
terraform import duplocloud_k8_secret_provider_class.spc v3/subscriptions/*TENANT_ID*/k8s/secretproviderclass/*NAME*