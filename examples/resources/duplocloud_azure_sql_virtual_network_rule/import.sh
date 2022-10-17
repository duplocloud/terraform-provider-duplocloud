# Example: Importing an existing Azure SQL server Vnet Rule
#  - *TENANT_ID* is the tenant GUID
#  - *SERVER_NAME* is the name of the Azure Sql server
#  - *RULE_NAME* is the name of the Azure Sql server Vnet Rule

terraform import duplocloud_azure_sql_virtual_network_rule.sql_vnet_rule *TENANT_ID*/*SERVER_NAME*/*RULE_NAME*
