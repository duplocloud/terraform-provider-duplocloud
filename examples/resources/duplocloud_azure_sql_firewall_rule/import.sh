# Example: Importing an existing Azure SQL server Firewall Rule
#  - *TENANT_ID* is the tenant GUID
#  - *SERVER_NAME* is the name of the Azure Sql server
#  - *RULE_NAME* is the name of the Azure Sql server Firewall Rule

terraform import duplocloud_azure_sql_firewall_rule.sql_firewall_rule *TENANT_ID*/*SERVER_NAME*/*RULE_NAME*
