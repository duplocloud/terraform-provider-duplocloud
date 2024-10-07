# Example: Importing an existing Azure storage class blob
#  - *TENANT_ID* is the tenant GUID
#  - *STORAGE_ACCOUNT_NAME* is the name of the Azure storage class
#  - *NAME* is the name of the blob
terraform import duplocloud_azure_storageclass_blob.blob *TENANT_ID*/*STORAGE_ACCOUNT_NAME*/blob/*NAME*
