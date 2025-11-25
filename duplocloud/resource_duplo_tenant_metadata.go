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

func resourceTenantMetadata() *schema.Resource {
	return &schema.Resource{
		Description:   "`duplocloud_tenant_metadata` manages a tenant metadata in Duplo.",
		ReadContext:   resourceTenantMetadataRead,
		CreateContext: resourceTenantMetadataCreate,
		DeleteContext: resourceTenantMetadataDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The GUID of the tenant that the tags will be created in.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"key": {
				Description: "Specify key for tag.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"value": {
				Description: "Specify value for tag.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Description:  "Specify type of metadata, Valid values are `text`, `url`, `aws_console`",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"text", "url", "aws_console"}, false),
			},
		},
	}
}

func resourceTenantMetadataRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, key, err := parseTenantMetadataIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceTenantMetadataRead(%s): start", tenantID)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	tag, clientErr := c.TenantMetadatasGetByKey(tenantID, key)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant tags '%s': %s", tenantID, err)
	}
	if tag == nil {
		d.SetId("") // object missing
		return nil
	}

	d.Set("tenant_id", tenantID)
	d.Set("key", tag.Key)
	d.Set("value", tag.Value)
	d.Set("type", tag.Type)
	log.Printf("[TRACE] resourceTenantMetadataRead(%s): end", tenantID)
	return nil
}

func resourceTenantMetadataCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rq := duplosdk.DuploTenantMetadata{
		TenantId: d.Get("tenant_id").(string),
		Key:      d.Get("key").(string),
		Value:    d.Get("value").(string),
		Type:     d.Get("type").(string),
	}

	log.Printf("[TRACE] resourceTenantMetadataCreate(%s, %s): start", rq.TenantId, rq.Key)

	c := m.(*duplosdk.Client)

	err := c.TenantMetadataManage(rq.TenantId, &rq)
	if err != nil {
		if err.Status() == 404 {
			log.Printf("[TRACE] resourceTenantMetadataCreate(%s, %s): object not found", rq.TenantId, rq.Key)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to create tenant tag '%s', '%s': %s", rq.TenantId, rq.Key, err)
	}

	diags := waitForResourceToBePresentAfterCreate(ctx, d, "tenant metadata", rq.Key, func() (interface{}, duplosdk.ClientError) {
		return c.TenantMetadatasGetByKey(rq.TenantId, rq.Key)
	})
	if diags != nil {
		return diags
	}
	d.SetId(fmt.Sprintf("%s/%s", rq.TenantId, rq.Key))

	diags = resourceTenantMetadataRead(ctx, d, m)
	if diags != nil {
		return diags
	}
	log.Printf("[TRACE] resourceTenantMetadataCreate(%s, %s): end", rq.TenantId, rq.Key)
	return nil
}

func resourceTenantMetadataDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()

	// Parse the identifying attributes
	tenantID, key, err := parseTenantMetadataIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceTenantMetadataDelete(%s, %s): start", tenantID, key)

	rq := duplosdk.DuploTenantMetadata{
		TenantId: tenantID,
		Key:      key,
		Value:    d.Get("value").(string),
		State:    "delete",
	}

	c := m.(*duplosdk.Client)
	clientErr := c.TenantMetadataManage(tenantID, &rq)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[TRACE] resourceTenantMetadataDelete(%s, %s): object not found", tenantID, key)
			return nil
		}
		return diag.Errorf("Unable to delete tenant tag '%s', '%s': %s", tenantID, key, err)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "tenant tag", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantMetadatasGetByKey(tenantID, rq.Key)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceTenantMetadataDelete(%s, %s): end", tenantID, key)
	return nil
}

func parseTenantMetadataIdParts(id string) (tenantID, key string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, key = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
