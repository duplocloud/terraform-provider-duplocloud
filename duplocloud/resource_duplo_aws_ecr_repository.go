package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAwsEcrRepositorySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the aws ecr repository will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the ECR Repository.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"arn": {
			Description: "Full ARN of the ECR repository.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"registry_id": {
			Description: "The registry ID where the repository was created.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"repository_url": {
			Description: "The URL of the repository.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"enable_tag_immutability": {
			Description: "The tag mutability setting for the repository. ",
			Type:        schema.TypeBool,
			Computed:    true,
			Optional:    true,
		},
		"enable_scan_image_on_push": {
			Description: "Indicates whether images are scanned after being pushed to the repository (true) or not scanned (false).",
			Type:        schema.TypeBool,
			Computed:    true,
			Optional:    true,
		},
		"kms_encryption_key": {
			Description: "The ARN of the KMS key to use.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
		"force_delete": {
			Description: "Whether to force delete the repository on destroy operations",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
	}
}

func resourceAwsEcrRepository() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_ecr_repository` manages an aws ecr repository in Duplo.",

		ReadContext:   resourceAwsEcrRepositoryRead,
		CreateContext: resourceAwsEcrRepositoryCreate,
		UpdateContext: resourceAwsEcrRepositoryUpdate,
		DeleteContext: resourceAwsEcrRepositoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsEcrRepositorySchema(),
	}
}

func resourceAwsEcrRepositoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsEcrRepositoryIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsEcrRepositoryRead(%s, %s): start", tenantID, name)

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	repository, err := c.AwsEcrRepositoryGet(tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}

	flattenEcrRepository(d, repository, tenantID)

	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))

	log.Printf("[TRACE] resourceAwsEcrRepositoryRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsEcrRepositoryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsEcrRepositoryCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAwsEcrRepository(d)
	err = c.AwsEcrRepositoryCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s aws ecr repository '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws ecr repository", id, func() (interface{}, duplosdk.ClientError) {
		return c.AwsEcrRepositoryGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsEcrRepositoryRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsEcrRepositoryCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsEcrRepositoryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// TODO - Update is not handled in backend.
	return nil
}

func resourceAwsEcrRepositoryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsEcrRepositoryIdParts(id)
	forceDelete := d.Get("force_delete").(bool)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsEcrRepositoryDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.AwsEcrRepositoryDelete(tenantID, name, forceDelete)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s aws ecr repository '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "aws ecr repository", id, func() (interface{}, duplosdk.ClientError) {
		return c.AwsEcrRepositoryGet(tenantID, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsEcrRepositoryDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAwsEcrRepository(d *schema.ResourceData) *duplosdk.DuploAwsEcrRepositoryRequest {
	return &duplosdk.DuploAwsEcrRepositoryRequest{
		Name:                  d.Get("name").(string),
		KmsEncryption:         d.Get("kms_encryption_key").(string),
		EnableTagImmutability: d.Get("enable_tag_immutability").(bool),
		EnableScanImageOnPush: d.Get("enable_scan_image_on_push").(bool),
	}
}

func parseAwsEcrRepositoryIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
