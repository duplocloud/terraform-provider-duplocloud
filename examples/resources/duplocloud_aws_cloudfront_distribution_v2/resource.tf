resource "duplocloud_tenant" "duplo-app" {
  account_name = "duplo-app"
  plan_id      = "default"
}
locals {
  region      = "us-west2"
  tenant_name = "nonprod"
}


resource "duplocloud_s3_bucket" "forcloudfront" {
  tenant_id           = duplocloud_tenant.duplo-app.tenant_id
  name                = "cloudfrontbucket"
  allow_public_access = true
  enable_access_logs  = false
  enable_versioning   = true
  managed_policies    = ["ssl"]
  default_encryption {
    method = "Sse"
  }
}


resource "duplocloud_aws_cloudfront_distribution_v2" "cloudfront" {
  tenant_id                 = duplocloud_tenant.duplo-app.tenant_id
  name                      = "mydistribution"
  enabled                   = true
  http_version              = "http2"
  price_class               = "PriceClass_All"
  is_ipv6_enabled           = true
  use_origin_access_control = true
  viewer_certificate {
    acm_certificate_arn      = "arn:aws:acm:us-east-1:100000000000:certificate/75e94c99-b916-459c-b9a1-ed9dec0ae550"
    minimum_protocol_version = "TLSv1.2_2021"
    ssl_support_method       = "sni-only"
  }
  origin {
    connection_attempts = 3
    connection_timeout  = 10
    domain_name         = "${duplocloud_s3_bucket.forcloudfront.fullname}.s3.${local.region}.amazonaws.com"
    origin_id           = "${duplocloud_s3_bucket.forcloudfront.fullname}.s3.${local.region}.amazonaws.com"
    origin_path         = "/acme-portal.dev.shastacloud.io"
  }
  default_cache_behavior {
    target_origin_id = "${duplocloud_s3_bucket.forcloudfront.fullname}.s3.${local.region}.amazonaws.com"
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    #cache_policy_id        = "658327ea-f89d-4fab-a63d-7e88639e58f6"
    compress               = true
    viewer_protocol_policy = "redirect-to-https"
    max_ttl                = 3
    min_ttl                = 2
    default_ttl            = 3
    forwarded_values {
      query_string = true

      cookies {
        forward           = "whitelist"
        whitelisted_names = ["session-id", "user-id"]
      }

      headers                 = ["Authorization", "Host"]
      query_string_cache_keys = ["lang", "user"]
    }

  }

  ordered_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "${duplocloud_s3_bucket.forcloudfront.fullname}.s3.${local.region}.amazonaws.com"
    compress               = false
    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 1
    default_ttl            = 2
    max_ttl                = 3
    path_pattern           = "/api/*"
    forwarded_values {
      cookies {
        forward = "all"
      }
      query_string = false
    }
  }
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }
  cors_allowed_host_names = ["https://app.example.com"]
  lifecycle {
    ignore_changes = [
      wait_for_deployment,
    ]
  }
}
