# Example: Importing an existing queue on which task are published
#  - *TENANT_ID* is the tenant GUID
#  - *QUEUE_NAME* is the  name of the queue where task is published


terraform import duplocloud_gcp_cloud_queue.queue *TENANT_ID/*QUEUE_NAME*