package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func targetGroupSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the target group will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Name of the target group.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"target_type": {
			Description: "Type of target that you must specify when registering targets with this target group.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"instance",
				"ip",
				"lambda",
				"alb",
			}, false),
		},
		"protocol": {
			Description: "Protocol to use to connect with the target. Not applicable when `target_type` is `lambda`.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Default:     "HTTP",
			ValidateFunc: validation.StringInSlice([]string{
				"HTTP",
				"HTTPS",
				"TCP",
				"TCP_UDP",
				"TLS",
				"UDP",
			}, false),
		},
		"port": {
			Description: "Port to use to connect with the target. Valid values are either ports 1-65535.",
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
		},
		"vpc_id": {
			Description: "Identifier of the VPC in which to create the target group. Required when `target_type` is `instance`, `ip` or `alb`. Does not apply when `target_type` is `lambda`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"ip_address_type": {
			Description: "The type of IP addresses used by the target group, only supported when target type is set to `ip`. Possible values are `ipv4` or `ipv6`",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"ipv4",
				"ipv6",
			}, false),
		},
		"protocol_version": {
			Description: "Only applicable when protocol is `HTTP` or `HTTPS`. The protocol version. Specify GRPC to send requests to targets using gRPC. Specify HTTP2 to send requests to targets using HTTP/2. The default is HTTP1, which sends requests to targets using HTTP/1.1",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			StateFunc: func(v interface{}) string {
				return strings.ToUpper(v.(string))
			},
			ValidateFunc: validation.StringInSlice([]string{
				"GRPC",
				"HTTP1",
				"HTTP2",
			}, true),
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				if d.Get("target_type").(string) == "lambda" {
					return true
				}
				switch d.Get("protocol").(string) {
				case "HTTP", "HTTPS":
					return false
				}
				return true
			},
		},

		"health_check": {
			Description: "Health Check configuration block.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enabled": {
						Description: "Whether health checks are enabled.",
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     true,
					},
					"healthy_threshold": {
						Description:  "Number of consecutive health checks successes required before considering an unhealthy target healthy.",
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      3,
						ValidateFunc: validation.IntBetween(2, 10),
					},
					"interval": {
						Description: "Approximate amount of time, in seconds, between health checks of an individual target. Minimum value 5 seconds, Maximum value 300 seconds. For lambda target groups, it needs to be greater as the `timeout` of the underlying `lambda`.",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     30,
					},
					"matcher": {
						Description: "Response codes to use when checking for a healthy responses from a target. You can specify multiple values (for example, \"200,202\" for HTTP(s) or \"0,12\" for GRPC) or a range of values (for example, \"200-299\" or \"0-99\"). Required for HTTP/HTTPS/GRPC ALB. Only applies to Application Load Balancers (i.e., HTTP/HTTPS/GRPC) not Network Load Balancers (i.e., TCP).",
						Type:        schema.TypeString,
						Computed:    true,
						Optional:    true,
					},
					"path": {
						Description: "Destination for the health check request. Required for HTTP/HTTPS ALB and HTTP NLB. Only applies to HTTP/HTTPS.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"port": {
						Description:      "Port to use to connect with the target. Valid values are either ports 1-65535, or traffic-port.",
						Type:             schema.TypeString,
						Optional:         true,
						Default:          "traffic-port",
						ValidateFunc:     validTargetGroupHealthCheckPort,
						DiffSuppressFunc: suppressIfTargetType("lambda"),
					},
					"protocol": {
						Description: " Protocol to use to connect with the target. Defaults to HTTP. Not applicable when target_type is lambda",
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "HTTP",
						StateFunc: func(v interface{}) string {
							return strings.ToUpper(v.(string))
						},
						ValidateFunc: validation.StringInSlice([]string{
							"HTTP",
							"HTTPS",
							"TCP",
						}, true),
						DiffSuppressFunc: suppressIfTargetType("lambda"),
					},
					"timeout": {
						Description:  "Amount of time, in seconds, during which no response means a failed health check.",
						Type:         schema.TypeInt,
						Optional:     true,
						Computed:     true,
						ValidateFunc: validation.IntBetween(2, 120),
					},
					"unhealthy_threshold": {
						Description:  "Number of consecutive health check failures required before considering the target unhealthy.",
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      3,
						ValidateFunc: validation.IntBetween(2, 10),
					},
				},
			},
		},
		"arn": {
			Description: "ARN of the Target Group.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceTargetGroup() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_lb_target_group` manages a target group in a Duplo tenant.",

		ReadContext:   resourceTargetGroupRead,
		CreateContext: resourceTargetGroupCreate,
		UpdateContext: resourceTargetGroupUpdate,
		DeleteContext: resourceTargetGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: targetGroupSchema(),
	}
}

func resourceTargetGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseTargetGroupIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceTargetGroupRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.DuploTargetGroupGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s target group %s : %s", tenantID, name, clientErr)
	}
	if rp == nil || rp.TargetGroupName == "" {
		d.SetId("")
		return nil
	}

	err = flattenTargetGroupResource(tenantID, d, rp)
	if err != nil {
		return diag.Errorf("Unable to flatten tenant %s target group %s : %s", tenantID, name, err)
	}
	log.Printf("[TRACE] resourceTargetGroupRead(%s, %s): end", tenantID, name)
	return nil
}

// CREATE resource
func resourceTargetGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] resourceTargetGroupCreate(%s, %s): start", tenantID, name)

	params := &duplosdk.DuploTargetGroup{
		Name:       name,
		TargetType: &duplosdk.DuploStringValue{Value: d.Get("target_type").(string)},
	}

	if d.Get("target_type").(string) != "lambda" {
		if _, ok := d.GetOk("port"); !ok {
			return diag.FromErr(fmt.Errorf("port should be set when target type is %s", d.Get("target_type").(string)))
		}

		if _, ok := d.GetOk("protocol"); !ok {
			return diag.FromErr(fmt.Errorf("protocol should be set when target type is %s", d.Get("target_type").(string)))
		}

		if _, ok := d.GetOk("vpc_id"); !ok {
			return diag.FromErr(fmt.Errorf("vpc_id should be set when target type is %s", d.Get("target_type").(string)))
		}
		params.Port = d.Get("port").(int)
		params.Protocol = &duplosdk.DuploStringValue{Value: d.Get("protocol").(string)}
		switch d.Get("protocol").(string) {
		case "HTTP", "HTTPS":
			params.ProtocolVersion = d.Get("protocol_version").(string)
		}
		params.VpcID = d.Get("vpc_id").(string)

		if d.Get("target_type").(string) == "ip" {
			if _, ok := d.GetOk("ip_address_type"); ok {
				params.IPAddressType.Value = d.Get("ip_address_type").(string)
			}
		}
		//Health check enabled must be true for target groups with target type "instance","ip" and "alb",
		params.HealthCheckEnabled = true
	}

	if healthChecks := d.Get("health_check").([]interface{}); len(healthChecks) == 1 {
		healthCheck := healthChecks[0].(map[string]interface{})

		params.HealthCheckEnabled = healthCheck["enabled"].(bool)

		params.HealthCheckIntervalSeconds = healthCheck["interval"].(int)

		params.HealthyThresholdCount = healthCheck["healthy_threshold"].(int)
		params.UnhealthyThresholdCount = healthCheck["unhealthy_threshold"].(int)
		t := healthCheck["timeout"].(int)
		if t != 0 {
			params.HealthCheckTimeoutSeconds = t
		}
		healthCheckProtocol := healthCheck["protocol"].(string)

		if healthCheckProtocol != "TCP" {
			p := healthCheck["path"].(string)
			if len(p) > 0 {
				params.HealthCheckPath = p
			}

			m := healthCheck["matcher"].(string)
			protocolVersion := d.Get("protocol_version").(string)
			if len(m) > 0 {
				if protocolVersion == "GRPC" {
					params.Matcher = &duplosdk.DuploTargetGroupMatcher{
						GRPCCode: m,
					}
				} else {
					params.Matcher = &duplosdk.DuploTargetGroupMatcher{
						HTTPCode: m,
					}
				}
			}
		}
		if d.Get("target_type").(string) != "lambda" {
			params.HealthCheckPort = healthCheck["port"].(string)
			params.HealthCheckProtocol = &duplosdk.DuploStringValue{
				Value: healthCheckProtocol,
			}
		}
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	cerr := c.DuploTargetGroupCreate(tenantID, params)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "target group", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploTargetGroupGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}

	d.SetId(id)

	diags = resourceTargetGroupRead(ctx, d, m)
	log.Printf("[TRACE] resourceTargetGroupCreate(%s, %s): end", tenantID, name)
	return diags
}

