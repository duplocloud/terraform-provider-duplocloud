package duplocloud

import (
	"context"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDatafactorySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure node pool will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the datafactory",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"public_access": {
			Description: "Enable or disable public access to datafactory",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
			Default:     false,
		},
		"version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"location": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"type": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"etag": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceAzureDataFactory() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_datafactory` manages datafactory in Duplo.",

		ReadContext:   resourceAzureDataFactoryRead,
		CreateContext: resourceAzureDataFactoryCreate,
		DeleteContext: resourceAzureDataFactoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: resourceDatafactorySchema(),
	}
}

func resourceAzureDataFactoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantId, name := idParts[0], idParts[2]
	log.Printf("[TRACE] resourceAzureDataFactoryRead(%s, %s): start", tenantId, name)

	c := m.(*duplosdk.Client)
	rp, err := c.GetAzureDataFactory(tenantId, name)
	if err != nil {
		if err.Status() == 404 {
			d.SetId("")
		}
		return diag.Errorf("Unable to retrieve datafactory details error: %s", err.Error())
	}
	if rp == nil {
		d.SetId("")
		return nil
	}
	flattenDataFactory(d, *rp)
	log.Printf("[TRACE] resourceAzureDataFactoryRead(%s, %s): end", tenantId, name)
	return nil
}

func resourceAzureDataFactoryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	rq := expandDatafactory(d)
	log.Printf("[TRACE] resourceAzureDataFactoryCreate(%s, %s): start", tenantId, rq.Name)
	c := m.(*duplosdk.Client)
	err := c.CreateAzureDataFactory(tenantId, rq)
	if err != nil {
		return diag.Errorf("error creating datafactory : %s", err.Error())
	}
	werr := waitUntilDataFactoryReady(ctx, c, tenantId, rq.Name, d.Timeout("create"))
	if werr != nil {
		return diag.Errorf("%s", werr.Error())
	}
	d.SetId(tenantId + "/datafactory/" + rq.Name)
	diag := resourceAzureDataFactoryRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureDataFactoryCreate(%s, %s): end", tenantId, rq.Name)

	return diag
}

func resourceAzureDataFactoryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantId, name := idParts[0], idParts[2]
	log.Printf("[TRACE] resourceAzureDataFactoryDelete(%s, %s): start", tenantId, name)

	c := m.(*duplosdk.Client)
	err := c.DeleteAzureDataFactory(tenantId, name)
	if err != nil {
		return diag.Errorf("error deleting datafactory: %s", err.Error())
	}
	log.Printf("[TRACE] resourceAzureDataFactoryDelete(%s, %s): end", tenantId, name)

	return nil
}

func expandDatafactory(d *schema.ResourceData) duplosdk.DuplocloudAzureDataFactoryRequest {
	o := duplosdk.DuplocloudAzureDataFactoryRequest{}
	o.Name = d.Get("name").(string)
	o.PublicEndPoint = d.Get("public_access").(bool)
	return o
}

func flattenDataFactory(d *schema.ResourceData, rp duplosdk.DuplocloudAzureDataFactoryResponse) {
	d.Set("name", rp.Name)
	d.Set("public_access", strings.EqualFold(rp.PublicAccess, "ENABLED"))
	d.Set("location", rp.Location)
	d.Set("type", rp.Type)
	d.Set("version", rp.Version)
	d.Set("etag", rp.ETag)
}

func waitUntilDataFactoryReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.GetAzureDataFactory(tenantID, name)
			status := "pending"
			if err == nil {
				if rp != nil && rp.State == "Succeeded" {
					status = "ready"
				} else {
					status = "pending"
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] waitUntilDataFactoryReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
