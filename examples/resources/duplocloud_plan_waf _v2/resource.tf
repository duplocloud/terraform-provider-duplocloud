resource "duplocloud_plan_waf_v2" "myplan" {
  plan_id = "plan-name"
  waf {
    name          = "WebAcl name"
    arn           = "WebAcl ARN"
    dashboard_url = "dashboard url"
  }
}
 