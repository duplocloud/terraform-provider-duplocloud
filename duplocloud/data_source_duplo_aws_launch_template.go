package duplocloud

import (
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsLaunchTemplate() *schema.Resource {
	return &schema.Resource{
		Read: datasourceAwsLaunchTemplateRead,

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Description: "Any of the existing version of the launch template",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"default_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"latest_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ami": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func datasourceAwsLaunchTemplateRead(d *schema.ResourceData, m interface{}) error {
	tenantId := d.Get("tenant_id").(string)
	asgName := d.Get("name").(string)
	ver := "1"
	c := m.(*duplosdk.Client)
	rp, err := c.GetAwsLaunchTemplate(tenantId, asgName)
	if err != nil {
		if err.Status() == 404 {
			d.SetId("")
		}
		return err

	}
	if rp == nil {
		d.SetId("")
		return nil
	}
	d.SetId(tenantId + "/launch-template/" + asgName + "/" + ver)

	fErr := flattenLaunchTemplateData(d, rp, ver)
	if fErr != nil {
		return err
	}
	return nil
}

func flattenLaunchTemplateData(d *schema.ResourceData, rp *[]duplosdk.DuploLaunchTemplateResponse, ver string) error {

	var name, cver, insType, verDesc, dver, imgId string
	max := 0
	for _, v := range *rp {
		if strconv.Itoa(int(v.VersionNumber)) == ver {
			name = v.LaunchTemplateName
			cver = strconv.Itoa(int(v.VersionNumber))
		}
		if v.DefaultVersion {
			dver = strconv.Itoa(int(v.VersionNumber))
		}
		if max < int(v.VersionNumber) {
			max = int(v.VersionNumber)
			insType = v.LaunchTemplateData.InstanceType.Value
			verDesc = v.VersionDescription
			imgId = v.LaunchTemplateData.ImageId
		}
	}
	d.Set("instance_type", insType)
	d.Set("version_description", verDesc)
	d.Set("name", name)
	d.Set("version", cver)
	d.Set("latest_version", strconv.Itoa(max))
	d.Set("default_version", dver)
	d.Set("ami", imgId)
	return nil
}
