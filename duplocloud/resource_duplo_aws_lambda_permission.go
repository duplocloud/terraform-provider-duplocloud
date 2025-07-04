package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsLambdaPermissionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the lambda permission will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"action": {
			Description: "The AWS Lambda action you want to allow in this statement. (e.g. `lambda:InvokeFunction`)",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"event_source_token": {
			Description: "The Event Source Token to validate.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"function_name": {
			Description: "Name of the Lambda function whose resource policy you are updating.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"principal": {
			Description: "The principal who is getting this permission.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"qualifier": {
			Description: "Query parameter to specify function version or alias name. The permission will then apply to the specific qualified ARN.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"source_account": {
			Description: "This parameter is used for S3 and SES. The AWS account ID (without a hyphen) of the source owner.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"source_arn": {
			Description: "When the principal is an AWS service, the ARN of the specific resource within that service to grant permission to.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"statement_id": {
			Description: "A unique statement identifier.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
	}
}

func resourceAwsLambdaPermission() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_lambda_permission` manages an AWS lambda permissions in Duplo.",

		ReadContext:   resourceAwsLambdaPermissionRead,
		CreateContext: resourceAwsLambdaPermissionCreate,
		//UpdateContext: resourceAwsLambdaPermissionUpdate,
		DeleteContext: resourceAwsLambdaPermissionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsLambdaPermissionSchema(),
	}
}

func resourceAwsLambdaPermissionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, functionName, sid, err := parseAwsLambdaPermissionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsLambdaPermissionRead(%s, %s, %s): start", tenantID, functionName, sid)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.LambdaPermissionGet(tenantID, functionName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("") // object missing
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s lambda permission '%s': %s", tenantID, functionName, clientErr)
	}
	if rp == nil {
		d.SetId("") // object missing
		return nil
	}
	for _, permission := range *rp {
		if permission.Sid == sid {
			d.Set("tenant_id", tenantID)
			d.Set("action", permission.Action)
			d.Set("function_name", functionName)
			if permission.Principal != nil {
				d.Set("principal", permission.Principal.Service)
			}
			d.Set("qualifier", d.Get("qualifier").(string))
			d.Set("source_account", d.Get("source_account").(string))
			if permission.Condition != nil {
				d.Set("source_arn", permission.Condition.Arn["AWS:SourceArn"])
			}
			d.Set("statement_id", permission.Sid)
		}
	}

	log.Printf("[TRACE] resourceAwsLambdaPermissionRead(%s, %s): end", tenantID, functionName)
	return nil
}

// CREATE resource
func resourceAwsLambdaPermissionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	tenantID := d.Get("tenant_id").(string)
	functionName := d.Get("function_name").(string)
	statementId := d.Get("statement_id").(string)
	log.Printf("[TRACE] resourceAwsLambdaPermissionCreate(%s, %s, %s): start", tenantID, functionName, statementId)

	rq := expandAwsLambdaPermission(d)
	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	_, err := c.LambdaPermissionCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s lambda permission '%s': %s", tenantID, functionName, err)
	}

	// Wait for Duplo to be able to return the cluster's details.
	id := fmt.Sprintf("%s/%s/%s", tenantID, functionName, statementId)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "lambda permission", id, func() (interface{}, duplosdk.ClientError) {
		return c.LambdaPermissionGet(tenantID, functionName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsLambdaPermissionRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsLambdaPermissionCreate(%s, %s): end", tenantID, functionName)
	return diags
}

// DELETE resource
func resourceAwsLambdaPermissionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, functionName, sid, err := parseAwsLambdaPermissionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsLambdaPermissionDelete(%s, %s, %s): start", tenantID, functionName, sid)

	c := m.(*duplosdk.Client)
	clientErr := c.LambdaPermissionDelete(tenantID, functionName, sid)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s lambda permission '%s': %s", tenantID, functionName, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "lambda permission", id, func() (interface{}, duplosdk.ClientError) {
		return c.LambdaPermissionGet(tenantID, functionName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsLambdaPermissionDelete(%s, %s, %s): end", tenantID, functionName, sid)
	return nil
}

func parseAwsLambdaPermissionIdParts(id string) (tenantID, name, statementId string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		tenantID, name, statementId = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func expandAwsLambdaPermission(d *schema.ResourceData) *duplosdk.DuploLambdaPermissionRequest {
	obj := duplosdk.DuploLambdaPermissionRequest{}
	if v, ok := d.GetOk("action"); ok && v != nil {
		obj.Action = v.(string)
	}
	if v, ok := d.GetOk("function_name"); ok && v != nil {
		obj.FunctionName = v.(string)
	}
	if v, ok := d.GetOk("principal"); ok && v != nil {
		obj.Principal = v.(string)
	}
	if v, ok := d.GetOk("event_source_token"); ok && v != nil {
		obj.EventSourceToken = v.(string)
	}
	if v, ok := d.GetOk("qualifier"); ok && v != nil {
		obj.Qualifier = v.(string)
	}
	if v, ok := d.GetOk("source_account"); ok && v != nil {
		obj.SourceAccount = v.(string)
	}
	if v, ok := d.GetOk("source_arn"); ok && v != nil {
		obj.SourceArn = v.(string)
	}
	if v, ok := d.GetOk("statement_id"); ok && v != nil {
		obj.StatementId = v.(string)
	}
	return &obj
}
