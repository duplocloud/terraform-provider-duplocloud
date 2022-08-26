# Example: Importing an existing kubernetes ingress
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the ingress name
#
terraform import duplocloud_k8_ingress.ingress v3/subscriptions/*TENANT_ID*/k8s/ingress/*NAME*