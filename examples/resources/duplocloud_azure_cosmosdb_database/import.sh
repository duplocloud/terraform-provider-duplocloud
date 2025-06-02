# Example: Importing an existing Azure CosmosDB Database
#  - *TENANT_ID* is the tenant GUID
#  - *ACCOUNT_NAME* is the name of the Azure Cosmosdb Account
#  - *NAME* is the name of the Azure Cosmosdb database belonging to the account
#
terraform import duplocloud_azure_cosmos_db_account.account *TENANT_ID*/cosmosdb/*ACCOUNT_NAME*/database/*NAME*