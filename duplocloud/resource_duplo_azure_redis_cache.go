package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureRedisCacheSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure redis cache will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the Redis instance. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"capacity": {
			Description: "The size of the Redis cache to deploy. Valid values for a SKU `family` of C (Basic/Standard) are `0, 1, 2, 3, 4, 5, 6`, and for P (Premium) `family` are `1, 2, 3, 4`",
			Type:        schema.TypeInt,
			Required:    true,
		},

		"family": {
			Description: "The SKU family/pricing group to use. Valid values are `C` (for Basic/Standard SKU family) and `P` (for `Premium`)",
			Type:        schema.TypeString,
			Required:    true,
		},

		"sku_name": {
			Description: "The SKU of Redis to use. Possible values are `Basic`, `Standard` and `Premium`.",
			Type:        schema.TypeString,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"Basic",
				"Standard",
				"Premium",
			}, true),
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until Redis cache instance to be ready, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},

		"minimum_tls_version": {
			Description: "The minimum TLS version.",
			Type:        schema.TypeString,
			Optional:    true,
		},

		"shard_count": {
			Description: "Only available when using the Premium SKU The number of Shards to create on the Redis Cluster.",
			Type:        schema.TypeInt,
			Optional:    true,
		},

		"enable_non_ssl_port": {
			Description: "Enable the non-SSL port (6379)",
			Type:        schema.TypeBool,
			Default:     false,
			Optional:    true,
		},

		"subnet_id": {
			Description: "Only available when using the Premium SKU The ID of the Subnet within which the Redis Cache should be deployed.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"hostname": {
			Type:     schema.TypeString,
			Computed: true,
		},

		"port": {
			Type:     schema.TypeInt,
			Computed: true,
		},

		"ssl_port": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"tags": {
			Type:     schema.TypeMap,
			Computed: true,
			Elem:     schema.TypeString,
		},
		"redis_version": {
			Description:  "Redis version. Only major version needed. Valid values: `4`, `6`.",
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice([]string{"4", "6"}, false),
			DiffSuppressFunc: func(_, old, new string, d *schema.ResourceData) bool {
				n := strings.Split(old, ".")
				if len(n) >= 1 {
					newMajor := n[0]
					return new == newMajor
				}
				return false
			},
		},
	}
}

func resourceAzureRedisCache() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_redis_cache` manages an Azure redis cache in Duplo.",

		ReadContext:   resourceAzureRedisCacheRead,
		CreateContext: resourceAzureRedisCacheCreate,
		UpdateContext: resourceAzureRedisCacheUpdate,
		DeleteContext: resourceAzureRedisCacheDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureRedisCacheSchema(),
	}
}

func resourceAzureRedisCacheRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureRedisCacheIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureRedisCacheRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.RedisCacheGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure redis cache %s : %s", tenantID, name, clientErr)
	}

	flattenAzureRedisCache(d, duplo)

	log.Printf("[TRACE] resourceAzureRedisCacheRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureRedisCacheCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureRedisCacheCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAzureRedisCache(d)
	err = c.RedisCacheCreate(tenantID, name, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure redis cache '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure redis cache", id, func() (interface{}, duplosdk.ClientError) {
		return c.RedisCacheGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	//By default, wait until the cache instances to be healthy.
	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = redisCacheWaitUntilReady(ctx, c, tenantID, name, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzureRedisCacheRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureRedisCacheCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureRedisCacheUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureRedisCacheDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureRedisCacheIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureRedisCacheDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.RedisCacheDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure redis cache '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure redis cache", id, func() (interface{}, duplosdk.ClientError) {
		if rp, err := c.RedisCacheExists(tenantID, name); rp || err != nil {
			return rp, err
		}
		return nil, nil
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureRedisCacheDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAzureRedisCache(d *schema.ResourceData) *duplosdk.DuploAzureRedisCacheRequest {
	return &duplosdk.DuploAzureRedisCacheRequest{
		Properties: duplosdk.DuploAzureRedisCacheProperties{
			ShardCount:       d.Get("shard_count").(int),
			EnableNonSslPort: d.Get("enable_non_ssl_port").(bool),
			SubnetID:         d.Get("subnet_id").(string),
			Sku: duplosdk.DuploAzureRedisCacheSku{
				Name:     d.Get("sku_name").(string),
				Family:   d.Get("family").(string),
				Capacity: d.Get("capacity").(int),
			},
		},
	}
}

func parseAzureRedisCacheIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureRedisCache(d *schema.ResourceData, duplo *duplosdk.DuploAzureRedisCache) {
	d.Set("shard_count", duplo.PropertiesShardCount)
	d.Set("enable_non_ssl_port", duplo.PropertiesEnableNonSslPort)
	d.Set("subnet_id", duplo.PropertiesSubnetID)
	d.Set("hostname", duplo.PropertiesHostName)
	d.Set("port", duplo.PropertiesPort)
	d.Set("ssl_port", duplo.PropertiesSslPort)
	d.Set("redis_version", duplo.PropertiesRedisVersion)
	d.Set("tags", duplo.Tags)
}

func redisCacheWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.RedisCacheGet(tenantID, name)
			log.Printf("[TRACE] Redis cache provisioning state is (%s).", rp.PropertiesProvisioningState)
			status := "pending"
			if err == nil {
				if rp.PropertiesProvisioningState == "Succeeded" {
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
	log.Printf("[DEBUG] redisCacheWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
