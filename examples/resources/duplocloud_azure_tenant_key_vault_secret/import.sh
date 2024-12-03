# Example: Importing an existing Azure Key Vault Secret
#  - *TENANT_ID* is the tenant GUID
#  - *VAULT_NAME* is the name of the Azure Key Vault
#  - *SECRET_NAME* is the name of the Azure Key Vault Secret

terraform import duplocloud_azure_tenant_key_vault_secret.kv_secret *TENANT_ID*/*VAULT_NAME*/*SECRET_NAME*
