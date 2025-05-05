resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


resource "duplocloud_asg_instance_refresh" "name" {
  tenant_id              = duplocloud_tenant.myapp.tenant_id
  asg_name               = "asg-name"
  refresh_identifier     = "1" #any identifier values can be used
  max_healthy_percentage = 100
  min_healthy_percentage = 90
}