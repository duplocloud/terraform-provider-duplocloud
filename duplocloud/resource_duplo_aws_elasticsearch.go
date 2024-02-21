package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	// TLSSecurityPolicyPolicyMinTLS10201907 is a TLSSecurityPolicy enum value
	TLSSecurityPolicyPolicyMinTLS10201907 = "Policy-Min-TLS-1-0-2019-07"

	// TLSSecurityPolicyPolicyMinTLS12201907 is a TLSSecurityPolicy enum value
	TLSSecurityPolicyPolicyMinTLS12201907 = "Policy-Min-TLS-1-2-2019-07"
)

func awsElasticSearchSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the ElasticSearch instance will be created in.",
			Type:         schema.TypeString,
			Optional:     false,
			Required:     true,
			ForceNew:     true, //switch tenant
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the ElasticSearch instance.  Duplo will add a prefix to the name.  You can retrieve the full name from the `domain_name` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][0-9a-z\-]{1,5}$`),
				"must start with a lowercase alphabet and be at least 2 and no more than 6 characters long."+
					" Valid characters are a-z (lowercase letters), 0-9, and - (hyphen)."),
		},
		"arn": {
			Description: "The ARN of the ElasticSearch instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"domain_id": {
			Description: "The domain ID of the ElasticSearch instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"domain_name": {
			Description: "The full name of the ElasticSearch instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"access_policies": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"advanced_options": {
			Type:     schema.TypeMap,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"require_ssl": {
			Description: "Whether or not to require SSL for accessing this ElasticSearch instance.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"use_latest_tls_cipher": {
			Description: "Whether or not to use the latest TLS cipher for this ElasticSearch instance.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"elasticsearch_version": {
			Description: "The version of the ElasticSearch instance.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Default:     "7.9",
		},
		"endpoints": {
			Description: "The endpoints to use when connecting to the ElasticSearch instance.",
			Type:        schema.TypeMap,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"storage_size": {
			Description: "The storage volume size, in GB, for the ElasticSearch instance.",
			Type:        schema.TypeInt,
			Optional:    true,
			ForceNew:    true,
			Default:     20,
		},
		"ebs_options": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"ebs_enabled": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"iops": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"volume_size": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"volume_type": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"encrypt_at_rest": {
			Description: "The storage encryption settings for the ElasticSearch instance.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enabled": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"kms_key_name": {
						Description: "The name of a KMS key to use with the ElasticSearch instance.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						ForceNew:    true,
					},
					"kms_key_id": {
						Description: "The ID of a KMS key to use with the ElasticSearch instance.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						ForceNew:    true,
					},
				},
			},
		},
		"enable_node_to_node_encryption": {
			Description: "Whether or not to use the enable node-to-node encryption for this ElasticSearch instance.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"selected_zone": {
			Description: "The numerical index of the zone to launch this ElasticSearch instance in.",
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"cluster_config": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			ForceNew: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"dedicated_master_count": {
						Type:             schema.TypeInt,
						Optional:         true,
						DiffSuppressFunc: isDedicatedMasterDisabled,
						Default:          0,
					},
					"dedicated_master_enabled": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
					"dedicated_master_type": {
						Type:             schema.TypeString,
						Optional:         true,
						DiffSuppressFunc: isDedicatedMasterDisabled,
						Default:          "t2.small.elasticsearch",
					},
					"instance_count": {
						Type:     schema.TypeInt,
						Optional: true,
						Default:  1,
					},
					"instance_type": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "t2.small.elasticsearch",
					},
					"cold_storage_options": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"enabled": {
									Type:     schema.TypeBool,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
					"warm_count": {
						Type:         schema.TypeInt,
						Optional:     true,
						ValidateFunc: validation.IntBetween(2, 150),
					},
					"warm_enabled": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"warm_type": {
						Type:     schema.TypeString,
						Optional: true,
						ValidateFunc: validation.StringInSlice([]string{
							"ultrawarm1.medium.search",
							"ultrawarm1.large.search",
							"ultrawarm1.xlarge.search",
						}, false),
					},
				},
			},
		},
		"snapshot_options": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"automated_snapshot_start_hour": {
						Type:     schema.TypeInt,
						Required: true,
					},
				},
			},
		},
		"vpc_options": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			ForceNew: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"availability_zones": {
						Type:     schema.TypeList,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"security_group_ids": {
						Type:     schema.TypeList,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"subnet_ids": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"vpc_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
	}
}

// Resource for managing an AWS ElasticSearch instance
func resourceDuploAwsElasticSearch() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_elasticsearch` manages an AWS ElasticSearch instance in Duplo.",

		ReadContext:   resourceDuploAwsElasticSearchRead,
		CreateContext: resourceDuploAwsElasticSearchCreate,
		UpdateContext: resourceDuploAwsElasticSearchUpdate,
		DeleteContext: resourceDuploAwsElasticSearchDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(75 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsElasticSearchSchema(),
	}
}

