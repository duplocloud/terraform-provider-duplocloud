package duplocloud

import (
	"context"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strings"

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
				Type:      schema.TypeList,
				Computed:  true,
				Sensitive: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"secret_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"secret_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"client_secret_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"secret_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"secret_data": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"secret_annotations": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"secret_labels": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceK8SecretsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	log.Printf("[TRACE] dataSourceK8SecretsRead(%s): start", tenantID)

	c := m.(*duplosdk.Client)
	usrResp, err := c.UserInfo()
	if err != nil {
		return diag.FromErr(err)
	}
	tennantAccess, err := c.TenantAccessGet(usrResp.Username)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(tennantAccess) > 0 && !usrResp.IsReadOnly {
		for _, tenantAccessInfo := range tennantAccess {
			if tenantAccessInfo.TenantId == tenantID {
				if tenantAccessInfo.Policy.IsReadOnly {
					usrResp.IsReadOnly = true
				}
				break
			}
		}
	}
	access, err := c.SystemSettingGet("AllowReadonlyK8sSecrets")
	if err != nil {
		return diag.FromErr(err)
	}
	if access != nil && strings.ToLower(access.Value) == "true" {
		usrResp.IsReadOnly = false
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
			"secret_name":        duplo.SecretName,
			"secret_type":        duplo.SecretType,
			"secret_version":     duplo.SecretVersion,
			"secret_annotations": duplo.SecretAnnotations,
			"secret_labels":      duplo.SecretLabels,
		}
		if usrResp.IsReadOnly {
			for key := range duplo.SecretData {
				duplo.SecretData[key] = "**********"
			}
		}
		// Next, set the JSON encoded strings.
		toJsonStringField("secret_data", duplo.SecretData, sc)
		list = append(list, sc)
	}
	log.Printf("[TRACE] K8SecretGetList(%s): received response: %s", tenantID, list)

	if err := d.Set("secrets", list); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tenantID)

	log.Printf("[TRACE] dataSourceK8SecretsRead(%s): start", tenantID)

	return nil
}
