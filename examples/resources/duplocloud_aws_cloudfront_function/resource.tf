resource "duplocloud_tenant" "duplo-app" {
  account_name = "duplo-app"
  plan_id      = "default"
}


resource "duplocloud_aws_cloudfront_function" "func" {
  tenant_id = duplocloud_tenant.duplo-app.tenant_id
  name      = "test-cloudfront-func"
  runtime   = "cloudfront-js-2.0"
  code      = <<EOF
function handler(event) {
  var request = event.request;
  var headers = request.headers;

  // Add a custom header
  headers['x-custom-header'] = { value: 'Hello from CloudFront Function!' };

  return request;     
}
EOF
  comment   = "This is a test CloudFront function"

}