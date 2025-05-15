package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsCloudfrontFunction() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudfrontFunctionCreate,
		ReadContext:   resourceCloudfrontFunctionRead,
		UpdateContext: resourceCloudfrontFunctionUpdate,
		DeleteContext: resourceCloudfrontFunctionDelete,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The GUID of the tenant in DuploCloud.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the CloudFront function.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"code": {
				Description: "The JavaScript code for the CloudFront function.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"comment": {
				Description: "A comment to describe the CloudFront function.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"runtime": {
				Description: "The runtime environment for the CloudFront function.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"function_arn": {
				Description: "The ARN of the CloudFront function.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"publish": {
				Description: "Whether to publish creation/change as Live CloudFront Function Version",
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCloudfrontFunctionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Construct the request
	req := expandCloudFrontFunction(d)
	// Call the Duplo API to create the CloudFront function
	log.Printf("[TRACE] Creating CloudFront function in tenant %s: %v", tenantID, request)
	resp, err := c.CreateCloudFrontFunction(tenantID, req)
	if err != nil {
		return diag.Errorf("failed to create CloudFront function: %s", err)
	}
	if d.Get("publish").(bool) {
		log.Printf("[TRACE] Publishing CloudFront function %s in tenant %s", resp.Name, tenantID)
		err = c.PublishCloudFrontFunction(tenantID, resp.Name)
		if err != nil {
			return diag.Errorf("failed to publish CloudFront function: %s", err)
		}
	}
	// Set the resource ID
	d.SetId(fmt.Sprintf("%s/%s", tenantID, d.Get("name").(string)))

	return resourceCloudfrontFunctionRead(ctx, d, m)
}

func resourceCloudfrontFunctionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*duplosdk.Client)

	tenantID, name := parseCloudfrontFunctionId(d.Id())

	// Call the Duplo API to get the CloudFront function
	log.Printf("[TRACE] Reading CloudFront function %s in tenant %s", name, tenantID)
	function := map[string]interface{}{}
	rp, err := c.GetCloudFrontFunction(tenantID, name)
	if err != nil {
		if err.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("failed to read CloudFront function: %s", err)
	}

	// Update the state
	d.Set("tenant_id", tenantID)
	d.Set("name", function["Name"])
	d.Set("code", function["Code"])
	d.Set("comment", function["Comment"])
	d.Set("runtime", function["Runtime"])
	d.Set("function_arn", function["FunctionArn"])

	return nil
}

func resourceCloudfrontFunctionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*duplosdk.Client)
	tenantID, name := parseCloudfrontFunctionId(d.Id())

	// Construct the request
	request := map[string]interface{}{
		"Name":    name,
		"Code":    d.Get("code").(string),
		"Comment": d.Get("comment").(string),
		"Runtime": d.Get("runtime").(string),
	}

	// Call the Duplo API to update the CloudFront function
	log.Printf("[TRACE] Updating CloudFront function %s in tenant %s: %v", name, tenantID, request)
	err := c.PutAPI(fmt.Sprintf("v3/subscriptions/%s/aws/cloudfrontFunction/%s", tenantID, name), request)
	if err != nil {
		return fmt.Errorf("failed to update CloudFront function: %s", err)
	}

	return resourceCloudfrontFunctionRead(ctx, d, m)
}

func resourceCloudfrontFunctionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*duplosdk.Client)
	tenantID, name := parseDuploAwsCloudfrontFunctionId(d.Id())

	// Call the Duplo API to delete the CloudFront function
	log.Printf("[TRACE] Deleting CloudFront function %s in tenant %s", name, tenantID)
	err := c.DeleteAPI(fmt.Sprintf("v3/subscriptions/%s/aws/cloudfrontFunction/%s", tenantID, name), nil)
	if err != nil {
		return fmt.Errorf("failed to delete CloudFront function: %s", err)
	}

	// Remove the resource from state
	d.SetId("")

	return nil
}

func parseCloudfrontFunctionId(id string) (tenantID, name string) {
	parts := strings.SplitN(id, "/", 2)
	return parts[0], parts[1]
}

func expandCloudFrontFunction(d *schema.ResourceData) *duplosdk.DuploCloudFrontFunction {
	return &DuploCloudFrontFunction{
		Name:    d.Get("name").(string),
		Runtime: d.Get("runtime").(string),
		Code:    d.Get("code").(string),
		Comment: d.Get("comment").(string),
	}
}

func FlattenCloudFrontFunction(d *schema.ResourceData, function *duplosdk.DuploCloudFrontFunction) {
	if function == nil {
		return
	}

	d.Set("name", function.Name)
	d.Set("runtime", function.Runtime)
	d.Set("code", function.Code)
	d.Set("comment", function.Comment)
	d.Set("status", function.Status)
	if function.Metadata != nil {
		d.Set("function_arn", function.Metadata.ARN)
	}
}
