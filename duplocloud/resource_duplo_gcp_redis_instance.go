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

func gcpRedisInstanceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "GUID of the tenant the Redis instance will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Short name of the Redis instance. Duplo adds a prefix; retrieve the full name from `fullname`.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"memory_size_gb": {
			Description: "Redis memory size in GiB.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"display_name": {
			Description: "Optional user-provided name for the instance.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"read_replicas_enabled": {
			Description: "Optional. Enable read replica mode (can only be set during instance creation).",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"redis_configs": {
			Description: "Redis configuration parameters. See https://cloud.google.com/memorystore/docs/redis/reference/rest/v1/projects.locations.instances#Instance.FIELDS.redis_configs for supported parameters.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"redis_version": {
			Description: "Version of Redis software. Defaults to the latest supported version.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"REDIS_3_2", "REDIS_4_0", "REDIS_5_0", "REDIS_6_X",
			}, false),
		},
		"replica_count": {
			Description: "Number of replica nodes. Valid range for Standard Tier with read replicas enabled is [1-5], default is 2. For basic tier, valid value is 0, default is 0.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     0,
		},
		"auth_enabled": {
			Description: "Optional. Enable OSS Redis AUTH. Defaults to false (AUTH disabled).",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"transit_encryption_enabled": {
			Description: "Optional. Enable TLS for the Redis instance. Defaults to disabled.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"tier": {
			Description: "Service tier. Must be one of ['BASIC', 'STANDARD_HA'].",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"BASIC", "STANDARD_HA",
			}, false),
		},
		"labels": {
			Description: "Resource labels for user-provided metadata.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

func resourceRedisInstance() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_redis_instance` manages a GCP redis instance in Duplo.",

		ReadContext:   resourceGcpRedisInstanceRead,
		CreateContext: resourceGcpRedisInstanceCreate,
		UpdateContext: resourceGcpRedisInstanceUpdate,
		DeleteContext: resourceGcpRedisInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
		Schema: gcpRedisInstanceSchema(),
	}
}

func resourceGcpRedisInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpRedisInstanceRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, fullname := idParts[0], idParts[1]

	// Get the client and retrieve the Redis instance
	c := m.(*duplosdk.Client)
	duplo, err := c.RedisInstanceGet(tenantID, fullname)
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s redis instance '%s': %s", tenantID, fullname, err)
	}
	if duplo == nil {
		// Object not found, remove from state
		d.SetId("")
		return nil
	}

	// Set the resource data
	d.SetId(fmt.Sprintf("%s/%s", tenantID, fullname))
	name := d.Get("name").(string)
	resourceGcpRedisInstanceSetData(d, tenantID, name, duplo)

	log.Printf("[TRACE] resourceGcpRedisInstanceRead ******** end")
	return nil
}

// CREATE resource
func resourceGcpRedisInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpRedisInstanceCreate ******** start")

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	// Create the request object.
	rq := expandGcpRedisInstance(d)

	// Post the object to Duplo and handle errors.
	duplo, err := c.RedisInstanceCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s redis instance '%s': %s", tenantID, rq.Name, err)
	}

	// Extract the instance name from Duplo response.
	parts := strings.Split(duplo.Name, "/")
	fullName := parts[len(parts)-1]
	id := fmt.Sprintf("%s/%s", tenantID, fullName)

	// Wait for the Redis instance details to be available.
	if diags := waitForResourceToBePresentAfterCreate(ctx, d, "redis instance", id, func() (interface{}, duplosdk.ClientError) {
		return c.RedisInstanceGet(tenantID, fullName)
	}); diags != nil {
		return diags
	}

	// Set resource data.
	duplo.Name = fullName
	d.SetId(id)
	resourceGcpRedisInstanceSetData(d, tenantID, name, duplo)

	log.Printf("[TRACE] resourceGcpRedisInstanceCreate ******** end")
	return nil
}

// UPDATE resource
func resourceGcpRedisInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("[TRACE] resourceGcpRedisInstanceUpdate start")

	// Parse the identifying attributes
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 {
		return diag.Errorf("Invalid resource ID: %s", d.Id())
	}
	tenantID, fullName := idParts[0], idParts[1]
	log.Printf("[DEBUG] tenantID: %s, fullName: %s", tenantID, fullName)

	// Prepare request for update
	rq := expandGcpRedisInstance(d)
	rq.Name = fullName // Set the name for the update request
	log.Printf("[DEBUG] Redis update request: %+v", rq)

	// Update the Redis instance using the Duplo client
	client := m.(*duplosdk.Client)
	if _, err := client.RedisInstanceUpdate(tenantID, fullName, rq); err != nil {
		return diag.Errorf("Failed to update Redis instance '%s' for tenant '%s': %s", fullName, tenantID, err)
	}

	log.Println("[TRACE] resourceGcpRedisInstanceUpdate end")
	return nil
}

// DELETE resource
func resourceGcpRedisInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpRedisInstanceDelete started")

	// Delete the Redis instance using the Duplo client
	client := m.(*duplosdk.Client)
	instanceID := d.Id()
	idParts := strings.SplitN(instanceID, "/", 2)
	fullName := d.Get("name").(string)

	if err := client.RedisInstanceDelete(idParts[0], fullName); err != nil {
		return diag.Errorf("Error deleting Redis instance '%s': %s", instanceID, err)
	}

	// Wait for the Redis instance to be deleted (timeout 60 seconds)
	if diag := waitForResourceToBeMissingAfterDelete(ctx, d, "redis instance", instanceID, func() (interface{}, duplosdk.ClientError) {
		return client.GcpCloudFunctionGet(idParts[0], idParts[1])
	}); diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceGcpRedisInstanceDelete completed successfully")
	return nil
}

func resourceGcpRedisInstanceSetData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploRedisInstanceBody) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("display_name", duplo.DisplayName)
	d.Set("memory_size_gb", duplo.MemorySizeGb)
	d.Set("read_replicas_enabled", duplo.ReadReplicasEnabled)
	d.Set("redis_configs", duplo.RedisConfigs)
	d.Set("redis_version", duplo.RedisVersion)
	d.Set("replica_count", duplo.ReplicaCount)
	d.Set("auth_enabled", duplo.AuthEnabled)
	d.Set("transit_encryption_enabled", duplo.TransitEncryptionEnabled)
	d.Set("tier", duplo.Tier)
	flattenGcpLabels(d, duplo.Labels)
}

func expandGcpRedisInstance(d *schema.ResourceData) *duplosdk.DuploRedisInstanceBody {
	duplo := duplosdk.DuploRedisInstanceBody{}
	if val, ok := d.GetOk("name"); ok {
		duplo.Name = val.(string)
	}
	if val, ok := d.GetOk("memory_size_gb"); ok {
		duplo.MemorySizeGb = val.(int)
	}

	if val, ok := d.GetOk("display_name"); ok {
		duplo.DisplayName = val.(string)
	}

	if val, ok := d.GetOk("read_replicas_enabled"); ok {
		duplo.ReadReplicasEnabled = val.(bool)
	}

	if val, ok := d.Get("redis_configs").(map[string]interface{}); ok {
		redisConfigs := make(map[string]string)
		for key, value := range val {
			if strVal, ok := value.(string); ok {
				redisConfigs[key] = strVal
			}
		}
		duplo.RedisConfigs = redisConfigs
	}

	if val, ok := d.GetOk("redis_version"); ok {
		duplo.RedisVersion = val.(string)
	}

	if val, ok := d.GetOk("replica_count"); ok {
		duplo.ReplicaCount = val.(int)
	}

	if val, ok := d.GetOk("auth_enabled"); ok {
		duplo.AuthEnabled = val.(bool)
	}

	if val, ok := d.GetOk("transit_encryption_enabled"); ok {
		duplo.TransitEncryptionEnabled = val.(bool)
	}

	if val, ok := d.GetOk("tier"); ok {
		strValue := val.(string)
		duplo.Tier = 0
		if strValue == "STANDARD_HA" {
			duplo.Tier = 1
		}
	}

	return &duplo
}
