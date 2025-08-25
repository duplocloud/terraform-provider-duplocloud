package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ecacheReplicationGroupSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the elasticache instance datastore has been created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"global_datastore_id": {
			Description: "Specify the global datastore name with which the secondary regional cluster should be associated.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"description": {
			Description:      "The description for secondary cluster",
			Type:             schema.TypeString,
			Computed:         true,
			Optional:         true,
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
		"secondary_tenant_id": {
			Description:  "The tenant_id where secondary cluster need to be created. **NOTE** The tenant_id must belong to a region different from that of the primary cluster.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"port": {
			Description: "Specify port for secondary cluster",
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
		},
		"secondary_cluster_name": {
			Description: "The name of the elasticache instance that need to be created as secondary regional cluster.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
	}
}

// SCHEMA for resource crud
func resourceDuploEcacheReplicationGroup() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_ecache_global_secondary_cluster` used to manage Amazon ElastiCache global datastore's secondary regional cluster" +
			"<p>This resource currently supports, define and manage only Redis based regional cluster" +
			"</p>",

		ReadContext:   resourceDuploEcacheReplicationGroupRead,
		CreateContext: resourceDuploEcacheReplicationGroupCreate,
		DeleteContext: resourceDuploEcacheReplicationGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(29 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: ecacheReplicationGroupSchema(),
	}
}

// CREATE resource
func resourceDuploEcacheReplicationGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	tenantID := d.Get("tenant_id").(string)
	rq := duplosdk.DuploEcacheReplicationGroup{
		Description:              d.Get("description").(string),
		SecondaryTenantId:        d.Get("secondary_tenant_id").(string),
		GlobalReplicationGroupId: d.Get("global_datastore_id").(string),
		Port:                     d.Get("port").(int),
		ReplicationGroupId:       d.Get("secondary_cluster_name").(string),
	}
	log.Printf("[TRACE] resourceDuploEcacheReplicationGroupCreate(%s): start", tenantID)

	c := m.(*duplosdk.Client)
	rp, cerr := c.DuploEcacheReplicationGroupCreate(tenantID, &rq)
	if cerr != nil {
		return diag.Errorf("%s", cerr)
	}
	id := fmt.Sprintf("%s/ecacheReplicationGroup/%s/%s/%s", tenantID, rq.SecondaryTenantId, rq.GlobalReplicationGroupId, rq.ReplicationGroupId)
	err := replicationGroupWaitUntilAvailable(ctx, c, tenantID, rq.GlobalReplicationGroupId, rq.ReplicationGroupId)
	if err != nil {
		return diag.Errorf("%s", cerr)

	}
	d.SetId(id)

	diags := resourceDuploEcacheReplicationGroupRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploEcacheReplicationGroupCreate(%s, %s): end", tenantID, rp.GlobalReplicationGroup.GlobalReplicationGroupId)
	return diags
}

func replicationGroupWaitUntilAvailable(ctx context.Context, c *duplosdk.Client, tenantID, gDatastore, name string) error {
	stateConf := &retry.StateChangeConf{
		Pending:      []string{"pending"},
		Target:       []string{"ready"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      20 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			resp, err := c.DuploEcacheReplicationGroupGet(tenantID, gDatastore, name)
			status := "pending"
			if resp.Status == "available" {
				status = "ready"
			}

			return resp, status, err
		},
	}
	log.Printf("[DEBUG] replicationGroupWaitUntilAvailable (%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func resourceDuploEcacheReplicationGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) != 5 {
		return diag.Errorf("invalid resource id")
	}
	tenantID, secTenantId, globalDatastore, name := idParts[0], idParts[2], idParts[3], idParts[4]
	log.Printf("[TRACE] resourceDuploEcacheReplicationGroupRead(%s,%s,%s, %s): start", tenantID, secTenantId, globalDatastore, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.DuploEcacheReplicationGroupGet(tenantID, globalDatastore, name)
	if err != nil {
		if err.Status() == 404 {
			log.Printf("Unable to fetch Ecache Replication Group")
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if duplo == nil {
		d.SetId("")
		return nil
	}

	// Convert the object into Terraform resource data
	d.Set("tenant_id", tenantID)
	d.Set("description", duplo.GlobalReplicationGroupDescription)
	d.Set("secondary_cluster_name", name)
	d.Set("global_datastore_id", duplo.GlobalReplicationGroupId)
	log.Printf("[TRACE] resourceDuploEcacheReplicationGroupRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceDuploEcacheReplicationGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) != 5 {
		return diag.Errorf("invalid resource id %s", id)
	}
	tenantID, secTenantId, globalDatastore, name := idParts[0], idParts[2], idParts[3], idParts[4]
	c := m.(*duplosdk.Client)
	cerr := c.DuploEcacheReplicationGroupDisassociate(tenantID, secTenantId, globalDatastore, name)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("Unable to disassociate Replication group from Datastore %s", cerr.Error())
			return nil
		}
		return diag.FromErr(cerr)
	}
	time.Sleep(2 * time.Minute)
	err := replicationGroupWaitUntilAvailable(ctx, c, tenantID, globalDatastore, name)
	if err != nil {
		return diag.Errorf("%s", cerr)

	}
	err = c.EcacheInstanceDelete(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait up to 60 seconds for Duplo to show the object as deleted.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "Secondary ecache instance", id, func() (interface{}, duplosdk.ClientError) {
		return c.EcacheInstanceGet(secTenantId, name)
	})

	// Wait for some time to deal with consistency issues.
	if diag == nil {
		time.Sleep(time.Duration(90) * time.Second)
	}

	return nil
}
