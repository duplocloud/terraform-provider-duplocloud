resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


# Simple Example 1:  S3 event destination as lambda function

resource "duplocloud_s3_event_notification" "event" {
  tenant_id   = duplocloud_tenant.myapp.tenant_id
  bucket_name = "duploservices-<tenantname>-<bucket>"
  event {
    destination_type = "lambda"
    destination_arn  = "<arn-of-destination-type>"
    event_types      = ["s3:ObjectCreated:Put", "s3:ObjectRemoved:DeleteMarkerCreated"]
  }
  enable_event_bridge = true

}

# Simple Example 2:  S3 event destination as SQS

resource "duplocloud_s3_event_notification" "event" {
  tenant_id   = duplocloud_tenant.myapp.tenant_id
  bucket_name = "duploservices-<tenantname>-<bucket>"
  event {
    destination_type = "sqs"
    destination_arn  = "<arn-of-destination-type>"
    event_types      = ["s3:ObjectCreated:Put", "s3:ObjectRemoved:DeleteMarkerCreated"]
  }
  enable_event_bridge = true

}

# Simple Example 3:  S3 event destination as SNS

resource "duplocloud_s3_event_notification" "event" {
  tenant_id   = duplocloud_tenant.myapp.tenant_id
  bucket_name = "duploservices-<tenantname>-<bucket>"
  event {
    destination_type = "sns"
    destination_arn  = "<arn-of-destination-type>"
    event_types      = ["s3:ObjectCreated:Put", "s3:ObjectRemoved:DeleteMarkerCreated"]
  }
  enable_event_bridge = true

}