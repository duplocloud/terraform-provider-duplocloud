package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAwsS3TableNamespaceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the S3 table namespace will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description:  "The name of the S3 table namespace. Must be 1–255 characters, lowercase letters, numbers, hyphens, and underscores, starting with a letter or number.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,254}$`), "Invalid namespace name: must be 1–255 characters, lowercase letters, numbers, hyphens, and underscores, starting with a letter or number."),
		},
		"table_bucket_name": {
			Description: "The full name of the S3 table bucket in which to create the namespace.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"owner_account_id": {
			Description: "The AWS account ID of the namespace owner.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"created_at": {
			Description: "The timestamp when the namespace was created.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceAwsS3TableNamespace() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_s3_table_namespace` manages an S3 table namespace in Duplo.",

		ReadContext:   resourceAwsS3TableNamespaceRead,
		CreateContext: resourceAwsS3TableNamespaceCreate,
		DeleteContext: resourceAwsS3TableNamespaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: duploAwsS3TableNamespaceSchema(),
	}
}

func resourceAwsS3TableNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, bucketName, name, err := parseAwsS3TableNamespaceIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsS3TableNamespaceRead(%s, %s, %s): start", tenantID, bucketName, name)

	c := m.(*duplosdk.Client)
	duplo, cerr := c.S3TableNamespaceGet(tenantID, bucketName, name)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[TRACE] resourceAwsS3TableNamespaceRead(%s, %s, %s): not found", tenantID, bucketName, name)
			d.SetId("")
			return nil
		}
		return diag.FromErr(cerr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("table_bucket_name", bucketName)
	d.Set("name", duplo.Name)
	d.Set("owner_account_id", duplo.OwnerAccountId)
	d.Set("created_at", duplo.CreatedAt)

	log.Printf("[TRACE] resourceAwsS3TableNamespaceRead(%s, %s, %s): end", tenantID, bucketName, name)
	return nil
}

func resourceAwsS3TableNamespaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	bucketName := d.Get("table_bucket_name").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsS3TableNamespaceCreate(%s, %s, %s): start", tenantID, bucketName, name)

	c := m.(*duplosdk.Client)
	_, cerr := c.S3TableNamespaceCreate(tenantID, bucketName, &duplosdk.DuploS3TableNamespaceRequest{Name: name})
	if cerr != nil {
		return diag.Errorf("Error creating S3 table namespace '%s' in tenant %s: %s", name, tenantID, cerr)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", tenantID, bucketName, name))

	diags := resourceAwsS3TableNamespaceRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsS3TableNamespaceCreate(%s, %s, %s): end", tenantID, bucketName, name)
	return diags
}

func resourceAwsS3TableNamespaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, bucketName, name, err := parseAwsS3TableNamespaceIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsS3TableNamespaceDelete(%s, %s, %s): start", tenantID, bucketName, name)

	c := m.(*duplosdk.Client)
	cerr := c.S3TableNamespaceDelete(tenantID, bucketName, name)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[TRACE] resourceAwsS3TableNamespaceDelete(%s, %s, %s): not found", tenantID, bucketName, name)
			return nil
		}
		return diag.Errorf("Unable to delete S3 table namespace '%s' in tenant %s: %s", name, tenantID, cerr)
	}

	log.Printf("[TRACE] resourceAwsS3TableNamespaceDelete(%s, %s, %s): end", tenantID, bucketName, name)
	return nil
}

// ID format: <tenant_id>/<table_bucket_name>/<namespace_name>
func parseAwsS3TableNamespaceIdParts(id string) (tenantID, bucketName, name string, err error) {
	parts := strings.SplitN(id, "/", 3)
	if len(parts) == 3 {
		tenantID, bucketName, name = parts[0], parts[1], parts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
