resource "duplocloud_plan_kms" "myplan" {
  plan_id = "plan-name"
  kms {
    id   = "kms-id"
    arn  = "kms-arn"
    name = "kms-name"
  }
}
