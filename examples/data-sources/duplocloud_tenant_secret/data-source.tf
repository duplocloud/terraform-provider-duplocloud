resource "duplocloud_tenant_secret" "mysecret" {
  tenant_id = "f4bf01f0-5077-489e-aa51-95fb77049608"

  # The full name will be:  duploservices-myapp-mysecret
  name_suffix = "mysecret"

  data = "hi"
}

# To view secret, use the following data source and run `terraform output secret_value`

data "aws_secretsmanager_secret" "duplo_secret" {
  arn = duplocloud_tenant_secret.mysecret.arn
}

data "aws_secretsmanager_secret_version" "duplo_secret_version" {
  secret_id = data.aws_secretsmanager_secret.duplo_secret.id
}
