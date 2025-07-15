package duplocloud

import (
	"context"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsLaunchTemplateDefaultVersionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the launch template will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The fullname of the asg group",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"default_version": {
			Description: "The default version of the launch template to be set.",
			Type:        schema.TypeString,
			Required:    true,
		},
	}
}
func resourceAwsLaunchTemplateDefaultVersion() *schema.Resource {
	return &schema.Resource{
		Description:   "duplocloud_aws_launch_template_default_version helps to set or update default version of launch template",
		ReadContext:   resourceAwsLaunchTemplateDefaultVersionRead,
		CreateContext: resourceAwsLaunchTemplateDefaultVersionCreateAndUpdate,
		UpdateContext: resourceAwsLaunchTemplateDefaultVersionCreateAndUpdate,
		DeleteContext: resourceAwsLaunchTemplateDefaultVersionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: awsLaunchTemplateDefaultVersionSchema(),
	}
}

func resourceAwsLaunchTemplateDefaultVersionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantId, asgName := idParts[0], idParts[2]
	c := m.(*duplosdk.Client)
	rp, err := c.GetAwsLaunchTemplate(tenantId, asgName)
	if err != nil {
		if err.Status() == 404 {
			d.SetId("")
		}
		return diag.Errorf("%s", err.Error())

	}
	if rp == nil {
		d.SetId("")
		return nil
	}
	fErr := flattenLaunchTemplateDefaultVersion(d, *rp)
	if fErr != nil {
		return diag.Errorf("%s", fErr.Error())
	}
	return nil
}

func resourceAwsLaunchTemplateDefaultVersionCreateAndUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	rq := duplosdk.DuploAwsLaunchTemplateRequest{
		LaunchTemplateName: d.Get("name").(string),
		DefaultVersion:     d.Get("default_version").(string),
	}
	c := m.(*duplosdk.Client)

	err := c.UpdateAwsLaunchTemplateVersion(tenantId, &rq)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	d.SetId(tenantId + "/launch-template/" + rq.LaunchTemplateName)
	diag := resourceAwsLaunchTemplateDefaultVersionRead(ctx, d, m)
	return diag
}

func resourceAwsLaunchTemplateDefaultVersionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil

}

func flattenLaunchTemplateDefaultVersion(d *schema.ResourceData, rp []duplosdk.DuploLaunchTemplateResponse) error {

	var name, dver string

	for _, v := range rp {
		name = v.LaunchTemplateName
		if v.DefaultVersion {
			dver = strconv.Itoa(int(v.VersionNumber))
		}
	}
	d.Set("name", name)
	d.Set("default_version", dver)

	return nil
}
