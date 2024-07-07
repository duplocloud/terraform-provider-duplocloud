---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "duplocloud_aws_apigateway_event Resource - terraform-provider-duplocloud"
subcategory: ""
description: |-
  duplocloud_aws_apigateway_event manages an AWS API Gateway events with integration in Duplo.
---

# duplocloud_aws_apigateway_event (Resource)

`duplocloud_aws_apigateway_event` manages an AWS API Gateway events with integration in Duplo.

## Example Usage

```terraform
resource "duplocloud_tenant" "myapp" {
  account_name = "myapp"
  plan_id      = "default"
}

resource "duplocloud_aws_apigateway_event" "apigateway_event" {
  tenant_id          = duplocloud_tenant.myapp.tenant_id
  api_gateway_id     = "t84tb3skz0"
  method             = "POST"
  path               = "/v2/docs"
  cors               = true
  authorization_type = "COGNITO_USER_POOLS"
  authorizer_id      = "gto03x"

  integration {
    type    = "AWS_PROXY"
    uri     = "arn:aws:apigateway:us-west-2:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:1234567890:function:duploservices-dev-valuation-test/invocations"
    timeout = 29000
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `api_gateway_id` (String) The ID of the REST API.
- `integration` (Block List, Min: 1, Max: 1) Specify API gateway integration. (see [below for nested schema](#nestedblock--integration))
- `method` (String) HTTP Method.
- `path` (String) The path segment of API resource.
- `tenant_id` (String) The GUID of the tenant that the API gateway event will be created in.

### Optional

- `api_key_required` (Boolean) Specify if the method requires an API key.
- `authorization_type` (String) Type of authorization used for the method. (`NONE`, `CUSTOM`, `AWS_IAM`, `COGNITO_USER_POOLS`)
- `authorizer_id` (String) Authorizer id to be used when the authorization is `CUSTOM` or `COGNITO_USER_POOLS`.
- `cors` (Boolean) Enable handling of preflight requests.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--integration"></a>
### Nested Schema for `integration`

Required:

- `type` (String) Integration input's type. Valid values are `HTTP` (for HTTP backends), `MOCK` (not calling any real backend), `AWS` (for AWS services), `AWS_PROXY` (for Lambda proxy integration) and `HTTP_PROXY` (for HTTP proxy integration).
- `uri` (String) Input's URI. Required if type is `AWS`, `AWS_PROXY`, `HTTP` or `HTTP_PROXY`. For AWS integrations, the URI should be of the form `arn:aws:apigateway:{region}:{subdomain.service|service}:{path|action}/{service_api}`. `region`, `subdomain` and `service` are used to determine the right endpoint.

Optional:

- `timeout` (Number) Custom timeout between 50 and 300,000 milliseconds.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)

## Import

Import is supported using the following syntax:

```shell
# Example: Importing an existing AWS API Gateway Event
#  - *TENANT_ID* is the tenant GUID
#  - *API_GATEWAY_ID* The API Gateway ID.
#  - *METHOD* The HTTP Method.
#  - *PATH* The API endpoint path.

terraform import duplocloud_aws_ssm_parameter.ssm_param *TENANT_ID*/*API_GATEWAY_ID*/*METHOD*/*PATH*
```