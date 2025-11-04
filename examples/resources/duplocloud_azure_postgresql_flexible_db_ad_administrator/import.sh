# Example: Importing an existing Azure PostgreSQL Database
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the Azure PostgreSQL Flexible Database
#  - *OBJECT_ID* Object Id of user in Active Directory

terraform import duplocloud_azure_postgresql_flexible_db_ad_administrator.adauth *TENANT_ID*/*DB_NAME*/*OBJECT_ID*
