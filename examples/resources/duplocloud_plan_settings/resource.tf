resource "duplocloud_plan_settings" "myplanSettings" {
  plan_id             = "myplan"
  unrestricted_ext_lb = true
  dns_setting {
    domain_id           = "Z02791752705G9GHH8CYF"
    internal_dns_suffix = ".test.duplocloud.net"
    external_dns_suffix = ".test.duplocloud.net"
  }
}

