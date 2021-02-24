package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// DuploEcsTaskDefinitionSchema returns a Terraform resource schema for an ECS Task Definition
func ecsTaskDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"family": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"revision": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"status": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"container_definitions": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"volumes": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Default:  "[]",
		},
		"cpu": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"task_role_arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"execution_role_arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"memory": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"network_mode": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"bridge", "host", "awsvpc", "none"}, false),
			Default:      "awsvpc",
		},
		"placement_constraints": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			MaxItems: 10,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:         schema.TypeString,
						ForceNew:     true,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"memberOf"}, false),
					},
					"expression": {
						Type:     schema.TypeString,
						ForceNew: true,
						Optional: true,
					},
				},
			},
		},
		"requires_attributes": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						ForceNew: true,
						Required: true,
					},
				},
			},
		},
		"requires_compatibilities": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"ipc_mode": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"host", "none", "task"}, false),
		},
		"pid_mode": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"host", "task"}, false),
		},
		"proxy_configuration": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"container_name": {
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
					},
					"properties": {
						Type:     schema.TypeMap,
						Elem:     &schema.Schema{Type: schema.TypeString},
						Optional: true,
						ForceNew: true,
					},
					"type": {
						Type:         schema.TypeString,
						Default:      "APPMESH",
						Optional:     true,
						ForceNew:     true,
						ValidateFunc: validation.StringInSlice([]string{"APPMESH"}, false),
					},
				},
			},
		},
		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			Required: false,
			Elem:     KeyValueSchema(),
		},
		"inference_accelerator": {
			Type:     schema.TypeSet,
			Optional: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"device_name": {
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
					},
					"device_type": {
						Type:     schema.TypeString,
						Required: true,
						ForceNew: true,
					},
				},
			},
		},
	}
}

// SCHEMA for resource crud
func resourceDuploEcsTaskDefinition() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceDuploEcsTaskDefinitionRead,
		CreateContext: resourceDuploEcsTaskDefinitionCreate,
		//UpdateContext: resourceDuploEcsTaskDefinitionUpdate,
		DeleteContext: resourceDuploEcsTaskDefinitionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: ecsTaskDefinitionSchema(),
	}
}

/// READ resource
func resourceDuploEcsTaskDefinitionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionRead ******** start")

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.EcsTaskDefinitionGet(d.Id())
	if duplo == nil {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	// Convert the object into Terraform resource data
	jo := duplosdk.EcsTaskDefToState(duplo, d)
	for key := range jo {
		d.Set(key, jo[key])
	}
	d.SetId(fmt.Sprintf("subscriptions/%s/EcsTaskDefinition/%s", duplo.TenantID, duplo.Arn))

	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionRead ******** end")
	return nil
}

/// CREATE resource
func resourceDuploEcsTaskDefinitionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionCreate ******** start")

	// Convert the Terraform resource data into a Duplo object
	duploObject, err := duplosdk.EcsTaskDefFromState(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	arn, err := c.EcsTaskDefinitionCreate(tenantID, duploObject)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("subscriptions/%s/EcsTaskDefinition/%s", tenantID, arn))

	diags := resourceDuploEcsTaskDefinitionRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionCreate ******** end")
	return diags
}

/// DELETE resource
func resourceDuploEcsTaskDefinitionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionDelete ******** start")
	// FIXME: NO-OP
	log.Printf("[TRACE] resourceDuploEcsTaskDefinitionDelete ******** end")
	return nil
}
