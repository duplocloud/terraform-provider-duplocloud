package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceK8sCronJob() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceK8sCronJobRead,
		Schema:      resourceKubernetesJobV1Schema(),
	}
}

func dataSourceK8sCronJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] dataSourceK8sCronJobRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rp, err := c.K8sCronJobGet(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	// convert the results into TF state
	flattenK8sCronJob(d, rp)
	d.SetId(fmt.Sprintf("v3/subscriptions/%s/k8s/cronjob/%s", tenantID, name))

	log.Printf("[TRACE] dataSourceK8sCronJobRead(%s, %s): end", tenantID, name)

	return nil
}
