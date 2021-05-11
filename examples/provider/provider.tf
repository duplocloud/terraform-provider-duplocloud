provider "duplocloud" {
  duplo_host  = "https://mycompany.duplocloud.net" # optionally use the `duplo_host` env var
  duplo_token = "...MY API TOKEN..."               # recommended to use the `duplo_token` env var
  # (to avoid accidentally committing secrets into your SCM)
}
