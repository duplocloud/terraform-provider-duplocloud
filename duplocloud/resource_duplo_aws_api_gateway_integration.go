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

func duploAwsApiGatewayIntegrationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the aws api gateway integration will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the api gateway.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the api gateway.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"lambda_function_name": {
			Description: "Name of the lambda function to be integrated with API gateway.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"metadata": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceAwsApiGatewayIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_api_gateway_integration` manages an aws api gateway integration in Duplo.",

		ReadContext:   resourceAwsApiGatewayIntegrationRead,
		CreateContext: resourceAwsApiGatewayIntegrationCreate,
		UpdateContext: resourceAwsApiGatewayIntegrationUpdate,
		DeleteContext: resourceAwsApiGatewayIntegrationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsApiGatewayIntegrationSchema(),
	}
}

func resourceAwsApiGatewayIntegrationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, fullname, err := parseAwsApiGatewayIntegrationIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsApiGatewayIntegrationRead(%s, %s): start", tenantID, fullname)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.TenantGetAPIGateway(tenantID, fullname)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s aws api gateway integration %s : %s", tenantID, fullname, clientErr)
	}

	flattenAwsApiGatewayIntegration(d, duplo)

	log.Printf("[TRACE] resourceAwsApiGatewayIntegrationRead(%s, %s): end", tenantID, fullname)
	return nil
}

func resourceAwsApiGatewayIntegrationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsApiGatewayIntegrationCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAwsApiGatewayIntegration(d)
	err = c.TenantCreateAPIGateway(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s aws api gateway integration '%s': %s", tenantID, name, err)
	}

	fullname, err := c.GetDuploServicesName(tenantID, name)
	if err != nil {
		return diag.Errorf("Error creating tenant %s aws api gateway integration '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, fullname)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws api gateway integration", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetAPIGateway(tenantID, fullname)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsApiGatewayIntegrationRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsApiGatewayIntegrationCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsApiGatewayIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAwsApiGatewayIntegrationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, fullname, err := parseAwsApiGatewayIntegrationIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsApiGatewayIntegrationDelete(%s, %s): start", tenantID, fullname)

	c := m.(*duplosdk.Client)
	clientErr := c.TenantDeleteAPIGateway(tenantID, fullname)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s aws api gateway integration '%s': %s", tenantID, fullname, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "aws api gateway integration", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetAPIGateway(tenantID, fullname)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsApiGatewayIntegrationDelete(%s, %s): end", tenantID, fullname)
	return nil
}

func expandAwsApiGatewayIntegration(d *schema.ResourceData) duplosdk.DuploApiGatewayRequest {
	return duplosdk.DuploApiGatewayRequest{
		Name:           d.Get("name").(string),
		LambdaFunction: d.Get("lambda_function_name").(string),
	}
}

func parseAwsApiGatewayIntegrationIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAwsApiGatewayIntegration(d *schema.ResourceData, duplo *duplosdk.DuploApiGatewayResource) {
	d.Set("metadata", duplo.MetaData)
	d.Set("fullname", duplo.Name)
}
