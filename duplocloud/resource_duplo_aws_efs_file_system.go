package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsEFSFileSystem() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the efs file system will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the EFS, this needs to be unique within a region.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"creation_token": {
			Description:  "A unique name (a maximum of 64 characters are allowed) used as reference when creating the Elastic File System to ensure idempotent file system creation.",
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringLenBetween(0, 64),
		},
		"fullname": {
			Description: "Duplo generated name of the EFS.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"lifecycle_policy": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 2,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"transition_to_ia": {
						Description:  "Indicates how long it takes to transition files to the IA storage class. Valid values: `AFTER_1_DAY`, `AFTER_7_DAYS`, `AFTER_14_DAYS`, `AFTER_30_DAYS`, `AFTER_60_DAYS`, or `AFTER_90_DAYS`",
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice(TransitionToIARules_Values(), false),
					},
					"transition_to_primary_storage_class": {
						Description:  "Describes the policy used to transition a file from infequent access storage to primary storage. Valid values: `AFTER_1_ACCESS`",
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice(TransitionToPrimaryStorageClassRules_Values(), false),
					},
				},
			},
		},
		"file_system_arn": {
			Description: "Amazon Resource Name of the file system.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"file_system_id": {
			Description: "The ID that identifies the file system.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"owner_id": {
			Description: "The AWS account that created the file system.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"size_in_bytes": {
			Description: "The latest known metered size (in bytes) of data stored in the file system.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"number_of_mount_targets": {
			Description: "The current number of mount targets that the file system has.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"performance_mode": {
			Description: "The file system performance mode. Can be either `generalPurpose` or `maxIO`.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "generalPurpose",
			ValidateFunc: validation.StringInSlice([]string{
				"generalPurpose",
				"maxIO",
			}, false),
		},
		"throughput_mode": {
			Description: "Throughput mode for the file system. When using `provisioned`, also set `provisioned_throughput_in_mibps`.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "bursting",
			ValidateFunc: validation.StringInSlice([]string{
				"bursting",
				"provisioned",
				"elastic",
			}, false),
		},
		"provisioned_throughput_in_mibps": {
			Description: "The throughput, measured in MiB/s, that you want to provision for the file system. Only applicable with `throughput_mode` set to `provisioned`.",
			Type:        schema.TypeFloat,
			Optional:    true,
			Computed:    true,
		},
		"tag": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until EFS to be available, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"encrypted": {
			Description: "If true, the disk will be encrypted.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"backup": {
			Description: "Specifies whether automatic backups are enabled on the file system that you are creating.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
	}
}

