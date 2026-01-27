// Example 1 - look up an S3 bucket by tenant ID and name.
data "duplocloud_s3_bucket" "mybucket" {
  tenant_id = var.tenant_id
  name      = "bucket-name"
}

// Output the bucket ARN
output "bucket_arn" {
  value = data.duplocloud_s3_bucket.mybucket.arn
}

// Output the full bucket name
output "bucket_fullname" {
  value = data.duplocloud_s3_bucket.mybucket.fullname
}

// Output the bucket domain name
output "bucket_domain_name" {
  value = data.duplocloud_s3_bucket.mybucket.domain_name
}
