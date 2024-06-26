---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_gcp_firestore Resource - terraform-provider-duplocloud"
subcategory: ""
description: |-
  duplocloud_gcp_firestore manages a GCP firestore in Duplo.
---

# duplocloud_gcp_firestore (Resource)

`duplocloud_gcp_firestore` manages a GCP firestore in Duplo.

## Example Usage

```terraform
resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_gcp_firestore" "app" {
  tenant_id                     = duplocloud_tenant.myapp.tenant_id
  name                          = "firestore-tf-2"
  type                          = "FIRESTORE_NATIVE"
  location_id                   = "us-west2"
  enable_delete_protection      = false
  enable_point_in_time_recovery = false
}

resource "duplocloud_gcp_firestore" "firestore-app" {

}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `location_id` (String) Location for firestore
- `name` (String) The short name of the firestore.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.
- `tenant_id` (String) The GUID of the tenant that the firestore will be created in.
- `type` (String) Firestore supports type `FIRESTORE_NATIVE` and `DATASTORE_MODE`

### Optional

- `enable_delete_protection` (Boolean) Delete protection prevents accidental deletion of firestore. Defaults to `false`.
- `enable_point_in_time_recovery` (Boolean) Restores data to a specific moment in time, enhancing data protection and recovery capabilities. Defaults to `false`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `app_engine_integration_mode` (String)
- `concurrency_mode` (String)
- `earliest_version_time` (String)
- `etag` (String)
- `fullname` (String) The full name of the firestore.
- `id` (String) The ID of this resource.
- `uid` (String)
- `version_retention_period` (String)

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)

## Import

Import is supported using the following syntax:

```shell
# Example: Importing an existing GCP Firestore
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the  name of the Firestore
#

terraform import duplocloud_gcp_firestore.firestore-app *TENANT_ID*/*NAME*
```
