# Example: Importing an existing Azure Security Group Rule
#  - *INFRA_NAME* is the Duplo Infra
#  - *SG_FULL_NAME* is the fullname of the Security Group, Example- "duploinfra-<SG_SHORT_NAME>"
#  - *RULE_FULL_NAME* is the fullname of the Security Group Rule
#
terraform import duplocloud_azure_network_security_rule.security_rule *INFRA_NAME*/*SG_FULL_NAME*/*RULE_FULL_NAME*
