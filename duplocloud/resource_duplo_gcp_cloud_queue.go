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

func gcpCloudTasksQueueSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant through which the gcp cloud tasks queue will be registered.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the cloud tasks queue",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"location": {
			Description: "The name of the cloud tasks queue",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"fullname": {
			Description: "The full name of the cloud function.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceGcpCloudQueue() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_cloud_queue` manages a GCP queue for publishing tasks.",

		ReadContext:   resourceGcpCloudQueueRead,
		CreateContext: resourceGcpCloudQueueCreate,
		DeleteContext: resourceGcpCloudQueueDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
		Schema: gcpCloudTasksQueueSchema(),
	}
}

func resourceGcpCloudQueueRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpCloudQueueRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) != 3 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, qName := idParts[0], idParts[2]

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.GCPCloudTasksQueueGet(tenantID, qName)
	if duplo == nil || (err != nil && err.Status() == 404) {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s cloud task's queue '%s' : %s", tenantID, qName, err)
	}

	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/tasks/queue/%s", tenantID, qName))
	resourceGcpCloudQueueSetData(d, tenantID, qName, duplo)
	log.Printf("[TRACE] resourceGcpCloudQueueRead ******** end")
	return nil
}

// CREATE resource
func resourceGcpCloudQueueCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpCloudQueueCreate ******** start")

	// Create the request object.
	rq := expandGcpCloudTasksQueue(d)

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	// Post the object to Duplo
	err := c.GcpCloudTasksQueueCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s cloud task's queue %s : %s", tenantID, rq.QueueName, err)
	}

	id := fmt.Sprintf("%s/queue/%s", tenantID, rq.QueueName)
	d.SetId(id)

	resourceGcpCloudQueueRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpCloudQueueCreate ******** end")
	return nil
}

// DELETE resource
func resourceGcpCloudQueueDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpCloudQueueDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) != 3 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, qName := idParts[0], idParts[2]

	err := c.GCPCloudTasksQueueDelete(tenantID, qName)
	if err != nil {
		if err.Status() == 404 {
			return nil
		}
		return diag.Errorf("Error deleting cloud tasks queue %s : %s", qName, err)
	}

	log.Printf("[TRACE] resourceGcpCloudQueueDelete ******** end")
	return nil
}

func resourceGcpCloudQueueSetData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploGCPCloudTaskQueue) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("location", duplo.Location)
	d.Set("fullname", duplo.QueueName)
}

func expandGcpCloudTasksQueue(d *schema.ResourceData) *duplosdk.DuploGCPCloudTaskQueue {

	duplo := duplosdk.DuploGCPCloudTaskQueue{
		QueueName: d.Get("name").(string),
		Location:  d.Get("location").(string),
	}
	return &duplo
}