// READ resource
func resourceDuploAwsElasticSearchRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] rresourceDuploAwsElasticSearchRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	// Get the object from Duplo, detecting a missing or deleted object
	c := m.(*duplosdk.Client)
	duplo, err := c.TenantGetElasticSearchDomain(tenantID, name, false)
	if duplo == nil || duplo.Deleted {
		d.SetId("") // object missing or deleted
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s AWS ElasticSearch Domain '%s': %s", tenantID, name, err)
	}

	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/%s", duplo.TenantID, duplo.Name))
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("arn", duplo.Arn)
	d.Set("domain_id", duplo.DomainID)
	d.Set("domain_name", duplo.DomainName)
	d.Set("elasticsearch_version", duplo.ElasticSearchVersion)
	d.Set("access_policies", duplo.AccessPolicies)
	d.Set("advanced_options", duplo.AdvancedOptions)
	d.Set("endpoints", duplo.Endpoints)
	d.Set("enable_node_to_node_encryption", duplo.NodeToNodeEncryptionOptions.Enabled)
	d.Set("require_ssl", duplo.DomainEndpointOptions.EnforceHTTPS)
	d.Set("use_latest_tls_cipher", duplo.DomainEndpointOptions.TLSSecurityPolicy.Value == TLSSecurityPolicyPolicyMinTLS12201907)

	// Set more complex fields next.
	d.Set("cluster_config", awsElasticSearchDomainClusterConfigToState(&duplo.ClusterConfig))
	d.Set("encrypt_at_rest", awsElasticSearchDomainEncryptionAtRestToState(c, tenantID, &duplo.EncryptionAtRestOptions))
	d.Set("ebs_options", awsElasticSearchDomainEBSOptionsToState(&duplo.EBSOptions))
	if duplo.EBSOptions.EBSEnabled {
		d.Set("storage_size", duplo.EBSOptions.VolumeSize)
	}
	d.Set("snapshot_options", []map[string]interface{}{{
		"automated_snapshot_start_hour": duplo.SnapshotOptions.AutomatedSnapshotStartHour,
	}})
	d.Set("vpc_options", []map[string]interface{}{{
		"vpc_id":             duplo.VPCOptions.VpcID,
		"availability_zones": duplo.VPCOptions.AvailabilityZones,
		"security_group_ids": duplo.VPCOptions.SecurityGroupIDs,
		"subnet_ids":         duplo.VPCOptions.SubnetIDs,
	}})

	// Interpret the selected zone.
	if duplo.ClusterConfig.InstanceCount == 1 && len(duplo.VPCOptions.SubnetIDs) == 1 {
		subnetIDs, err := c.TenantGetInternalSubnets(tenantID)
		if err != nil {
			return diag.Errorf("Internal error: failed to get internal subnets for tenant '%s': %s", tenantID, err)
		}

		// Find the selected subnet in the list, then use this as the zone.
		for i, subnetID := range subnetIDs {
			if subnetID == duplo.VPCOptions.SubnetIDs[0] {
				d.Set("selected_zone", i+1)
				break
			}
		}
	}

	log.Printf("[TRACE] resourceDuploAwsElasticSearchRead ******** end")
	return nil
}

