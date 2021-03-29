package duplocloud

import (
	"context"
	"fmt"

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
			Required: true,
		},
		"other_docker_host_config": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"other_docker_config": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"extra_config": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"allocation_tags": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"volumes": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"commands": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"cloud": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"agent_platform": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"replicas": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"replicas_matching_asg_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"docker_image": {
			Type:     schema.TypeString,
			Computed: true,
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

func dataSourceDuploService() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploServiceRead,
		Schema:      duploServiceComputedSchema(),
	}
}

func dataSourceDuploServicesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceDuploServicesRead ******** start")

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

	log.Printf("[TRACE] dataSourceDuploServicesRead ******** end")
	return nil
}

func dataSourceDuploServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceDuploServiceRead ******** start")

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.DuploServiceGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to read tenant %s service '%s': %s", tenantID, err)
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))

	// Apply TF state
	if duplo.Name != "" {
		d.Set("other_docker_host_config", duplo.OtherDockerHostConfig)
		d.Set("other_docker_config", duplo.OtherDockerConfig)
		d.Set("allocation_tags", duplo.AllocationTags)
		d.Set("extra_config", duplo.ExtraConfig)
		d.Set("commands", duplo.Commands)
		d.Set("volumes", duplo.Volumes)
		d.Set("docker_image", duplo.DockerImage)
		d.Set("agent_platform", duplo.AgentPlatform)
		d.Set("replicas_matching_asg_name", duplo.ReplicasMatchingAsgName)
		d.Set("replicas", duplo.Replicas)
		d.Set("cloud", duplo.Cloud)
		d.Set("tags", duplosdk.KeyValueToState("tags", duplo.Tags))
	}

	log.Printf("[TRACE] dataSourceDuploServiceRead ******** end")
	return nil
}
