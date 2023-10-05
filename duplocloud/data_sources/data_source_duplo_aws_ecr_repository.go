package data_sources

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource data/search
func dataSourceEcrRepository() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_ecr_repository` get ecr repository in a Duplo tenant.",
		ReadContext: dataSourceEcrRepositoryRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The GUID of the tenant in which to list the hosts.",
				Type:        schema.TypeString,
				Computed:    false,
				Required:    true,
			},
			"name": {
				Description: "The name of the ECR Repository.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_tag_immutability": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_scan_image_on_push": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"kms_encryption_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// READ/SEARCH resources
func dataSourceEcrRepositoryRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] dataSourceEcrRepositoryRead(%s): start", tenantID)

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	repository, err := c.AwsEcrRepositoryGet(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}

	flattenEcrRepository(d, repository, tenantID)

	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))

	log.Printf("[TRACE] dataSourceEmrClusterRead(%s): end", tenantID)
	return nil
}

func flattenEcrRepository(d *schema.ResourceData, duplo *duplosdk.DuploAwsEcrRepository, tenantID string) {
	log.Printf("[TRACE] dataSourceEmrClusterRead(%s): end", tenantID)
	d.Set("tenant_id", tenantID)
	d.Set("registry_id", duplo.RegistryId)
	d.Set("name", duplo.Name)
	d.Set("repository_url", duplo.RepositoryUri)
	d.Set("arn", duplo.Arn)
	d.Set("enable_tag_immutability", duplo.EnableTagImmutability)
	d.Set("enable_scan_image_on_push", duplo.EnableScanImageOnPush)
	d.Set("kms_encryption_key", duplo.KmsEncryption)
}
