package duplocloud

import (
	"context"
	"fmt"
	"log"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceS3Bucket() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_s3_bucket` retrieves an S3 bucket in a Duplo tenant.",
		ReadContext: dataSourceS3BucketRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The GUID of the tenant that the S3 bucket resides in.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Description: "The short name of the S3 bucket.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"fullname": {
				Description: "The full name of the S3 bucket.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"arn": {
				Description: "The ARN of the S3 bucket.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"domain_name": {
				Description: "The domain name of the S3 bucket.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"enable_versioning": {
				Description: "Whether or not versioning is enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"enable_access_logs": {
				Description: "Whether or not access logs are enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"allow_public_access": {
				Description: "Whether or not public access is allowed.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"default_encryption": {
				Description: "Default encryption settings for objects uploaded to the bucket.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"method": {
							Description: "Default encryption method.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"region": {
				Description: "The region of the S3 bucket.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"managed_policies": {
				Description: "The managed policies applied to the bucket.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tags": {
				Description: "The tags assigned to this S3 bucket.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        KeyValueSchema(),
			},
		},
	}
}

func dataSourceS3BucketRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] dataSourceS3BucketRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)

	// Figure out the full resource name.
	features, _ := c.AdminGetSystemFeatures()
	fullName, err := c.GetDuploServicesNameWithAws(tenantID, name)
	if err != nil {
		return diag.Errorf("dataSourceS3BucketRead: Unable to retrieve duplo service name (tenant: %s, bucket: %s: error: %s)", tenantID, name, err)
	}
	if features != nil && features.IsTagsBasedResourceMgmtEnabled {
		fullName = features.S3BucketNamePrefix + name
	}

	// Get the object from Duplo
	duplo, clientErr := c.TenantGetV3S3Bucket(tenantID, fullName)
	if clientErr != nil && !clientErr.PossibleMissingAPI() {
		return diag.Errorf("dataSourceS3BucketRead: Unable to retrieve s3 bucket (tenant: %s, bucket: %s: error: %s)", tenantID, name, clientErr)
	}

	// Fallback on older api
	if clientErr != nil && clientErr.PossibleMissingAPI() {
		duplo, clientErr = c.TenantGetS3BucketSettings(tenantID, name)
		if clientErr != nil {
			return diag.Errorf("dataSourceS3BucketRead: Unable to retrieve s3 bucket settings (tenant: %s, bucket: %s: error: %s)", tenantID, name, clientErr)
		}
	}

	if duplo == nil {
		return diag.Errorf("dataSourceS3BucketRead: S3 bucket not found (tenant: %s, bucket: %s)", tenantID, name)
	}

	// Set the ID and data
	id := fmt.Sprintf("%s/%s", tenantID, name)
	d.SetId(id)
	flattenS3BucketData(d, tenantID, name, duplo)

	log.Printf("[TRACE] dataSourceS3BucketRead(%s, %s): end", tenantID, name)
	return nil
}

func flattenS3BucketData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploS3Bucket) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("domain_name", duplo.DomainName)
	d.Set("arn", duplo.Arn)
	d.Set("enable_versioning", duplo.EnableVersioning)
	d.Set("enable_access_logs", duplo.EnableAccessLogs)
	d.Set("allow_public_access", duplo.AllowPublicAccess)
	d.Set("default_encryption", []map[string]interface{}{{
		"method": duplo.DefaultEncryption,
	}})
	d.Set("managed_policies", duplo.Policies)
	d.Set("tags", keyValueToState("tags", duplo.Tags))
	d.Set("region", duplo.Region)
}
