# Example: Importing an existing Azure MS SQL databse
#  - *TENANT_ID* is the tenant GUID
#  - *SERVER_NAME* is the short name of the Azure MS SQL Server
#  - *DB_NAME* is the short name of the Azure MS SQL Database
terraform import duplocloud_azure_mssql_database.myMsSqlDb *TENANT_ID*/*SERVER_NAME*/*DB_NAME*
