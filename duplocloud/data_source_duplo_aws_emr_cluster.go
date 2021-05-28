package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource data/search
func dataSourceEmrClusters() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_emr_cluster` lists EmrClusters in a Duplo tenant.",
		ReadContext: dataSourceEmrClusterRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The GUID of the tenant in which to list the hosts.",
				Type:        schema.TypeString,
				Computed:    false,
				Required:    true,
			},
			"data": {
				Description: "The list of native hosts.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "The  name of the emrCluster.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"tenant_id": {
							Description: "The GUID of the tenant that the emrCluster will be created in.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"arn": {
							Description: "The ARN of the emrCluster.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"status": {
							Description: "The status of the emrCluster.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"job_flow_id": {
							Description: "The job flow id of the emrCluster.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

/// READ/SEARCH resources
func dataSourceEmrClusterRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] dataSourceEmrClusterRead(%s): start", tenantID)

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	list, err := c.DuploEmrClusterGetList(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Populate the results from the list.
	clusters := make([]interface{}, 0, len(*list))
	for _, duplo := range *list {
		clusters = append(clusters, flattenEmrCluster(&duplo, tenantID))
	}

	if err := d.Set("data", clusters); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tenantID)

	log.Printf("[TRACE] dataSourceEmrClusterRead(%s): end", tenantID)
	return nil
}

func flattenEmrCluster(duplo *duplosdk.DuploEmrClusterSummary, tenantID string) map[string]interface{} {
	log.Printf("[TRACE] dataSourceEmrClusterRead(%s): end", tenantID)
	return map[string]interface{}{
		"tenant_id":   tenantID,
		"job_flow_id": duplo.JobFlowId,
		"name":        duplo.Name,
		"status":      duplo.Status,
		"arn":         duplo.Arn,
	}
}
