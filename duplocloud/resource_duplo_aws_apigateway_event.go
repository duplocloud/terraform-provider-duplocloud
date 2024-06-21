package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func AwsApiGatewayEventSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the API gateway event will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"api_gateway_id": {
			Description: "The ID of the REST API.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"method": {
			Description: "HTTP Method.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"GET",
				"POST",
				"PUT",
				"DELETE",
				"HEAD",
				"OPTIONS",
				"ANY",
			}, false),
		},
		"path": {
			Description: "The path segment of API resource.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"cors": {
			Description: "Enable handling of preflight requests.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"authorizer_id": {
			Description: "Authorizer id to be used when the authorization is `CUSTOM` or `COGNITO_USER_POOLS`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"authorization_type": {
			Description: "Type of authorization used for the method. (`NONE`, `CUSTOM`, `AWS_IAM`, `COGNITO_USER_POOLS`)",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"NONE",
				"CUSTOM",
				"AWS_IAM",
				"COGNITO_USER_POOLS",
			}, false),
		},
		"integration": {
			Description: "Specify API gateway integration.",
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Description: "Integration input's type. Valid values are `HTTP` (for HTTP backends), `MOCK` (not calling any real backend), `AWS` (for AWS services), `AWS_PROXY` (for Lambda proxy integration) and `HTTP_PROXY` (for HTTP proxy integration).",
						Type:        schema.TypeString,
						Required:    true,
						ValidateFunc: validation.StringInSlice([]string{
							"HTTP",
							"MOCK",
							"AWS",
							"AWS_PROXY",
							"HTTP_PROXY",
						}, false),
					},
					"uri": {
						Description: "Input's URI. Required if type is `AWS`, `AWS_PROXY`, `HTTP` or `HTTP_PROXY`. For AWS integrations, the URI should be of the form `arn:aws:apigateway:{region}:{subdomain.service|service}:{path|action}/{service_api}`. `region`, `subdomain` and `service` are used to determine the right endpoint.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"timeout": {
						Description: "Custom timeout between 50 and 300,000 milliseconds.",
						Type:        schema.TypeInt,
						Optional:    true,
					},
				},
			},
		},
	}
}

