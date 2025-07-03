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

func gcpCloudTasksSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant through which the gcp cloud tasks queue will be registered.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"queue_name": {
			Description: "The name of the cloud tasks queue",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"name": {
			Description: "The name of the cloud tasks queue",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the cloud function.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"http_target": {
			Description: "Specifies an HTTPS trigger for the cloud function.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Description: "Specify the endpoint URL to which the HTTP request will be sent when the Cloud Tasks queue triggers the HTTP target.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"method": {
						Description:  "The HTTP method to use for the request. Must be one of: `POST`, `PUT`, `DELETE`, `PATCH`, `HEAD`.",
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"POST", "PUT", "DELETE", "PATCH", "HEAD"}, false),
					},
					"headers": {
						Description: "A map of HTTP headers to include in the request. Each key is a header name, and each value is the corresponding header value.",
						Type:        schema.TypeMap,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Required:    true,
					},
					"body": {
						Description: "The body of the HTTP request. This field is required and must be base64 string.",
						Type:        schema.TypeString,
						Optional:    true,
						ForceNew:    true,
					},
				},
			},
			ConflictsWith: []string{"app_engine"},
		},
		"app_engine": {
			Description: "Specifies an HTTPS trigger for the cloud function.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"relative_uri": {
						Description: "Specify the relative URL path to which the HTTP request will be sent when the Cloud Tasks queue triggers the App Engine target.",
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
					},
					"method": {
						Description:  "The HTTP method to use for the request. Must be one of: `POST`, `PUT`, `PATCH`.",
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"POST", "PUT", "PATCH"}, false),
						ForceNew:     true,
					},
					"headers": {
						Description: "A map of HTTP headers to include in the request. Each key is a header name, and each value is the corresponding header value.",
						Type:        schema.TypeMap,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Required:    true,
						ForceNew:    true,
					},
					"body": {
						Description: "The body of the HTTP request. This field is optional and can be used to send additional data in the request should be base64 encoded.",
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
					},
				},
			},
			ConflictsWith: []string{"http_target"},
		},
	}
}