// CREATE resource
func resourceDuploAwsElasticSearchCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploAwsElasticSearchCreate ******** start")

	// Set simple fields first.
	duploVPCOptions := duplosdk.DuploElasticSearchDomainVPCOptions{}
	duploObject := duplosdk.DuploElasticSearchDomainRequest{
		Name:                       d.Get("name").(string),
		Version:                    d.Get("elasticsearch_version").(string),
		RequireSSL:                 d.Get("require_ssl").(bool),
		UseLatestTLSCipher:         d.Get("use_latest_tls_cipher").(bool),
		EnableNodeToNodeEncryption: d.Get("enable_node_to_node_encryption").(bool),
		EBSOptions: duplosdk.DuploElasticSearchDomainEBSOptions{
			VolumeSize: d.Get("storage_size").(int),
		},
		VPCOptions: duploVPCOptions,
	}

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	id := fmt.Sprintf("%s/%s", tenantID, duploObject.Name)

	// Set encryption-at-rest
	encryptAtRest, err := getOptionalBlockAsMap(d, "encrypt_at_rest")
	if err != nil {
		return diag.FromErr(err)
	}
	if kmsKeyName, ok := encryptAtRest["kms_key_name"]; ok && kmsKeyName != nil && kmsKeyName.(string) != "" {
		if kmsKeyID, ok := encryptAtRest["kms_key_id"]; ok && kmsKeyID != nil && kmsKeyID.(string) != "" {
			return diag.Errorf("encrypt_at_rest.kms_key_name and encrypt_at_rest.kms_key_id are mutually exclusive")
		}

		// Let the user pick a KMS key by name, just like the UI.
		kmsKey, err := c.TenantGetKmsKeyByName(tenantID, kmsKeyName.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		if kmsKey != nil {
			duploObject.KmsKeyID = kmsKey.KeyID
		}

	} else if kmsKeyID, ok := encryptAtRest["kms_key_id"]; ok {
		if kmsKeyID != nil {
			duploObject.KmsKeyID = kmsKeyID.(string)
		}
	}

	// Set cluster config
	clusterConfig, err := getOptionalBlockAsMap(d, "cluster_config")
	if err != nil {
		return diag.FromErr(err)
	}
	awsElasticSearchDomainClusterConfigFromState(clusterConfig, &duploObject.ClusterConfig)

	// Set VPC options
	vpcOptions, err := getOptionalBlockAsMap(d, "vpc_options")
	if err != nil {
		return diag.FromErr(err)
	}
	selectedSubnetIDs, ok := vpcOptions["subnet_ids"]
	if ok && selectedSubnetIDs != nil {
		for _, subnetId := range selectedSubnetIDs.([]interface{}) {
			duploObject.VPCOptions.SubnetIDs = append(duploObject.VPCOptions.SubnetIDs, subnetId.(string))
		}
		// duploObject.VPCOptions.SubnetIDs = selectedSubnetIDs.([]string)
	}

	// Handle subnet selection: either a single zone domain, or explicit subnet IDs
	selectedZone := d.Get("selected_zone").(int)
	subnetIDs, err := c.TenantGetInternalSubnets(tenantID)
	if err != nil {
		return diag.Errorf("Internal error: failed to get internal subnets for tenant '%s': %s", tenantID, err)
	}
	if selectedZone > 0 {
		if selectedZone > len(subnetIDs) {
			return diag.Errorf("Invalid ElasticSearch domain '%s': selected_zone == %d but Duplo only has %d zones", id, selectedZone, len(subnetIDs))
		}
		if duploObject.ClusterConfig.InstanceCount > 1 {
			return diag.Errorf("Invalid ElasticSearch domain '%s': selected_zone not supported when cluster_config.instance_count > 1", id)
		}
		if len(duploObject.VPCOptions.SubnetIDs) > 0 {
			return diag.Errorf("Invalid ElasticSearch domain '%s': selected_zone and vpc_options.subnet_ids are mutually exclusive", id)
		}

		// Populate a single subnet ID automatically, just like the UI
		duploObject.VPCOptions.SubnetIDs = []string{subnetIDs[selectedZone-1]}

	} else if len(duploObject.VPCOptions.SubnetIDs) == 0 {
		if duploObject.ClusterConfig.InstanceCount > 1 {
			// Populate the subnet IDs automatically, just like the UI
			duploObject.VPCOptions.SubnetIDs = subnetIDs
		} else {
			// Require a zone to be selected, just like the UI
			return diag.Errorf("Invalid ElasticSearch domain '%s': vpc_options.subnet_ids or selected_zone must be set", id)
		}
	}

	// Post the object to Duplo
	err = c.TenantUpdateElasticSearchDomain(tenantID, &duploObject)
	if err != nil {
		return diag.Errorf("Error creating ElasticSearch domain '%s': %s", id, err)
	}

	// Wait up to 60 seconds for Duplo to be able to return the domain's details.
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "ElasticSearch domain", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetElasticSearchDomain(tenantID, duploObject.Name, false)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	// Wait for the instance to become available.
	err = awsElasticSearchDomainWaitUntilAvailable(ctx, c, tenantID, duploObject.Name, d.Timeout("create"))
	if err != nil {
		return diag.Errorf("Error waiting for ElasticSearch domain '%s' to be available: %s", id, err)
	}

	diags = resourceDuploAwsElasticSearchRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploAwsElasticSearchCreate ******** end")
	return diags
}

