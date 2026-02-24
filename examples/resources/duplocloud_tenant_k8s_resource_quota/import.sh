# Example: Importing an existing tenant K8s Quota
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the name given to K8s quota in specified tenant
#

terraform import duplocloud_tenant_k8s_resource_quota.quota *TENANT_ID*/k8-quota/*NAME*
