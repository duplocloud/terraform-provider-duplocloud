---

# Resource: duplocloud_tenant



`duplocloud_tenant` manages a tenant in Duplo.<p>A **DuploCloud tenant** is an isolated environment within the DuploCloud platform where you can manage and provision cloud resources. It essentially represents a distinct organizational unit or environment for deploying and managing infrastructure and applications.</p>


## Example Usage

### Create a DuploCloud tenant named 'prod'.

```terraform
# Before creating a tenant, you must first set up the infrastructure. Below is the resource for creating the infrastructure.
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0             # AWS Cloud
  region            = "us-west-2"
  enable_k8_cluster = false
  address_prefix    = "10.11.0.0/16"
}

# Use the infrastructure name as the 'plan_id' from the 'duplocloud_infrastructure' resource while creating tenant.
resource "duplocloud_tenant" "tenant" {
 account_name = "prod"
 plan_id      = duplocloud_infrastructure.infra.infra_name
}
```

### Create a DuploCloud tenant named 'prod' inside the following prod infra.

```terraform
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0             # AWS Cloud
  region            = "us-west-2"
  enable_k8_cluster = false
  address_prefix    = "10.11.0.0/16"
}
```

```terraform
# Use the infrastructure name as the 'plan_id' from the 'duplocloud_infrastructure' resource.
resource "duplocloud_tenant" "tenant" {
 account_name = "prod"
 plan_id      = duplocloud_infrastructure.infra.infra_name
}
```

### Create a DuploCloud tenant named 'dev' within the 'nonprod' infrastructure.

```terraform
# Ensure the 'nonprod' infrastructure is already created before setting up the tenant.
data "duplocloud_infrastructure" "infra" {
 infra_name = "nonprod"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "dev"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
}
```

### Create a DuploCloud tenant named 'dev' with infra name variable and tenant id as output.

```terraform
variable "infra_name" {
  type    = string
  default = "nonprod"
}

# Ensure the 'nonprod' infrastructure is already created before setting up the tenant.
data "duplocloud_infrastructure" "infra" {
 infra_name = "nonprod"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "dev"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
}

output "tenant_id" {
  description = "A GUID identifying the tenant."
  value       = duplocloud_tenant.tenant.tenant_id
} 
```

### Create a duplocloud tenant named dev with AWS Cognito power user access in the nonprod infrastructure.

```terraform

# A prerequisite for creating a tenant is having an existing infrastructure. Here’s how you can reference an existing infrastructure.
data "duplocloud_infrastructure" "infra" {
 infra_name = "nonprod"
}

# Here’s how to create a tenant by providing the infrastructure name for the plan_id field.
resource "duplocloud_tenant" "tenant" {
 account_name = "dev"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
}

# Attaches a managed IAM policy to an IAM role.
resource "aws_iam_role_policy_attachment" "AmazonCognitoPowerUser" {
 role       = "duploservices-${duplocloud_tenant.tenant.account_name}"
 policy_arn = "arn:aws:iam::aws:policy/AmazonCognitoPowerUser"
}

```

### Create a DuploCloud tenant named 'qa' with full access to invoke AWS API Gateway in the nonprod infrastructure.

```terraform
# A prerequisite for creating a tenant is having an existing infrastructure. Here’s how you can reference an existing infrastructure.
data "duplocloud_infrastructure" "infra" {
 infra_name = "nonprod"
}

# Here’s how to create a tenant by providing the infrastructure name for the plan_id field.
resource "duplocloud_tenant" "tenant" {
 account_name = "qa"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
}

# Attaches a managed IAM policy to an IAM role.
resource "aws_iam_role_policy_attachment" "AmazonAPIGatewayInvokeFullAccess" {
 role       = "duploservices-${duplocloud_tenant.tenant.account_name}"
 policy_arn = "arn:aws:iam::aws:policy/AmazonAPIGatewayInvokeFullAccess"
}

```

### Create duplocloud tenant named dev with security group rule to allow access from 10.220.0.0/16 on port 5432 in nonprod infra’

