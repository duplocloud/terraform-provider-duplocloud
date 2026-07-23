# Example: Importing an existing kubernetes HTTPRoute
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the HTTPRoute name
#
terraform import duplocloud_k8s_http_route.myroute v3/subscriptions/*TENANT_ID*/k8s/httproute/*NAME*
