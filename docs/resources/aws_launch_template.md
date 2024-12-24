---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_aws_launch_template Resource - terraform-provider-duplocloud"
subcategory: ""
description: |-
  duplocloudawslaunch_template creates the new version over current launch template version
---

# duplocloud_aws_launch_template (Resource)

duplocloud_aws_launch_template creates the new version over current launch template version

## Example Usage

```terraform
resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_launch_template" "lt" {
  tenant_id           = duplocloud_tenant.myapp.tenant_id
  name                = "launch-template-name"
  instance_type       = "t3a.medium"
  version             = "1"
  version_description = "launch template description"
  ami                 = "ami-123test"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The fullname of the asg group
- `tenant_id` (String) The GUID of the tenant that the launch template will be created in.
- `version` (String) Any of the existing version of the launch template
- `version_description` (String) The version of the launch template

### Optional

- `ami` (String) Asg ami to be used to update the version from the current version
- `instance_type` (String) Asg instance type to be used to update the version from the current version
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `default_version` (String) The current default version of the launch template.
- `id` (String) The ID of this resource.
- `latest_version` (String) The latest launch template version
- `version_metadata` (String)

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)

## Import

Import is supported using the following syntax:

```shell
# Example: Importing an existing AWS launch template
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the name of the AWS launch template
#  - *VERSION* available version of launch template

terraform import duplocloud_aws_launch_template.lt *TENANT_ID*/launch-template/*NAME*/*VERSION*
```