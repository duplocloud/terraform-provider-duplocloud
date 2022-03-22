package duplocloud

import (
	"context"
	"encoding/json"
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
		"hpa_specs": {
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
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
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
		"lb_synced_deployment": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"any_host_allowed": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"cloud_creds_from_k8s_service_account": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"is_daemonset": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
	}
}

func dataSourceDuploService() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploServiceRead,
		Schema:      duploServiceComputedSchema(),
	}
}

func dataSourceDuploServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourceDuploServiceRead: start")

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.ReplicationControllerGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to read tenant %s service '%s': %s", tenantID, name, err)
	}
	if duplo == nil {
		return diag.Errorf("Unable to read tenant %s service '%s': not found", tenantID, name)
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))

	// Read the object into state
	flattenDuploService(d, duplo)

	log.Printf("[TRACE] dataSourceDuploServiceRead: end")
	return nil

}

func flattenDuploService(d *schema.ResourceData, duplo *duplosdk.DuploReplicationController) {

	// Apply TF state
	d.Set("name", duplo.Name)
	d.Set("volumes", duplo.Volumes)
	d.Set("lb_synced_deployment", duplo.IsLBSyncedDeployment)
	d.Set("any_host_allowed", duplo.IsAnyHostAllowed)
	d.Set("cloud_creds_from_k8s_service_account", duplo.IsCloudCredsFromK8sServiceAccount)
	d.Set("is_daemonset", duplo.IsDaemonset)
	d.Set("replicas_matching_asg_name", duplo.ReplicasMatchingAsgName)
	d.Set("replicas", duplo.Replicas)
	d.Set("tags", keyValueToState("tags", duplo.Tags))
	if len(duplo.HPASpecs) > 0 {
		flattenHPASpecs("hpa_specs", duplo.HPASpecs, d)
	}

	// If we have a pod template, read data from it
	if duplo.Template != nil {
		d.Set("agent_platform", duplo.Template.AgentPlatform)
		d.Set("cloud", duplo.Template.Cloud)
		d.Set("other_docker_host_config", duplo.Template.OtherDockerHostConfig)
		d.Set("other_docker_config", duplo.Template.OtherDockerConfig)
		d.Set("allocation_tags", duplo.Template.AllocationTags)
		d.Set("extra_config", duplo.Template.ExtraConfig)
		d.Set("commands", duplo.Template.Commands)

		// If there is at least one container, get the first docker image from it.
		if duplo.Template.Containers != nil && len(*duplo.Template.Containers) > 0 {
			d.Set("docker_image", (*duplo.Template.Containers)[0].Image)
		}
	}
}

func flattenHPASpecs(field string, from interface{}, to *schema.ResourceData) {
	var err error
	var encoded []byte

	if encoded, err = json.Marshal(from); err == nil {
		err = to.Set(field, string(encoded))
	}

	if err != nil {
		log.Printf("[DEBUG] flattenHPASpecs: failed to serialize %s to JSON: %s", field, err)
	}
}
