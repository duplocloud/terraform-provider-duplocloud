# Example 1: Importing a "source_tenant" rule:
#  - TENANT_ID is the target tenant GUID
#  - TYPE (0/1) : If the source in the Duplo portal is a tenant name, represent the TYPE as 0. If the source is an IP address, represent the TYPE as 1.
#  - SOURCE_TENANT : Name of security group's source tenant
#  - SOURCE_ADDRESS IP address or CIDR block, format 10.220.32.192-32.  
#  - PROTOCOL is the protocol (tcp, udp, icmp)
#  - FROM_PORT is the starting port (0-65535)
#  - TO_PORT is the ending port (0-65535)
terraform import duplocloud_tenant_network_security_rule.myrule *TENANT_ID*/*TYPE*/*SOURCE_TENANT*/*PROTOCOL*/*FROM_PORT*/*TO_PORT*
#Example : terraform import duplocloud_tenant_network_security_rule.myrule {GUID}/0/abc/tcp/443/443


terraform import duplocloud_tenant_network_security_rule.myrule *TENANT_ID*/*TYPE*/*SOURCE_ADDRESS*/*PROTOCOL*/*FROM_PORT*/*TO_PORT*
#Example : terraform import duplocloud_tenant_network_security_rule.myrule {GUID}/1/10.34.0.1-32/tcp/443/443
