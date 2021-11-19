# Example: Importing an existing Azure Key Vault Secret
#  - *TENANT_ID* is the tenant GUID
#  - *SHORT_NAME* is the short name of the Azure Key Vault Secret
#
terraform import duplocloud_azure_key_vault_secret.mykvsecret *TENANT_ID*/*SHORT_NAME*
