package duplocloud

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

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
			Description: "Any of the existing version of the launch template, if not provided, the latest version will be used",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"default_version": {
			Description: "The current default version of the launch template.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"latest_version": {
			Description: "The latest launch template version",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"version_description": {
			Description: "The version of the launch template",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Computed:    true,
		},

		"instance_type": {
			Description: "Asg instance type to be used to update the version from the current version",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"ami": {
			Description: "Asg ami to be used to update the version from the current version",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Computed:    true,
		},
		"version_metadata": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}
func resourceAwsLaunchTemplate() *schema.Resource {
	return &schema.Resource{
		Description:   "duplocloud_aws_launch_template creates the new version over current launch template version",
		ReadContext:   resourceAwsLaunchTemplateRead,
		CreateContext: resourceAwsLaunchTemplateCreate,
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
	ver := ""
	tenantId, asgName := idParts[0], idParts[2]
	if len(idParts) == 4 {
		ver = idParts[3]
	}

	c := m.(*duplosdk.Client)
	fullName := asgName
	var err1 error
	if !strings.Contains(asgName, "duploservices") {
		fullName, err1 = c.GetResourceName("duploservices", tenantId, asgName, false)
		if err1 != nil {
			diag.FromErr(err1)
		}
	}
	rp, err := c.GetAwsLaunchTemplate(tenantId, fullName)
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
	fErr := flattenLaunchTemplate(d, rp, ver)
	if fErr != nil {
		return diag.Errorf("%s", fErr.Error())
	}
	return nil
}
func resourceAwsLaunchTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)

	rq, cerr := expandLaunchTemplate(d, c, tenantId)
	if cerr != nil {
		return diag.Errorf("%s", cerr.Error())
	}
	name := rq.LaunchTemplateName
	var err duplosdk.ClientError
	if !strings.Contains(name, "duploservices") {
		rq.LaunchTemplateName, err = c.GetResourceName("duploservices", tenantId, name, false)
		if err != nil {
			diag.FromErr(err)
		}
	}
	err = c.CreateAwsLaunchTemplate(tenantId, rq)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	d.SetId(tenantId + "/launch-template/" + rq.LaunchTemplateName)
	diag := resourceAwsLaunchTemplateRead(ctx, d, m)
	return diag

}

func resourceAwsLaunchTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil

}

func expandLaunchTemplate(d *schema.ResourceData, c *duplosdk.Client, tenantId string) (*duplosdk.DuploAwsLaunchTemplateRequest, error) {
	sv := d.Get("version").(string)
	name := d.Get("name").(string)
	if sv == "" {
		rp, err := c.GetAwsLaunchTemplate(tenantId, name)
		if err != nil {
			return nil, err
		}
		_, _, _, _, sv, _ = extractASGTemplateDetails(rp)
		log.Printf("Setting the version to latest version %s since source version not provided ", sv)
	}
	return &duplosdk.DuploAwsLaunchTemplateRequest{
		LaunchTemplateName: name,
		SourceVersion:      sv,
		VersionDescription: d.Get("version_description").(string),
		LaunchTemplateData: &duplosdk.DuploLaunchTemplateData{
			InstanceType: duplosdk.DuploStringValue{
				Value: d.Get("instance_type").(string),
			},
			ImageId: d.Get("ami").(string),
		},
	}, nil

}

func flattenLaunchTemplate(d *schema.ResourceData, rp *[]duplosdk.DuploLaunchTemplateResponse, ver string) error {

	b, err := json.Marshal(rp)
	if err != nil {
		return err
	}
	name, insType, verDesc, dver, lver, imgId := extractASGTemplateDetails(rp)
	d.Set("version_metadata", string(b))
	d.Set("instance_type", insType)
	d.Set("version_description", verDesc)
	n := d.Get("name").(string)
	d.Set("name", name)
	if !strings.Contains(n, "duploservices") {
		d.Set("name", n)
	}

	if v, ok := d.GetOk("version"); ok && v.(string) != "" {
		d.Set("version", v.(string))
	}
	d.Set("latest_version", lver)
	d.Set("default_version", dver)
	d.Set("ami", imgId)
	return nil
}

func extractASGTemplateDetails(rp *[]duplosdk.DuploLaunchTemplateResponse) (string, string, string, string, string, string) {
	var name, insType, verDesc, dver, imgId string
	max := 0
	for _, v := range *rp {

		if v.DefaultVersion {
			dver = strconv.Itoa(int(v.VersionNumber))
		}
		if max < int(v.VersionNumber) {
			max = int(v.VersionNumber)
			insType = v.LaunchTemplateData.InstanceType.Value
			verDesc = v.VersionDescription
			imgId = v.LaunchTemplateData.ImageId
			name = v.LaunchTemplateName
		}
	}
	lver := strconv.Itoa(max)
	return name, insType, verDesc, dver, lver, imgId
}
