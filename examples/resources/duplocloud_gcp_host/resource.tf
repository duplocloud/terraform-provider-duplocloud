resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_gcp_host" "host" {
  tenant_id      = duplocloud_tenant.myapp.tenant_id
  friendly_name  = "tfnewhost"
  capacity       = "e2-medium"
  zone           = "us-west2-a"
  agent_platform = 0
  metadata = {
    "OsDiskSize"     = "10"
    "startup_script" = "echo \"Hello from test startup script!\" > /test.txt\n"
  }
  tags     = ["networktag"]
  image_id = "projects/{project}/global/images/{image}"
  labels = {
    "resource" : "label"
  }
  user_account        = "abc@xyz.com"
  allocated_public_ip = true
}