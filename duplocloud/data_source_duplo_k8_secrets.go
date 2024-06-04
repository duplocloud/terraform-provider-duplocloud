package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource data/search
func dataSourceK8Secrets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceK8SecretsRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"secrets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: k8sSecretSchemaComputed(),
				},
			},
		},
	}
}

func dataSourceK8SecretsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] dataSourceK8SecretsRead(%s): start", tenantID)

	c := m.(*duplosdk.Client)
	usrrp, err := c.UserInfo()
	if err != nil {
		return diag.FromErr(err)
	}
	rp, err := c.K8SecretGetList(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Convert the results into TF state.
	list := make([]map[string]interface{}, 0, len(*rp))
	for _, duplo := range *rp {

		// First, set the simple fields.
		sc := map[string]interface{}{
			"tenant_id":      duplo.TenantID,
			"secret_name":    duplo.SecretName,
			"secret_type":    duplo.SecretType,
			"secret_version": duplo.SecretVersion,
		}
		if usrrp.IsReadOnly {
			for key := range duplo.SecretData {
				duplo.SecretData[key] = "**********"
			}
		}
		// Next, set the JSON encoded strings.
		toJsonStringField("secret_data", duplo.SecretData, sc)
		toJsonStringField("secret_annotations", duplo.SecretAnnotations, sc)

		list = append(list, sc)
	}

	if err := d.Set("secrets", list); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tenantID)

	log.Printf("[TRACE] dataSourceK8SecretsRead(%s): start", tenantID)

	return nil
}
