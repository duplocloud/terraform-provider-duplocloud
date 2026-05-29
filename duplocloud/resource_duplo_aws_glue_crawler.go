package duplocloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	glueCrawlerTypedFields = []glueTypedField{
		{tfKey: "name", jsonKey: "Name", identity: true},
		{tfKey: "role", jsonKey: "Role"},
	}
	glueCrawlerWrap = glueWrap{response: "Crawler"}
)

func resourceDuploAwsGlueCrawler() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_glue_crawler` manages an AWS Glue Crawler in a DuploCloud tenant. " +
			"The backend prepends `duploservices-<tenant>-` to the crawler name automatically. " +
			"The IAM role must have a trust relationship with `glue.amazonaws.com`.",
		CreateContext: resourceDuploAwsGlueCrawlerCreate,
		ReadContext:   resourceDuploAwsGlueCrawlerRead,
		UpdateContext: resourceDuploAwsGlueCrawlerUpdate,
		DeleteContext: resourceDuploAwsGlueCrawlerDelete,
		Importer:      &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": glueCommonTenantIDSchema(),
			"name": {
				Description: "The short name of the Glue crawler. The backend prepends `duploservices-<tenant>-`.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"role": {
				Description:      "The ARN (or short name) of the IAM role the crawler assumes to access source data.",
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: glueRoleARNSuppressDiff,
			},
			"fullname":  glueFullnameSchema(),
			"body_json": glueBodyJSONSchema(),
		},
	}
}

func resourceDuploAwsGlueCrawlerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceDuploAwsGlueCrawlerCreate(%s, %s): start", tenantID, name)

	body, _, err := glueBuildRequestBody(d, glueCrawlerTypedFields, glueCrawlerWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)
	if _, cerr := c.AwsGlueCrawlerCreate(tenantID, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("%s/%s", tenantID, name))
	return resourceDuploAwsGlueCrawlerRead(ctx, d, m)
}

func resourceDuploAwsGlueCrawlerRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueCrawlerRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rp, cerr := c.AwsGlueCrawlerGet(tenantID, name)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	if rp == nil {
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	if err := glueApplyResponse(d, rp, glueCrawlerTypedFields, glueCrawlerWrap); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDuploAwsGlueCrawlerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueCrawlerUpdate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	body, _, err := glueBuildRequestBody(d, glueCrawlerTypedFields, glueCrawlerWrap)
	if err != nil {
		return diag.FromErr(err)
	}
	prefixed, err := gluePrefixedName(c, tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	glueOverrideIdentity(body, glueCrawlerTypedFields, glueCrawlerWrap, prefixed)
	if _, cerr := c.AwsGlueCrawlerUpdate(tenantID, name, body); cerr != nil {
		return diag.FromErr(cerr)
	}
	return resourceDuploAwsGlueCrawlerRead(ctx, d, m)
}

func resourceDuploAwsGlueCrawlerDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := glueTwoPartID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceDuploAwsGlueCrawlerDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	if cerr := c.AwsGlueCrawlerDelete(tenantID, name); cerr != nil {
		return diag.FromErr(cerr)
	}
	return nil
}
