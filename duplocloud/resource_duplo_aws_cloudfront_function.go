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
		Description:   "duplocloud_aws_cloudfront_function allows you to create and manage AWS CloudFront Functions in DuploCloud.",
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
	log.Printf("[TRACE] Creating CloudFront function in tenant %s: %v", tenantID, req)
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
	d.SetId(fmt.Sprintf("%s/cloudfront-function/%s", tenantID, req.Name))

	return resourceCloudfrontFunctionRead(ctx, d, m)
}

func resourceCloudfrontFunctionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*duplosdk.Client)

	tenantID, name := parseCloudfrontFunctionId(d.Id())
	fullName, err := c.GetDuploServicesName(tenantID, name)
	if err != nil {
		return diag.Errorf("resourceCloudfrontFunctionRead: Unable to retrieve duplo service name (tenant: %s, function: %s: error: %s)", tenantID, name, err)
	}

	// Call the Duplo API to get the CloudFront function
	log.Printf("[TRACE] Reading CloudFront function %s in tenant %s", name, tenantID)
	rp, err := c.GetCloudFrontFunction(tenantID, fullName)
	if err != nil {
		if err.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("failed to read CloudFront function: %s", err)
	}
	d.Set("name", name)
	flattenCloudFrontFunction(d, rp)
	// Update the state
	return nil
}

func resourceCloudfrontFunctionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*duplosdk.Client)
	tenantID, name := parseCloudfrontFunctionId(d.Id())
	fullName, err := c.GetDuploServicesName(tenantID, name)
	if err != nil {
		return diag.Errorf("resourceCloudfrontFunctionUpdate: Unable to retrieve duplo service name (tenant: %s, function: %s: error: %s)", tenantID, name, err)
	}

	// Construct the request
	req := expandCloudFrontFunction(d)
	req.Name = fullName
	// Call the Duplo API to update the CloudFront function
	log.Printf("[TRACE] Updating CloudFront function %s in tenant %s: %v", name, tenantID, req)
	err = c.UpdateCloudFrontFunction(tenantID, fullName, req)
	if err != nil {
		return diag.Errorf("failed to update CloudFront function: %s", err)
	}

	if d.Get("publish").(bool) {
		log.Printf("[TRACE] Publishing CloudFront function %s in tenant %s", name, tenantID)
		err = c.PublishCloudFrontFunction(tenantID, fullName)
		if err != nil {
			return diag.Errorf("failed to publish CloudFront function: %s", err)
		}
	}

	return resourceCloudfrontFunctionRead(ctx, d, m)
}

func resourceCloudfrontFunctionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*duplosdk.Client)
	tenantID, name := parseCloudfrontFunctionId(d.Id())
	fullName, err := c.GetDuploServicesName(tenantID, name)
	if err != nil {
		return diag.Errorf("resourceCloudfrontFunctionDelete: Unable to retrieve duplo service name (tenant: %s, function: %s: error: %s)", tenantID, name, err)
	}

	// Call the Duplo API to delete the CloudFront function
	log.Printf("[TRACE] Deleting CloudFront function %s in tenant %s", name, tenantID)
	err = c.DeleteCloudFrontFunction(tenantID, fullName)
	if err != nil {
		return diag.Errorf("failed to delete CloudFront function: %s", err)
	}

	// Remove the resource from state
	d.SetId("")

	return nil
}

func parseCloudfrontFunctionId(id string) (tenantID, name string) {
	parts := strings.Split(id, "/")
	return parts[0], parts[2]
}

func expandCloudFrontFunction(d *schema.ResourceData) *duplosdk.DuploCloudFrontFunction {
	return &duplosdk.DuploCloudFrontFunction{
		Name:    d.Get("name").(string),
		Runtime: d.Get("runtime").(string),
		Code:    d.Get("code").(string),
		Comment: d.Get("comment").(string),
	}
}

func flattenCloudFrontFunction(d *schema.ResourceData, function *duplosdk.DuploCloudFrontFunctionResponse) {
	if function == nil {
		return
	}

	d.Set("runtime", function.FunctionSummary.FunctionConfig.Runtime.Value)
	d.Set("code", function.FunctionCode)
	d.Set("comment", function.FunctionSummary.FunctionConfig.Comment)
	d.Set("status", function.FunctionSummary.Status)
	if function.FunctionSummary.FunctionMetadata != nil {
		d.Set("function_arn", function.FunctionSummary.FunctionMetadata.FunctionARN)
	}
}