```terraform
data "duplocloud_infrastructure" "infra" {
 infra_name = "nonprod"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "dev"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
}

# Allow communication on port 5432 for the PostgreSQL database from the 10.220.0.0/16 subnet
resource "duplocloud_tenant_network_security_rule" "allow_from_vpn" {
 tenant_id      = duplocloud_tenant.tenant.tenant_id
 source_address = "10.220.0.0/16"
 protocol       = "tcp"
 from_port      = 5432
 to_port        = 5432
 description    = "Allow communication from 10.220.0.0/16 on port 5432."
}
```

### Setup duplocloud tenant named dev with security group rule to allow access from 10.220.0.0/16 on port 22 in nonprod infra’

```terraform
data "duplocloud_infrastructure" "infra" {
 infra_name = "nonprod"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "dev"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
}

# Allow communication on port 22 from the 10.220.0.0/16 subnet.
resource "duplocloud_tenant_network_security_rule" "allow_from_vpn" {
 tenant_id      = duplocloud_tenant.tenant.tenant_id
 source_address = "10.220.0.0/16"
 protocol       = "tcp"
 from_port      = 22
 to_port        = 22
 description    = "Allow communication from 10.220.0.0/16 on port 22."
}
```

### Provision a tenant named 'myapp' within the infrastructure 'myinfra' and disable deletion protection.

```terraform
data "duplocloud_infrastructure" "infra" {
 infra_name = "myinfra"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "myapp"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
 allow_deletion = true
}

# Reference the tenant_id field from the duplocloud_tenant resource.
resource "duplocloud_tenant_config" "tenant_config" {
  tenant_id = duplocloud_tenant.tenant.tenant_id

  setting {
    key   = "delete_protection"
    value = "false"
  }
}

```

### Provision a tenant named 'myapp' within the infrastructure 'myinfra', and ensure that the S3 bucket has public access blocked and SSL enforcement enabled in the S3 bucket policy.

```terraform
data "duplocloud_infrastructure" "infra" {
 infra_name = "myinfra"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "myapp"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
 allow_deletion = true
}

# Reference the tenant_id field from the duplocloud_tenant resource.
resource "duplocloud_tenant_config" "tenant_config" {
  tenant_id = duplocloud_tenant.tenant.tenant_id

  // This turns on a default public access block for S3 buckets.
  setting {
    key   = "block_public_access_to_s3"
    value = "true"
  }

  // This turns on a default SSL enforcement S3 bucket policy.
  setting {
    key   = "enforce_ssl_for_s3"
    value = "true"
  }
}

```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `account_name` (String) The name of the tenant. Tenant names are globally unique, and cannot be a prefix of any other tenant name.
- `plan_id` (String) The name of the plan under which the tenant will be created.

### Optional

- `allow_deletion` (Boolean) Whether or not to even try and delete the tenant. *NOTE: This only works if you have disabled deletion protection for the tenant.* Defaults to `false`.
- `existing_k8s_namespace` (String) Existing kubernetes namespace to use by the tenant. *NOTE: This is an advanced feature, please contact your DuploCloud administrator for help if you want to use this field.*
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `wait_until_created` (Boolean) Whether or not to wait until Duplo has created the tenant. Defaults to `true`.
- `wait_until_deleted` (Boolean) Whether or not to wait until Duplo has destroyed the tenant. Defaults to `false`.

### Read-Only

- `id` (String) The ID of this resource.
- `infra_owner` (String)
- `policy` (List of Object) (see [below for nested schema](#nestedatt--policy))
- `tags` (List of Object) (see [below for nested schema](#nestedatt--tags))
- `tenant_id` (String) A GUID identifying the tenant. This is automatically generated by Duplo.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)


<a id="nestedatt--policy"></a>
### Nested Schema for `policy`

Read-Only:

- `allow_volume_mapping` (Boolean)
- `block_external_ep` (Boolean)


<a id="nestedatt--tags"></a>
### Nested Schema for `tags`

Read-Only:

- `key` (String)
- `value` (String)

## Import

Import is supported using the following syntax:

```
terraform import duplocloud_tenant.myapp v2/admin/TenantV2/*TENANT_ID*
```