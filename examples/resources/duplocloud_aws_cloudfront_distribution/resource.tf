resource "duplocloud_tenant" "duplo-app" {
  account_name = "duplo-app"
  plan_id      = "default"
}

resource "duplocloud_aws_cloudfront_distribution" "cfd" {
  tenant_id           = duplocloud_tenant.duplo-app.tenant_id
  comment             = "duploservices-dev01-app"
  default_root_object = "index.html"
  enabled             = true
  http_version        = "http2"
  is_ipv6_enabled     = true
  aliases             = ["app-dev.abc.xyz"]
  price_class         = "PriceClass_All"
  wait_for_deployment = true

  origin {
    domain_name         = "pa-api-dev01.abc.xyz"
    origin_id           = "pa-api-dev01.abc.xyz"
    connection_attempts = 3
    connection_timeout  = 10
    custom_origin_config {
      http_port                = 80
      https_port               = 443
      origin_keepalive_timeout = 30
      origin_read_timeout      = 30
      origin_protocol_policy   = "https-only"
      origin_ssl_protocols     = ["TLSv1.2"]
    }
  }

  origin {
    domain_name         = "duploservices-dev01-abc-app-1234567890.s3.us-west-2.amazonaws.com"
    origin_id           = "duploservices-dev01-abc-app-1234567890.s3.us-west-2.amazonaws.com"
    connection_attempts = 3
    connection_timeout  = 10
  }


  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "duploservices-dev01-abc-app-1234567890.s3.us-west-2.amazonaws.com"
    compress               = false
    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 0
    max_ttl                = 0
  }

  ordered_cache_behavior {
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "pa-api-dev01.abc.xyz"
    compress               = false
    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 0
    max_ttl                = 0
    path_pattern           = "/api/*"
  }

  viewer_certificate {
    acm_certificate_arn      = "arn:aws:acm:us-east-1:1234567890:certificate/49c7a151-14b1-486e-801f-cf6bc9a43804"
    minimum_protocol_version = "TLSv1.2_2019"
    ssl_support_method       = "sni-only"
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  lifecycle {
    ignore_changes = [
      wait_for_deployment,
    ]
  }

}
