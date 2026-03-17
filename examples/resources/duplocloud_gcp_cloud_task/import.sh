# Example: Importing an existing published gcp cloud task
#  - *TENANT_ID* is the tenant GUID
#  - *QUEUE_NAME* is the  name of the queue where task is published
# - *TASK_NAME* is the name of the task to be imported

terraform import duplocloud_gcp_cloud_task.task *TENANT_ID*/queue/*QUEUE_NAME*/task/*TASK_NAME*"
