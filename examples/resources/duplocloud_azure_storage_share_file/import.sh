# Example: Importing an existing Azure storage account share file
#  - *TENANT_ID* is the tenant GUID
#  - *STORAGE_ACCOUNT_NAME* is the name of the Azure storage account
#  - *NAME* is the name of the share file
terraform import duplocloud_azure_storage_share_file.share_file *TENANT_ID*/*STORAGE_ACCOUNT_NAME*/*NAME*
