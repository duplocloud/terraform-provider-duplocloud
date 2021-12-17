package duplocloud

import (
	"context"

	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource data/search
func dataSourceDuploServices() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploServicesRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"services": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: duploServiceComputedSchema(),
				},
			},
		},
	}
}

func dataSourceDuploServicesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceDuploServicesRead: start")

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	list, err := c.ReplicationControllerList(tenantID)
	if err != nil {
		return diag.Errorf("Unable to list tenant %s services: %s", tenantID, err)
	}
	d.SetId(tenantID)

	// Build the list for TF state
	itemCount := 0
	if list != nil {
		itemCount = len(*list)
	}
	services := make([]map[string]interface{}, 0, itemCount)
	if itemCount > 0 {
		for _, duplo := range *list {

			// Get the base information from the replication controller.
			rpc := map[string]interface{}{
				"name":                                 duplo.Name,
				"tenant_id":                            duplo.TenantId,
				"volumes":                              duplo.Volumes,
				"any_host_allowed":                     duplo.IsAnyHostAllowed,
				"cloud_creds_from_k8s_service_account": duplo.IsCloudCredsFromK8sServiceAccount,
				"lb_synced_deployment":                 duplo.IsLBSyncedDeployment,
				"agent_platform":                       duplo.AgentPlatform,
				"replicas_matching_asg_name":           duplo.ReplicasMatchingAsgName,
				"replicas":                             duplo.Replicas,
				"cloud":                                duplo.Cloud,
				"tags":                                 keyValueToState("tags", duplo.Tags),
			}

			// If there is a pod template, get additional information from that.
			if duplo.Template != nil {
				rpc["other_docker_host_config"] = duplo.Template.OtherDockerHostConfig
				rpc["other_docker_config"] = duplo.Template.OtherDockerConfig
				rpc["allocation_tags"] = duplo.Template.AllocationTags
				rpc["extra_config"] = duplo.Template.ExtraConfig
				rpc["commands"] = duplo.Template.Commands

				// If there is at least one container, get the first docker image from it.
				if duplo.Template.Containers != nil && len(*duplo.Template.Containers) > 0 {
					rpc["docker_image"] = (*duplo.Template.Containers)[0].Image
				}
			}

			services = append(services, rpc)
		}
	}
	if err := d.Set("services", services); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] dataSourceDuploServicesRead: end")
	return nil
}
