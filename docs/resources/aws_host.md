---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_aws_host Resource - terraform-provider-duplocloud"
subcategory: ""
description: |-
  duplocloud_aws_host manages a native AWS host in Duplo.
---

# duplocloud_aws_host (Resource)

`duplocloud_aws_host` manages a native AWS host in Duplo.

## Example Usage

```terraform
resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

# Simple Example 1:  Deploy a host to be used with Duplo's native container agent
resource "duplocloud_aws_host" "native" {
  tenant_id     = duplocloud_tenant.myapp.tenant_id
  friendly_name = "host1"

  image_id       = "ami-abcd1234" # <== put the AWS duplo docker AMI ID here
  capacity       = "t3a.medium"
  agent_platform = 0 # Duplo native container agent
  zone           = 0 # Zone A
  user_account   = duplocloud_tenant.myapp.account_name

  metadata {
    key   = "OsDiskSize" # <== This is the size of the OS disk in GB
    value = "100"
  }
}

# Simple Example 2:  Deploy a host to be used with EKS
resource "duplocloud_aws_host" "eks2" {
  tenant_id     = "81f6043b-1480-4c92-a0d8-4d3d3a6ae13a"
  friendly_name = "tf-v2only"

  image_id       = "ami-0b896ca73d6b87976" # <== put the AWS EKS 1.18 AMI ID here
  capacity       = "t3.small"
  agent_platform = 7 # Duplo EKS agent
  zone           = 0 # Zone A
  user_account   = "jt-1303"
  keypair_type   = "1"
}

# Simple Example 3:  Create a host with instance metadata service
resource "duplocloud_aws_host" "host" {
  tenant_id     = "81f6043b-1480-4c92-a0d8-4d3d3a6ae13a"
  friendly_name = "tf-v2only"

  image_id       = "ami-0b896ca73d6b87976" # <== put the AWS EKS 1.18 AMI ID here
  capacity       = "t3.small"
  agent_platform = 7 # Duplo EKS agent
  zone           = 0 # Zone A
  user_account   = "jt-1303"
  keypair_type   = "1"

  metadata {
    key   = "OsDiskSize" # <== This is the size of the OS disk in GB
    value = "100"
  }

  # Create a host with instance metadata v2 only
  metadata {
    key   = "MetadataServiceOption"
    value = "enabled_v2_only"
  }

  # Create a host with instance metadata v1 and v2
  /* metadata {
    key   = "MetadataServiceOption"
    value = "enabled"
  } */

  # Create a host with instance metadata disabled
  /* metadata {
     key   = "MetadataServiceOption"
     value = "disabled"
   } */
}

# Simple Example 4:  Enabling Hibernation for ec2 host

resource "duplocloud_aws_host" "myhost" {
  capacity      = "t3a.small"
  cloud         = 0
  friendly_name = "host-5"
  image_id      = "ami-0adf7efba3226042e"
  tenant_id     = "abd5fb54-306a-46e7-a0e3-39a4de441bfd"

  metadata {
    key   = "EnableHibernation"
    value = "True"
  }

}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `capacity` (String) The AWS EC2 instance type.
- `friendly_name` (String) The short name of the host.
- `image_id` (String) The AMI ID to use.
- `tenant_id` (String) The GUID of the tenant that the host will be created in.

### Optional

- `agent_platform` (Number) The numeric ID of the container agent pool that this host is added to. Defaults to `0`.
- `allocated_public_ip` (Boolean) Whether or not to allocate a public IP. Defaults to `false`.
- `base64_user_data` (String) Base64 encoded EC2 user data to associated with the host.
- `cloud` (Number) The numeric ID of the cloud provider to launch the host in. Defaults to `0`.
- `encrypt_disk` (Boolean) Defaults to `false`.
- `is_ebs_optimized` (Boolean) Defaults to `false`.
- `is_minion` (Boolean) Defaults to `true`.
- `keypair_type` (Number) The numeric ID of the keypair type being used.Should be one of:

   - `0` : Default
   - `1` : ED25519
   - `2` : RSA (deprecated - some operating systems no longer support it)
- `metadata` (Block List) Configuration metadata used when creating the host. (see [below for nested schema](#nestedblock--metadata))
- `minion_tags` (Block List) A map of tags to assign to the resource. Example - `AllocationTags` can be passed as tag key with any value. (see [below for nested schema](#nestedblock--minion_tags))
- `network_interface` (Block List) An optional list of custom network interface configurations to use when creating the host. (see [below for nested schema](#nestedblock--network_interface))
- `prepend_user_data` (Boolean) Bootstrap an EKS host with Duplo's user data, prepending it to custom user data if also provided. Defaults to `true`.
- `tags` (Block List) (see [below for nested schema](#nestedblock--tags))
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `user_account` (String) The name of the tenant that the host will be created in.
- `volume` (Block List) (see [below for nested schema](#nestedblock--volume))
- `wait_until_connected` (Boolean) Whether or not to wait until Duplo can connect to the host, after creation. Defaults to `true`.
- `zone` (Number) The availability zone to launch the host in, expressed as a number and starting at 0. Defaults to `0`.

### Read-Only

- `id` (String) The ID of this resource.
- `identity_role` (String) The name of the IAM role associated with this host.
- `initial_base64_user_data` (String)
- `instance_id` (String) The AWS EC2 instance ID of the host.
- `private_ip_address` (String) The primary private IP address assigned to the host.
- `public_ip_address` (String) The primary public IP address assigned to the host.
- `status` (String) The current status of the host.

<a id="nestedblock--metadata"></a>
### Nested Schema for `metadata`

Required:

- `key` (String)
- `value` (String)


<a id="nestedblock--minion_tags"></a>
### Nested Schema for `minion_tags`

Required:

- `key` (String)
- `value` (String)


<a id="nestedblock--network_interface"></a>
### Nested Schema for `network_interface`

Optional:

- `associate_public_ip` (Boolean) Whether or not to associate a public IP with the newly created ENI.  Cannot be specified if `network_interface_id` is specified.
- `device_index` (Number) The device index to pass to AWS for attaching the ENI.  Starts at zero.
- `groups` (List of String)
- `metadata` (Block List) (see [below for nested schema](#nestedblock--network_interface--metadata))
- `network_interface_id` (String) The ID of an ENI to attach to this host.  Cannot be specified if `subnet_id` or `associate_public_ip` is specified.
- `subnet_id` (String) The ID of a subnet in which to create a new ENI.  Cannot be specified if `network_interface_id` is specified.

<a id="nestedblock--network_interface--metadata"></a>
### Nested Schema for `network_interface.metadata`

Required:

- `key` (String)
- `value` (String)



<a id="nestedblock--tags"></a>
### Nested Schema for `tags`

Required:

- `key` (String)
- `value` (String)


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)


<a id="nestedblock--volume"></a>
### Nested Schema for `volume`

Optional:

- `iops` (Number)
- `name` (String)
- `size` (Number)
- `volume_id` (String)
- `volume_type` (String)

## Import

Import is supported using the following syntax:

```shell
# Example: Importing an existing AWS host
#  - *TENANT_ID* is the tenant GUID
#  - *INSTANCE_ID* is the AWS EC2 instance ID
#
terraform import duplocloud_aws_host.myhost v2/subscriptions/*TENANT_ID*/NativeHostV2/*INSTANCE_ID*
```