// Resource for managing an AWS SSM parameter
func resourceAwsApiGatewayEvent() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_apigateway_event` manages an AWS API Gateway events with integration in Duplo.",

		ReadContext:   resourceAwsApiGatewayEventRead,
		CreateContext: resourceAwsApiGatewayEventCreate,
		UpdateContext: resourceAwsApiGatewayEventUpdate,
		DeleteContext: resourceAwsApiGatewayEventDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: AwsApiGatewayEventSchema(),
	}
}

// READ resource
func resourceAwsApiGatewayEventRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, apigatewayid, method, path, err := parseAwsApiGatewayEventIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsApiGatewayEventRead(%s, %s): start", tenantID, apigatewayid+"--"+method+"--"+path)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, clientErr := c.ApiGatewayEventGet(tenantID, apigatewayid, method, path)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("") // object missing
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s API Gateway Event '%s': %s", tenantID, apigatewayid+"--"+method+"--"+path, clientErr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("api_gateway_id", duplo.APIGatewayID)
	d.Set("method", duplo.Method)
	d.Set("path", duplo.Path)
	d.Set("cors", duplo.Cors)
	d.Set("authorizer_id", duplo.AuthorizerId)
	d.Set("authorization_type", duplo.AuthorizationType)
	if duplo.Integration != nil {
		d.Set("integration", []interface{}{
			map[string]interface{}{
				"type":    duplo.Integration.Type,
				"uri":     duplo.Integration.URI,
				"timeout": duplo.Integration.Timeout,
			},
		})
	}
	log.Printf("[TRACE] resourceAwsApiGatewayEventRead(%s, %s): end", tenantID, apigatewayid+"--"+method+"--"+path)
	return nil
}

// CREATE resource
func resourceAwsApiGatewayEventCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	apigatewayid := d.Get("api_gateway_id").(string)
	method := d.Get("method").(string)
	path := d.Get("path").(string)

	log.Printf("[TRACE] resourceAwsApiGatewayEventCreate(%s, %s): start", tenantID, apigatewayid+"--"+method+"--"+path)

	// Create the request object.
	rq := duplosdk.DuploApiGatewayEvent{
		APIGatewayID: apigatewayid,
		Method:       method,
		Path:         path,
		Cors:         d.Get("cors").(bool),
	}
	if v, ok := d.GetOk("authorizer_id"); ok && v != nil {
		rq.AuthorizerId = v.(string)
	}
	if v, ok := d.GetOk("authorization_type"); ok && v != nil {
		rq.AuthorizationType = v.(string)
	}
	if v, ok := d.GetOk("integration"); ok {
		if s := v.([]interface{}); len(s) > 0 {
			rq.Integration = expandAwsApiGatewayEventIntegration(s[0].(map[string]interface{}))
		}
	}
	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	_, err = c.ApiGatewayEventCreate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s API Gateway Event '%s': %s", tenantID, apigatewayid+"--"+method+"--"+path, err)
	}

	// Wait for Duplo to be able to return the table's details.
	id := fmt.Sprintf("%s/%s/%s/%s", tenantID, apigatewayid, method, path)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "API Gateway Event", id, func() (interface{}, duplosdk.ClientError) {
		return c.ApiGatewayEventGet(tenantID, apigatewayid, method, path)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsApiGatewayEventRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsApiGatewayEventCreate(%s, %s): end", tenantID, apigatewayid+"--"+method+"--"+path)
	return diags
}

// UPDATE resource
func resourceAwsApiGatewayEventUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	apigatewayid := d.Get("api_gateway_id").(string)
	method := d.Get("method").(string)
	path := d.Get("path").(string)
	log.Printf("[TRACE] resourceAwsApiGatewayEventUpdate(%s, %s): start", tenantID, apigatewayid+"--"+method+"--"+path)

	// Create the request object.
	rq := duplosdk.DuploApiGatewayEvent{
		APIGatewayID: apigatewayid,
		Method:       method,
		Path:         path,
		Cors:         d.Get("cors").(bool),
	}

	if v, ok := d.GetOk("authorizer_id"); ok && v != nil {
		rq.AuthorizerId = v.(string)
	}
	if v, ok := d.GetOk("authorization_type"); ok && v != nil {
		rq.AuthorizationType = v.(string)
	}
	if v, ok := d.GetOk("integration"); ok && ok {
		if s := v.([]interface{}); len(s) > 0 {
			rq.Integration = expandAwsApiGatewayEventIntegration(s[0].(map[string]interface{}))
		}
	}

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	_, err = c.ApiGatewayEventUpdate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("Error update tenant %s API Gateway Event '%s': %s", tenantID, apigatewayid+"--"+method+"--"+path, err)
	}

	diags := resourceAwsApiGatewayEventRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsApiGatewayEventUpdate(%s, %s): end", tenantID, apigatewayid+"--"+method+"--"+path)
	return diags
}

// DELETE resource
func resourceAwsApiGatewayEventDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()

	tenantID := d.Get("tenant_id").(string)
	apigatewayid := d.Get("api_gateway_id").(string)
	method := d.Get("method").(string)
	path := d.Get("path").(string)

	log.Printf("[TRACE] resourceAwsApiGatewayEventDelete(%s, %s): start", tenantID, apigatewayid+"--"+method+"--"+path)

	// Delete the function.
	c := m.(*duplosdk.Client)
	clientErr := c.ApiGatewayEventDelete(tenantID, apigatewayid, method, path)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s API Gateway Event '%s': %s", tenantID, apigatewayid+"--"+method+"--"+path, clientErr)
	}

	// Wait up to 60 seconds for Duplo to delete the cluster.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "API Gateway Event", id, func() (interface{}, duplosdk.ClientError) {
		return c.ApiGatewayEventGet(tenantID, apigatewayid, method, path)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsApiGatewayEventDelete(%s, %s): end", tenantID, apigatewayid+"--"+method+"--"+path)
	return nil
}

func parseAwsApiGatewayEventIdParts(id string) (tenantID, apigatewayid, method, path string, err error) {
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) == 4 {
		tenantID, apigatewayid, method, path = idParts[0], idParts[1], idParts[2], idParts[3]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func expandAwsApiGatewayEventIntegration(m map[string]interface{}) *duplosdk.DuploApiGatewayEventIntegration {
	integration := &duplosdk.DuploApiGatewayEventIntegration{
		Type:    m["type"].(string),
		URI:     m["uri"].(string),
		Timeout: m["timeout"].(int),
	}

	return integration
}
