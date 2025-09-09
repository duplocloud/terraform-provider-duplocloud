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

func duploAwsTagSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the custom tag for a resource will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"arn": {
			Description: "The resource arn of which custom tag need to be created.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
		},
		"key": {
			Description: "The tag name.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
		},
		"value": {
			Description: "The value of the tag.",
			Type:        schema.TypeString,
			Required:    true,
		},
	}
}

func resourceAwsCustomTag() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_tag` manages an AWS custom tag for resources in Duplo.",

		ReadContext:   resourceAwsTagRead,
		CreateContext: resourceAwsTagCreate,
		UpdateContext: resourceAwsTagUpdate,
		DeleteContext: resourceAwsTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsTagSchema(),
	}
}

func resourceAwsTagRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantId, arn, key, err := parseAwsTagIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsTagRead(%s, %s, %s): start", tenantId, arn, key)

	c := m.(*duplosdk.Client)

	tag, clientErr := c.GetAWSTag(tenantId, arn, key)
	if tag == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve AWS tag - (Tenant: %s,  Arn: %s, TagKey: %s) : %s", tenantId, arn, key, clientErr)
	}

	d.Set("tenant_id", tenantId)
	d.Set("arn", arn)
	d.Set("key", key)
	d.Set("value", tag.Value)
	log.Printf("[TRACE] resourceAwsTagRead(%s, %s, %s): end", tenantId, arn, key)
	return nil
}

func resourceAwsTagCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	arn := d.Get("arn").(string)
	tag := &duplosdk.DuploAWSTag{
		Key:   d.Get("key").(string),
		Value: d.Get("value").(string),
	}
	log.Printf("[TRACE] resourceAwsTagCreate(%s,%s,%s): start", tenantID, arn, tag.Key)
	c := m.(*duplosdk.Client)

	err := c.CreateAWSTag(tenantID, arn, tag)
	if err != nil {
		return diag.Errorf("Error creating aws tag - (Tenant: %s,  arn: %s,  TagKey: %s) : %s", tenantID, arn, tag.Key, err)
	}
	id := fmt.Sprintf("%s/%s/%s", tenantID, arn, tag.Key)

	diags := waitForResourceToBePresentAfterCreate(ctx, d, "AWS Tag", id, func() (interface{}, duplosdk.ClientError) {
		return c.GetAWSTag(tenantID, arn, tag.Key)
	})
	if diags != nil {
		return diags
	}

	d.SetId(id)

	diags = resourceAwsTagRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsTagCreate(%s,%s,%s): end", tenantID, arn, tag.Key)
	return diags
}

func resourceAwsTagUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, arn, key, err := parseAwsTagIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsTagUpdate(%s, %s, %s): start", tenantID, arn, key)

	c := m.(*duplosdk.Client)

	tag, clientErr := c.GetAWSTag(tenantID, arn, key)
	if tag == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve aws tag - (Tenant: %s,  arn: %s,  TagKey: %s) : %s", tenantID, arn, key, clientErr)
	}
	if d.HasChange("value") {
		newTag := &duplosdk.DuploAWSTag{
			Key:   key,
			Value: d.Get("value").(string),
		}
		err := c.UpdateAWSTag(tenantID, arn, key, newTag)
		if err != nil {
			return diag.Errorf("Error updating aws tag - (Tenant: %s,  arn: %s,  TagKey: %s) : %s", tenantID, arn, key, err)
		}

		diags := waitForResourceToBePresentAfterCreate(ctx, d, "AWS Tag", id, func() (interface{}, duplosdk.ClientError) {
			return c.GetAWSTag(tenantID, arn, key)
		})
		if diags != nil {
			return diags
		}

		diags = resourceAwsTagRead(ctx, d, m)

		log.Printf("[TRACE] resourceAwsTagUpdate(%s, %s, %s): end", tenantID, arn, key)

		return diags

	}
	log.Printf("[TRACE] resourceAwsTagUpdate(%s, %s, %s): ed", tenantID, arn, key)
	return nil
}

func resourceAwsTagDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, arn, key, err := parseAwsTagIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsTagDelete(%s, %s, %s): start", tenantID, arn, key)

	c := m.(*duplosdk.Client)
	clientErr := c.DeleteAWSTag(tenantID, arn, key)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete aws tag - (Tenant: %s,  Arn: %s, TagKey: %s) : %s", tenantID, arn, key, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "AWS Tag", id, func() (interface{}, duplosdk.ClientError) {
		return c.GetAWSTag(tenantID, arn, key)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsTagDelete(%s, %s, %s): end", tenantID, arn, key)
	return nil
}

func parseAwsTagIdParts(id string) (tenantID, arn, tagKey string, err error) {
	idParts := strings.Split(id, "/")
	if len(idParts) == 3 {
		tenantID, arn, tagKey = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
