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