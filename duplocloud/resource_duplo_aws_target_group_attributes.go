package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAwsTargetGroupAttributesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the aws target group attributes will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"target_group_arn": {
			Description:   "ARN of the Target Group.",
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"role_name", "port", "is_ecs_lb", "is_passthrough_lb"},
		},
		"role_name": {
			Description:   "Name of the ecs service or replication controller.",
			Type:          schema.TypeString,
			Optional:      true,
			ConflictsWith: []string{"target_group_arn"},
		},
		"port": {
			Description:   "Port used to connect with the target.",
			Type:          schema.TypeInt,
			Optional:      true,
			ConflictsWith: []string{"target_group_arn"},
		},
		"is_ecs_lb": {
			Description:   "Whether or not to look up the LB via an ECS service name instead of replication controller name.",
			Type:          schema.TypeBool,
			Optional:      true,
			ConflictsWith: []string{"target_group_arn"},
		},
		"is_passthrough_lb": {
			Description:   "Whether or not to look up the LB via the LB name instead of replication controller name.",
			Type:          schema.TypeBool,
			Optional:      true,
			ConflictsWith: []string{"target_group_arn"},
		},
		"attribute": {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"value": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
				},
			},
		},
	}
}

func resourceAwsTargetGroupAttributes() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_target_group_attributes` manages an aws target group attributes in Duplo.",

		ReadContext:   resourceAwsTargetGroupAttributesRead,
		CreateContext: resourceAwsTargetGroupAttributesCreate,
		UpdateContext: resourceAwsTargetGroupAttributesUpdate,
		DeleteContext: resourceAwsTargetGroupAttributesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsTargetGroupAttributesSchema(),
	}
}

func resourceAwsTargetGroupAttributesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var tenantID, targetGrpArn, roleName, port, isEcs, isPassthrough string
	var getReq duplosdk.DuploTargetGroupAttributesGetReq
	var err error
	id := d.Id()
	log.Printf("[TRACE] resourceAwsTargetGroupAttributesRead(%s): start", tenantID)

	log.Printf("[TRACE] Parsing Id (%s): start", id)
	if strings.Contains(id, "arn:aws:elasticloadbalancing") {
		tenantID, targetGrpArn, err = parseAwsTargetGroupAttributesIdPartsForTargetGrpArn(id)
		if err != nil {
			return diag.FromErr(err)
		}
		getReq = duplosdk.DuploTargetGroupAttributesGetReq{
			TargetGroupArn: targetGrpArn,
		}
	} else {
		tenantID, roleName, port, isEcs, isPassthrough, err = parseAwsTargetGroupAttributesIdParts(id)
		if err != nil {
			return diag.FromErr(err)
		}
		targetPort, _ := strconv.Atoi(port)
		isEcsBool, _ := strconv.ParseBool(isEcs)
		isPassthroughBool, _ := strconv.ParseBool(isPassthrough)
		getReq = duplosdk.DuploTargetGroupAttributesGetReq{
			RoleName:     roleName,
			Port:         targetPort,
			IsEcsLB:      isEcsBool,
			IsPassThruLB: isPassthroughBool,
		}
	}

	log.Printf("[TRACE] Parsing Id (%s): end", id)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.DuploAwsTargetGroupAttributesGet(tenantID, getReq)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s aws target group attributes %s", tenantID, clientErr)
	}

	if len(getReq.TargetGroupArn) > 0 {
		d.Set("target_group_arn", getReq.TargetGroupArn)
	} else {
		d.Set("role_name", getReq.RoleName)
		d.Set("port", getReq.Port)
		d.Set("is_ecs_lb", getReq.IsEcsLB)
		d.Set("is_passthrough_lb", getReq.IsPassThruLB)
	}
	d.Set("attribute", flattenAwsTargetGroupKeyValueAttributes(duplo))

	log.Printf("[TRACE] resourceAwsTargetGroupAttributesRead(%s): end", tenantID)
	return nil
}

func resourceAwsTargetGroupAttributesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	var id string
	tenantID := d.Get("tenant_id").(string)
	targteGrpArn := d.Get("target_group_arn").(string)
	roleName := d.Get("role_name").(string)
	port := d.Get("port").(int)
	isEcsLB := d.Get("is_ecs_lb").(bool)
	isPassThruLB := d.Get("is_passthrough_lb").(bool)
	log.Printf("[TRACE] resourceAwsTargetGroupAttributesCreate(%s): start", tenantID)
	c := m.(*duplosdk.Client)

	rq := expandAwsTargetGroupAttributes(d)
	err = c.DuploAwsTargetGroupAttributesCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s aws target group attributes: %s", tenantID, err)
	}

	if len(targteGrpArn) > 0 {
		id = fmt.Sprintf("%s/%s", tenantID, targteGrpArn)
	} else {
		id = fmt.Sprintf("%s/%s/%s/%s/%s", tenantID, roleName, strconv.Itoa(port), strconv.FormatBool(isEcsLB), strconv.FormatBool(isPassThruLB))
	}

	diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws target group attributes", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploAwsTargetGroupAttributesGet(tenantID, duplosdk.DuploTargetGroupAttributesGetReq{
			TargetGroupArn: targteGrpArn,
			RoleName:       roleName,
			Port:           port,
			IsEcsLB:        isEcsLB,
			IsPassThruLB:   isPassThruLB,
		})
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsTargetGroupAttributesRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsTargetGroupAttributesCreate(%s): end", tenantID)
	return diags
}

func resourceAwsTargetGroupAttributesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAwsTargetGroupAttributesCreate(ctx, d, m)
}

func resourceAwsTargetGroupAttributesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func expandAwsTargetGroupAttributes(d *schema.ResourceData) *duplosdk.DuploTargetGroupAttributes {
	attrs := &duplosdk.DuploTargetGroupAttributes{
		TargetGroupArn: d.Get("target_group_arn").(string),
		RoleName:       d.Get("role_name").(string),
		Port:           d.Get("port").(int),
		IsEcsLB:        d.Get("is_ecs_lb").(bool),
		IsPassThruLB:   d.Get("is_passthrough_lb").(bool),
	}

	if v, ok := d.GetOk("attribute"); ok {
		attrs.Attributes = expandAwsTargetGroupKeyValueAttributes(v.(*schema.Set).List())
	}
	return attrs
}

func expandAwsTargetGroupKeyValueAttributes(v interface{}) *[]duplosdk.DuploKeyStringValue {
	s := v.([]interface{})
	items := make([]duplosdk.DuploKeyStringValue, 0, len(s))
	for _, i := range s {
		kv := i.(map[string]interface{})
		items = append(items, duplosdk.DuploKeyStringValue{
			Key:   kv["key"].(string),
			Value: kv["value"].(string),
		})
	}
	return &items
}

func flattenAwsTargetGroupKeyValueAttributes(kv *[]duplosdk.DuploKeyStringValue) []map[string]interface{} {
	s := []map[string]interface{}{}
	for _, v := range *kv {
		m := map[string]interface{}{}
		m["key"] = v.Key
		m["value"] = v.Value
		s = append(s, m)
	}
	return s
}

func parseAwsTargetGroupAttributesIdPartsForTargetGrpArn(id string) (tenantID, targetGrpArn string, err error) {

	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, targetGrpArn = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func parseAwsTargetGroupAttributesIdParts(id string) (tenantID, roleName, port, isEcs, isPassthrough string, err error) {
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) == 5 {
		tenantID, roleName, port, isEcs, isPassthrough = idParts[0], idParts[1], idParts[2], idParts[3], idParts[4]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
