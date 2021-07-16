resource "duplocloud_plan_configs" "myplan" {
  plan_id = "myplan"

  config {
    key   = "foo"
    type  = ""
    value = "bar"
  }

  config {
    key  = "my docker repo"
    type = "DockerRegistryCreds"
    value = jsonencode({
      registry = "https://index.docker.io/v1/"
      username = "MY-USERNAME"
      password = "MY-PASSWORD"
    })
  }
}
