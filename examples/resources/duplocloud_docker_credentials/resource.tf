resource "duplocloud_tenant" "duplo-app" {
  account_name   = "duplo-app"
  plan_id        = "default"
  allow_deletion = true
}

resource "duplocloud_docker_credentials" "docker_creds" {
  tenant_id = duplocloud_tenant.duplo-app.tenant_id
  user_name = "myname"
  email     = "abc@xyz.com"
  password  = "p@assW0rd"
}