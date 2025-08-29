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

func ecacheGlobalDatastoreSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the global datastore  will be created in.",
			Type:         schema.TypeString,
			Optional:     false,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"global_replication_group_name": {
			Description: "Specify name of global datastore",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"description": {
			Description:      "Specify descritpion for global datastore",
			Type:             schema.TypeString,
			Computed:         true,
			Optional:         true,
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
		"primary_instance_name": {
			Description: "Specify primary instance name",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// SCHEMA for resource crud
func resourceDuploEcacheGlobalDatastore() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_ecache_global_datastore` used to manage Amazon ElastiCache Global Datastore within a DuploCloud-managed environment. " +
			"<p>This resource currently supports, define and manage only Redis based global datastore" +
			"</p>",

		ReadContext:   resourceDuploEcacheGlobalDatastoreRead,
		CreateContext: resourceDuploEcacheGlobalDatastoreCreate,
		DeleteContext: resourceDuploEcacheGlobalDatastoreDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(29 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: ecacheGlobalDatastoreSchema(),
	}
}

// CREATE resource
func resourceDuploEcacheGlobalDatastoreCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	tenantID := d.Get("tenant_id").(string)
	rq := duplosdk.DuploEcacheGlobalDatastore{
		Description:              d.Get("description").(string),
		PrimaryInstance:          d.Get("primary_instance_name").(string),
		GlobalReplicationGroupId: d.Get("global_replication_group_name").(string),
	}
	log.Printf("[TRACE] resourceDuploEcacheGlobalDatastoreCreate(%s): start", tenantID)

	c := m.(*duplosdk.Client)
	rp, cerr := c.DuploEcacheGlobalDatastoreCreate(tenantID, &rq)
	if cerr != nil {
		return diag.Errorf("DuploEcacheGlobalDatastoreCreate failed to create global datastore %s", cerr)
	}
	id := fmt.Sprintf("%s/ecacheGlobalDatastore/%s", tenantID, rp.GlobalReplicationGroup.GlobalReplicationGroupId)
	err := globalDatastoreWaitUntilAvailable(ctx, c, tenantID, rp.GlobalReplicationGroup.GlobalReplicationGroupId)
	if err != nil {
		return diag.Errorf("globalDatastoreWaitUntilAvailable Waiting for global datastore to be available failed %s", cerr)

	}
	d.SetId(id)

	diags := resourceDuploEcacheGlobalDatastoreRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploEcacheInstanceCreate(%s, %s): end", tenantID, rp.GlobalReplicationGroup.GlobalReplicationGroupId)
	return diags
}

func globalDatastoreWaitUntilAvailable(ctx context.Context, c *duplosdk.Client, tenantID, name string) error {
	stateConf := &retry.StateChangeConf{
		Pending:      []string{"pending"},
		Target:       []string{"ready"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      20 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			resp, err := c.DuploEcacheGlobalDatastoreGet(tenantID, name)
			status := "pending"
			if resp != nil && resp.Status == "primary-only" {
				status = "ready"
			}

			return resp, status, err
		},
	}
	log.Printf("[DEBUG] globalDatastoreWaitUntilAvailable (%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func resourceDuploEcacheGlobalDatastoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) != 3 {
		return diag.Errorf("invalid resource id %s", id)
	}
	tenantID, name := idParts[0], idParts[2]
	log.Printf("[TRACE] resourceDuploEcacheGlobalDatastoreRead(%s, %s): start", tenantID, name)
	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.DuploEcacheGlobalDatastoreGet(tenantID, name)
	if err != nil {
		if err.Status() == 404 {
			log.Printf("Unable to fetch Ecache Global Datastore")
			d.SetId("")
			return diag.Errorf("DuploEcacheGlobalDatastoreGet Unable to fetch Ecache Global Datastore %s", err)
		}
		return diag.FromErr(err)
	}
	if duplo == nil {
		d.SetId("")
		return nil
	}
	sn := strings.Split(name, "duplo-")
	// Convert the object into Terraform resource data
	d.Set("tenant_id", tenantID)
	d.Set("global_replication_group_name", sn[1])
	d.Set("fullname", duplo.GlobalReplicationGroupId)
	d.Set("description", duplo.GlobalReplicationGroupDescription)
	for _, member := range duplo.Members {
		if member.Role == "PRIMARY" {
			d.Set("primary_instance_name", member.ReplicationGroupId)
		}
	}
	log.Printf("[TRACE] resourceDuploEcacheGlobalDatastoreRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceDuploEcacheGlobalDatastoreDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) != 3 {
		return diag.Errorf("invalid resource id %s", id)
	}
	tenantID, name := idParts[0], idParts[2]
	log.Printf("[TRACE] resourceDuploEcacheGlobalDatastoreDelete(%s, %s): start", tenantID, name)
	fullName := d.Get("fullname").(string)
	c := m.(*duplosdk.Client)
	cerr := c.DuploEcacheGlobalDatastoreDelete(tenantID, fullName)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("Unable to delete Ecache Global Datastore %s", cerr.Error())
			return nil
		}
		return diag.FromErr(cerr)
	}
	err := globalDatastoreWaitUntilUnAvailable(ctx, c, tenantID, fullName)
	if err != nil {
		return diag.Errorf("Unable to delete secondary redis cluster after disassociation %s", cerr)

	}
	time.Sleep(10 * time.Minute)
	log.Printf("[TRACE] resourceDuploEcacheGlobalDatastoreDelete(%s, %s): end", tenantID, name)

	return nil
}

func globalDatastoreWaitUntilUnAvailable(ctx context.Context, c *duplosdk.Client, tenantID, name string) error {
	stateConf := &retry.StateChangeConf{
		Pending:      []string{"pending"},
		Target:       []string{"ready"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      20 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			resp, err := c.DuploEcacheGlobalDatastoreGet(tenantID, name)
			status := "pending"
			if resp == nil {
				status = "ready"
			}

			return resp, status, err
		},
	}
	log.Printf("[DEBUG] globalDatastoreWaitUntilAvailable (%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
