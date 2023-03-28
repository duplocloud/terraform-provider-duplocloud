# Example: Importing an existing kubernetes storage class
#  - *TENANT_ID* is the tenant GUID
#  - *FULL_NAME* is the Duplo provided storage class name
#
terraform import duplocloud_k8_storage_class.sc v3/subscriptions/*TENANT_ID*/k8s/storageclass/*FULL_NAME*