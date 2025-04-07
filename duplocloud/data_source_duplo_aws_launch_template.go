package duplocloud

import (
	"strings"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
				Optional:    true,
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
	name := d.Get("name").(string)
	ver := d.Get("version").(string)
	asgName := name
	var err duplosdk.ClientError
	c := m.(*duplosdk.Client)
	if !strings.Contains(name, "duploservices") {
		asgName, err = c.GetResourceName("duploservices", tenantId, name, false)
		if err != nil {
			diag.FromErr(err)
		}
	}

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
	d.SetId(tenantId + "/launch-template/" + asgName)

	fErr := flattenLaunchTemplate(d, rp, ver)
	if fErr != nil {
		return err
	}
	return nil
}
