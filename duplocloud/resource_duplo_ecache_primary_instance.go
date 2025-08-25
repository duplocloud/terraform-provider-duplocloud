package duplocloud

/*
import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ecacheGlobalPrimaryInstanceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the elasticache instance will be created in.",
			Type:         schema.TypeString,
			Optional:     false,
			Required:     true,
			ForceNew:     true, //switch tenant
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the elasticache instance.  Duplo will add a prefix to the name.  You can retrieve the full name from the `identifier` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 50-MAX_DUPLO_LENGTH),
				validation.StringMatch(regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`), "Invalid AWS Elasticache cluster name"),
				validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "AWS Elasticache cluster names cannot end with a hyphen"),
				validation.StringNotInSlice([]string{"--"}, true),
			),
		},
		"identifier": {
			Description: "The full name of the elasticache instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"port": {
			Description: "The listening port of the elasticache instance.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"cache_type": {
			Description: "The numerical index of elasticache instance type.\n" +
				"Should be :\n\n" +
				"   - `0` : Redis\n",
			Type:         schema.TypeInt,
			Optional:     true,
			ForceNew:     true,
			Default:      0,
			ValidateFunc: validation.IntBetween(0, 2),
		},
		"engine_version": {
			Description: "The engine version of the elastic instance.\n" +
				"See AWS documentation for the [available Redis and Valkey instance types](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/supported-engine-versions.html) " +
				"or the [available Memcached instance types](https://docs.aws.amazon.com/AmazonElastiCache/latest/mem-ug/supported-engine-versions-mc.html).",
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Computed: true,
			//DiffSuppressFunc: suppressEnginePatchVersion,
		},
		"actual_engine_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"size": {
			Description: "The instance type of the elasticache instance.\n" +
				"See AWS documentation for the [available instance types](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/CacheNodes.SupportedTypes.html).",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^cache\.`), "Elasticache instance types must start with 'cache.'"),
		},
		"replicas": {
			Description:  "The number of replicas to create. Supported number of replicas is 1 to 6",
			Type:         schema.TypeInt,
			Optional:     true,
			ForceNew:     true,
			Default:      1,
			ValidateFunc: validation.IntBetween(1, 6),
		},
		"automatic_failover_enabled": {
			Description: "Enables automatic failover.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
			Default:     false,
		},
		"parameter_group_name": {
			Description: "The REDIS/Valkey parameter group to supply.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
		},
		"enable_cluster_mode": {
			Description: "Flag to enable/disable redis/valkey cluster mode.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
		},
		"number_of_shards": {
			Description:      "The number of shards to create. Applicable only if enable_cluster_mode is set to true",
			Type:             schema.TypeInt,
			Optional:         true,
			DiffSuppressFunc: suppressNoOfShardsDiff,
			Computed:         true,
			ValidateFunc:     validation.IntBetween(1, 500),
		},
		"is_primary": {
			Description: "Flag to indicate if this is primary replication group",
			Type:        schema.TypeBool,
			Computed:    true,
		},
	}
}

// SCHEMA for resource crud
func resourceDuploEcacheGlobalPrimaryInstance() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_ecache_instance` used to manage Amazon ElastiCache instances within a DuploCloud-managed environment. " +
			"<p>This resource allows you to define and manage Redis or Memcached instances on AWS through Terraform, with DuploCloud handling the underlying infrastructure and integration aspects." +
			"</p>",

		ReadContext:   resourceDuploEcacheInstanceRead,
		CreateContext: resourceDuploEcacheGlobalPrimaryInstanceCreate,
		DeleteContext: resourceDuploEcacheInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(29 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema:        ecacheInstanceSchema(),
		CustomizeDiff: validateEcacheParameters,
	}
}

// CREATE resource
func resourceDuploEcacheGlobalPrimaryInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceDuploEcacheInstanceCreate(%s): start", tenantID)

	duplo, diagErr := expandEcacheGlobalPrimaryInstance(d)
	if diagErr != nil {
		return diagErr
	}
	c := m.(*duplosdk.Client)

	fullName := "duplo-" + duplo.Name
	if !validateStringLength(fullName, TOTALECACHENAMELENGTH) {
		return diag.Errorf("resourceDuploEcacheInstanceCreate: fullname %s exceeds allowable ecache name length %d)", fullName, TOTALECACHENAMELENGTH)

	}

	duplo.Identifier = duplo.Name
	id := fmt.Sprintf("v2/subscriptions/%s/ECacheDBGlobalPrimaryInstance/%s", tenantID, duplo.Name)

	// Perform additional validation.

	if duplo.Replicas < 2 && duplo.AutomaticFailoverEnabled {
		return diag.Errorf("Invalid automatic_failover_enabled '%s': To enable automatic failover, replicas must be 2 or more", id)
	}

	// Post the object to Duplo
	err = c.DuploPrimaryEcacheCreate(tenantID, duplo)
	if err != nil {
		return diag.Errorf("Error updating ECache instance '%s': %s", id, err)
	}
	d.SetId(id)

	// Wait up to 60 seconds for Duplo to be able to return the instance details.
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "ECache Global Primary instance", id, func() (interface{}, duplosdk.ClientError) {
		return c.EcacheInstanceGet(tenantID, duplo.Name)
	})
	if diags != nil {
		return diags
	}

	// Wait for the instance to become available.
	err = ecacheInstanceWaitUntilAvailable(ctx, c, tenantID, duplo.Name)
	if err != nil {
		return diag.Errorf("Error waiting for ECache instance '%s' to be available: %s", id, err)
	}

	// Read the resource state
	diags = resourceDuploEcacheInstanceRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploEcacheInstanceCreate(%s, %s): end", tenantID, duplo.Name)
	return diags
}

// expand Ecache Instance converts resource data respresenting an ECache instance to a Duplo SDK object.
func expandEcacheGlobalPrimaryInstance(d *schema.ResourceData) (*duplosdk.DuploEcacheGlobalPrimaryInstance, diag.Diagnostics) {
	data := &duplosdk.DuploEcacheGlobalPrimaryInstance{
		Name:                     d.Get("name").(string),
		Identifier:               d.Get("identifier").(string),
		CacheType:                d.Get("cache_type").(int),
		EngineVersion:            d.Get("engine_version").(string),
		Size:                     d.Get("size").(string),
		Replicas:                 d.Get("replicas").(int),
		ParameterGroupName:       d.Get("parameter_group_name").(string),
		AutomaticFailoverEnabled: d.Get("automatic_failover_enabled").(bool),
		EnableClusterMode:        d.Get("enable_cluster_mode").(bool),
		IsGlobal:                 true,
	}

	if data.EnableClusterMode {
		data.AutomaticFailoverEnabled = true
		if v, ok := d.GetOk("number_of_shards"); !ok || (v.(int) < 1 && v.(int) > 500) {
			data.NumberOfShards = 2
		} else {
			data.NumberOfShards = v.(int) //number of shards accepted if cluster mode is enabled
		}
	}
	return data, nil
}

// flattenEcacheInstance converts a Duplo SDK object respresenting an ECache instance to terraform resource data.
*/