// UPDATE jresource
func resourceDuploAwsElasticSearchUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	log.Printf("[TRACE] resourceDuploAwsElasticSearchUpdate ******** start")

	// Set simple fields first.
	duploObject := duplosdk.DuploElasticSearchDomainRequest{
		State:              "update",
		Name:               d.Get("name").(string),
		RequireSSL:         d.Get("require_ssl").(bool),
		UseLatestTLSCipher: d.Get("use_latest_tls_cipher").(bool),
	}

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	id := fmt.Sprintf("%s/%s", tenantID, duploObject.Name)

	// Post the object to Duplo
	err = c.TenantUpdateElasticSearchDomain(tenantID, &duploObject)
	if err != nil {
		return diag.Errorf("Error updating ElasticSearch domain '%s': %s", id, err)
	}

	// Wait up to 60 seconds for the ES domain to start processing.
	err = awsElasticSearchDomainWaitUntilUnavailable(ctx, c, tenantID, duploObject.Name, d.Timeout("update"))
	if err != nil {
		log.Printf("[TRACE] resourceDuploAwsElasticSearchUpdate: Error waiting for ElasticSearch domain '%s' changes to begin processing: %s", id, err)
		// return diag.Errorf("Error waiting for ElasticSearch domain '%s' changes to begin processing: %s", id, err)
	}

	// Wait for the instance to become available.
	err = awsElasticSearchDomainWaitUntilAvailable(ctx, c, tenantID, duploObject.Name, d.Timeout("update"))
	if err != nil {
		return diag.Errorf("Error waiting for ElasticSearch domain '%s' to be available: %s", id, err)
	}

	diags := resourceDuploAwsElasticSearchRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploAwsElasticSearchUpdate ******** end")
	return diags
}

// DELETE resource
func resourceDuploAwsElasticSearchDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	log.Printf("[TRACE] resourceDuploAwsElasticSearchDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	domainName := d.Get("domain_name").(string)
	err = c.TenantDeleteElasticSearchDomain(tenantID, domainName)
	if err != nil {
		return diag.Errorf("Error deleting ElasticSearch domain '%s': %s", id, err)
	}

	// Wait for the instance to become deleted.
	err = awsElasticSearchDomainWaitUntilDeleted(ctx, c, tenantID, name, d.Timeout("delete"))
	if err != nil {
		return diag.Errorf("Error waiting for ElasticSearch domain '%s' to be deleted: %s", id, err)
	}

	log.Printf("[TRACE] resourceDuploAwsElasticSearchDelete ******** end")
	return nil
}

