---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_tenant_secret Data Source - terraform-provider-duplocloud"
subcategory: ""
description: |-
  
---

# duplocloud_tenant_secret (Data Source)



## Example Usage

```terraform
data "duplocloud_tenant_secret" "mysecret" {
  tenant_id = "f4bf01f0-5077-489e-aa51-95fb77049608"

  # The full name will be:  duploservices-myapp-mysecret
  name_suffix = "mysecret"
}

# To view secret, use the following data source and run `terraform output secret_value`

data "aws_secretsmanager_secret" "duplo_secret" {
  arn = duplocloud_tenant_secret.mysecret.arn
}

data "aws_secretsmanager_secret_version" "duplo_secret_version" {
  secret_id = data.aws_secretsmanager_secret.duplo_secret.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `tenant_id` (String)

### Optional

- `arn` (String)
- `first_match` (Boolean) Defaults to `true`.
- `name` (String)
- `name_suffix` (String)

### Read-Only

- `id` (String) The ID of this resource.
- `rotation_enabled` (Boolean)
- `tags` (List of Object) (see [below for nested schema](#nestedatt--tags))

<a id="nestedatt--tags"></a>
### Nested Schema for `tags`

Read-Only:

- `key` (String)
- `value` (String)
