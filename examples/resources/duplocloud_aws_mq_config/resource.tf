resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_mq_config" "c" {
  tenant_id               = duplocloud_tenant.myapp.tenant_id
  name                    = "conf1"
  authentication_strategy = "SIMPLE"
  engine_type             = "ACTIVEMQ"
  engine_version          = "5.18"
}

// Update example: Data can only be updated after config is created
// data must be base64 encoded
resource "duplocloud_aws_mq_config" "c" {
  tenant_id               = duplocloud_tenant.myapp.tenant_id
  name                    = "conf1"
  authentication_strategy = "SIMPLE"
  engine_type             = "ACTIVEMQ"
  engine_version          = "5.18"
  description             = "second version"
  data                    = "PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiIHN0YW5kYWxvbmU9InllcyI/Pgo8YnJva2VyIHhtbG5zPSJodHRwOi8vYWN0aXZlbXEuYXBhY2hlLm9yZy9zY2hlbWEvY29yZSIgc2NoZWR1bGVQZXJpb2RGb3JEZXN0aW5hdGlvblB1cmdlPSIxMDAwMCI+CiAgPGRlc3RpbmF0aW9uUG9saWN5PgogICAgPHBvbGljeU1hcD4KICAgICAgPHBvbGljeUVudHJpZXM+CiAgICAgICAgPHBvbGljeUVudHJ5IHRvcGljPSImZ3Q7IiBnY0luYWN0aXZlRGVzdGluYXRpb25zPSJ0cnVlIiBpbmFjdGl2ZVRpbW91dEJlZm9yZUdDPSI2MDAwMDAiPgogICAgICAgICAgPHBlbmRpbmdNZXNzYWdlTGltaXRTdHJhdGVneT4KICAgICAgICAgICAgPGNvbnN0YW50UGVuZGluZ01lc3NhZ2VMaW1pdFN0cmF0ZWd5IGxpbWl0PSIxMzAwIi8+CiAgICAgICAgICA8L3BlbmRpbmdNZXNzYWdlTGltaXRTdHJhdGVneT4KICAgICAgICA8L3BvbGljeUVudHJ5PgogICAgICAgIDxwb2xpY3lFbnRyeSBxdWV1ZT0iJmd0OyIgZ2NJbmFjdGl2ZURlc3RpbmF0aW9ucz0idHJ1ZSIgaW5hY3RpdmVUaW1vdXRCZWZvcmVHQz0iNjAwMDAwIiAvPgogICAgICA8L3BvbGljeUVudHJpZXM+CiAgICA8L3BvbGljeU1hcD4KICA8L2Rlc3RpbmF0aW9uUG9saWN5PgogIDxwbHVnaW5zPgogIDwvcGx1Z2lucz4KICAKPC9icm9rZXI+"
}