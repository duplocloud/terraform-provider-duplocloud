# Example: Importing an existing Azure CosmosDB Container of a Database in an Account
#  - *TENANT_ID* is the tenant GUID
#  - *ACCOUNT_NAME* is the name of the Azure Cosmosdb Account
#  - *DATABASE_NAME* is the name of the Azure Cosmosdb database belonging to the account
#  - *NAME* is the name of the Azure Cosmosdb Container belonging to the Database
#
terraform import duplocloud_azure_cosmos_db_container.acontainer *TENANT_ID*/cosmosdb/*ACCOUNT_NAME*/*DATABASE_NAME*/container/*NAME*