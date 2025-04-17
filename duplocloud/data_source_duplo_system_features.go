package duplocloud

import (
	"context"
	"log"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDuploSystemFeatures() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploSystemFeaturesRead,
		Schema: map[string]*schema.Schema{
			"is_katkit_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_signup_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_compliance_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_billing_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_siem_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_aws_cloud_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"aws_regions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"default_aws_account": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_aws_partition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_aws_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_azure_cloud_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_google_cloud_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_on_prem_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"eks_versions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"supported_versions": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"is_otp_needed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_aws_admin_jit_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_duplo_ops_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"devops_manager_hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenant_name_max_length": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"is_tags_based_resource_mgmt_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"duplo_shell_fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"s3_bucket_name_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled_flags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"app_configs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"disable_oob_data": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"default_infra_cloud": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDuploSystemFeaturesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceDuploSystemFeaturesRead ******** start")
	c := m.(*duplosdk.Client)

	// Fetch system features from the API
	resp, err := c.AdminGetSystemFeatures()
	log.Printf("[TRACE] dataSourceDuploSystemFeaturesRead: resp: %v", resp)
	if err != nil {
		return diag.Errorf("Unable to get system features: %s", err)
	}

	// Parse the response into the schema
	if err := d.Set("is_katkit_enabled", resp.IsKatkitEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_signup_enabled", resp.IsSignupEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_compliance_enabled", resp.IsComplianceEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_billing_enabled", resp.IsBillingEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_siem_enabled", resp.IsSiemEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_aws_cloud_enabled", resp.IsAwsCloudEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("aws_regions", resp.AwsRegions); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_aws_account", resp.DefaultAwsAccount); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_aws_partition", resp.DefaultAwsPartition); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_aws_region", resp.DefaultAwsRegion); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_azure_cloud_enabled", resp.IsAzureCloudEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_google_cloud_enabled", resp.IsGoogleCloudEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_on_prem_enabled", resp.IsOnPremEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("eks_versions", []interface{}{map[string]interface{}{
		"default_version":    resp.EksVersions.DefaultVersion,
		"supported_versions": resp.EksVersions.SupportedVersions,
	}}); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_otp_needed", resp.IsOtpNeeded); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_aws_admin_jit_enabled", resp.IsAwsAdminJITEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_duplo_ops_enabled", resp.IsDuploOpsEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("devops_manager_hostname", resp.DevopsManagerHostname); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tenant_name_max_length", resp.TenantNameMaxLength); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_tags_based_resource_mgmt_enabled", resp.IsTagsBasedResourceMgmtEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("duplo_shell_fqdn", resp.DuploShellFqdn); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("s3_bucket_name_prefix", resp.S3BucketNamePrefix); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled_flags", resp.EnabledFlags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("app_configs", flattenDuploSystemFeaturesAppConfigs(resp.AppConfigs)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("disable_oob_data", resp.DisableOobData); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_infra_cloud", resp.DefaultInfraCloud); err != nil {
		return diag.FromErr(err)
	}

	// Set the ID for the resource
	d.SetId("duplo-system-features")
	log.Printf("[TRACE] dataSourceDuploSystemFeaturesRead ******** end")
	return nil
}

func flattenDuploSystemFeaturesAppConfigs(appConfigs []duplosdk.DuploSystemFeaturesAppConfigs) []map[string]interface{} {
	result := make([]map[string]interface{}, len(appConfigs))
	for i, config := range appConfigs {
		result[i] = map[string]interface{}{
			"type":  config.Type,
			"key":   config.Key,
			"value": config.Value,
		}
	}
	return result
}
