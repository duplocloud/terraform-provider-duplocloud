
//recurring maintenance example
resource "duplocloud_gcp_infra_maintenance_window" "mw" {
  infra_name = "infra-name"
  recurring_window {
    start_time = "2024-10-05T00:10:00"
    end_time   = "2024-11-06T00:10:00"
    recurrence = "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR,SA,SU"
  }
}

//example with exclusion
resource "duplocloud_gcp_infra_maintenance_window" "mw" {
  infra_name = "me2"
  recurring_window {
    start_time = "2024-10-05T00:10:00"
    end_time   = "2024-11-06T00:10:00"
    recurrence = "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR,SA,SU"
  }
  exclusions {
    start_time = "2024-10-06T00:00:00"
    end_time   = "2024-10-07T00:00:00"
    scope      = "NO_UPGRADES"
  }
}

//recurrencing maintenance example
resource "duplocloud_gcp_infra_maintenance_window" "mw" {
  infra_name                   = "infra-name"
  daily_maintenance_start_time = "18:06"

}
