package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceRedisInstance() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGcpRedisInstanceRead,
		Schema:      dataGcpRedisInstanceSchema(),
	}
}

func dataGcpRedisInstanceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the redis instance will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the redis instance.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"fullname": {
			Description: "The full name of the of the Redis instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"memory_size_gb": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: `Redis memory size in GiB.`,
		},
		"display_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: `An arbitrary and optional user-provided name for the instance.`,
		},
		"read_replicas_enabled": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: `Optional. Read replica mode. Can only be specified when trying to create the instance.`,
		},
		"redis_configs": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: `Redis configuration parameters, according to http://redis.io/topics/config. Please check Memorystore documentation for the list of supported parameters: https://cloud.google.com/memorystore/docs/redis/reference/rest/v1/projects.locations.instances#Instance.FIELDS.redis_configs`,
		},
		"redis_version": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: `The version of Redis software. If not provided, latest supported version will be used. Please check the API documentation linked at the top for the latest valid values.`,
		},
		"replica_count": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: `The number of replica nodes. The valid range for the Standard Tier with read replicas enabled is [1-5] and defaults to 2. If read replicas are not enabled for a Standard Tier instance, the only valid value is 1 and the default is 1. The valid value for basic tier is 0 and the default is also 0.`,
		},
		"auth_enabled": {
			Type:        schema.TypeBool,
			Description: `Indicates whether OSS Redis AUTH is enabled for the instance. If set to "true" AUTH is enabled on the instance. Default value is "false" meaning AUTH is disabled.`,
			Computed:    true,
		},
		"transit_encryption_enabled": {
			Type:        schema.TypeBool,
			Description: `The TLS mode of the Redis instance, If not provided, TLS is disabled for the instance.`,
			Computed:    true,
		},
		"tier": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: `The service tier of the instance. Must be one of these values: BASIC: standalone instance or STANDARD_HA: highly available primary/replica instances Default value: "BASIC" Possible values: ["BASIC", "STANDARD_HA"]`,
		},
		"labels": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: `Resource labels to represent user provided metadata.`,
		},
	}
}

func dataSourceGcpRedisInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceGcpRedisInstanceRead ******** start")

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	// Validate the required name parameter
	if name == "" {
		return diag.Errorf("error fetching detail: name is required")
	}

	// Initialize the Duplo client
	c := m.(*duplosdk.Client)

	// Fetch the Redis instance from Duplo
	duplo, err := c.RedisInstanceGet(tenantID, name)
	if err != nil {
		return diag.Errorf("unable to retrieve Redis instance '%s' for tenant '%s': %s", name, tenantID, err)
	}

	// Handle missing object case
	if duplo == nil {
		d.SetId("") // Redis instance is missing
		return nil
	}

	// Set the resource data and ID
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))
	resourceGcpRedisInstanceSetData(d, tenantID, name, duplo)

	log.Printf("[TRACE] dataSourceGcpRedisInstanceRead ******** end")
	return nil
}
