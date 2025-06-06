---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_aws_tag Resource - terraform-provider-duplocloud"
subcategory: ""
description: |-
  duplocloud_aws_tag manages an AWS custom tag for resources in Duplo.
---

# duplocloud_aws_tag (Resource)

`duplocloud_aws_tag` manages an AWS custom tag for resources in Duplo.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `arn` (String) The resource arn of which custom tag need to be created.
- `key` (String) The tag name.
- `tenant_id` (String) The GUID of the tenant that the custom tag for a resource will be created in.
- `value` (String) The value of the tag.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)

## Import

Import is supported using the following syntax:

```shell
# Example: Importing an existing AWS tag
#  - *TENANT_ID* is the tenant GUID
#  - *ARN* The resource arn. Should be encoded thrice with URL encoding.
#  - *TAGKEY* Key of the tag
#
terraform import duplocloud_aws_tag.custom *TENANT_ID*/*ARN*/*TAGKEY*
```