func isDedicatedMasterDisabled(k, old, new string, d *schema.ResourceData) bool {
	v, ok := d.GetOk("cluster_config")
	if ok {
		clusterConfig := v.([]interface{})[0].(map[string]interface{})
		return !clusterConfig["dedicated_master_enabled"].(bool)
	}
	return true
}

func awsElasticSearchDomainEncryptionAtRestToState(c *duplosdk.Client, tenantID string, duplo *duplosdk.DuploElasticSearchDomainEncryptAtRestOptions) []map[string]interface{} {

	// Finally, set the fields.
	encryptAtRest := map[string]interface{}{}
	if duplo.Enabled {
		encryptAtRest["enabled"] = true
		if duplo.KmsKeyID != "" {
			encryptAtRest["kms_key_id"] = duplo.KmsKeyID
			if kmsKeyName, err := c.TenantGetKmsKeyByID(tenantID, duplo.KmsKeyID); err == nil {
				encryptAtRest["kms_key_name"] = kmsKeyName
			}
		} else {
			encryptAtRest["kms_key_id"] = nil
		}
	} else {
		encryptAtRest["enabled"] = false
		encryptAtRest["kms_key_id"] = nil
	}
	return []map[string]interface{}{encryptAtRest}
}

func awsElasticSearchDomainEBSOptionsToState(duplo *duplosdk.DuploElasticSearchDomainEBSOptions) []map[string]interface{} {
	ebsOptions := map[string]interface{}{}
	if duplo.EBSEnabled {
		ebsOptions["ebs_enabled"] = true
		ebsOptions["iops"] = duplo.IOPS
		ebsOptions["volume_type"] = duplo.VolumeType.Value
		ebsOptions["volume_size"] = duplo.VolumeSize
	} else {
		ebsOptions["ebs_enabled"] = false
	}
	return []map[string]interface{}{ebsOptions}
}

func awsElasticSearchDomainClusterConfigToState(duplo *duplosdk.DuploElasticSearchDomainClusterConfig) []map[string]interface{} {
	clusterConfig := map[string]interface{}{}
	clusterConfig["dedicated_master_enabled"] = duplo.DedicatedMasterEnabled
	if duplo.DedicatedMasterEnabled {
		clusterConfig["dedicated_master_count"] = duplo.DedicatedMasterCount
		clusterConfig["dedicated_master_type"] = duplo.DedicatedMasterType.Value
	}
	clusterConfig["instance_count"] = duplo.InstanceCount
	clusterConfig["instance_type"] = duplo.InstanceType.Value

	if duplo.WarmEnabled {
		clusterConfig["warm_enabled"] = duplo.WarmEnabled
		clusterConfig["warm_count"] = duplo.WarmCount
		clusterConfig["warm_type"] = duplo.WarmType.Value
	} else {
		clusterConfig["warm_enabled"] = false
	}

	if duplo.ColdStorageOptions != nil && duplo.ColdStorageOptions.Enabled {
		cold_storage_options := make(map[string]bool)
		cold_storage_options["enabled"] = duplo.ColdStorageOptions.Enabled
		clusterConfig["cold_storage_options"] = cold_storage_options
	}

	return []map[string]interface{}{clusterConfig}
}

