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

func duploAwsSnsTopicSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the SNS topic will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the topic. Topic names must be made up of only uppercase and lowercase ASCII letters, numbers, underscores, and hyphens, and must be between 1 and 256 characters long.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
		},
		"kms_key_id": {
			Description: "The ID of an AWS-managed customer master key (CMK) for Amazon SNS or a custom CMK.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"arn": {
			Description: "The ARN of the SNS topic.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"fullname": {
			Description: "The full name of the SNS topic.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceAwsSnsTopic() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_sns_topic` manages a SNS topic in Duplo.",

		ReadContext:   resourceAwsSnsTopicRead,
		CreateContext: resourceAwsSnsTopicCreate,
		UpdateContext: resourceAwsSnsTopicUpdate,
		DeleteContext: resourceAwsSnsTopicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsSnsTopicSchema(),
	}
}

func resourceAwsSnsTopicRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, arn, err := parseAwsSnsTopicIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsSnsTopicRead(%s, %s): start", tenantID, arn)

	c := m.(*duplosdk.Client)

	accountID, err := c.TenantGetAwsAccountID(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	topic, clientErr := c.TenantGetSnsTopic(tenantID, arn)
	if topic == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s sns topic %s : %s", tenantID, arn, clientErr)
	}

	prefix, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("tenant_id", tenantID)
	d.Set("arn", topic.Name)
	parts := strings.Split(topic.Name, ":"+accountID+":")
	fullname := parts[1]
	d.Set("fullname", fullname)
	name, _ := duplosdk.UnprefixName(prefix, fullname)
	d.Set("name", name)
	// d.Set("kms_key_id", "") // TODO - Backend is not persisting this value.
	log.Printf("[TRACE] resourceAwsSnsTopicRead(%s, %s): end", tenantID, arn)
	return nil
}

func resourceAwsSnsTopicCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsSnsTopicCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAwsSnsTopic(d)
	resp, err := c.DuploSnsTopicCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s SNP topic '%s': %s", tenantID, name, err)
	}
	id := fmt.Sprintf("%s/%s", tenantID, resp.Name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "SNS Topic", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetSnsTopic(tenantID, resp.Name)
	})
	if diags != nil {
		return diags
	}

	d.SetId(id)

	diags = resourceAwsSnsTopicRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsSnsTopicCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsSnsTopicUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAwsSnsTopicDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, arn, err := parseAwsSnsTopicIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsSnsTopicDelete(%s, %s): start", tenantID, arn)

	c := m.(*duplosdk.Client)
	clientErr := c.DuploSnsTopicDelete(tenantID, arn)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s snp topic '%s': %s", tenantID, arn, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "SNS Topic", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetSnsTopic(tenantID, arn)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsSnsTopicDelete(%s, %s): end", tenantID, arn)
	return nil
}

func expandAwsSnsTopic(d *schema.ResourceData) *duplosdk.DuploSnsTopic {
	return &duplosdk.DuploSnsTopic{
		Name:     d.Get("name").(string),
		KmsKeyId: d.Get("kms_key_id").(string),
	}
}

func parseAwsSnsTopicIdParts(id string) (tenantID, arn string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, arn = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
