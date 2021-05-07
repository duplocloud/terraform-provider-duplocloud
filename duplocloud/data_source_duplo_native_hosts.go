package duplocloud

import (
	"context"

	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource data/search
func dataSourceNativeHosts() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_native_hosts` lists native hosts in a Duplo tenant.",

		ReadContext: dataSourceNativeHostsRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The GUID of the tenant in which to list the hosts.",
				Type:        schema.TypeString,
				Computed:    false,
				Required:    true,
			},
			"hosts": {
				Description: "The list of native hosts.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: nativeHostSchema(),
				},
			},
		},
	}
}

/// READ/SEARCH resources
func dataSourceNativeHostsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] dataSourceNativeHostRead(%s): start", tenantID)

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	list, err := c.NativeHostGetList(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Populate the results from the list.
	hosts := make([]interface{}, 0, len(*list))
	for _, duplo := range *list {
		// TODO: ability to filter
		hosts = append(hosts, flattenNativeHost(&duplo))
	}

	if err := d.Set("hosts", hosts); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tenantID)

	log.Printf("[TRACE] dataSourceNativeHostRead(%s): end", tenantID)
	return nil
}

func flattenNativeHost(duplo *duplosdk.DuploNativeHost) map[string]interface{} {
	return map[string]interface{}{
		"instance_id":         duplo.InstanceID,
		"user_account":        duplo.UserAccount,
		"tenant_id":           duplo.TenantID,
		"friendly_name":       duplo.FriendlyName,
		"capacity":            duplo.Capacity,
		"zone":                duplo.Zone,
		"is_minion":           duplo.IsMinion,
		"image_id":            duplo.ImageID,
		"base64_user_data":    duplo.Base64UserData,
		"agent_platform":      duplo.AgentPlatform,
		"is_ebs_optimized":    duplo.IsEbsOptimized,
		"allocated_public_ip": duplo.AllocatedPublicIP,
		"cloud":               duplo.Cloud,
		"encrypt_disk":        duplo.EncryptDisk,
		"status":              duplo.Status,
		"identity_role":       duplo.IdentityRole,
		"private_ip_address":  duplo.PrivateIPAddress,
		"metadata":            duplosdk.KeyValueToState("metadata", duplo.MetaData),
		"tags":                duplosdk.KeyValueToState("tags", duplo.Tags),
		"minion_tags":         duplosdk.KeyValueToState("minion_tags", duplo.MinionTags),
		"volumes":             flattenNativeHostVolumes(duplo.Volumes),
		"network_interfaces":  flattenNativeHostNetworkInterfaces(duplo.NetworkInterfaces),
	}
}
