resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


resource "duplocloud_gcp_cloud_queue_task" "task" {
  tenant_id  = duplocloud_tenant.myapp.tenant_id
  queue_name = duplocloud_gcp_cloud_tasks_queue.queue.name
  name       = "task1"
  app_engine {
    relative_uri = "/create"
    body         = "ewogICAibmFtZSI6ICJBcHBsZSBNYWNCb29rIFBybyAxNiIsCiAgICJkYXRhIjogewogICAgICAieWVhciI6IDIwMTksCiAgICAgICJwcmljZSI6IDE4NDkuOTksCiAgICAgICJDUFUgbW9kZWwiOiAiSW50ZWwgQ29yZSBpOSIsCiAgICAgICJIYXJkIGRpc2sgc2l6ZSI6ICIxIFRCIgogICB9Cn0="
    method       = "POST"
    headers = {
      Content-Type = "application/json"
    }
  }

}

resource "duplocloud_gcp_cloud_queue_task" "task2" {
  tenant_id  = duplocloud_tenant.myapp.tenant_id
  queue_name = duplocloud_gcp_cloud_tasks_queue.queue.name
  name       = "task2"
  http_target {
    url    = "https://catfact.ninja/fact"
    method = "POST"
    body   = "ewogICAibmFtZSI6ICJBcHBsZSBNYWNCb29rIFBybyAxNiIsCiAgICJkYXRhIjogewogICAgICAieWVhciI6IDIwMTksCiAgICAgICJwcmljZSI6IDE4NDkuOTksCiAgICAgICJDUFUgbW9kZWwiOiAiSW50ZWwgQ29yZSBpOSIsCiAgICAgICJIYXJkIGRpc2sgc2l6ZSI6ICIxIFRCIgogICB9Cn0="
    headers = {
      Content-Type = "application/json"
    }
  }
}
