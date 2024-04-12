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

func gcpFirestoreSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the firestore will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the firestore.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the firestore.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"enable_delete_protection": {
			Description: "Delete protection prevents accidental deletion of firestore.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"enable_point_in_time_recovery": {
			Description: "Restores data to a specific moment in time, enhancing data protection and recovery capabilities.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"location_id": {
			Description: "Location for firestore",
			Type:        schema.TypeString,
			Required:    true,
		},
		"type": {
			Description:  "Firestore type",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validateFirestoreOrDatastoreMode,
		},
		"etag": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"uid": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"version_retention_period": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"earliest_version_time": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"concurrency_mode": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"app_engine_integration_mode": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceFirestore() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_firestore` manages a GCP firestore in Duplo.",

		ReadContext:   resourceGcpFirestoreRead,
		CreateContext: resourceGcpFirestoreCreate,
		UpdateContext: resourceGcpFirestoreUpdate,
		DeleteContext: resourceGcpFirestoreDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
		Schema: gcpFirestoreSchema(),
	}
}

func resourceGcpFirestoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpFirestoreRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, fullname := idParts[0], idParts[1]
	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)

	duplo, err := c.FirestoreGet(tenantID, fullname)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s firestore '%s': %s", tenantID, fullname, err)
	}
	name := d.Get("name").(string)
	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/%s", tenantID, fullname))
	resourceGcpFirestoreSetData(d, tenantID, name, duplo)
	log.Printf("[TRACE] resourceGcpFirestoreRead ******** end")
	return nil
}

// CREATE resource
func resourceGcpFirestoreCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpFirestoreCreate ******** start")

	// Create the request object.
	rq := expandGcpFirestore(d)

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	duplo, err := c.FirestoreCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s firestore '%s': %s", tenantID, rq.Name, err)
	}
	strSlice := strings.Split(duplo.Name, "/")
	fullName := strSlice[len(strSlice)-1]
	id := fmt.Sprintf("%s/%s", tenantID, fullName)
	name := d.Get("name").(string)
	// Wait for Duplo to be able to return the firestore details.
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "firestore", id, func() (interface{}, duplosdk.ClientError) {
		return c.FirestoreGet(tenantID, fullName)
	})
	if diags != nil {
		return diags
	}
	duplo.Name = fullName
	d.SetId(id)
	resourceGcpFirestoreSetData(d, tenantID, name, duplo)

	//resourceGcpFirestoreRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpFirestoreCreate ******** end")
	return diags
}

// UPDATE resource
func resourceGcpFirestoreUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpFirestoreUpdate ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, _ := idParts[0], idParts[1]

	// Create the request object.
	rq := expandGcpFirestore(d)
	rq.Name = d.Get("fullname").(string)

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	_, err := c.FirestoreUpdate(tenantID, rq.Name, rq)
	if err != nil {
		return diag.Errorf("Error updating tenant %s firesote '%s': %s", tenantID, rq.Name, err)
	}
	resourceGcpCloudFunctionRead(ctx, d, m)

	log.Printf("[TRACE] resourceGcpFirestoreUpdate ******** end")
	return nil
}

// DELETE resource
func resourceGcpFirestoreDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpFirestoreDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	fullName := d.Get("fullname").(string)
	err := c.FirestoreDelete(idParts[0], fullName)
	if err != nil {
		return diag.Errorf("Error deleting firestore '%s': %s", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the firestore.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "firestore", id, func() (interface{}, duplosdk.ClientError) {
		return c.GcpCloudFunctionGet(idParts[0], idParts[1])
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceGcpFirestoreDelete ******** end")
	return nil
}

func resourceGcpFirestoreSetData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploFirestoreBody) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("enable_delete_protection", duplo.DeleteProtectionState == "DELETE_PROTECTION_ENABLED")
	d.Set("enable_point_in_time_recovery", duplo.PointInTimeRecoveryEnablement == "POINT_IN_TIME_RECOVERY_ENABLED")
	d.Set("location_id", duplo.LocationId)
	d.Set("etag", duplo.Etag)
	d.Set("uid", duplo.UID)
	d.Set("version_retention_period", duplo.VersionRetentionPeriod)
	d.Set("earliest_version_time", duplo.EarliestVersionTime)
	d.Set("concurrency_mode", duplo.ConcurrencyMode)
	d.Set("type", duplo.Type)
	d.Set("app_engine_integration_mode", duplo.AppEngineIntegrationMode)

}

func expandGcpFirestore(d *schema.ResourceData) *duplosdk.DuploFirestoreBody {
	duplo := duplosdk.DuploFirestoreBody{}

	if val, ok := d.GetOk("name"); ok {
		duplo.Name = val.(string)
	}
	if val, ok := d.GetOk("location_id"); ok {
		duplo.LocationId = val.(string)
	}
	if val, ok := d.GetOk("type"); ok {
		duplo.Type = val.(string)
	}
	duplo.DeleteProtectionState = "DELETE_PROTECTION_DISABLED"

	if val, ok := d.GetOk("enable_delete_protection"); ok {
		if val.(bool) {
			duplo.DeleteProtectionState = "DELETE_PROTECTION_ENABLED"
		}
	}
	duplo.PointInTimeRecoveryEnablement = "POINT_IN_TIME_RECOVERY_DISABLED"

	if val, ok := d.GetOk("enable_point_in_time_recovery"); ok {
		if val.(bool) {
			duplo.PointInTimeRecoveryEnablement = "POINT_IN_TIME_RECOVERY_DISABLED_ENABLED"
		}
	}

	return &duplo
}

func validateFirestoreOrDatastoreMode(value interface{}, key string) (warns []string, errs []error) {
	// Convert the input value to a string
	strValue := value.(string)

	// Check if the input value is either "FIRESTORE_NATIVE" or "DATASTORE_MODE"
	if strValue != "FIRESTORE_NATIVE" && strValue != "DATASTORE_MODE" {
		errs = append(errs, fmt.Errorf("%q must be either 'FIRESTORE_NATIVE' or 'DATASTORE_MODE'", key))
	}
	return
}
