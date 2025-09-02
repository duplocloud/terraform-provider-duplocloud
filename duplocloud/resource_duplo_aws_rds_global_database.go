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

func duploAwsRdsGlobalDatabaseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the RDS tag will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"identifier": {
			Description: "The identifier of the primary Database.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
		},
		"secondary_tenant_id": {
			Description:  "The GUID of the tenant that the secondary RDS Global Database will be created in.",
			Type:         schema.TypeString,
			ForceNew:     true,
			Required:     true,
			ValidateFunc: validation.IsUUID,
		},
		"global_id": {
			Description: "The identifier of the Global Database.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"secondary_identifier": {
			Description: "The identifier of the secondary Database.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"region": {
			Description: "The region of the secondary Database.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
		},
	}
}

func resourceAwsRdsGlobalDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_rds_gloabal_database` manages secondary AWS RDS Global Database in Duplo.",

		ReadContext:   resourceAwsRdsGlobalDatabaseRead,
		CreateContext: resourceAwsRdsGlobalDatabaseCreate,
		//UpdateContext: resourceAwsRdsGlobalDatabaseCreate,
		DeleteContext: resourceAwsRdsGlobalDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: duploAwsRdsGlobalDatabaseSchema(),
		//CustomizeDiff: validateGlobalDatabaseParameters,
	}
}

func resourceAwsRdsGlobalDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) != 3 {
		return diag.Errorf("Invalid ID format for RDS Global Database: %s", id)
	}
	tenantId, gclusterId, region := idParts[0], idParts[1], idParts[2]
	log.Printf("[TRACE] resourceAwsRdsGlobalDatabaseRead(%s, %s, %s): start", tenantId, gclusterId, region)

	c := m.(*duplosdk.Client)

	rp, cerr := c.GetGloabalRegion(tenantId, gclusterId, region)

	if cerr != nil {
		if cerr.Status() == 404 {
			d.SetId("")
			return diag.Errorf("%s", cerr.Error())
		}
		return diag.Errorf("Unable to fetch details of secondary region cluster %s", cerr.Error())
	}

	d.Set("tenant_id", tenantId)
	d.Set("secondary_tenant_id", rp.Region.TenantId)
	d.Set("global_id", rp.GlobalInfo.GlobalClusterId)
	d.Set("secondary_identifier", rp.Region.ClusterId)
	d.Set("region", rp.Region.Region)
	d.Set("identifier", rp.GlobalInfo.PrimaryClusterId)
	log.Printf("[TRACE] resourceAwsRdsGlobalDatabaseRead(%s, %s, %s): end", tenantId, gclusterId, region)
	return nil
}

func resourceAwsRdsGlobalDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	region := d.Get("region").(string)
	identifier := d.Get("identifier").(string)
	secTenantId := d.Get("secondary_tenant_id").(string)
	req := duplosdk.DuploRDSGlobalDatabase{
		TenantID: secTenantId,
	}
	log.Printf("[TRACE] resourceAwsRdsGlobalDatabaseCreate(%s, %s, %s, %s): start", tenantID, identifier, region, secTenantId)
	c := m.(*duplosdk.Client)

	rp, cerr := c.CreateRDSDBSecondaryDB(tenantID, identifier, region, req)
	if cerr != nil {
		return diag.Errorf("Error creating secondary region RDS Global Database: %s", cerr)
	}
	id := fmt.Sprintf("%s/%s/%s", tenantID, rp.GlobalInfo.GlobalClusterId, rp.Region.Region)
	time.Sleep(120 * time.Second)
	//err := waitUntilSecondoryDBReady(ctx, c, tenantID, rp.GlobalInfo.ClusterId, rp.Region.Region, d.Timeout("create"), true)
	//if err != nil {
	//	return diag.Errorf("Error waiting for global region RDS Global Database group to be ready: %s", err)
	//}
	err := waitUntilSecondoryDBReady(ctx, c, secTenantId, rp.Region.ClusterId, rp.Region.Region, d.Timeout("create"), false)
	if err != nil {
		return diag.Errorf("Error waiting for secondary region RDS Global Database to be ready: %s", err)
	}
	d.SetId(id)

	diags := resourceAwsRdsGlobalDatabaseRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsRdsGlobalDatabaseCreate(%s, %s, %s, %s): end", tenantID, identifier, region, secTenantId)
	return diags
}

func resourceAwsRdsGlobalDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) != 3 {
		return diag.Errorf("Invalid ID format for RDS Global Database: %s", id)
	}
	tenantId, clusterId, region := idParts[0], idParts[1], idParts[2]
	log.Printf("[TRACE] resourceAwsRdsGlobalDatabaseDelete(%s, %s, %s): start", tenantId, clusterId, region)

	c := m.(*duplosdk.Client)
	clientErr := c.DeleteRDSDBSecondaryDB(tenantId, clusterId, region)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete rds tag - (Tenant: %s,  Cluster: %s, Region: %s) : %s", tenantId, clusterId, region, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "RDS Secondary region cluster", id, func() (interface{}, duplosdk.ClientError) {
		return c.GetGloabalRegion(tenantId, clusterId, region)

	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsRdsGlobalDatabaseDelete(%s, %s, %s): end", tenantId, clusterId, region)
	return nil
}

/*
func validateGlobalDatabaseParameters(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	c := &duplosdk.Client{}
	tenantId := diff.Get("tenant_id").(string)
	region := diff.Get("region").(string)
	identifier := diff.Get("identifier").(string)
	rp, err := c.GetGloabalRegions(tenantId, identifier)
	if err != nil {
		return fmt.Errorf("error retrieving global regions for identifier %s: %s", identifier, err)
	}
	if rp == nil {
		return fmt.Errorf("no global regions found for identifier %s", identifier)
	}
	for _, rg := range rp.GlobalInfo.Regions {
		if !strings.EqualFold(rg, region) {
			return fmt.Errorf("region %s is not a valid global region for identifier %s, valid regions are: %v", region, identifier, rp.GlobalInfo.Regions)
		}
		if strings.EqualFold(rp.GlobalInfo.PrimaryRegion, rg) {
			return fmt.Errorf("region %s is the primary region for identifier %s, please use a secondary region", region, identifier)
		}
	}
	return nil
}*/

func waitUntilSecondoryDBReady(ctx context.Context, c *duplosdk.Client, tenantID, identifier, region string, timeout time.Duration, global bool) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.GetGloabalRegion(tenantID, identifier, region)
			status := "pending"
			if err == nil {
				if global && rp.GlobalInfo.Status == "available" {
					status = "ready"
				} else if !global && rp.Region.Status == "available" {
					status = "ready"

				} else {
					status = "pending"
				}
			} else if strings.Contains(err.Error(), "not found") {
				status = "pending"
				err = nil
			}
			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] waitUntilSecondoryDBReady(%s, %s,%s)", tenantID, identifier, region)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
