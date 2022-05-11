# Example: Importing an existing Azure MS SQL databse
#  - *TENANT_ID* is the tenant GUID
#  - *SERVER_NAME* is the short name of the Azure MS SQL Server
#  - *EP_NAME* is the short name of the Azure MS SQL Elastic Pool
terraform import duplocloud_azure_mssql_elasticpool.myMsSqlEP *TENANT_ID*/*SERVER_NAME*/*EP_NAME*
