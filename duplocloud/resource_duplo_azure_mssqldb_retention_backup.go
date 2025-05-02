package duplocloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func mssqlDBRetentionBackup() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant in which retention backup will be applied for mssql db of an mssql server",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"server_name": {
			Description: "The name of mssql server",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"database_name": {
			Description: "The name of mssql database",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"retention_backup": {
			Description:  "Specify retention backup number of days",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.IntBetween(7, 35),
		},
	}
}
func resourceMsSQLDBRetentionBackup() *schema.Resource {
	return &schema.Resource{
		Description:   "duplocloud_azure_mssqldb_retention_backup sets retention backup days to mssql db belonging to mssql server",
		ReadContext:   resourceMssqlDBRetentionBackupRead,
		CreateContext: resourceMssqlDBRetentionBackupCreateOrUpdate,
		UpdateContext: resourceMssqlDBRetentionBackupCreateOrUpdate,
		DeleteContext: resourceMssqlDBRetentionBackupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: mssqlDBRetentionBackup(),
	}
}

func resourceMssqlDBRetentionBackupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	fmt.Println("Which context got called ", ctx)
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) != 4 {
		return diag.Errorf("")
	}
	tenantID, server, db := idParts[0], idParts[2], idParts[3]
	c := m.(*duplosdk.Client)
	rp, err := c.GetMsSqlDBRetention(tenantID, server, db)
	if rp == nil {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("resourceMssqlDBRetentionBackupCreateOrUpdate could not fetch retention days for %s db belonging to server %s : %s", db, server, err.Error())
	}
	d.Set("retention_backup", rp.RetentionDays)
	d.Set("tenant_id", tenantID)
	d.Set("server_name", server)
	d.Set("database_name", db)
	return nil
}
func resourceMssqlDBRetentionBackupCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	tenantId := d.Get("tenant_id").(string)
	server := d.Get("server_name").(string)
	db := d.Get("database_name").(string)
	rq := duplosdk.DuploMyssqlDBRetention{
		RetentionDays: d.Get("retention_backup").(int),
	}
	c := m.(*duplosdk.Client)
	err := c.SetMsSqlDBRetention(tenantId, server, db, rq)
	if err != nil {
		return diag.Errorf("resourceMssqlDBRetentionBackupCreateOrUpdate could not update retention days to %s db belonging to server %s : %s", db, server, err.Error())
	}
	d.SetId(tenantId + "/retention-backup/" + server + "/" + db)

	return resourceMssqlDBRetentionBackupRead(ctx, d, m)

}

func resourceMssqlDBRetentionBackupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	return nil

}
