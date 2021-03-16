package duplocloud

import (
	"context"
	"fmt"
	"strconv"

	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DuploServiceSchema returns a Terraform resource schema for a service's parameters
func duploServiceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"tenant_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"other_docker_host_config": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"other_docker_config": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"extra_config": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"allocation_tags": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"volumes": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"commands": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"cloud": {
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
		},
		"agent_platform": {
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
		},
		"replicas": {
			Type:     schema.TypeInt,
			Optional: false,
			Required: true,
		},
		"replicas_matching_asg_name": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"docker_image": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
		},
		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
	}
}

// SCHEMA for resource crud
func resourceDuploService() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceDuploServiceRead,
		CreateContext: resourceDuploServiceCreate,
		UpdateContext: resourceDuploServiceUpdate,
		DeleteContext: resourceDuploServiceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploServiceSchema(),
	}
}

// SCHEMA for resource data/search
func dataSourceDuploService() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploServiceRead,
		Schema: map[string]*schema.Schema{
			"filter": FilterSchema(), // todo: search specific to this object... may be api should support filter?
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: false,
				Optional: true,
			},
			"data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: duploServiceSchema(),
				},
			},
		},
	}
}

/// READ/SEARCH resources
func dataSourceDuploServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-dataSourceDuploServiceRead ******** start")

	c := m.(*duplosdk.Client)
	var diags diag.Diagnostics
	duploObjs, err := c.DuploServiceGetList(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	itemList := c.DuploServicesFlatten(duploObjs, d)
	if err := d.Set("data", itemList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] duplo-dataSourceDuploServiceRead ******** end")

	return diags
}

/// READ resource
func resourceDuploServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceDuploServiceRead ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	err := c.DuploServiceGet(d, m)
	if d.Get("name").(string) == "" {
		return diag.Diagnostics{
			{
				Severity: diag.Warning,
				Summary:  "Service does not exist",
				Detail:   fmt.Sprintf("Service: %v does not exist. It may have been deleted outside of Terraform", d.Id()),
			},
		}
	}
	if err != nil {
		return diag.FromErr(err)
	}

	c.DuploServiceSetID(d)
	log.Printf("[TRACE] duplo-resourceDuploServiceRead ******** end")
	return diags
}

/// CREATE resource
func resourceDuploServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceDuploServiceCreate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.DuploServiceCreate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.DuploServiceSetID(d)
	resourceDuploServiceRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceDuploServiceCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceDuploServiceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceDuploServiceUpdate ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.DuploServiceUpdate(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	c.DuploServiceSetID(d)
	resourceDuploServiceRead(ctx, d, m)
	log.Printf("[TRACE] duplo-resourceDuploServiceUpdate ******** end")

	return diags
}

/// DELETE resource
func resourceDuploServiceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] duplo-resourceDuploServiceDelete ******** start")

	c := m.(*duplosdk.Client)

	var diags diag.Diagnostics
	_, err := c.DuploServiceDelete(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	//todo: wait for it completely deleted

	log.Printf("[TRACE] duplo-resourceDuploServiceDelete ******** end")

	return diags
}
