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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	// TLSSecurityPolicyPolicyMinTLS10201907 is a TLSSecurityPolicy enum value
	TLSSecurityPolicyPolicyMinTLS10201907 = "Policy-Min-TLS-1-0-2019-07"

	// TLSSecurityPolicyPolicyMinTLS12201907 is a TLSSecurityPolicy enum value
	TLSSecurityPolicyPolicyMinTLS12201907 = "Policy-Min-TLS-1-2-2019-07"
)

// DuploEcacheInstanceSchema returns a Terraform resource schema for an ECS Service
func awsElasticSearchSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][0-9a-z\-]{1,5}$`),
				"must start with a lowercase alphabet and be at least 2 and no more than 6 characters long."+
					" Valid characters are a-z (lowercase letters), 0-9, and - (hyphen)."),
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"domain_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"domain_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"access_policies": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"advanced_options": {
			Type: schema.TypeMap,
			// Optional: true,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"domain_endpoint_options": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enforce_https": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"tls_security_policy": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"elasticsearch_version": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
			Default:  "7.9",
		},
		"endpoint": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"kibana_endpoint": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"storage_size": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
			Default:  20,
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
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			ForceNew: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enabled": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"kms_key_id": {
						Type:     schema.TypeString,
						Optional: true,
						ForceNew: true,
					},
				},
			},
		},
		"enable_node_to_node_encryption": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"selected_zone": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
			ForceNew: true,
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
		ReadContext:   resourceDuploAwsElasticSearchRead,
		CreateContext: resourceDuploAwsElasticSearchCreate,
		DeleteContext: resourceDuploAwsElasticSearchDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsElasticSearchSchema(),
	}
}

/// READ resource
func resourceDuploAwsElasticSearchRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] rresourceDuploAwsElasticSearchRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	tenantID, name := idParts[0], idParts[1]

	// Get the object from Duplo, detecting a missing or deleted object
	c := m.(*duplosdk.Client)
	duplo, err := c.TenantGetElasticSearchDomain(tenantID, name)
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
	d.Set("enable_node_to_node_encryption", duplo.NodeToNodeEncryptionOptions.Enabled)

	// Interpret the cluster config options.
	clusterConfig := map[string]interface{}{}
	clusterConfig["dedicated_master_enabled"] = duplo.ClusterConfig.DedicatedMasterEnabled
	if duplo.ClusterConfig.DedicatedMasterEnabled {
		clusterConfig["dedicated_master_count"] = duplo.ClusterConfig.DedicatedMasterCount
		clusterConfig["dedicated_master_type"] = duplo.ClusterConfig.DedicatedMasterType.Value
	}
	clusterConfig["instance_count"] = duplo.ClusterConfig.InstanceCount
	clusterConfig["instance_type"] = duplo.ClusterConfig.InstanceType.Value
	d.Set("cluster_config", []map[string]interface{}{clusterConfig})

	// Interpret the EBS options.
	ebsOptions := map[string]interface{}{}
	if duplo.EBSOptions.EBSEnabled {
		ebsOptions["ebs_enabled"] = true
		ebsOptions["iops"] = duplo.EBSOptions.IOPS
		ebsOptions["volume_type"] = duplo.EBSOptions.VolumeType.Value
		ebsOptions["volume_size"] = duplo.EBSOptions.VolumeSize
		d.Set("storage_size", duplo.EBSOptions.VolumeSize)
	} else {
		ebsOptions["ebs_enabled"] = false
	}
	d.Set("ebs_options", []map[string]interface{}{ebsOptions})

	// Interpret the encryption at rest options.
	encryptAtRest := map[string]interface{}{}
	if duplo.EncryptionAtRestOptions.Enabled {
		encryptAtRest["enabled"] = true
		if duplo.EncryptionAtRestOptions.KmsKeyID != "" {
			encryptAtRest["kms_key_id"] = duplo.EncryptionAtRestOptions.KmsKeyID
		} else if duplo.KmsKeyID != "" {
			encryptAtRest["kms_key_id"] = duplo.KmsKeyID
		} else {
			encryptAtRest["kms_key_id"] = nil
		}
	} else {
		encryptAtRest["enabled"] = false
		encryptAtRest["kms_key_id"] = nil
	}
	d.Set("encrypt_at_rest", []map[string]interface{}{encryptAtRest})

	// Interpret the snapshot options.
	snapshotOptions := map[string]interface{}{}
	snapshotOptions["automated_snapshot_start_hour"] = duplo.SnapshotOptions.AutomatedSnapshotStartHour
	d.Set("snapshot_options", []map[string]interface{}{snapshotOptions})

	// Interpret the domain endpoint options.
	endpointOptions := map[string]interface{}{}
	endpointOptions["enforce_https"] = duplo.DomainEndpointOptions.EnforceHTTPS
	endpointOptions["tls_security_policy"] = duplo.DomainEndpointOptions.TLSSecurityPolicy.Value
	d.Set("domain_endpoint_options", []map[string]interface{}{endpointOptions})

	// Interpret the VPC options.
	vpcOptions := map[string]interface{}{}
	vpcOptions["vpc_id"] = duplo.VPCOptions.VpcID
	vpcOptions["availability_zones"] = duplo.VPCOptions.AvailabilityZones
	vpcOptions["security_group_ids"] = duplo.VPCOptions.SecurityGroupIDs
	vpcOptions["subnet_ids"] = duplo.VPCOptions.SubnetIDs
	d.Set("vpc_options", []map[string]interface{}{vpcOptions})

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

/// CREATE resource
func resourceDuploAwsElasticSearchCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploAwsElasticSearchCreate ******** start")

	// Set simple fields first.
	duploObject := duplosdk.DuploElasticSearchDomainRequest{
		Name:    d.Get("name").(string),
		Version: d.Get("elasticsearch_version").(string),
		EBSOptions: duplosdk.DuploElasticSearchDomainEBSOptions{
			VolumeSize: d.Get("storage_size").(int),
		},
	}

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	id := fmt.Sprintf("%s/%s", tenantID, duploObject.Name)

	// Set encryption-at-rest
	encryptAtRest, err := getOptionalBlockAsMap(d, "encrypt_at_rest")
	if err != nil {
		return diag.FromErr(err)
	}
	kmsKeyID, ok := encryptAtRest["kms_key_id"]
	if ok {
		duploObject.KmsKeyID = kmsKeyID.(string)
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
	if ok {
		duploObject.VPCOptions.SubnetIDs = selectedSubnetIDs.([]string)
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
	d.SetId(id)

	// Wait up to 60 seconds for Duplo to be able to return the domain's details.
	err = resource.Retry(time.Minute, func() *resource.RetryError {
		resp, errget := c.TenantGetElasticSearchDomain(tenantID, duploObject.Name)

		if errget != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting ElasticSearch domain '%s': %s", id, err))
		}

		if resp == nil {
			return resource.RetryableError(fmt.Errorf("Expected ElasticSearch domain '%s' to be retrieved, but got: nil", id))
		}

		return nil
	})

	// Wait for the instance to become available.
	err = awsElasticSearchDomainWaitUntilAvailable(c, tenantID, duploObject.Name)
	if err != nil {
		return diag.Errorf("Error waiting for ElasticSearch domain '%s' to be available: %s", id, err)
	}

	diags := resourceDuploAwsElasticSearchRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploAwsElasticSearchCreate ******** end")
	return diags
}

/// DELETE resource
func resourceDuploAwsElasticSearchDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	if v, ok := m["dedicated_master_enabled"]; ok {
		isEnabled := v.(bool)
		duplo.DedicatedMasterEnabled = isEnabled

		if isEnabled {
			if v, ok := m["dedicated_master_count"]; ok && v.(int) > 0 {
				duplo.DedicatedMasterCount = v.(int)
			}
			if v, ok := m["dedicated_master_type"]; ok && v.(string) != "" {
				duplo.DedicatedMasterType.Value = v.(string)
			}
		}
	}
}

// awsElasticSearchDomainWaitUntilAvailable waits until an ECache instance is available.
//
// It should be usable both post-creation and post-modification.
func awsElasticSearchDomainWaitUntilAvailable(c *duplosdk.Client, tenantID string, name string) error {
	stateConf := &resource.StateChangeConf{
		Pending:      []string{"new", "processing", "upgrade-processing"},
		Target:       []string{"created"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      60 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			status := "new"
			resp, err := c.TenantGetElasticSearchDomain(tenantID, name)
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
	log.Printf("[DEBUG] awsElasticSearchDomainWaitUntilAvailable (%s/%s)", tenantID, name)
	_, err := stateConf.WaitForState()
	return err
}