func resourceAwsEFS() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_efs_file_system` Provides an Elastic File System (EFS) File System resource in DuploCloud.",

		ReadContext:   resourceAwsEFSRead,
		CreateContext: resourceAwsEFSCreate,
		UpdateContext: resourceAwsEFSUpdate,
		DeleteContext: resourceAwsEFSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsEFSFileSystem(),
	}
}

func resourceAwsEFSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, efsId, err := parseAwsEFSIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsEFSRead(%s, %s): start", tenantID, efsId)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.DuploEFSGet(tenantID, efsId)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s EFS  '%s': %s", tenantID, efsId, clientErr)
	}

	d.Set("tenant_id", tenantID)
	flattenAwsEfs(tenantID, d, duplo, c)
	log.Printf("[TRACE] resourceAwsEFSRead(%s, %s): end", tenantID, efsId)
	return nil
}

func resourceAwsEFSCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsEFSCreate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rq := expandAwsEfs(d)
	resp, err := c.DuploEFSCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s EFS '%s': %s", tenantID, name, err)
	}

	// Wait for Duplo to be able to return the table's details.
	id := fmt.Sprintf("%s/%s", tenantID, resp.FileSystemID)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "AWS EFS", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploEFSGet(tenantID, resp.FileSystemID)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	//By default, wait until the cache instances to be healthy.
	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = efsWaitUntilReady(ctx, c, tenantID, resp.FileSystemID, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := d.GetOk("lifecycle_policy"); ok {
		input := &duplosdk.PutLifecycleConfigurationInput{
			FileSystemId:      resp.FileSystemID,
			LifecyclePolicies: expandFileSystemLifecyclePolicies(v.([]interface{})),
		}

		_, err := c.DuploAwsLifecyclePolicyUpdate(tenantID, input)

		if err != nil {
			return diag.Errorf("putting EFS file system (%s) lifecycle configuration: %s", d.Id(), err)
		}
	}

	diags = resourceAwsEFSRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsEFSCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsEFSUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	efsId := d.Get("file_system_id").(string)
	c := m.(*duplosdk.Client)

	if d.HasChanges("provisioned_throughput_in_mibps", "throughput_mode") {

		throughputMode := d.Get("throughput_mode").(string)

		input := &duplosdk.DuploEFSUpdateReq{
			FileSystemId: d.Get("file_system_id").(string),
			ThroughputMode: &duplosdk.DuploStringValue{
				Value: throughputMode,
			},
		}

		if throughputMode == "provisioned" {
			input.ProvisionedThroughputInMibps = d.Get("provisioned_throughput_in_mibps").(float64)
		}

		_, err := c.DuploEFSUpdate(tenantID, input)

		if err != nil {
			return diag.Errorf("updating EFS file system (%s): %s", d.Id(), err)
		}

		//By default, wait until the cache instances to be healthy.
		if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
			err := efsWaitUntilReady(ctx, c, tenantID, efsId, d.Timeout("create"))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("lifecycle_policy") {
		input := &duplosdk.PutLifecycleConfigurationInput{
			FileSystemId:      efsId,
			LifecyclePolicies: expandFileSystemLifecyclePolicies(d.Get("lifecycle_policy").([]interface{})),
		}

		if input.LifecyclePolicies == nil {
			input.LifecyclePolicies = []*duplosdk.LifecyclePolicy{}
		}

		_, err := c.DuploAwsLifecyclePolicyUpdate(tenantID, input)

		if err != nil {
			return diag.Errorf("putting EFS file system (%s) lifecycle configuration: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsEFSDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, efsId, err := parseAwsEFSIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsEFSDelete(%s, %s): start", tenantID, efsId)

	// Delete the function.
	c := m.(*duplosdk.Client)
	_, clientErr := c.DuploEFSDelete(tenantID, efsId)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s EFS '%s': %s", tenantID, efsId, clientErr)
	}

	// Wait up to 60 seconds for Duplo to delete the cluster.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "AWS EFS", id, func() (interface{}, duplosdk.ClientError) {
		return c.DynamoDBTableGet(tenantID, efsId)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsEFSDelete(%s, %s): end", tenantID, efsId)
	return nil
}

func expandAwsEfs(d *schema.ResourceData) *duplosdk.DuploEFSCreateReq {
	throughputMode := d.Get("throughput_mode").(string)

	req := &duplosdk.DuploEFSCreateReq{
		Name:            d.Get("name").(string),
		PerformanceMode: d.Get("performance_mode").(string),
		ThroughputMode:  throughputMode,
		Backup:          d.Get("backup").(bool),
		Encrypted:       d.Get("encrypted").(bool),
	}
	if v, ok := d.GetOk("provisioned_throughput_in_mibps"); ok && throughputMode == "provisioned" {
		req.ProvisionedThroughputInMibps = v.(float64)
	}
	if v, ok := d.GetOk("creation_token"); ok {
		req.CreationToken = v.(string)
	}
	return req
}

func flattenAwsEfs(tenantId string, d *schema.ResourceData, efs *duplosdk.DuploEFSGetResp, c *duplosdk.Client) diag.Diagnostics {
	prefix, err := c.GetDuploServicesPrefix(tenantId)
	if err != nil {
		return diag.FromErr(err)
	}
	name, _ := duplosdk.UnprefixName(prefix, efs.Name)
	d.Set("name", name)
	d.Set("file_system_arn", efs.FileSystemArn)
	d.Set("file_system_id", efs.FileSystemID)
	d.Set("owner_id", efs.OwnerID)
	d.Set("size_in_bytes", efs.SizeInBytes.Value)
	d.Set("number_of_mount_targets", efs.NumberOfMountTargets)
	d.Set("fullname", efs.Name)
	d.Set("throughput_mode", efs.ThroughputMode.Value)

	//d.Set("backup","")
	d.Set("encrypted", efs.Encrypted)
	d.Set("performance_mode", efs.PerformanceMode.Value)
	d.Set("tag", keyValueToState("tag", efs.Tags))
	d.Set("provisioned_throughput_in_mibps", efs.ProvisionedThroughputInMibps)
	d.Set("creation_token", efs.CreationToken)
	return nil
}

func parseAwsEFSIdParts(id string) (tenantID, efsId string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, efsId = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func efsWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, efsId string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.DuploEFSGet(tenantID, efsId)
			log.Printf("[TRACE] EFS status is (%s).", rp.LifeCycleState.Value)
			status := "pending"
			if err == nil {
				if rp.LifeCycleState.Value == "available" {
					status = "ready"
				} else {
					status = "pending"
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] dynamodbWaitUntilReady(%s, %s)", tenantID, efsId)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
