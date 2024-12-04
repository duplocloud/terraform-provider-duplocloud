---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_k8_helm_release Resource - terraform-provider-duplocloud"
subcategory: ""
description: |-
  duplocloud_helm_release manages helm release at duplocloud
---

# duplocloud_k8_helm_release (Resource)

`duplocloud_helm_release` manages helm release at duplocloud

## Example Usage

```terraform
resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}


resource "duplocloud_k8_helm_release" "release" {
  tenant_id    = duplocloud_tenant.myapp.tenant_id
  name         = "helm-release-name"
  interval     = "05m00s"
  release_name = "helm-release-1"
  chart {
    name               = "chart-name"
    version            = "v1"
    reconcile_strategy = "ChartVersion"
    source_type        = "HelmRepository"
    source_name        = duplocloud_k8_helm_repository.repo.name
  }
  values = jsonencode({
    "replicaCount" : 2,
    "serviceAccount" : {
      "create" : false
    }
    }
  )
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the helm chart
- `release_name` (String) Provide release name to identify specific deployment of helm chart.
- `tenant_id` (String) The GUID of the tenant that the storage bucket will be created in.

### Optional

- `chart` (Block List) Helm chart (see [below for nested schema](#nestedblock--chart))
- `interval` (String) Interval related to helm release Defaults to `5m0s`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `values` (String) Customise an helm chart.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--chart"></a>
### Nested Schema for `chart`

Required:

- `name` (String) Provide unique name for the helm chart.
- `source_name` (String) The name of the source, referred from helm repository resource.
- `version` (String) The helm chart version

Optional:

- `interval` (String) The interval associated to helm chart Defaults to `5m0s`.
- `reconcile_strategy` (String) The reconcile strategy should be chosen from ChartVersion or Revision. No new chart artifact is produced on updates to the source unless the version is changed in HelmRepository. Use `Revision` to produce new chart artifact on change in source revision. Defaults to `ChartVersion`.
- `source_type` (String) The helm chart source, currently only HelmRepository as source is supported Defaults to `HelmRepository`.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)

## Import

Import is supported using the following syntax:

```shell
# Example: Importing an existing helm release
#  - *TENANT_ID* is the tenant GUID
#  - *NAME* is the helm release name
#
terraform import duplocloud_k8_helm_repository.release *TENANT_ID*/helm-release/*NAME*
```