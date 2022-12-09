# Example: Importing an existing kubernetes Persistent Volume Claim
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the name of Persistent Volume Claim
#
terraform import duplocloud_k8_persistent_volume_claim.pvc v3/subscriptions/*TENANT_ID*/k8s/pvc/*NAME*