// Resource for managing a GCP cloud function
func resourceGcpCloudQueueTask() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_cloud_queue_task` manages a GCP cloud task for respective queue in Duplo.",

		ReadContext:   resourceGcpCloudQueueTaskRead,
		CreateContext: resourceGcpCloudQueueTaskCreate,
		DeleteContext: resourceGcpCloudQueueTaskDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
		Schema: gcpCloudTasksSchema(),
	}
}

// READ resource
func resourceGcpCloudQueueTaskRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpCloudQueueTaskRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) < 5 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, qName, task := idParts[0], idParts[2], idParts[4]

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.GCPCloudTasksGet(tenantID, qName, task)
	if duplo == nil || (err != nil && err.Status() == 404) {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s cloud queue '%s' task %s : %s", tenantID, qName, task, err)
	}

	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/queue/%s/task/%s", tenantID, qName, task))
	resourceGcpCloudQueueTaskSetData(d, tenantID, qName, task, duplo)
	log.Printf("[TRACE] resourceGcpCloudQueueTaskRead ******** end")
	return nil
}

// CREATE resource
func resourceGcpCloudQueueTaskCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpCloudQueueTaskCreate ******** start")

	// Create the request object.
	rq := expandGcpCloudQueueTask(d)

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	queueName := d.Get("queue_name").(string)
	// Post the object to Duplo
	err := c.GcpCloudTasksCreate(tenantID, queueName, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s cloud queue %s task %s: %s", tenantID, queueName, rq.TaskName, err)
	}

	// Wait for Duplo to be able to return the cloud function's details.
	id := fmt.Sprintf("%s/queue/%s/task/%s", tenantID, queueName, rq.TaskName)
	d.SetId(id)

	resourceGcpCloudQueueTaskRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpCloudQueueTaskCreate ******** end")
	return nil
}

// DELETE resource
func resourceGcpCloudQueueTaskDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpCloudQueueTaskDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) < 5 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, qName, task := idParts[0], idParts[2], idParts[4]
	err := c.GCPCloudTasksDelete(tenantID, qName, task)
	if err != nil {
		return diag.Errorf("Error deleting cloud queue %s task '%s': %s", qName, task, err)
	}

	log.Printf("[TRACE] resourceGcpCloudQueueTaskDelete ******** end")
	return nil
}

func resourceGcpCloudQueueTaskSetData(d *schema.ResourceData, tenantID string, qName, name string, duplo *duplosdk.DuploGCPCloudTasks) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.TaskName)
	d.Set("queue_name", qName)
	if duplo.TaskType == 1 {
		m := map[string]interface{}{
			"relative_uri": duplo.RelativeUri,
			"method":       duplo.Method,
			"body":         duplo.Body,
			"headers":      duplo.Headers,
		}
		d.Set("app_engine", []interface{}{m})
	}
	if duplo.TaskType == 0 {
		m := map[string]interface{}{
			"url":     duplo.Url,
			"method":  duplo.Method,
			"body":    duplo.Body,
			"headers": duplo.Headers,
		}
		d.Set("http_target", []interface{}{m})
	}
}

func expandGcpCloudQueueTask(d *schema.ResourceData) *duplosdk.DuploGCPCloudTasks {

	duplo := duplosdk.DuploGCPCloudTasks{
		TaskName: d.Get("name").(string),
	}
	appConfig := d.Get("app_engine")
	if appConfig != nil && len(appConfig.([]interface{})) > 0 {
		mp := appConfig.([]interface{})[0].(map[string]interface{})
		duplo.TaskType = 1
		duplo.RelativeUri = mp["relative_uri"].(string)
		duplo.Method = mp["method"].(string)
		duplo.Body = mp["body"].(string)
		if v, ok := mp["headers"].(map[string]interface{}); ok && len(v) > 0 {
			m := map[string]string{}
			for k, val := range v {
				if val == nil {
					m[k] = ""
				} else {
					m[k] = val.(string)
				}
			}
			duplo.Headers = m
		}

	}
	httpTrgt := d.Get("http_target")

	if httpTrgt != nil && len(httpTrgt.([]interface{})) > 0 {
		mp := httpTrgt.([]interface{})[0].(map[string]interface{})
		duplo.TaskType = 0
		duplo.Url = mp["url"].(string)
		duplo.Method = mp["method"].(string)
		duplo.Body = mp["body"].(string)
		if v, ok := mp["headers"].(map[string]interface{}); ok && len(v) > 0 {
			m := map[string]string{}
			for k, val := range v {
				if val == nil {
					m[k] = ""
				} else {
					m[k] = val.(string)
				}
			}
			duplo.Headers = m
		}
	}

	return &duplo
}

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

func resourceGcpCloudTasksQueue() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_cloud_tasks_queue` manages a GCP queue for publishing tasks.",

		ReadContext:   resourceGcpCloudTasksQueueRead,
		CreateContext: resourceGcpCloudTasksQueueCreate,
		DeleteContext: resourceGcpCloudTasksQueueDelete,
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

func resourceGcpCloudTasksQueueRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpCloudTasksQueueRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) < 4 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, qName := idParts[0], idParts[3]

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
	resourceGcpCloudTasksQueueSetData(d, tenantID, qName, duplo)
	log.Printf("[TRACE] resourceGcpCloudTasksQueueRead ******** end")
	return nil
}

// CREATE resource
func resourceGcpCloudTasksQueueCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpCloudTasksQueueCreate ******** start")

	// Create the request object.
	rq := expandGcpCloudTasksQueue(d)

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	// Post the object to Duplo
	err := c.GcpCloudTasksQueueCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s cloud task's queue %s : %s", tenantID, rq.QueueName, err)
	}

	id := fmt.Sprintf("%s/tasks/queue/%s", tenantID, rq.QueueName)
	d.SetId(id)

	resourceGcpCloudTasksQueueRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpCloudTasksQueueCreate ******** end")
	return nil
}

// DELETE resource
func resourceGcpCloudTasksQueueDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpCloudTasksQueueDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) < 4 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, qName := idParts[0], idParts[3]

	err := c.GCPCloudTasksQueueDelete(tenantID, qName)
	if err != nil {
		return diag.Errorf("Error deleting cloud tasks queue %s : %s", qName, err)
	}

	log.Printf("[TRACE] resourceGcpCloudTasksQueueDelete ******** end")
	return nil
}

func resourceGcpCloudTasksQueueSetData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploGCPCloudTaskQueue) {
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
