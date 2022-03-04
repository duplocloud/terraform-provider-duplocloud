# Example: Importing an existing Autoscaling Policy
#  - *TENANT_ID* is the tenant GUID
#  - *SERVICE_NAMESPACE* The AWS service namespace of the scalable target
#  - *SCALABLE_DIMENSION*  The scalable dimension of the scalable target.
#  - *RESOURCE_ID* is the duploservices-<account_name>-<resource_name>
#  - *POLICY_NAME* is the name of policy
terraform import duplocloud_aws_appautoscaling_policy.asgAppPolicy *TENANT_ID*/*SERVICE_NAMESPACE*/*SCALABLE_DIMENSION*/*RESOURCE_ID*/*POLICY_NAME*
