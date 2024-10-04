# Example: Importing an existing Azure storage class table
#  - *TENANT_ID* is the tenant GUID
#  - *STORAGE_ACCOUNT_NAME* is the name of the Azure storage class
#  - *NAME* is the name of the table
terraform import duplocloud_azure_storageclass_table.tbl *TENANT_ID*/*STORAGE_ACCOUNT_NAME*/table/*NAME*
