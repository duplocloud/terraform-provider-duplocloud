# Example: Importing an existing kubernetes ReferenceGrant
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the ReferenceGrant name
#
terraform import duplocloud_k8s_reference_grant.mygrant v3/subscriptions/*TENANT_ID*/k8s/referencegrant/*NAME*
