# Example: Importing an existing kubernetes daemonset
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the k8s daemonset name
#
terraform import duplocloud_k8s_daemon_set.myapp v3/subscriptions/*TENANT_ID*/k8s/daemonSet/*NAME*
