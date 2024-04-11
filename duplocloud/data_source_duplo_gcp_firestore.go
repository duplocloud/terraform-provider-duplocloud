package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceFirestore() *schema.Resource {
	return &schema.Resource{

		ReadContext: dataSourceGcpFirestoreRead,
		Schema:      dataGcpFirestoreSchema(),
	}
}

func dataGcpFirestoreSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the firestore will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the firestore.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
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

func dataSourceGcpFirestoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceGcpFirestoreRead ******** start")

	tenantID := d.Get("tenant_id").(string)

	name := d.Get("name").(string)
	if name == "" {
		return diag.Errorf("error fetching detail name required ")

	}
	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)

	duplo, err := c.FirestoreGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s firestore '%s': %s", tenantID, name, err)
	}

	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))
	resourceGcpFirestoreSetData(d, tenantID, name, duplo)
	log.Printf("[TRACE] dataSourceGcpFirestoreRead ******** end")
	return nil
}
