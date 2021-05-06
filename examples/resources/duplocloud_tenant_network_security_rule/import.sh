# Example 1: Importing a "source_tenant" rule:
#  - TENANT_ID is the target tenant GUID
#  - 0 is the rule type
#  - SOURCE_TENANT is the source tenant NAME
terraform import duplocloud_tenant_network_security_rule.myrule TENANT_ID/0/SOURCE_TENANT/PROTOCOL/FROM_PORT/TO_PORT
