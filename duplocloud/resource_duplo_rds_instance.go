package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DuploRdsInstanceSchema returns a Terraform resource schema for an RDS instance.
func rdsInstanceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description: "The GUID of the tenant that the RDS instance will be created in.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, //switch tenant
		},
		"name": {
			Description: "The short name of the RDS instance.  Duplo will add a prefix to the name.  You can retrieve the full name from the `identifier` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"identifier": {
			Description: "The full name of the RDS instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"arn": {
			Description: "The ARN of the RDS instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"endpoint": {
			Description: "The endpoint of the RDS instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"host": {
			Description: "The DNS hostname of the RDS instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"port": {
			Description: "The listening port of the RDS instance.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"master_username": {
			Description: "The master username of the RDS instance.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"master_password": {
			Description: "The master password of the RDS instance.",
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
		},
		"engine": {
			Description: "The numerical index of database engine to use the for the RDS instance.\n" +
				"Should be one of:\n\n" +
				"   - `0` : MySQL\n" +
				"   - `1` : PostgreSQL\n" +
				"   - `2` : MsftSQL-Express\n" +
				"   - `3` : MsftSQL-Standard\n" +
				"   - `8` : Aurora-MySQL\n" +
				"   - `9` : Aurora-PostgreSQL\n" +
				"   - `10` : MsftSQL-Web\n" +
				"   - `11` : Aurora-Serverless-MySql\n" +
				"   - `12` : Aurora-Serverless-PostgreSql\n" +
				"   - `13` : DocumentDB\n",
			Type:     schema.TypeInt,
			Required: true,
			ForceNew: true,
		},
		"engine_version": {
			Description: "The database engine version to use the for the RDS instance.\n" +
				"If you don't know the available engine versions for your RDS instance, you can use the [AWS CLI](https://docs.aws.amazon.com/cli/latest/reference/rds/describe-db-engine-versions.html) to retrieve a list.",
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"snapshot_id": {
			Description:   "A database snapshot to initialize the RDS instance from, at launch.",
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			ConflictsWith: []string{"master_username"},
		},
		"parameter_group_name": {
			Description: "A RDS parameter group name to apply to the RDS instance.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"store_details_in_secret_manager": {
			Description: "Whether or not to store RDS details in the AWS secrets manager.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
		},
		"size": {
			Description: "The instance type of the RDS instance.\n" +
				"See AWS documentation for the [available instance types](https://aws.amazon.com/rds/instance-types/).",
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"encrypt_storage": {
			Description: "Whether or not to encrypt the RDS instance storage.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
		},
		"instance_status": {
			Description: "The current status of the RDS instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

// SCHEMA for resource crud
func resourceDuploRdsInstance() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_rds_instance` manages an AWS RDS instance in Duplo.",

		ReadContext:   resourceDuploRdsInstanceRead,
		CreateContext: resourceDuploRdsInstanceCreate,
		UpdateContext: resourceDuploRdsInstanceUpdate,
		DeleteContext: resourceDuploRdsInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: rdsInstanceSchema(),
	}
}

/// READ resource
func resourceDuploRdsInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploRdsInstanceRead ******** start")

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.RdsInstanceGet(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if duplo == nil {
		d.SetId("")
		return nil
	}

	// Convert the object into Terraform resource data
	jo := rdsInstanceToState(duplo, d)
	for key := range jo {
		d.Set(key, jo[key])
	}
	d.SetId(fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/%s", duplo.TenantID, duplo.Name))

	log.Printf("[TRACE] resourceDuploRdsInstanceRead ******** end")
	return nil
}

/// CREATE resource
func resourceDuploRdsInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploRdsInstanceCreate ******** start")

	// Convert the Terraform resource data into a Duplo object
	duploObject, err := rdsInstanceFromState(d)
	if err != nil {
		return diag.Errorf("Internal error: %s", err)
	}

	// Populate the identifier field, and determine some other fields
	duploObject.Identifier = duploObject.Name
	tenantID := d.Get("tenant_id").(string)
	id := fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/%s", tenantID, duploObject.Name)

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	_, err = c.RdsInstanceCreate(tenantID, duploObject)
	if err != nil {
		return diag.Errorf("Error creating RDS DB instance '%s': %s", id, err)
	}
	d.SetId(id)

	// Wait up to 60 seconds for Duplo to be able to return the instance details.
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "RDS DB instance", id, func() (interface{}, duplosdk.ClientError) {
		return c.RdsInstanceGet(id)
	})
	if diags != nil {
		return diags
	}

	// Wait for the instance to become available.
	err = rdsInstanceWaitUntilAvailable(ctx, c, id, d.Timeout("create"))
	if err != nil {
		return diag.Errorf("Error waiting for RDS DB instance '%s' to be available: %s", id, err)
	}

	diags = resourceDuploRdsInstanceRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploRdsInstanceCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceDuploRdsInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	log.Printf("[TRACE] resourceDuploRdsInstanceUpdate ******** start")

	// Request the password change in Duplo
	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	id := d.Id()
	err = c.RdsInstanceChangePassword(tenantID, duplosdk.DuploRdsInstancePasswordChange{
		Identifier:     d.Get("identifier").(string),
		MasterPassword: d.Get("master_password").(string),
		StorePassword:  true,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for the instance to become unavailable.
	err = rdsInstanceWaitUntilUnavailable(ctx, c, id, 2*time.Minute)
	if err != nil {
		return diag.Errorf("Error waiting for RDS DB instance '%s' to be unavailable: %s", id, err)
	}

	// Wait for the instance to become available.
	err = rdsInstanceWaitUntilAvailable(ctx, c, id, d.Timeout("update"))
	if err != nil {
		return diag.Errorf("Error waiting for RDS DB instance '%s' to be available: %s", id, err)
	}

	diags := resourceDuploRdsInstanceRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploRdsInstanceUpdate ******** end")
	return diags
}

/// DELETE resource
func resourceDuploRdsInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploRdsInstanceDelete ******** start")

	// Delete the object from Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	_, err := c.RdsInstanceDelete(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	diags := waitForResourceToBeMissingAfterDelete(ctx, d, "RDS DB instance", id, func() (interface{}, duplosdk.ClientError) {
		return c.RdsInstanceGet(id)
	})

	// Wait 1 more minute to deal with consistency issues.
	if diags == nil {
		time.Sleep(time.Minute)
	}

	log.Printf("[TRACE] resourceDuploRdsInstanceDelete ******** end")
	return diags
}

// RdsInstanceWaitUntilAvailable waits until an RDS instance is available.
//
// It should be usable both post-creation and post-modification.
func rdsInstanceWaitUntilAvailable(ctx context.Context, c *duplosdk.Client, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"processing", "backing-up", "backtracking", "configuring-enhanced-monitoring", "configuring-iam-database-auth", "configuring-log-exports", "creating",
			"maintenance", "modifying", "moving-to-vpc", "rebooting", "renaming",
			"resetting-master-credentials", "starting", "stopping", "storage-optimization", "upgrading",
		},
		Target:       []string{"available"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
		Refresh: func() (interface{}, string, error) {
			resp, err := c.RdsInstanceGet(id)
			if err != nil {
				return 0, "", err
			}
			if resp.InstanceStatus == "" {
				resp.InstanceStatus = "processing"
			}
			return resp, resp.InstanceStatus, nil
		},
	}
	log.Printf("[DEBUG] RdsInstanceWaitUntilAvailable (%s)", id)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// RdsInstanceWaitUntilUnavailable waits until an RDS instance is unavailable.
//
// It should be usable post-modification.
func rdsInstanceWaitUntilUnavailable(ctx context.Context, c *duplosdk.Client, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Target: []string{
			"processing", "backing-up", "backtracking", "configuring-enhanced-monitoring", "configuring-iam-database-auth", "configuring-log-exports", "creating",
			"maintenance", "modifying", "moving-to-vpc", "rebooting", "renaming",
			"resetting-master-credentials", "starting", "stopping", "storage-optimization", "upgrading",
		},
		Pending:      []string{"available"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
		Refresh: func() (interface{}, string, error) {
			resp, err := c.RdsInstanceGet(id)
			if err != nil {
				return 0, "", err
			}
			if resp.InstanceStatus == "" {
				resp.InstanceStatus = "available"
			}
			return resp, resp.InstanceStatus, nil
		},
	}
	log.Printf("[DEBUG] RdsInstanceWaitUntilUnavailable (%s)", id)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

/*************************************************
 * DATA CONVERSIONS to/from duplo/terraform
 */

// RdsInstanceFromState converts resource data respresenting an RDS instance to a Duplo SDK object.
func rdsInstanceFromState(d *schema.ResourceData) (*duplosdk.DuploRdsInstance, error) {
	duploObject := new(duplosdk.DuploRdsInstance)

	// First, convert things into simple scalars
	duploObject.Name = d.Get("name").(string)
	duploObject.Identifier = d.Get("identifier").(string)
	duploObject.Arn = d.Get("arn").(string)
	duploObject.Endpoint = d.Get("endpoint").(string)
	duploObject.MasterUsername = d.Get("master_username").(string)
	duploObject.MasterPassword = d.Get("master_password").(string)
	duploObject.Engine = d.Get("engine").(int)
	duploObject.EngineVersion = d.Get("engine_version").(string)
	duploObject.SnapshotID = d.Get("snapshot_id").(string)
	duploObject.DBParameterGroupName = d.Get("parameter_group_name").(string)
	duploObject.Cloud = 0 // AWS
	duploObject.SizeEx = d.Get("size").(string)
	duploObject.EncryptStorage = d.Get("encrypt_storage").(bool)
	duploObject.InstanceStatus = d.Get("instance_status").(string)

	return duploObject, nil
}

// RdsInstanceToState converts a Duplo SDK object respresenting an RDS instance to terraform resource data.
func rdsInstanceToState(duploObject *duplosdk.DuploRdsInstance, d *schema.ResourceData) map[string]interface{} {
	if duploObject == nil {
		return nil
	}
	jsonData, _ := json.Marshal(duploObject)
	log.Printf("[TRACE] duplo-RdsInstanceToState ******** 1: INPUT <= %s ", jsonData)

	jo := make(map[string]interface{})

	// First, convert things into simple scalars
	jo["tenant_id"] = duploObject.TenantID
	jo["name"] = duploObject.Name
	jo["identifier"] = duploObject.Identifier
	jo["arn"] = duploObject.Arn
	jo["endpoint"] = duploObject.Endpoint
	if duploObject.Endpoint != "" {
		uriParts := strings.SplitN(duploObject.Endpoint, ":", 2)
		jo["host"] = uriParts[0]
		if len(uriParts) == 2 {
			jo["port"], _ = strconv.Atoi(uriParts[1])
		}
	}
	jo["master_username"] = duploObject.MasterUsername
	jo["master_password"] = duploObject.MasterPassword
	jo["engine"] = duploObject.Engine
	jo["engine_version"] = duploObject.EngineVersion
	jo["snapshot_id"] = duploObject.SnapshotID
	jo["parameter_group_name"] = duploObject.DBParameterGroupName
	jo["size"] = duploObject.SizeEx
	jo["encrypt_storage"] = duploObject.EncryptStorage
	jo["instance_status"] = duploObject.InstanceStatus

	jsonData2, _ := json.Marshal(jo)
	log.Printf("[TRACE] duplo-RdsInstanceToState ******** 2: OUTPUT => %s ", jsonData2)

	return jo
}
