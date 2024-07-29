package duplocloud

import (
	"encoding/base64"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource crud
func dataSourceEksCredentials() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceEksCredentialsRead,

		Schema: map[string]*schema.Schema{
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_certificate_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// READ resource
func dataSourceEksCredentialsRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[TRACE] dataSourceEksCredentialsRead ******** start")

	// Get the data from Duplo.
	planID := d.Get("plan_id").(string)
	c := m.(*duplosdk.Client)
	infra, err := c.InfrastructureGetConfig(planID)
	if err != nil {
		return fmt.Errorf("failed to get plan %s kubernetes JIT access: %s", planID, err)
	}
	k8sConfig := &duplosdk.DuploEksCredentials{}
	// Now we know infra is not nil, proceed with checks
	if infra != nil && planID != "default" {
		if !infra.EnableK8Cluster && infra.Cloud != 2 {
			return fmt.Errorf("no kubernetes cluster for this plan %s", planID)
		} else if infra.Cloud == 2 {
			// Check for AksConfig only if Cloud is 2 (relevant scenario)
			if infra.AksConfig == nil || !infra.AksConfig.CreateAndManage {
				return fmt.Errorf("no kubernetes cluster for plan %s", planID)
			}
		}
		k8sConfig, err = c.GetPlanK8sJitAccess(planID)
		if err != nil && !err.PossibleMissingAPI() {
			return fmt.Errorf("failed to get plan %s kubernetes JIT access: %s", planID, err)
		}

		// If it failed, try the fallback method.
		if k8sConfig == nil {
			k8sConfig, err = c.GetK8sCredentials(planID)
			if err != nil {
				return fmt.Errorf("failed to read EKS credentials: %s", err)
			}
		}
	} else {
		// First, try the newer method of obtaining a JIT access token.

		k8sConfig, err = c.GetPlanK8sJitAccess(planID)
		if err != nil && !err.PossibleMissingAPI() {
			return fmt.Errorf("failed to get plan %s kubernetes JIT access: %s", planID, err)
		}

		// If it failed, try the fallback method.
		if k8sConfig == nil {
			k8sConfig, err = c.GetK8sCredentials(planID)
			if err != nil {
				return fmt.Errorf("failed to read EKS credentials: %s", err)
			}
		}
	}
	d.SetId(planID)

	// Set the Terraform resource data
	d.Set("plan_id", planID)
	d.Set("name", k8sConfig.Name)
	d.Set("endpoint", k8sConfig.APIServer)
	d.Set("token", k8sConfig.Token)
	d.Set("region", k8sConfig.AwsRegion)
	d.Set("version", k8sConfig.K8sVersion)

	bytes, err64 := base64.StdEncoding.DecodeString(k8sConfig.CertificateAuthorityDataBase64)
	if err64 == nil {
		d.Set("ca_certificate_data", string(bytes))
	}

	log.Printf("[TRACE] dataSourceEksCredentialsRead ******** end")
	return nil
}
