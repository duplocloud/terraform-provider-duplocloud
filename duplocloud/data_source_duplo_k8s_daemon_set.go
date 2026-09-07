package duplocloud

import (
	"context"
	"fmt"
	"log"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceK8sDaemonSet() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceK8sDaemonSetRead,
		Schema:      resourceKubernetesDaemonSetV1Schema(true),
	}
}

func dataSourceK8sDaemonSetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name, err := getK8sDaemonSetName(d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] dataSourceK8sDaemonSetRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	ds, cerr := c.K8sDaemonSetGet(tenantID, name)
	if cerr != nil {
		return diag.Errorf("Failed to read DaemonSet %s/%s. API error: %s", tenantID, name, cerr)
	}
	if ds == nil {
		return diag.Errorf("DaemonSet %s not found in tenant %s", name, tenantID)
	}

	d.Set("tenant_id", tenantID)
	d.Set("is_tenant_local", ds.IsTenantLocal)

	if metaErr := d.Set("metadata", flattenMetadata(ds.Metadata, d, m)); metaErr != nil {
		return diag.FromErr(metaErr)
	}

	specOut, specErr := flattenDaemonSetSpec(ds.Spec, d, m)
	if specErr != nil {
		return diag.FromErr(specErr)
	}
	if err := d.Set("spec", specOut); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("v3/subscriptions/%s/k8s/daemonSet/%s", tenantID, name))

	log.Printf("[TRACE] dataSourceK8sDaemonSetRead(%s, %s): end", tenantID, name)
	return nil
}
