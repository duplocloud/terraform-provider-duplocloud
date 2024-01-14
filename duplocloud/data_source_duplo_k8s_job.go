package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceK8sJob() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceK8sJobRead,
		Schema:      resourceKubernetesJobV1Schema(true),
	}
}

func dataSourceK8sJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name, err := getK8sJobName(d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] dataSourceK8sJobRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rp, err := c.K8sJobGet(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	// convert the results into TF state
	err = flattenK8sJob(d, rp, m)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("v3/subscriptions/%s/k8s/job/%s", tenantID, name))

	log.Printf("[TRACE] dataSourceK8sJobRead(%s, %s): end", tenantID, name)

	return nil
}
