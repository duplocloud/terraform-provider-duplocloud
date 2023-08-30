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

func duploAwsRdsTagSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the RDS tag will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"resource_type": {
			Description: "The type of the RDS resource to manage the tag for. Valid values are `cluster` and `db`.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Required:    true,
			ValidateFunc: validation.StringInSlice([]string{
				duplosdk.RDS_TYPE_CLUSTER,
				duplosdk.RDS_TYPE_DB,
			}, false),
		},
		"resource_id": {
			Description: "The ID of the RDS resource to manage the tag for.",
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

func resourceAwsRdsTag() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_rds_tag` manages a AWS RDS tags in Duplo.",

		ReadContext:   resourceAwsRdsTagRead,
		CreateContext: resourceAwsRdsTagCreate,
		UpdateContext: resourceAwsRdsTagUpdate,
		DeleteContext: resourceAwsRdsTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsRdsTagSchema(),
	}
}

func resourceAwsRdsTagRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, resourceType, resourceId, tagKey, err := parseAwsRdsTagIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsRdsTagRead(%s, %s, %s, %s): start", tenantID, resourceType, resourceId, tagKey)

	c := m.(*duplosdk.Client)

	tag, clientErr := c.RdsTagGetV3(tenantID, duplosdk.DuploRDSTag{
		ResourceType: resourceType,
		ResourceId:   resourceId,
		Key:          tagKey,
	})
	if tag == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve rds tag - (Tenant: %s,  ResourceType: %s, ResourceId: %s, TagKey: %s) : %s", tenantID, resourceType, resourceId, tagKey, clientErr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("resource_type", resourceType)
	d.Set("resource_id", resourceId)
	d.Set("key", tagKey)
	d.Set("value", tag.Value)
	log.Printf("[TRACE] resourceAwsRdsTagRead(%s, %s, %s, %s): end", tenantID, resourceType, resourceId, tagKey)
	return nil
}

func resourceAwsRdsTagCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	tag := duplosdk.DuploRDSTag{
		ResourceType: d.Get("resource_type").(string),
		ResourceId:   d.Get("resource_id").(string),
		Key:          d.Get("key").(string),
		Value:        d.Get("value").(string),
	}
	log.Printf("[TRACE] resourceAwsRdsTagCreate(%s, %s, %s, %s): start", tenantID, tag.ResourceType, tag.ResourceId, tag.Key)
	c := m.(*duplosdk.Client)

	err := c.RdsTagCreateV3(tenantID, tag)
	if err != nil {
		return diag.Errorf("Error creating rds tag - (Tenant: %s,  ResourceType: %s, ResourceId: %s, TagKey: %s) : %s", tenantID, tag.ResourceType, tag.ResourceId, tag.Key, err)
	}
	id := fmt.Sprintf("%s/%s/%s/%s", tenantID, tag.ResourceType, tag.ResourceId, tag.Key)

	diags := waitForResourceToBePresentAfterCreate(ctx, d, "RDS Tag", id, func() (interface{}, duplosdk.ClientError) {
		return c.RdsTagGetV3(tenantID, tag)
	})
	if diags != nil {
		return diags
	}

	d.SetId(id)

	diags = resourceAwsRdsTagRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsRdsTagCreate(%s, %s, %s, %s): end", tenantID, tag.ResourceType, tag.ResourceId, tag.Key)
	return diags
}

func resourceAwsRdsTagUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, resourceType, resourceId, tagKey, err := parseAwsRdsTagIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsRdsTagUpdate(%s, %s, %s, %s): start", tenantID, resourceType, resourceId, tagKey)

	c := m.(*duplosdk.Client)

	tag, clientErr := c.RdsTagGetV3(tenantID, duplosdk.DuploRDSTag{
		ResourceType: resourceType,
		ResourceId:   resourceId,
		Key:          tagKey,
	})
	if tag == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve rds tag - (Tenant: %s,  ResourceType: %s, ResourceId: %s, TagKey: %s) : %s", tenantID, resourceType, resourceId, tagKey, clientErr)
	}
	if d.HasChange("value") {
		newTag := duplosdk.DuploRDSTag{
			ResourceType: resourceType,
			ResourceId:   resourceId,
			Key:          tagKey,
			Value:        d.Get("value").(string),
		}
		err := c.RdsTagUpdateV3(tenantID, newTag)
		if err != nil {
			return diag.Errorf("Error updating rds tag - (Tenant: %s,  ResourceType: %s, ResourceId: %s, TagKey: %s) : %s", tenantID, newTag.ResourceType, newTag.ResourceId, newTag.Key, err)
		}

		diags := waitForResourceToBePresentAfterCreate(ctx, d, "RDS Tag", id, func() (interface{}, duplosdk.ClientError) {
			return c.RdsTagGetV3(tenantID, newTag)
		})
		if diags != nil {
			return diags
		}

		diags = resourceAwsRdsTagRead(ctx, d, m)

		log.Printf("[TRACE] resourceAwsRdsTagUpdate(%s, %s, %s, %s): end", tenantID, resourceType, resourceId, tagKey)

		return diags

	}
	log.Printf("[TRACE] resourceAwsRdsTagUpdate(%s, %s, %s, %s): end", tenantID, resourceType, resourceId, tagKey)
	return nil
}

func resourceAwsRdsTagDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, resourceType, resourceId, tagKey, err := parseAwsRdsTagIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsRdsTagDelete(%s, %s, %s, %s): start", tenantID, resourceType, resourceId, tagKey)

	c := m.(*duplosdk.Client)
	tag := duplosdk.DuploRDSTag{
		ResourceType: resourceType,
		ResourceId:   resourceId,
		Key:          tagKey,
	}
	clientErr := c.RdsTagDeleteV3(tenantID, tag)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete rds tag - (Tenant: %s,  ResourceType: %s, ResourceId: %s, TagKey: %s) : %s", tenantID, resourceType, resourceId, tagKey, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "RDS Tag", id, func() (interface{}, duplosdk.ClientError) {
		return c.RdsTagGetV3(tenantID, tag)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsRdsTagDelete(%s, %s, %s, %s): end", tenantID, resourceType, resourceId, tagKey)
	return nil
}

func parseAwsRdsTagIdParts(id string) (tenantID, resourceType, resourceId, tagKey string, err error) {
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) == 4 {
		tenantID, resourceType, resourceId, tagKey = idParts[0], idParts[1], idParts[2], idParts[3]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
