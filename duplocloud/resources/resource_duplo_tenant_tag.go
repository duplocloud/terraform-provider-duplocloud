package resources

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplocloud"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTenantTag() *schema.Resource {
	return &schema.Resource{
		Description:   "`duplocloud_tenant_tag` manages a tenant tag in Duplo.",
		ReadContext:   resourceTenantTagRead,
		CreateContext: resourceTenantTagCreate,
		UpdateContext: resourceTenantTagRead, // NO-OP
		DeleteContext: resourceTenantTagDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
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
			},
		},
	}
}

func resourceTenantTagRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, key, err := parseTenantTagIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceTenantTagRead(%s): start", tenantID)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	tag, clientErr := c.TenantTagsGetByKey(tenantID, key)
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

	log.Printf("[TRACE] resourceTenantTagRead(%s): end", tenantID)
	return nil
}

func resourceTenantTagCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rq := duplosdk.DuploTenantConfigUpdateRequest{
		TenantID: d.Get("tenant_id").(string),
		Key:      d.Get("key").(string),
		Value:    d.Get("value").(string),
	}

	log.Printf("[TRACE] resourceTenantTagCreate(%s, %s): start", rq.TenantID, rq.Key)

	c := m.(*duplosdk.Client)

	err := c.TenantTagCreate(rq)
	if err != nil {
		return diag.Errorf("Unable to create tenant tag '%s', '%s': %s", rq.TenantID, rq.Key, err)
	}

	diags := duplocloud.waitForResourceToBePresentAfterCreate(ctx, d, "tenant tag", rq.Key, func() (interface{}, duplosdk.ClientError) {
		return c.TenantTagsGetByKey(rq.TenantID, rq.Key)
	})
	if diags != nil {
		return diags
	}
	d.SetId(fmt.Sprintf("%s/%s", rq.TenantID, rq.Key))

	diags = resourceTenantTagRead(ctx, d, m)
	if diags != nil {
		return diags
	}
	log.Printf("[TRACE] resourceTenantTagCreate(%s, %s): end", rq.TenantID, rq.Key)
	return nil
}

func resourceTenantTagDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()

	// Parse the identifying attributes
	tenantID, key, err := parseTenantTagIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceTenantTagDelete(%s, %s): start", tenantID, key)

	rq := duplosdk.DuploTenantConfigUpdateRequest{
		TenantID: tenantID,
		Key:      key,
		Value:    d.Get("value").(string),
		State:    "delete",
	}

	c := m.(*duplosdk.Client)
	clientErr := c.TenantTagDelete(rq)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant tag '%s', '%s': %s", tenantID, key, err)
	}

	diag := duplocloud.waitForResourceToBeMissingAfterDelete(ctx, d, "tenant tag", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantTagsGetByKey(rq.TenantID, rq.Key)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceTenantTagDelete(%s, %s): end", tenantID, key)
	return nil
}

func parseTenantTagIdParts(id string) (tenantID, key string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, key = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
