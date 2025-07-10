# Example 1: Importing a "source_tenant" rule:
#  - TENANT_ID is the target tenant GUID
#  - TYPE SG rule type integer value Tenant SG = 0 IP Address SG = 1 
#  - SOURCE_TENANT for type 0, this is the source tenant name
#  - SOURCE_ADDRESS for TYPE = 1, this is the source IP address or CIDR block, format 10.220.32.192-32.  
#  - PROTOCOL is the protocol (tcp, udp, icmp)
#  - FROM_PORT is the starting port (0-65535)
#  - TO_PORT is the ending port (0-65535)
terraform import duplocloud_tenant_network_security_rule.myrule *TENANT_ID*/*TYPE*/*SOURCE_TENANT*/*PROTOCOL*/*FROM_PORT*/*TO_PORT*

terraform import duplocloud_tenant_network_security_rule.myrule *TENANT_ID*/*TYPE*/*SOURCE_ADDRESS*/*PROTOCOL*/*FROM_PORT*/*TO_PORT*
