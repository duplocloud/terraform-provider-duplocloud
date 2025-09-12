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
		"primary_region": {
			Type:     schema.TypeString,
			Computed: true,
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
	if len(idParts) != 4 {
		return diag.Errorf("Invalid ID format for RDS Global Database: %s", id)
	}
	tenantId, gclusterId, _, region := idParts[0], idParts[1], idParts[2], idParts[3]
	log.Printf("[TRACE] resourceAwsRdsGlobalDatabaseRead(%s, %s, %s): start", tenantId, gclusterId, region)

	c := m.(*duplosdk.Client)

	rp, cerr := c.GetGloabalRegion(tenantId, gclusterId, region)

	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[WARN] RDS Global Database %s/%s/%s not found, removing from state", tenantId, gclusterId, region)
			d.SetId("")
			return nil
		} else {
			return diag.Errorf("Unable to fetch details of secondary region cluster %s", cerr.Error())
		}
	}

	d.Set("tenant_id", tenantId)
	d.Set("global_id", rp.GlobalInfo.GlobalClusterId)
	if strings.EqualFold(rp.Region.Role, "secondary") {
		d.Set("secondary_identifier", rp.Region.ClusterId)
		d.Set("region", rp.Region.Region)
		d.Set("secondary_tenant_id", rp.Region.TenantId)
	} else {
		d.Set("secondary_identifier", d.Get("secondary_identifier"))
		d.Set("region", d.Get("region"))
		d.Set("secondary_tenant_id", d.Get("secondary_tenant_id"))
	}
	d.Set("identifier", rp.GlobalInfo.PrimaryClusterId)
	d.Set("primary_region", rp.GlobalInfo.PrimaryRegion)
	log.Printf("[TRACE] resourceAwsRdsGlobalDatabaseRead(%s, %s, %s): end", tenantId, gclusterId, region)
	return nil
}

func resourceAwsRdsGlobalDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	secRegion := d.Get("region").(string)
	identifier := d.Get("identifier").(string)
	secTenantId := d.Get("secondary_tenant_id").(string)
	req := duplosdk.DuploRDSGlobalDatabase{
		TenantID: secTenantId,
	}
	log.Printf("[TRACE] resourceAwsRdsGlobalDatabaseCreate(%s, %s, %s, %s): start", tenantID, identifier, secRegion, secTenantId)
	c := m.(*duplosdk.Client)
	trp, cerr := c.TenantFeaturesGet(tenantID)
	if cerr != nil {
		return diag.Errorf("Failed to get tenant %s details : %s", tenantID, cerr)
	}
	rp, cerr := c.CreateRDSDBSecondaryDB(tenantID, identifier, secRegion, req)
	if cerr != nil {
		return diag.Errorf("Error creating secondary region RDS Global Database: %s", cerr)
	}
	globalCluster := rp.GlobalInfo.GlobalClusterId
	pRegion := trp.Region
	secCluster := rp.Region.ClusterId
	id := fmt.Sprintf("%s/%s/%s/%s", tenantID, globalCluster, pRegion, secRegion)
	time.Sleep(120 * time.Second)
	//err := waitUntilSecondoryDBReady(ctx, c, tenantID, rp.GlobalInfo.ClusterId, rp.Region.Region, d.Timeout("create"), true)
	//if err != nil {
	//	return diag.Errorf("Error waiting for global region RDS Global Database group to be ready: %s", err)
	//}
	err := waitUntilGlobalConfigReady(ctx, c, secTenantId, secCluster, secRegion, d.Timeout("create"), false)
	if err != nil {
		return diag.Errorf("Error waiting for secondary region RDS Global Database to be ready: %s", err)
	}
	d.SetId(id)

	diags := resourceAwsRdsGlobalDatabaseRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsRdsGlobalDatabaseCreate(%s, %s, %s, %s): end", tenantID, identifier, secRegion, secTenantId)
	return diags
}

func resourceAwsRdsGlobalDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	if len(idParts) != 4 {
		return diag.Errorf("Invalid ID format for RDS Global Database: %s", id)
	}
	tenantId, clusterId, pRegion, region := idParts[0], idParts[1], idParts[2], idParts[3]
	log.Printf("[TRACE] resourceAwsRdsGlobalDatabaseDelete(%s, %s, %s): start", tenantId, clusterId, region)

	c := m.(*duplosdk.Client)
	// Disassociate secondary cluster
	clientErr := c.DisassociateRDSDRegionalCluster(tenantId, clusterId, region)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to dissassociate secondary cluster from - (Tenant: %s,  Cluster: %s, Region: %s) : %s", tenantId, clusterId, region, clientErr)
	}
	err := waitUntilGlobalConfigReady(ctx, c, tenantId, clusterId, pRegion, d.Timeout("delete"), true)
	if err != nil {
		return diag.Errorf("Error waiting for global region RDS Global Database group to be ready: %s", err)
	}
	secTenant := d.Get("secondary_tenant_id").(string)
	secClust := d.Get("secondary_identifier").(string)

	iId := fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/%s", secTenant, secClust)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "RDS DB instance", iId, func() (interface{}, duplosdk.ClientError) {
		return c.RdsInstanceGet(id)
	})
	if diags != nil {
		return diags
	}

	// Wait for the instance to become available.
	err = rdsInstanceWaitUntilAvailable(ctx, c, iId, d.Timeout("delete"))
	if err != nil {
		return diag.Errorf("Error waiting for RDS DB instance '%s' to be available: %s", id, err)
	}

	//Remove secondary cluster
	_, cerr := c.RdsInstanceDelete(fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/%s", secTenant, secClust))
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	diags = waitForResourceToBeMissingAfterDelete(ctx, d, "Secondary Aurora RDS cluster", id, func() (interface{}, duplosdk.ClientError) {
		return c.RdsInstanceGet(id)
	})
	if diags != nil {
		return diags
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

func waitUntilGlobalConfigReady(ctx context.Context, c *duplosdk.Client, tenantID, identifier, region string, timeout time.Duration, global bool) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			log.Printf("[DEBUG] waitUntilSecondoryDBReady(%s, %s,%s) attempt", tenantID, identifier, region)

			rp, err := c.GetGloabalRegion(tenantID, identifier, region)
			status := "pending"
			if err == nil {
				if global && rp.GlobalInfo.Status == "available" {
					status = "ready"
				} else if !global && rp.Region.Status == "created" {
					status = "ready"

				} else {
					status = "pending"
				}
			} else if err.Status() == 404 {
				status = "pending"
				err = nil
			}
			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] waitUntilSecondoryDBReady(%s, %s,%s) done", tenantID, identifier, region)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
