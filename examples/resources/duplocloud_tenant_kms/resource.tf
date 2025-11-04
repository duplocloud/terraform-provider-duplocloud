resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_tenant_kms" "mykey" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  kms {
    id   = "ebdc0518-e512-4cf8-b09a-5cbf2797a736"
    arn  = "asdewcasdcasawwaqwq"
    name = "tftest"
  }
  kms {
    id   = "ebdc0518-e512-4cf8-b09a-5cbf2797a737"
    arn  = "asdewcasdcasawwaqwq"
    name = "tftest1"
  }

}