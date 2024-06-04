package duplocloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func k8sSecretSchemaComputed() map[string]*schema.Schema {
	schema := k8sSecretSchema()
	delete(schema, "client_secret_version")

	for k, v := range schema {
		if k != "secret_name" && k != "tenant_id" {
			v.Computed = true
			v.Optional = false
			v.Required = false
			v.DiffSuppressFunc = nil
			v.ValidateDiagFunc = nil

			//nolint:staticcheck // even though it is deprecated, we still must nil it
			v.ValidateFunc = nil
		}
	}

	return schema
}

// SCHEMA for resource data/search
func dataSourceK8Secret() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceK8SecretRead,
		Schema:      k8sSecretSchemaComputed(),
	}
}

func dataSourceK8SecretRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("secret_name").(string)

	log.Printf("[TRACE] dataSourceK8SecretRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	usrrp, err := c.UserInfo()
	if err != nil {
		return diag.FromErr(err)
	}
	rp, err := c.K8SecretGet(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if rp == nil || rp.SecretName == "" {
		return diag.Errorf("tenant k8 secret '%s' not found", name)
	}
	// Convert the results into TF state.
	flattenK8sSecret(d, rp, usrrp.IsReadOnly)
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))

	log.Printf("[TRACE] dataSourceK8SecretRead(%s, %s): end", tenantID, name)

	return nil
}
