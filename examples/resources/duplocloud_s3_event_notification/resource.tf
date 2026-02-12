resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


# Simple Example 1:  S3 event destination as lambda function

resource "duplocloud_s3_event_notification" "event" {
  tenant_id   = duplocloud_tenant.myapp.tenant_id
  bucket_name = "duploservices-<tenantname>-<bucket1>"
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
  bucket_name = "duploservices-<tenantname>-<bucket2>"
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
  bucket_name = "duploservices-<tenantname>-<bucket3>"
  event {
    destination_type = "sns"
    destination_arn  = "<arn-of-destination-type>"
    event_types      = ["s3:ObjectCreated:Put", "s3:ObjectRemoved:DeleteMarkerCreated"]
  }
  enable_event_bridge = true

}

# Simple Example 4: Multiple event type with different destination for an S3 bucket

resource "duplocloud_s3_event_notification" "event4" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  bucket_name         = "duploservices-sep8-eventtests-182680712604"
  enable_event_bridge = false
  event {
    destination_type = "lambda"
    event_types      = ["s3:ObjectCreated:Put"]
    destination_arn  = "<arn-of-destination-type>"
  }
  event {
    destination_type = "sns"
    destination_arn  = "<arn-of-destination-type>"
    event_types      = ["s3:ObjectCreated:CompleteMultipartUpload"]
  }
  event {
    destination_type = "sqs"
    event_types      = ["s3:ObjectCreated:Post", "s3:ObjectRemoved:DeleteMarkerCreated"]
    destination_arn  = "<arn-of-destination-type>"
  }

}

# Simple Example 5: Multiple event type with same destination for an S3 bucket

resource "duplocloud_s3_event_notification" "event5" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  bucket_name         = "duploservices-sep8-eventtests-182680712604"
  enable_event_bridge = false
  event {
    destination_type = "sqs"
    event_types      = ["s3:ObjectCreated:Put"]
    destination_arn  = "<arn-of-destination-type>"
  }
  event {
    destination_type = "sqs"
    destination_arn  = "<arn-of-destination-type>"
    event_types      = ["s3:ObjectCreated:CompleteMultipartUpload"]
  }
  event {
    destination_type = "sqs"
    event_types      = ["s3:ObjectCreated:Post", "s3:ObjectRemoved:DeleteMarkerCreated"]
    destination_arn  = "<arn-of-destination-type>"
  }

}
