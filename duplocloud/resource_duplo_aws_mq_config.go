package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAwsMQConfigSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the custom tag for a resource will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Description: "The name of MQ configuration.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
		},
		"configuration_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"data": {
			Type:        schema.TypeString,
			Description: "Base64 configuration data, applicable for only updating",
			Optional:    true,
		},
		"description": {
			Type:        schema.TypeString,
			Description: "Descritpion of MQ configuration, can be set only during update",
			Computed:    true,
			Optional:    true,
		},
		"latest_version": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"authentication_strategy": {
			Description:  "Specify authentication strategy, valid values are `SIMPLE`, `LDAP`",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"SIMPLE", "LDAP"}, false),
		},
		"engine_type": {
			Description:  "Specify engine type, valid values are `ACTIVEMQ`, `RABBITMQ`",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"ACTIVEMQ", "RABBITMQ"}, false),
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Description: "A map of tags to assign to the resource.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

func resourceAwsMQConfig() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_mq_config` manages an configuration for amazon mq managed by Duplo.",

		ReadContext:   resourceAwsMQConfigRead,
		CreateContext: resourceAwsMQConfigCreate,
		UpdateContext: resourceAwsMQConfigUpdate,
		DeleteContext: resourceAwsMQConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema:        duploAwsMQConfigSchema(),
		CustomizeDiff: resourceAwsMQConfigCustomizeDiff,
	}
}

func resourceAwsMQConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		return diag.Errorf("invalid resource id")
	}
	log.Printf("[TRACE] resourceAwsMQConfigRead(%s, %s): start", idParts[0], idParts[1])

	c := m.(*duplosdk.Client)

	rp, clientErr := c.DuploAWSMQConfigGet(idParts[0], idParts[1])
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve AWS MQ configuration - (Tenant: %s,  Configuration Id: %s) : %s", idParts[0], idParts[1], clientErr)
	}
	if rp == nil {
		d.SetId("") // object missing
		return nil
	}

	d.Set("tenant_id", idParts[0])
	d.Set("configuration_id", idParts[1])
	d.Set("name", rp.Name)
	d.Set("engine_type", rp.EngineType.Value)
	d.Set("engine_version", rp.EngineVersion)
	d.Set("authentication_strategy", rp.AuthenticationStrategy.Value)
	d.Set("latest_version", rp.LatestRevision.Revision)
	d.Set("arn", rp.Arn)
	d.Set("tags", rp.Tags)
	d.Set("description", rp.LatestRevision.Description)
	log.Printf("[TRACE] resourceAwsMQConfigRead(%s, %s): end", idParts[0], idParts[1])
	return nil
}

func resourceAwsMQConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	req := duplosdk.DuploAwsMQConfig{
		AuthenticationStrategy: d.Get("authentication_strategy").(string),
		EngineType:             d.Get("engine_type").(string),
		EngineVersion:          d.Get("engine_version").(string),
		Name:                   d.Get("name").(string),
		Tags:                   d.Get("tags").(map[string]interface{}),
	}
	log.Printf("[TRACE] resourceAwsMQConfigCreate(%s,%s): start", tenantID, req.Name)
	c := m.(*duplosdk.Client)

	rp, err := c.DuploAWSMQConfigCreate(tenantID, req)
	if err != nil {
		return diag.Errorf("Error creating aws MQ configuration - (Tenant: %s,  Name: %s) : %s", tenantID, req.Name, err)
	}
	id := fmt.Sprintf("%s/%s", tenantID, rp["id"].(string))

	d.SetId(id)

	diags := resourceAwsMQConfigRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsMQConfigCreate(%s,%s): end", tenantID, req.Name)
	return diags
}

func resourceAwsMQConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 {
		return diag.Errorf("invalid resource id")
	}
	log.Printf("[TRACE] resourceAwsMQConfigUpdate(%s, %s): start", idParts[0], idParts[1])

	c := m.(*duplosdk.Client)
	update := false
	rq := duplosdk.DuploAwsMQConfigUpdate{}
	if d.HasChange("data") {
		rq.Data = d.Get("data").(string)
		update = true
	}
	if d.HasChange("description") {
		rq.Description = d.Get("description").(string)
		update = true
	}
	if update {
		rq.ConfigurationId = idParts[1]
		cerr := c.DuploAWSMQConfigUpdate(idParts[0], idParts[1], rq)
		if cerr != nil {
			return diag.Errorf("Updating MQ config %s for tenant %s failed : %s", rq.ConfigurationId, idParts[0], cerr.Error())
		}
	}
	log.Printf("[TRACE] resourceAwsMQConfigUpdate(%s, %s): end", idParts[0], idParts[1])
	return nil
}

func resourceAwsMQConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAwsMQConfigCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	id := d.Id()
	if id != "" {
		if d.Get("data").(string) == "" {
			return fmt.Errorf("the 'data' field can be set only during update")
		}
		if d.Get("description").(string) == "" {
			return fmt.Errorf("the 'description' field can be set only during update")
		}
	}
	return nil
}
