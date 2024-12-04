package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource data/search
func dataSourceAsgProfiles() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_asg_profiles` lists autoscaling group profiles in a Duplo tenant.",

		ReadContext: dataSourceAsgProfilesRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The GUID of the tenant in which to list the ASG profiles.",
				Type:        schema.TypeString,
				Computed:    false,
				Required:    true,
			},
			"asg_profiles": {
				Description: "The list of ASG profiles.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: autoscalingGroupSchema(),
				},
			},
		},
	}
}

// READ/SEARCH resources
func dataSourceAsgProfilesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] dataSourceAsgProfilesRead(%s): start", tenantID)

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	list, err := c.AsgProfileGetList(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Populate the results from the list.
	profiles := make([]interface{}, 0, len(*list))
	for _, duplo := range *list {
		// TODO: ability to filter
		profiles = append(profiles, flattenAsgProfile(&duplo))
	}
	d.SetId(tenantID)

	if err := d.Set("asg_profiles", profiles); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] dataSourceAsgProfilesRead(%s): end", tenantID)
	return nil
}

func flattenAsgProfile(duplo *duplosdk.DuploAsgProfile) map[string]interface{} {
	mp := map[string]interface{}{
		"instance_count":      duplo.DesiredCapacity,
		"min_instance_count":  duplo.MinSize,
		"max_instance_count":  duplo.MaxSize,
		"user_account":        duplo.AccountName,
		"tenant_id":           duplo.TenantId,
		"friendly_name":       duplo.FriendlyName,
		"capacity":            duplo.Capacity,
		"zone":                duplo.Zone,
		"is_minion":           duplo.IsMinion,
		"image_id":            duplo.ImageID,
		"base64_user_data":    duplo.Base64UserData,
		"prepend_user_data":   duplo.PrependUserData,
		"agent_platform":      duplo.AgentPlatform,
		"is_ebs_optimized":    duplo.IsEbsOptimized,
		"allocated_public_ip": duplo.AllocatedPublicIP,
		"encrypt_disk":        duplo.EncryptDisk,
		"metadata":            keyValueToState("metadata", duplo.MetaData),
		"tags":                keyValueToState("tags", duplo.Tags),
		"minion_tags":         keyValueToState("minion_tags", duplo.CustomDataTags),
		"volume":              flattenNativeHostVolumes(duplo.Volumes),
		"network_interface":   flattenNativeHostNetworkInterfaces(duplo.NetworkInterfaces),
	}
	if duplo.Taints != nil {
		mp["taints"] = flattenTaints(*duplo.Taints)
	}
	return mp
}
