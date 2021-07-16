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
)

func gcpPubsubTopicSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description: "The GUID of the tenant that the pubsub topic will be created in.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"name": {
			Description: "The short name of the pubsub topic.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the pubsub topic.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"self_link": {
			Description: "The SelfLink of the pubsub topic.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"labels": {
			Description: "The labels assigned to this pubsub topic.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

// Resource for managing a GCP pubsub topic
func resourceGcpPubsubTopic() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_pubsub_topic` manages a GCP pubsub topic in Duplo.",

		ReadContext:   resourceGcpPubsubTopicRead,
		CreateContext: resourceGcpPubsubTopicCreate,
		UpdateContext: resourceGcpPubsubTopicUpdate,
		DeleteContext: resourceGcpPubsubTopicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
		Schema: gcpPubsubTopicSchema(),
	}
}

/// READ resource
func resourceGcpPubsubTopicRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpPubsubTopicRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.GcpPubsubTopicGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s pubsub topic '%s': %s", tenantID, name, err)
	}

	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/%s", duplo.TenantID, name))
	resourceGcpPubsubTopicSetData(d, tenantID, name, duplo)
	log.Printf("[TRACE] resourceGcpPubsubTopicRead ******** end")
	return nil
}

/// CREATE resource
func resourceGcpPubsubTopicCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpPubsubTopicCreate ******** start")

	// Create the request object.
	rq := duplosdk.DuploGcpPubsubTopic{
		Name:   d.Get("name").(string),
		Labels: expandGcpLabels("labels", d),
	}

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	rp, err := c.GcpPubsubTopicCreate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s pubsub topic '%s': %s", tenantID, rq.Name, err)
	}

	// Wait for Duplo to be able to return the pubsub topic's details.
	id := fmt.Sprintf("%s/%s", tenantID, rq.Name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "pubsub topic", id, func() (interface{}, duplosdk.ClientError) {
		return c.GcpPubsubTopicGet(tenantID, rq.Name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	resourceGcpPubsubTopicSetData(d, tenantID, rq.Name, rp)
	log.Printf("[TRACE] resourceGcpPubsubTopicCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceGcpPubsubTopicUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpPubsubTopicUpdate ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	// Create the request object.
	rq := duplosdk.DuploGcpPubsubTopic{
		Name:   d.Get("name").(string),
		Labels: expandGcpLabels("labels", d),
	}

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	rp, err := c.GcpPubsubTopicUpdate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("Error updating tenant %s pubsub topic '%s': %s", tenantID, rq.Name, err)
	}
	resourceGcpPubsubTopicSetData(d, tenantID, name, rp)

	log.Printf("[TRACE] resourceGcpPubsubTopicUpdate ******** end")
	return nil
}

/// DELETE resource
func resourceGcpPubsubTopicDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpPubsubTopicDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	err := c.GcpPubsubTopicDelete(idParts[0], idParts[1])
	if err != nil {
		return diag.Errorf("Error deleting pubsub topic '%s': %s", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the pubsub topic.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "pubsub topic", id, func() (interface{}, duplosdk.ClientError) {
		return c.GcpPubsubTopicGet(idParts[0], idParts[1])
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceGcpPubsubTopicDelete ******** end")
	return nil
}

func resourceGcpPubsubTopicSetData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploGcpPubsubTopic) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("self_link", duplo.SelfLink)

	flattenGcpLabels(d, duplo.Labels)
}
