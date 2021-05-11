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
)

func awsLambdaFunctionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description: "The GUID of the tenant that the lambda function will be created in.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"name": {
			Description: "The short name of the lambda function cluster.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the lambda function.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"arn": {
			Description: "The ARN of the lambda function.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"timeout": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  3,
		},
		"s3_bucket": {
			Type:     schema.TypeString,
			Required: true,
		},
		"s3_key": {
			Type:     schema.TypeString,
			Required: true,
		},
		"environment": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"variables": {
						Type:     schema.TypeMap,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
		"last_modified": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"source_code_hash": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"source_code_size": {
			Type:     schema.TypeInt,
			Computed: true,
		},
	}
}

// Resource for managing an AWS lambda function
func resourceAwsLambdaFunction() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_lambda_function` manages an AWS lambda function in Duplo.",

		ReadContext:   resourceAwsLambdaFunctionRead,
		CreateContext: resourceAwsLambdaFunctionCreate,
		UpdateContext: resourceAwsLambdaFunctionUpdate,
		DeleteContext: resourceAwsLambdaFunctionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: kafkaClusterSchema(),
	}
}

/// READ resource
func resourceAwsLambdaFunctionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseAwsLambdaFunctionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsLambdaFunctionRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.LambdaFunctionGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s lambda function '%s': %s", tenantID, name, err)
	}

	// Set simple fields first.
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.FunctionName)
	d.Set("arn", duplo.FunctionArn)
	d.Set("description", duplo.Description)
	d.Set("last_modified", duplo.LastModified)
	d.Set("source_code_hash", duplo.CodeSha256)
	d.Set("source_code_size", duplo.CodeSize)

	log.Printf("[TRACE] resourceAwsLambdaFunctionRead(%s, %s): end", tenantID, name)
	return nil
}

/// CREATE resource
func resourceAwsLambdaFunctionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsLambdaFunctionCreate(%s, %s): start", tenantID, name)

	// Create the request object.
	rq := duplosdk.DuploLambdaFunction{
		FunctionName: name,
	}

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	err := c.LambdaFunctionCreate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s lambda function '%s': %s", tenantID, name, err)
	}

	// Wait for Duplo to be able to return the cluster's details.
	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "lambda function", id, func() (interface{}, error) {
		return c.LambdaFunctionGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsLambdaFunctionRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsLambdaFunctionCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsLambdaFunctionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseAwsLambdaFunctionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsLambdaFunctionUpdate(%s, %s): start", tenantID, name)

	//c := m.(*duplosdk.Client)

	log.Printf("[TRACE] resourceAwsLambdaFunctionUpdate(%s, %s): end", tenantID, name)
	return nil
}

/// DELETE resource
func resourceAwsLambdaFunctionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseAwsLambdaFunctionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsLambdaFunctionDelete(%s, %s): start", tenantID, name)

	// See if the object still exists in Duplo.
	c := m.(*duplosdk.Client)
	duplo, err := c.LambdaFunctionGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to get lambda function '%s': %s", id, err)
	}
	if duplo != nil {

		// Delete the function.
		err := c.LambdaFunctionDelete(tenantID, duplo.FunctionName)
		if err != nil {
			return diag.Errorf("Error deleting lambda function '%s': %s", id, err)
		}

		// Wait up to 60 seconds for Duplo to delete the cluster.
		diag := waitForResourceToBeMissingAfterDelete(ctx, d, "AWS lambda function", id, func() (interface{}, error) {
			return c.LambdaFunctionGet(tenantID, name)
		})
		if diag != nil {
			return diag
		}
	}

	// Wait 10 more seconds to deal with consistency issues.
	time.Sleep(10 * time.Second)

	log.Printf("[TRACE] resourceAwsLambdaFunctionDelete(%s, %s): end", tenantID, name)
	return nil
}

func parseAwsLambdaFunctionIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("Invalid resource ID: %s", id)
	}
	return
}

func needsAwsLambdaFunctionCodeUpdate(d schema.ResourceData) bool {
	return d.HasChange("filename") ||
		d.HasChange("source_code_hash") ||
		d.HasChange("source_code_size") ||
		d.HasChange("s3_bucket") ||
		d.HasChange("s3_key")
}
