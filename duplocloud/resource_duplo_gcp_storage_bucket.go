package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func gcpStorageBucketSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the storage bucket will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the storage bucket.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the storage bucket.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"self_link": {
			Description: "The SelfLink of the storage bucket.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"enable_versioning": {
			Description: "Whether or not versioning is enabled for the storage bucket.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"allow_public_access": {
			Description: "Whether or not public access might be allowed for the storage bucket.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"labels": {
			Description: "The labels assigned to this storage bucket.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

// Resource for managing a GCP storage bucket
func resourceGcpStorageBucket() *schema.Resource {
	return &schema.Resource{
		Description:        "`duplocloud_gcp_storage_bucket` manages a GCP storage bucket in Duplo.",
		DeprecationMessage: "duplocloud_gcp_storage_bucket is deprecated. Use duplocloud_gcp_storage_bucket_v2 instead.",

		ReadContext:   resourceGcpStorageBucketRead,
		CreateContext: resourceGcpStorageBucketCreate,
		UpdateContext: resourceGcpStorageBucketUpdate,
		DeleteContext: resourceGcpStorageBucketDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
		Schema: gcpStorageBucketSchema(),
	}
}

// READ resource
func resourceGcpStorageBucketRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpStorageBucketRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.GcpStorageBucketGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s storage bucket '%s': %s", tenantID, name, err)
	}

	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/%s", duplo.TenantID, name))
	resourceGcpStorageBucketSetData(d, tenantID, name, duplo)
	log.Printf("[TRACE] resourceGcpStorageBucketRead ******** end")
	return nil
}

// CREATE resource
func resourceGcpStorageBucketCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpStorageBucketCreate ******** start")

	// Create the request object.
	rq := duplosdk.DuploGcpStorageBucket{
		Name:             d.Get("name").(string),
		Labels:           expandAsStringMap("labels", d),
		EnableVersioning: d.Get("enable_versioning").(bool),
	}

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	rp, err := c.GcpStorageBucketCreate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s storage bucket '%s': %s", tenantID, rq.Name, err)
	}

	// Wait for Duplo to be able to return the storage bucket's details.
	id := fmt.Sprintf("%s/%s", tenantID, rq.Name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "storage bucket", id, func() (interface{}, duplosdk.ClientError) {
		return c.GcpStorageBucketGet(tenantID, rq.Name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	resourceGcpStorageBucketSetData(d, tenantID, rq.Name, rp)
	log.Printf("[TRACE] resourceGcpStorageBucketCreate ******** end")
	return diags
}

// UPDATE resource
func resourceGcpStorageBucketUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpStorageBucketUpdate ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	// Create the request object.
	rq := duplosdk.DuploGcpStorageBucket{
		Name:              d.Get("fullname").(string),
		Labels:            expandAsStringMap("labels", d),
		EnableVersioning:  d.Get("enable_versioning").(bool),
		AllowPublicAccess: d.Get("allow_public_access").(bool),
	}

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	rp, err := c.GcpStorageBucketUpdate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("Error updating tenant %s storage bucket '%s': %s", tenantID, rq.Name, err)
	}
	resourceGcpStorageBucketSetData(d, tenantID, name, rp)

	log.Printf("[TRACE] resourceGcpStorageBucketUpdate ******** end")
	return nil
}

// DELETE resource
func resourceGcpStorageBucketDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpStorageBucketDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	err := c.GcpStorageBucketDelete(idParts[0], idParts[1])
	if err != nil {
		return diag.Errorf("Error deleting storage bucket '%s': %s", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the storage bucket.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "storage bucket", id, func() (interface{}, duplosdk.ClientError) {
		return c.GcpStorageBucketGet(idParts[0], idParts[1])
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceGcpStorageBucketDelete ******** end")
	return nil
}

func resourceGcpStorageBucketSetData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploGcpStorageBucket) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("self_link", duplo.SelfLink)
	d.Set("enable_versioning", duplo.EnableVersioning)
	d.Set("allow_public_access", duplo.AllowPublicAccess)

	flattenGcpLabels(d, duplo.Labels)
}
