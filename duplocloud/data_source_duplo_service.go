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
		"replica_collocation_allowed": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"cloud_creds_from_k8s_service_account": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"force_stateful_set": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"is_daemonset": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"is_unique_k8s_node_required": {
			Description: "Whether or not the replicas must be scheduled on separate Kubernetes nodes.  Only supported on Kubernetes.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"should_spread_across_zones": {
			Description: "Whether or not the replicas must be spread across availability zones.  Only supported on Kubernetes.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
		"fqdn": {
			Description: "The fully qualified domain associated with the service",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"fqdn_ex": {
			Description: "External fully qualified domain associated with the service",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"parent_domain": {
			Description: "The service's parent domain",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"domain": {
			Description: "The service domain (whichever fqdn_ex or fqdn which is non empty)",
			Type:        schema.TypeString,
			Computed:    true,
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
		return diag.FromErr(err)
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
	_ = d.Set("name", duplo.Name)
	_ = d.Set("volumes", duplo.Template.Volumes)
	_ = d.Set("lb_synced_deployment", duplo.IsLBSyncedDeployment)
	_ = d.Set("any_host_allowed", duplo.IsAnyHostAllowed)
	_ = d.Set("replica_collocation_allowed", duplo.IsReplicaCollocationAllowed)
	_ = d.Set("cloud_creds_from_k8s_service_account", duplo.IsCloudCredsFromK8sServiceAccount)
	_ = d.Set("is_daemonset", duplo.IsDaemonset)
	_ = d.Set("is_unique_k8s_node_required", duplo.IsUniqueK8sNodeRequired)
	_ = d.Set("should_spread_across_zones", duplo.ShouldSpreadAcrossZones)
	_ = d.Set("force_stateful_set", duplo.ForceStatefulSet)
	_ = d.Set("replicas_matching_asg_name", duplo.ReplicasMatchingAsgName)
	_ = d.Set("replicas", duplo.Replicas)
	_ = d.Set("index", duplo.Index)
	_ = d.Set("tags", keyValueToState("tags", duplo.Tags))
	_ = d.Set("fqdn", duplo.Fqdn)
	_ = d.Set("parent_domain", duplo.ParentDomain)
	_ = d.Set("fqdn_ex", duplo.FqdnEx)
	if duplo.FqdnEx != "" {
		_ = d.Set("domain", duplo.FqdnEx)
	} else {
		_ = d.Set("domain", duplo.Fqdn)
	}
	if len(duplo.HPASpecs) > 0 {
		flattenHPASpecs("hpa_specs", duplo.HPASpecs, d)
	}

	// If we have a pod template, read data from it
	if duplo.Template != nil {
		_ = d.Set("agent_platform", duplo.Template.AgentPlatform)
		_ = d.Set("cloud", duplo.Template.Cloud)
		_ = d.Set("other_docker_host_config", duplo.Template.OtherDockerHostConfig)
		_ = d.Set("other_docker_config", duplo.Template.OtherDockerConfig)
		_ = d.Set("allocation_tags", duplo.Template.AllocationTags)
		_ = d.Set("extra_config", duplo.Template.ExtraConfig)
		_ = d.Set("commands", duplo.Template.Commands)

		// If there is at least one container, get the first docker image from it.
		if duplo.Template.Containers != nil && len(*duplo.Template.Containers) > 0 {
			_ = d.Set("docker_image", (*duplo.Template.Containers)[0].Image)
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
