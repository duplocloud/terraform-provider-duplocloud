# Example: Importing an existing  Retention Backup value for Azure MS SQL Server's Database
#  - *TENANT_ID* is the tenant GUID
#  - *SERVER_NAME* is the short name of the Azure MS SQL Server
#  - *DB_NAME* is the short name of the Azure MS SQL Database

terraform import duplocloud_azure_mssqldb_retention_backup.backup *TENANT_ID*/retention-backup/*SERVER_NAME*/*DB_NAME*