func awsElasticSearchDomainClusterConfigFromState(m map[string]interface{}, duplo *duplosdk.DuploElasticSearchDomainClusterConfig) {
	if v, ok := m["instance_count"]; ok {
		duplo.InstanceCount = v.(int)
	} else {
		duplo.InstanceCount = 1
	}
	if v, ok := m["instance_type"]; ok {
		duplo.InstanceType.Value = v.(string)
	} else {
		duplo.InstanceType.Value = "t2.small.elasticsearch"
	}
	if v, ok := m["cold_storage_options"]; ok {
		obj := v.([]interface{})
		log.Printf("cold storage option value %+v", obj)
		if len(obj) > 0 {
			coldStorageOptions := duplosdk.DuploElasticSearchDomainColdStorageOptions{
				Enabled: obj[0].(map[string]interface{})["enabled"].(bool),
			}
			duplo.ColdStorageOptions = &coldStorageOptions
		}
	}
	if v, ok := m["warm_count"]; ok {
		duplo.WarmCount = v.(int)
	}
	if v, ok := m["warm_enabled"]; ok {
		duplo.WarmEnabled = v.(bool)
	}
	if v, ok := m["warm_type"]; ok {
		obj := v.(string)
		if obj != "" {
			warmType := duplosdk.DuploStringValue{
				Value: obj,
			}
			duplo.WarmType = &warmType
		}
	}

	if v, ok := m["dedicated_master_enabled"]; ok {
		isEnabled := v.(bool)
		duplo.DedicatedMasterEnabled = isEnabled

		if isEnabled {
			if v, ok := m["dedicated_master_count"]; ok && v.(int) > 0 {
				duplo.DedicatedMasterCount = v.(int)
			}
			if v, ok := m["dedicated_master_type"]; ok && v.(string) != "" {
				obj := v.(string)
				duplo.DedicatedMasterType = nil
				if obj != "" {
					duplo.DedicatedMasterType.Value = obj
				}
			}
		}
	}
}

// awsElasticSearchDomainWaitUntilAvailable waits until an ES domain is unavailable.
//
// It should be usable both post-creation and post-modification.
func awsElasticSearchDomainWaitUntilAvailable(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:      []string{"new", "processing", "upgrade-processing", "created"},
		Target:       []string{"available"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
		Refresh: func() (interface{}, string, error) {
			status := "new"
			resp, err := c.TenantGetElasticSearchDomain(tenantID, name, false)
			if err != nil {
				return 0, "", err
			}
			if resp == nil {
				status = "missing"
			} else if resp.Processing {
				status = "processing"
			} else if resp.UpgradeProcessing {
				status = "upgrade-processing"
			} else if resp.Deleted {
				status = "deleted"
			} else if resp.Created {
				if len(resp.Endpoints) == 0 {
					status = "created"
				} else {
					status = "available"
				}
			}
			return resp, status, nil
		},
	}
	log.Printf("[DEBUG] awsElasticSearchDomainWaitUntilAvailable (%s/%s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// awsElasticSearchDomainWaitUntilUnavailable waits until an ES domain is unavailable.
//
// It should be usable post-modification.
func awsElasticSearchDomainWaitUntilUnavailable(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:      []string{"created"},
		Target:       []string{"processing", "upgrade-processing"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
		Refresh: func() (interface{}, string, error) {
			status := "new"
			resp, err := c.TenantGetElasticSearchDomain(tenantID, name, false)
			if err != nil {
				return 0, "", err
			}
			if resp == nil {
				status = "missing"
			} else if resp.Processing {
				status = "processing"
			} else if resp.UpgradeProcessing {
				status = "upgrade-processing"
			} else if resp.Deleted {
				status = "deleted"
			} else if resp.Created {
				status = "created"
			}
			return resp, status, nil
		},
	}
	log.Printf("[DEBUG] awsElasticSearchDomainWaitUntilUnavailable (%s/%s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// awsElasticSearchDomainWaitUntilDeleted waits until an ES domain is deleted.
//
// It should be usable both post-creation and post-modification.
func awsElasticSearchDomainWaitUntilDeleted(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:      []string{"waiting", "processing", "upgrade-processing"},
		Target:       []string{"deleted"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
		Refresh: func() (interface{}, string, error) {
			status := "waiting"
			resp, err := c.TenantGetElasticSearchDomain(tenantID, name, true)
			if err != nil {
				return 0, "", err
			}
			if resp == nil {
				status = "deleted"
			} else if resp.Processing {
				status = "processing"
			} else if resp.UpgradeProcessing {
				status = "upgrade-processing"
			} else if resp.Deleted {
				status = "deleted"
			} else if resp.Created {
				status = "created"
			}
			return resp, status, nil
		},
	}
	log.Printf("[DEBUG] awsElasticSearchDomainWaitUntilDeleted (%s/%s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
