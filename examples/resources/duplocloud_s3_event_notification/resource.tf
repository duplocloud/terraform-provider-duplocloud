resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


# Simple Example 1:  S3 event destination as lambda function

resource "duplocloud_s3_event_notification" "event" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  destination_type    = "lambda"
  destination_name    = "duploservices-<tenantname>-<lambdafunctionname>"
  event_types         = ["s3:ObjectCreated:Put", "s3:ObjectRemoved:DeleteMarkerCreated"]
  enable_event_bridge = true
}

resource "duplocloud_s3_event_notification" "event" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  destination_type    = "sqs"
  destination_name    = "duploservices-<tenantname>-<sqsname>"
  event_types         = ["s3:ObjectCreated:Put", "s3:ObjectRemoved:DeleteMarkerCreated"]
  enable_event_bridge = true
}

resource "duplocloud_s3_event_notification" "event" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  destination_type    = "sns"
  destination_name    = "duploservices-<tenantname>-<snsname>"
  event_types         = ["s3:ObjectCreated:Put", "s3:ObjectRemoved:DeleteMarkerCreated"]
  enable_event_bridge = true
}