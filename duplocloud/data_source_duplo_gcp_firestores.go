package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceFirestores() *schema.Resource {
	return &schema.Resource{

		ReadContext: dataSourceGcpFirestoreList,
		Schema:      dataGcpFirestoresSchema(),
	}
}

func dataGcpFirestoresSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the firestore will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsUUID,
		},

		"firestores": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: firestoreSchema(),
			},
		},
	}
}

func firestoreSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Description: "The short name of the firestore.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"fullname": {
			Description: "The full name of the firestore.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"enable_delete_protection": {
			Description: "Delete protection prevents accidental deletion of firestore.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"enable_point_in_time_recovery": {
			Description: "Restores data to a specific moment in time, enhancing data protection and recovery capabilities.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"location_id": {
			Description: "Location for firestore",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"type": {
			Description: "Firestore type",
			Type:        schema.TypeString,
			Computed:    true,
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

func dataSourceGcpFirestoreList(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceGcpFirestoreRead ******** start")

	tenantID := d.Get("tenant_id").(string)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)

	list, err := c.FirestoreList(tenantID)
	if list == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s firestore list: %s", tenantID, err)
	}

	// Set simple fields first.
	d.SetId(tenantID)
	firestores := make([]map[string]interface{}, 0, len(*list))
	for _, duplo := range *list {
		firestores = append(firestores, setFirestoreFieldList(duplo))
	}
	_ = d.Set("firestores", firestores)
	log.Printf("[TRACE] dataSourceGcpFirestoreRead ******** end")
	return nil
}

func setFirestoreFieldList(duplo duplosdk.DuploFirestoreBody) map[string]interface{} {
	// Set simple fields first.
	return map[string]interface{}{
		"name":                          duplo.Name,
		"fullname":                      duplo.Name,
		"enable_delete_protection":      duplo.DeleteProtectionState == "DELETE_PROTECTION_ENABLED",
		"enable_point_in_time_recovery": duplo.PointInTimeRecoveryEnablement == "POINT_IN_TIME_RECOVERY_ENABLED",
		"location_id":                   duplo.LocationId,
		"etag":                          duplo.Etag,
		"uid":                           duplo.UID,
		"version_retention_period":      duplo.VersionRetentionPeriod,
		"earliest_version_time":         duplo.EarliestVersionTime,
		"concurrency_mode":              duplo.ConcurrencyMode,
		"type":                          duplo.Type,
		"app_engine_integration_mode":   duplo.AppEngineIntegrationMode,
	}
}
