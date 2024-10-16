# Example: Importing an existing kubernetes job
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the k8s job name
#
terraform import duplocloud_k8s_job.myapp v3/subscriptions/*TENANT_ID*/k8s/job/*NAME*
