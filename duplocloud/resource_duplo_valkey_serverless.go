package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func valkeyServerlessInstanceSchema() map[string]*schema.Schema {
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
			Description: "The short name of the serverless valkey.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 50-MAX_DUPLO_LENGTH),
				validation.StringMatch(regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`), "Invalid AWS Valkey name"),
				validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "AWS Valkey names cannot end with a hyphen"),
				validation.StringNotInSlice([]string{"--"}, true),
			),
		},
		"fullname": {
			Description: "The full name of the elasticache instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"description": {
			Description: "The description for serverless valkey",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"endpoint": {
			Description: "The endpoint of the serverless valkey.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"port": {
			Description: "The listening port of the elasticache instance.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"engine_version": {
			Description: "The major version of the valkey serverless engine.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"actual_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"kms_key_id": {
			Description: "The globally unique identifier for the key.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			ForceNew:    true,
		},
		"arn": {
			Description: "The arn of serverless valkey",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"snapshot_retention_limit": {
			Description:  "Specify retention limit in days. Accepted values - 1-35.",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
			ValidateFunc: validation.IntBetween(1, 35),
		},
		"subnet_ids": {
			Description: "subnet ids allocated to serverless valkey",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"security_group_ids": {
			Description: "subnet ids allocated to serverless valkey",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

// SCHEMA for resource crud
func resourceDuploServerlessValkeyInstance() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_valkey_serverless` used to manage Serverless Valkey within a DuploCloud-managed environment.",

		ReadContext:   resourceDuploValkeyServerlessRead,
		CreateContext: resourceDuploValkeyServerlessCreate,
		DeleteContext: resourceDuploValkeyServerlessDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(29 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: valkeyServerlessInstanceSchema(),
	}
}

// READ resource
func resourceDuploValkeyServerlessRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseValkeyServerlessIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploValkeyServerlessRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, cerr := c.DuploValkeyServerlessGet(tenantID, name)
	if duplo == nil {
		d.SetId("")
		return nil
	}
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[DEBUG] resourceDuploValkeyServerlessRead: serverless valkey %s not found for tenantId %s, removing from state", name, tenantID)
			d.SetId("")
			return nil
		}
		return diag.FromErr(cerr)
	}
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	// Convert the object into Terraform resource data
	diag := flattenValkeyServerless(duplo, d)

	log.Printf("[TRACE] resourceDuploValkeyServerlessRead(%s, %s): end", tenantID, name)
	return diag
}

// CREATE resource
func resourceDuploValkeyServerlessCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceDuploValkeyServerlessCreate(%s): start", tenantID)

	duplo := expandValkeyServerless(d)
	c := m.(*duplosdk.Client)

	rp, err := c.DuploValkeyServerlessCreate(tenantID, duplo)
	if err != nil {
		return diag.Errorf("Error creating valkey serverless resource '%s': %s", duplo.Name, err)
	}
	id := fmt.Sprintf("%s/valkey/serverless/%s", tenantID, duplo.Name)
	d.SetId(id)
	d.Set("fullname", rp.Name)
	// Wait up to 60 seconds for Duplo to be able to return the instance details.
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "Valkey serverless", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploValkeyServerlessGet(tenantID, rp.Name)
	})
	if diags != nil {
		return diags
	}

	// Wait for the instance to become available.
	err = serverlessValkeyWaitUntilAvailable(ctx, c, tenantID, duplo.Name)
	if err != nil {
		return diag.Errorf("Error waiting for valkey serverless '%s' to be available: %s", id, err)
	}

	// Read the resource state
	diags = resourceDuploValkeyServerlessRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploValkeyServerlessCreate(%s, %s): end", tenantID, duplo.Name)
	return diags
}

// DELETE resource
func resourceDuploValkeyServerlessDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseECacheInstanceIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	fname := d.Get("fullname").(string)
	log.Printf("[TRACE] resourceDuploValkeyServerlessDelete(%s, %s): start", tenantID, name)

	// Delete the object from Duplo
	c := m.(*duplosdk.Client)
	cerr := c.DuploValkeyServerlessDelete(tenantID, name)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[DEBUG] resourceDuploValkeyServerlessDelete: Valkey serverless  %s not found for tenantId %s, removing from state", name, tenantID)
			return nil
		}
		return diag.FromErr(cerr)
	}

	// Wait up to 60 seconds for Duplo to show the object as deleted.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "valkey serverless", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploValkeyServerlessGet(tenantID, fname)
	})

	log.Printf("[TRACE] resourceDuploValkeyServerlessDelete(%s, %s): end", tenantID, name)
	return diag
}

// expand Ecache Instance converts resource data respresenting an ECache instance to a Duplo SDK object.
func expandValkeyServerless(d *schema.ResourceData) *duplosdk.DuploValkeyServerless {
	return &duplosdk.DuploValkeyServerless{
		Name:                   d.Get("name").(string),
		Description:            d.Get("description").(string),
		Engine:                 "Valkey",
		KMSKeyId:               d.Get("kms_key_id").(string),
		EngineVersion:          d.Get("engine_version").(string),
		SnapshotRetentionLimit: d.Get("snapshot_retention_limit").(int),
	}
}

// flattenEcacheInstance converts a Duplo SDK object respresenting an ECache instance to terraform resource data.
func flattenValkeyServerless(duplo *duplosdk.DuploValkeyServerlessResponse, d *schema.ResourceData) diag.Diagnostics {
	d.Set("fullname", duplo.Name)
	d.Set("arn", duplo.Arn)
	if duplo.Endpoint != nil {
		d.Set("endpoint", duplo.Endpoint.Address)
		d.Set("port", duplo.Endpoint.Port)
	}
	d.Set("actual_version", duplo.FullEngineVersion)
	d.Set("engine_version", duplo.MajorEngineVersion)
	d.Set("kms_key_id", duplo.KmsKeyId)
	d.Set("description", duplo.Description)
	d.Set("snapshot_retention_limit", duplo.SnapshotRetentionLimit)
	d.Set("kms_key_id", duplo.KmsKeyId)
	d.Set("snapshot_retention_limit", duplo.SnapshotRetentionLimit)
	sbnets := []interface{}{}
	for _, subnet := range duplo.SubnetIds {
		sbnets = append(sbnets, subnet)
	}
	d.Set("subnet_ids", sbnets)
	sgs := []interface{}{}
	for _, sg := range duplo.SecurityGroupIds {
		sgs = append(sgs, sg)
	}
	d.Set("security_group_ids", sgs)
	return nil
}

func serverlessValkeyWaitUntilAvailable(ctx context.Context, c *duplosdk.Client, tenantID, name string) error {
	stateConf := &retry.StateChangeConf{
		Pending:      []string{"pending"},
		Target:       []string{"available"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      20 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			resp, err := c.DuploValkeyServerlessGet(tenantID, name)
			if err != nil {
				return 0, "", err
			}
			status := "pending"
			if resp.Status == "available" {
				status = "available"
			}
			return resp, status, nil
		},
	}
	log.Printf("[DEBUG] serverlessValkeyWaitUntilAvailable (%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func parseValkeyServerlessIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) == 4 {
		tenantID, name = idParts[0], idParts[3]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
