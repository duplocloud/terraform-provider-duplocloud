package duplocloud

import (
	"context"

	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func duploServiceComputedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"tenant_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
		},
		"other_docker_host_config": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"other_docker_config": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"extra_config": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"allocation_tags": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"volumes": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"commands": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"cloud": {
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
		},
		"agent_platform": {
			Type:     schema.TypeInt,
			Optional: true,
			Required: false,
			Default:  0,
		},
		"replicas": {
			Type:     schema.TypeInt,
			Optional: false,
			Required: true,
		},
		"replicas_matching_asg_name": {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
		},
		"docker_image": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
		},
		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
	}
}

// SCHEMA for resource data/search
func dataSourceDuploServices() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploServiceRead,
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

/// READ/SEARCH resources
func dataSourceDuploServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceDuploServiceRead ******** start")

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	list, err := c.DuploServiceList(tenantID)
	if err != nil {
		return diag.Errorf("Unable to list tenant %s services '%s': %s", tenantID, err)
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
			services = append(services, map[string]interface{}{
				"name":                       duplo.Name,
				"tenant_id":                  duplo.TenantID,
				"other_docker_host_config":   duplo.OtherDockerHostConfig,
				"other_docker_config":        duplo.OtherDockerConfig,
				"allocation_tags":            duplo.AllocationTags,
				"extra_config":               duplo.ExtraConfig,
				"commands":                   duplo.Commands,
				"volumes":                    duplo.Volumes,
				"docker_image":               duplo.DockerImage,
				"agent_platform":             duplo.AgentPlatform,
				"replicas_matching_asg_name": duplo.ReplicasMatchingAsgName,
				"replicas":                   duplo.Replicas,
				"cloud":                      duplo.Cloud,
				"tags":                       duplosdk.KeyValueToState("tags", duplo.Tags),
			})
		}
	}
	d.Set("services", services)

	log.Printf("[TRACE] dataSourceDuploServiceRead ******** end")
	return nil
}
