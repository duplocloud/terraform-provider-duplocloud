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

func duploAwsTargetGroupTargetMngrSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant where duplo will manage register and deregister of aws targets in target group.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"target_group_arn": {
			Description: "ARN of the Target Group.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"targets": {
			Description: "List of target id to be associated with target group",
			Type:        schema.TypeList,
			Required:    true,
			ForceNew:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Description: "The ID of the target. If the target type of the target group is INSTANCE, this is an instance ID. If the target type is IP , this is an IP address. If the target type is LAMBDA, this is the ARN of the Lambda function. If the target type is ALB, this is the ARN of the Application Load Balancer.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"port": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"availability_zone": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
	}
}

func resourceAwsTargetGroupTargetRegister() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_target_group_target_register` manages register or deregister targets in a target group in Duplo",

		ReadContext:   resourceAwsTargetGroupTargetRegisterRead,
		CreateContext: resourceAwsTargetGroupTargetRegisterCreate,
		//	UpdateContext: resourceAwsTargetGroupTargetRegisterUpdate,
		DeleteContext: resourceAwsTargetGroupTargetRegisterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsTargetGroupTargetMngrSchema(),
	}
}

func resourceAwsTargetGroupTargetRegisterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tokens := strings.Split(id, "/")
	tenantID, tgName := tokens[0], tokens[1]
	log.Printf("[TRACE] resourceAwsTargetGroupTargetRegisterRead(%s,%s): start", tenantID, tgName)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.DuploAwsTargetGroupTargetGet(tenantID, tgName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve target from target group %s for tenant %s : %s", tgName, tenantID, clientErr)
	}

	if duplo != nil {
		d.Set("target_group_arn", d.Get("target_group_arn"))
		if len(duplo.Targets) > 0 {
			o := []interface{}{}
			for _, target := range duplo.Targets {
				m := map[string]interface{}{}
				m["id"] = target.Id
				m["availability_zone"] = target.AvailabilityZone
				m["port"] = target.Port
				o = append(o, m)
			}
			d.Set("targets", o)
		}
	}

	log.Printf("[TRACE] resourceAwsTargetGroupTargetRegisterRead(%s,%s): end", tenantID, tgName)
	return nil
}

func resourceAwsTargetGroupTargetRegisterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	var id string
	tenantID := d.Get("tenant_id").(string)
	targteGrpArn := d.Get("target_group_arn").(string)
	tgToken := strings.Split(targteGrpArn, "/")
	tgName := tgToken[1]
	log.Printf("[TRACE] resourceAwsTargetGroupTargetRegisterCreate(%s,%s): start", tenantID, tgName)

	targets := d.Get("targets").([]interface{})
	req := &duplosdk.DuploTargetGroupTargetRegister{}
	for _, target := range targets {
		mp := target.(map[string]interface{})
		tId := duplosdk.DuploTargetId{
			Id: mp["id"].(string),
		}
		req.Targets = append(req.Targets, tId)
	}
	req.TargetGroupARN = targteGrpArn
	c := m.(*duplosdk.Client)

	err = c.DuploAwsTargetGroupTargetCreate(tenantID, tgName, req)
	if err != nil {
		return diag.Errorf("Error adding targets to target group for tenant %s : %s", tenantID, err)
	}

	id = fmt.Sprintf("%s/%s", tenantID, tgName)

	d.SetId(id)

	diags := resourceAwsTargetGroupTargetRegisterRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsTargetGroupTargetRegisterCreate(%s,%s): end", tenantID, tgName)
	return diags
}

/*
	func resourceAwsTargetGroupTargetRegisterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		return resourceAwsTargetGroupAttributesCreate(ctx, d, m)
	}
*/
func resourceAwsTargetGroupTargetRegisterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tokens := strings.Split(id, "/")
	tenantID, tgName := tokens[0], tokens[1]
	log.Printf("[TRACE] resourceAwsTargetGroupTargetRegisterDelete(%s,%s): start", tenantID, tgName)

	targteGrpArn := d.Get("target_group_arn").(string)

	targets := d.Get("targets").([]interface{})
	req := &duplosdk.DuploTargetGroupTargetRegister{}
	for _, target := range targets {
		mp := target.(map[string]interface{})
		tId := duplosdk.DuploTargetId{
			Id:               mp["id"].(string),
			AvailabilityZone: mp["availability_zone"].(string),
		}
		req.Targets = append(req.Targets, tId)
	}
	req.TargetGroupARN = targteGrpArn
	c := m.(*duplosdk.Client)

	err := c.DuploAwsTargetGroupTargetDelete(tenantID, tgName, req)
	if err != nil {
		return diag.Errorf("Error deleting targets from target group for tenant %s : %s", tenantID, err)
	}

	log.Printf("[TRACE] resourceAwsTargetGroupTargetRegisterDelete(%s,%s): end", tenantID, tgName)

	return nil
}