// UPDATE resource
func resourceTargetGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseTargetGroupIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceTargetGroupUpdate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	// Convert the Terraform resource data into a Duplo object
	if d.HasChange("health_check") {
		var params *duplosdk.DuploTargetGroupUpdateReq
		healthChecks := d.Get("health_check").([]interface{})
		if len(healthChecks) == 1 {
			healthCheck := healthChecks[0].(map[string]interface{})

			params = &duplosdk.DuploTargetGroupUpdateReq{
				TargetGroupArn:          d.Get("arn").(string),
				HealthCheckEnabled:      healthCheck["enabled"].(bool),
				HealthyThresholdCount:   healthCheck["healthy_threshold"].(int),
				UnhealthyThresholdCount: healthCheck["unhealthy_threshold"].(int),
			}

			t := healthCheck["timeout"].(int)
			if t != 0 {
				params.HealthCheckTimeoutSeconds = t
			}

			healthCheckProtocol := healthCheck["protocol"].(string)
			protocolVersion := d.Get("protocol_version").(string)
			if healthCheckProtocol != "TCP" && !d.IsNewResource() {
				if protocolVersion == "GRPC" {
					params.Matcher = &duplosdk.DuploTargetGroupMatcher{
						GRPCCode: healthCheck["matcher"].(string),
					}
				} else {
					params.Matcher = &duplosdk.DuploTargetGroupMatcher{
						HTTPCode: healthCheck["matcher"].(string),
					}
				}
				params.HealthCheckPath = healthCheck["path"].(string)
				params.HealthCheckIntervalSeconds = healthCheck["interval"].(int)
			}
			if d.Get("target_type").(string) != "lambda" {
				params.HealthCheckPort = healthCheck["port"].(string)
				params.HealthCheckProtocol = &duplosdk.DuploStringValue{Value: healthCheckProtocol}
			}
		}

		if params != nil {
			cerr := c.DuploTargetGroupUpdate(tenantID, name, params)
			if cerr != nil {
				return diag.FromErr(cerr)
			}
		}
	}

	diags := resourceTargetGroupRead(ctx, d, m)
	log.Printf("[TRACE] resourceTargetGroupUpdate(%s, %s): end", tenantID, name)
	return diags
}

// DELETE resource
func resourceTargetGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseTargetGroupIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceTargetGroupDelete(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.DuploTargetGroupGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s target group %s : %s", tenantID, name, clientErr)
	}
	if rp != nil && rp.TargetGroupName != "" {
		clientErr := c.DuploTargetGroupDelete(tenantID, name)
		if clientErr != nil {
			if clientErr.Status() == 404 {
				d.SetId("")
				return nil
			}
			return diag.Errorf("Unable to delete tenant %s target group %s : %s", tenantID, name, clientErr)
		}
	}

	log.Printf("[TRACE] resourceTargetGroupDelete(%s, %s): end", tenantID, name)
	return nil
}

func parseTargetGroupIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func suppressIfTargetType(t string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		return d.Get("target_type").(string) == t
	}
}

func validTargetGroupHealthCheckPort(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if value == "traffic-port" {
		return
	}

	port, err := strconv.Atoi(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q must be a valid port number (1-65536) or %q", k, "traffic-port"))
	}

	if port < 1 || port > 65536 {
		errors = append(errors, fmt.Errorf("%q must be a valid port number (1-65536) or %q", k, "traffic-port"))
	}

	return
}

func flattenTargetGroupResource(tenantId string, d *schema.ResourceData, targetGroup *duplosdk.DuploTargetGroup) error {
	d.Set("tenant_id", tenantId)
	d.Set("arn", targetGroup.TargetGroupArn)
	d.Set("name", targetGroup.TargetGroupName)
	d.Set("target_type", targetGroup.TargetType.Value)

	if err := d.Set("health_check", flattenLbTargetGroupHealthCheck(targetGroup)); err != nil {
		return fmt.Errorf("setting health_check: %w", err)
	}

	if v, _ := d.Get("target_type").(string); v != "lambda" {
		d.Set("vpc_id", targetGroup.VpcID)
		d.Set("port", targetGroup.Port)
		d.Set("protocol", targetGroup.Protocol.Value)
	}

	switch d.Get("protocol").(string) {
	case "HTTP", "HTTPS":
		d.Set("protocol_version", targetGroup.ProtocolVersion)
	}

	return nil
}

func flattenLbTargetGroupHealthCheck(targetGroup *duplosdk.DuploTargetGroup) []interface{} {
	if targetGroup == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"enabled":             targetGroup.HealthCheckEnabled,
		"healthy_threshold":   targetGroup.HealthyThresholdCount,
		"interval":            targetGroup.HealthCheckIntervalSeconds,
		"port":                targetGroup.HealthCheckPort,
		"protocol":            targetGroup.HealthCheckProtocol.Value,
		"timeout":             targetGroup.HealthCheckTimeoutSeconds,
		"unhealthy_threshold": targetGroup.UnhealthyThresholdCount,
	}

	if len(targetGroup.HealthCheckPath) > 0 {
		m["path"] = targetGroup.HealthCheckPath
	}
	if targetGroup.Matcher != nil && len(targetGroup.Matcher.HTTPCode) > 0 {
		m["matcher"] = targetGroup.Matcher.HTTPCode
	}
	if targetGroup.Matcher != nil && len(targetGroup.Matcher.GRPCCode) > 0 {
		m["matcher"] = targetGroup.Matcher.GRPCCode
	}

	return []interface{}{m}
}
