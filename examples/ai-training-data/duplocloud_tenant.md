### Create a DuploCloud tenant named 'prod'.

```
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0             # AWS Cloud
  region            = "us-west-2"
  enable_k8_cluster = false
  address_prefix    = "10.11.0.0/16"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "prod"
 plan_id      = duplocloud_infrastructure.infra.infra_name
}
```

### Create a DuploCloud tenant named 'prod' inside the following prod infra.
```
resource "duplocloud_infrastructure" "infra" {
  infra_name        = "prod"
  cloud             = 0             # AWS Cloud
  region            = "us-west-2"
  enable_k8_cluster = false
  address_prefix    = "10.11.0.0/16"
}
```
```
resource "duplocloud_tenant" "tenant" {
 account_name = "prod"
 plan_id      = duplocloud_infrastructure.infra.infra_name
}
```

### Create a DuploCloud tenant named 'dev' within the nonprod infrastructure.

```
data "duplocloud_infrastructure" "infra" {
 infra_name = "nonprod"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "dev"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
}
```

### Create a DuploCloud tenant named 'dev' with infra name variable and tenant id as output.

```
variable "infra_name" {
  type    = string
  default = "nonprod"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "dev"
 plan_id      = var.infra_name
}

output "tenant_id" {
  description = "A GUID identifying the tenant."
  value       = duplocloud_tenant.tenant.tenant_id
} 
```

### Create a duplocloud tenant named dev with AWS Cognito power user access in the nonprod infrastructure.

```
data "duplocloud_infrastructure" "infra" {
 infra_name = "nonprod"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "dev"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
}


resource "aws_iam_role_policy_attachment" "AmazonCognitoPowerUser" {
 role       = "duploservices-${duplocloud_tenant.tenant.account_name}"
 policy_arn = "arn:aws:iam::aws:policy/AmazonCognitoPowerUser"
}

```

### Create a DuploCloud tenant named 'qa' with full access to invoke AWS API Gateway in the nonprod infrastructure.

```
data "duplocloud_infrastructure" "infra" {
 infra_name = "nonprod"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "qa"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
}

resource "aws_iam_role_policy_attachment" "AmazonAPIGatewayInvokeFullAccess" {
 role       = "duploservices-${duplocloud_tenant.tenant.account_name}"
 policy_arn = "arn:aws:iam::aws:policy/AmazonAPIGatewayInvokeFullAccess"
}

```

### Create duplocloud tenant named dev with security group rule to allow access from 10.220.0.0/16 on port 5432 in nonprod infra’

```
data "duplocloud_infrastructure" "infra" {
 infra_name = "nonprod"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "dev"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
}

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

```
data "duplocloud_infrastructure" "infra" {
 infra_name = "nonprod"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "dev"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
}

resource "duplocloud_tenant_network_security_rule" "allow_from_vpn" {
 tenant_id      = duplocloud_tenant.tenant.tenant_id
 source_address = "10.220.0.0/16"
 protocol       = "tcp"
 from_port      = 22
 to_port        = 22
 description    = "Allow communication from 10.220.0.0/16 on port 22."
}
```

### Create a tenant test with under test infrastructure

```
data "duplocloud_infrastructure" "infra" {
 infra_name = "test"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "test"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
}
```

### Provision a tenant named 'myapp' with the infrastructure 'myinfra' and deletion protection disabled.

```
data "duplocloud_infrastructure" "infra" {
 infra_name = "myinfra"
}

resource "duplocloud_tenant" "tenant" {
 account_name = "myapp"
 plan_id      = data.duplocloud_infrastructure.infra.infra_name
 allow_deletion = true
}

resource "duplocloud_tenant_config" "tenant_config" {
  tenant_id = duplocloud_tenant.tenant.tenant_id

  setting {
    key   = "delete_protection"
    value = "true"
  }
}

```