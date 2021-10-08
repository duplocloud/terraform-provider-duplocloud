# Example: Importing an existing AWS cloudwatch event target
#  - *TENANT_ID* is the tenant GUID
#  - *FRIENDLY_NAME* is the duploservices-<account_name>-<name_of_event_rule>
#  - *TARGET_ID* The unique target assignment ID.

terraform import duplocloud_aws_cloudwatch_event_target.myEventTarget *TENANT_ID*/*FRIENDLY_NAME*/*TARGET_ID*