package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAwsS3TableSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the S3 table bucket will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the S3 table bucket.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the S3 table bucket as assigned by Duplo.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"arn": {
			Description: "The ARN of the S3 table bucket.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"owner_account_id": {
			Description: "The AWS account ID of the S3 table bucket owner.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"created_at": {
			Description: "The timestamp when the S3 table bucket was created.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceAwsS3Table() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_s3_table` manages an S3 table bucket in Duplo.",

		ReadContext:   resourceAwsS3TableRead,
		CreateContext: resourceAwsS3TableCreate,
		DeleteContext: resourceAwsS3TableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: duploAwsS3TableSchema(),
	}
}

func resourceAwsS3TableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsS3TableIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsS3TableRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	fullName, cerr := c.GetDuploServicesNameWithAws(tenantID, name)
	if cerr != nil {
		return diag.Errorf("resourceAwsS3TableRead: Unable to retrieve duplo service name (tenant: %s, name: %s, error: %s)", tenantID, name, cerr)
	}

	duplo, cerr := c.S3TableGet(tenantID, fullName)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[TRACE] resourceAwsS3TableRead(%s, %s): not found", tenantID, name)
			d.SetId("")
			return nil
		}
		return diag.FromErr(cerr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("arn", duplo.Arn)
	d.Set("owner_account_id", duplo.OwnerAccountId)
	d.Set("created_at", duplo.CreatedAt)

	log.Printf("[TRACE] resourceAwsS3TableRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsS3TableCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsS3TableCreate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	fullName, cerr := c.GetDuploServicesNameWithAws(tenantID, name)
	if cerr != nil {
		return diag.Errorf("resourceAwsS3TableCreate: Unable to retrieve duplo service name (tenant: %s, name: %s, error: %s)", tenantID, name, cerr)
	}

	_, cerr = c.S3TableCreate(tenantID, &duplosdk.DuploS3TableRequest{Name: name})
	if cerr != nil {
		return diag.Errorf("Error creating S3 table bucket '%s' in tenant %s: %s", name, tenantID, cerr)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "S3 table bucket", id, func() (interface{}, duplosdk.ClientError) {
		return c.S3TableGet(tenantID, fullName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsS3TableRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsS3TableCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsS3TableDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsS3TableIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsS3TableDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	fullName := d.Get("fullname").(string)

	cerr := c.S3TableDelete(tenantID, fullName)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[TRACE] resourceAwsS3TableDelete(%s, %s): not found", tenantID, name)
			return nil
		}
		return diag.Errorf("Unable to delete S3 table bucket '%s' in tenant %s: %s", name, tenantID, cerr)
	}

	log.Printf("[TRACE] resourceAwsS3TableDelete(%s, %s): end", tenantID, name)
	return nil
}

// ID format: <tenant_id>/<name>
func parseAwsS3TableIdParts(id string) (tenantID, name string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) == 2 {
		tenantID, name = parts[0], parts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
