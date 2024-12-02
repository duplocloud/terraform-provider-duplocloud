package duplocloud

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsLaunchTemplateSchema() map[string]*schema.Schema {
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
		"version": {
			Description: "The existing latest version of the launch template",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"default_version": {
			Description: "The default version of the launch template to be set.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"latest_version": {
			Description: "The latest launch template version",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"version_description": {
			Description:      "The version of the launch template",
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},

		"instance_type": {
			Description:      "The version of the launch template",
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
		"version_metadata": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}
func resourceAwsLaunchTemplate() *schema.Resource {
	return &schema.Resource{
		Description:   "",
		ReadContext:   resourceAwsLaunchTemplateRead,
		CreateContext: resourceAwsLaunchTemplateCreate,
		UpdateContext: resourceAwsLaunchTemplateUpdate,
		DeleteContext: resourceAwsLaunchTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: awsLaunchTemplateSchema(),
	}
}

func resourceAwsLaunchTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantId, asgName, ver := idParts[0], idParts[2], idParts[3]
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
	}
	fErr := flattenLaunchTemplate(d, *rp, ver)
	if fErr != nil {
		return diag.Errorf("%s", fErr.Error())
	}
	return nil
}
func resourceAwsLaunchTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	rq := expandLaunchTemplate(d)
	c := m.(*duplosdk.Client)
	err := c.CreateAwsLaunchTemplate(tenantId, &rq)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	d.SetId(tenantId + "/launch-template/" + rq.LaunchTemplateName + "/" + rq.SourceVersion)
	diag := resourceAwsLaunchTemplateRead(ctx, d, m)
	return diag

}

func resourceAwsLaunchTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantId, asgName := idParts[0], idParts[2]
	rq := duplosdk.DuploAwsLaunchTemplateRequest{
		LaunchTemplateName: asgName,
		DefaultVersion:     d.Get("version").(string),
	}
	c := m.(*duplosdk.Client)

	err := c.UpdateAwsLaunchTemplateVersion(tenantId, &rq)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	d.SetId(tenantId + "/launch-template/" + rq.LaunchTemplateName + "/" + rq.SourceVersion)
	diag := resourceAwsLaunchTemplateRead(ctx, d, m)
	return diag
}

func resourceAwsLaunchTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil

}

func expandLaunchTemplate(d *schema.ResourceData) duplosdk.DuploAwsLaunchTemplateRequest {
	return duplosdk.DuploAwsLaunchTemplateRequest{
		LaunchTemplateName: d.Get("name").(string),
		SourceVersion:      d.Get("version").(string),
		VersionDescription: d.Get("version_description").(string),
		LaunchTemplateData: &duplosdk.DuploLaunchTemplateData{
			InstanceType: duplosdk.DuploStringValue{
				Value: d.Get("instance_type").(string),
			},
		},
	}

}

func flattenLaunchTemplate(d *schema.ResourceData, rp []duplosdk.DuploLaunchTemplateResponse, ver string) error {

	b, err := json.Marshal(rp)
	if err != nil {
		return err
	}
	var insType, verDesc, name, cver, dver string
	max := 0
	d.Set("version_metadata", string(b))
	for _, v := range rp {
		if strconv.Itoa(int(v.VersionNumber)) == ver {
			insType = v.LaunchTemplateData.InstanceType.Value
			verDesc = v.VersionDescription
			name = v.LaunchTemplateName
			cver = strconv.Itoa(int(v.VersionNumber))
		}
		if v.DefaultVersion {
			dver = strconv.Itoa(int(v.VersionNumber))
		}
		if max < int(v.VersionNumber) {
			max = int(v.VersionNumber)
		}
	}
	d.Set("instance_type", insType)
	d.Set("version_description", verDesc)
	d.Set("name", name)
	d.Set("version", cver)
	d.Set("latest_version", strconv.Itoa(max))
	d.Set("default_version", dver)

	return nil
}
