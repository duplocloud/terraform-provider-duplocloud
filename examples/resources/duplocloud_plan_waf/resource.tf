resource "duplocloud_plan_waf" "myplan" {
  plan_id       = "plan-name"
  waf_name      = "WebAcl name"
  waf_arn       = "WebAcl ARN"
  dashboard_url = "dashboard url"
}
 