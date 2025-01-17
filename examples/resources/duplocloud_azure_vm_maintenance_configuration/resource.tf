resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


resource "duplocloud_azure_vm_maintenance_configuration" "mt" {
  tenant_id = duplocloud_tenant.myapp.tenant_id
  vm_name   = "schedl"
  window {
    start_time      = "2024-11-12 00:00"
    expiration_time = "2024-11-19 00:00"
    duration        = "06:00"
    recur_every     = "1Month day1,day2,day3,day4,day5,day6,day7,day8,day9,day10,day11,day12,day13,day14,day15,day16,day17,day18,day19,day20,day21,day22,day23,day24,day25,day26,day27,day28,day29,day30,day31,day-1"
    time_zone       = "India Standard Time"
  }
}
