# Example: Importing an existing Azure storage account share file
#  - *TENANT_ID* is the tenant GUID
#  - *STORAGE_ACCOUNT_NAME* is the name of the Azure storage account
#  - *NAME* is the name of the queue
terraform import duplocloud_azure_storageclass_queue.qu *TENANT_ID*/*STORAGE_ACCOUNT_NAME*/queue/*NAME*
