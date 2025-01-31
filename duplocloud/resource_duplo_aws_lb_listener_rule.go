package duplocloud

import (
	"context"
	"errors"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	listenerRulePriorityMin     = 1
	listenerRulePriorityMax     = 50000
	listenerRulePriorityDefault = 99999

	listenerActionOrderMin = 1
	listenerActionOrderMax = 50000
)

func awsLbListenerRuleSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the aws lb listener rule will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"listener_arn": {
			Description: "The ARN of the listener to which to attach the rule.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"arn": {
			Description: "The ARN of the rule.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"priority": {
			Description:  "The priority for the rule between `1` and `50000`. Leaving it unset will automatically set the rule with next available priority after currently existing highest rule. A listener can't have multiple rules with the same priority.",
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			ForceNew:     false,
			ValidateFunc: validListenerRulePriority,
		},
		"tags": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     KeyValueSchema(),
		},
		"action": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Description:  "The type of routing action. Valid values are `redirect`, `forward`, `fixed-response`, `authenticate-cognito` and `authenticate-oidc`",
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"redirect", "forward", "fixed-response", "authenticate-cognito", "authenticate-oidc"}, true),
					},
					"order": {
						Type:         schema.TypeInt,
						Optional:     true,
						Computed:     true,
						ValidateFunc: validation.IntBetween(listenerActionOrderMin, listenerActionOrderMax),
					},

					"target_group_arn": {
						Description:      "The ARN of the Target Group to which to route traffic. Specify only if `type` is `forward` and you want to route to a single target group. To route to one or more target groups, use a `forward` block instead.",
						Type:             schema.TypeString,
						Optional:         true,
						Computed:         true,
						DiffSuppressFunc: suppressIfActionTypeNot("forward"),
					},

					"forward": {
						Description:      "Information for creating an action that distributes requests among one or more target groups. Specify only if `type` is `forward`. If you specify both `forward` block and `target_group_arn` attribute, you can specify only one target group.",
						Type:             schema.TypeList,
						Optional:         true,
						Computed:         true,
						DiffSuppressFunc: suppressIfActionTypeNot("forward"),
						MaxItems:         1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"target_group": {
									Type:     schema.TypeSet,
									MinItems: 2,
									MaxItems: 5,
									Required: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"arn": {
												Description: "The Amazon Resource Name (ARN) of the target group.",
												Type:        schema.TypeString,
												Required:    true,
											},
											"weight": {
												Description:  "The weight. The range is 0 to 999.",
												Type:         schema.TypeInt,
												ValidateFunc: validation.IntBetween(0, 999),
												Default:      1,
												Optional:     true,
											},
										},
									},
								},
								"stickiness": {
									Type:     schema.TypeList,
									Optional: true,
									Computed: true,
									MaxItems: 1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"enabled": {
												Description: "Indicates whether target group stickiness is enabled.",
												Type:        schema.TypeBool,
												Optional:    true,
												Default:     false,
											},
											"duration": {
												Description:  "The time period, in seconds, during which requests from a client should be routed to the same target group.",
												Type:         schema.TypeInt,
												Required:     true,
												ValidateFunc: validation.IntBetween(1, 604800),
											},
										},
									},
								},
							},
						},
					},

					"redirect": {
						Description:      "Information for creating a redirect action. Required if `type` is `redirect`.",
						Type:             schema.TypeList,
						Optional:         true,
						Computed:         true,
						DiffSuppressFunc: suppressIfActionTypeNot("redirect"),
						MaxItems:         1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"host": {
									Description:  "The hostname. This component is not percent-encoded. The hostname can contain `#{host}`.",
									Type:         schema.TypeString,
									Optional:     true,
									Default:      "#{host}",
									ValidateFunc: validation.StringLenBetween(1, 128),
								},

								"path": {
									Description:  "The absolute path, starting with the leading \"/\".",
									Type:         schema.TypeString,
									Optional:     true,
									Default:      "/#{path}",
									ValidateFunc: validation.StringLenBetween(1, 128),
								},

								"port": {
									Description: "The port. Specify a value from `1` to `65535`.",
									Type:        schema.TypeString,
									Optional:    true,
									Default:     "#{port}",
								},

								"protocol": {
									Description: "The protocol. Valid values are `HTTP`, `HTTPS`, or `#{protocol}`.",
									Type:        schema.TypeString,
									Optional:    true,
									Default:     "#{protocol}",
									ValidateFunc: validation.StringInSlice([]string{
										"#{protocol}",
										"HTTP",
										"HTTPS",
									}, false),
								},

								"query": {
									Description:  "The query parameters, URL-encoded when necessary.",
									Type:         schema.TypeString,
									Optional:     true,
									Default:      "#{query}",
									ValidateFunc: validation.StringLenBetween(0, 128),
								},

								"status_code": {
									Description:  "The HTTP redirect code. The redirect is either permanent or temporary",
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringInSlice([]string{"HTTP_301", "HTTP_302"}, false),
								},
							},
						},
					},

					"fixed_response": {
						Description:      "Information for creating an action that returns a custom HTTP response. Required if `type` is `fixed-response`.",
						Type:             schema.TypeList,
						Optional:         true,
						Computed:         true,
						DiffSuppressFunc: suppressIfActionTypeNot("fixed-response"),
						MaxItems:         1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"content_type": {
									Description: "The content type. Valid values are `text/plain`, `text/css`, `text/html`, `application/javascript` and `application/json`",
									Type:        schema.TypeString,
									Required:    true,
									ValidateFunc: validation.StringInSlice([]string{
										"text/plain",
										"text/css",
										"text/html",
										"application/javascript",
										"application/json",
									}, false),
								},

								"message_body": {
									Description:  "The message body.",
									Type:         schema.TypeString,
									Optional:     true,
									Computed:     true,
									ValidateFunc: validation.StringLenBetween(0, 1024),
								},

								"status_code": {
									Description:  "The HTTP response code. Valid values are `2XX`, `4XX`, or `5XX`.",
									Type:         schema.TypeString,
									Optional:     true,
									Computed:     true,
									ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[245]\d\d$`), ""),
								},
							},
						},
					},

					"authenticate_cognito": {
						Description:      "Information for creating an authenticate action using Cognito. Required if `type` is `authenticate-cognito`.",
						Type:             schema.TypeList,
						Optional:         true,
						Computed:         true,
						DiffSuppressFunc: suppressIfActionTypeNot("authenticate-cognito"),
						MaxItems:         1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"authentication_request_extra_params": {
									Description: "The query parameters to include in the redirect request to the authorization endpoint.",
									Type:        schema.TypeMap,
									Optional:    true,
									Computed:    true,
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
								"on_unauthenticated_request": {
									Description:  "The behavior if the user is not authenticated. Valid values: `deny`, `allow` and `authenticate`.",
									Type:         schema.TypeString,
									Optional:     true,
									Computed:     true,
									ValidateFunc: validation.StringInSlice([]string{"allow", "authenticate", "deny"}, true),
								},
								"scope": {
									Description: "The set of user claims to be requested from the IdP.",
									Type:        schema.TypeString,
									Optional:    true,
									Default:     "openid",
								},
								"session_cookie_name": {
									Description: "The name of the cookie used to maintain session information.",
									Type:        schema.TypeString,
									Optional:    true,
									Default:     "AWSELBAuthSessionCookie",
								},
								"session_timeout": {
									Description: "The maximum duration of the authentication session, in seconds.",
									Type:        schema.TypeInt,
									Optional:    true,
									Default:     604800,
								},
								"user_pool_arn": {
									Description: "The ARN of the Cognito user pool.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"user_pool_client_id": {
									Description: "The ID of the Cognito user pool client.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"user_pool_domain": {
									Description: "The domain prefix or fully-qualified domain name of the Cognito user pool.",
									Type:        schema.TypeString,
									Required:    true,
								},
							},
						},
					},

					"authenticate_oidc": {
						Description:      "Information for creating an authenticate action using OIDC. Required if `type` is `authenticate-oidc`.",
						Type:             schema.TypeList,
						Optional:         true,
						Computed:         true,
						DiffSuppressFunc: suppressIfActionTypeNot("authenticate-oidc"),
						MaxItems:         1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"authentication_request_extra_params": {
									Description: "The query parameters to include in the redirect request to the authorization endpoint. Max: 10",
									Type:        schema.TypeMap,
									Optional:    true,
									Computed:    true,
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
								"authorization_endpoint": {
									Description: "The authorization endpoint of the IdP.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"client_id": {
									Description: "The OAuth 2.0 client identifier.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"client_secret": {
									Description: "The OAuth 2.0 client secret.",
									Type:        schema.TypeString,
									Required:    true,
									Sensitive:   true,
								},
								"issuer": {
									Description: "The OIDC issuer identifier of the IdP.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"on_unauthenticated_request": {
									Description:  "The behavior if the user is not authenticated. Valid values: `deny`, `allow` and `authenticate`.",
									Type:         schema.TypeString,
									Optional:     true,
									Computed:     true,
									ValidateFunc: validation.StringInSlice([]string{"allow", "authenticate", "deny"}, true),
								},
								"scope": {
									Description: "The set of user claims to be requested from the IdP.",
									Type:        schema.TypeString,
									Optional:    true,
									Default:     "openid",
								},
								"session_cookie_name": {
									Description: "The name of the cookie used to maintain session information.",
									Type:        schema.TypeString,
									Optional:    true,
									Default:     "AWSELBAuthSessionCookie",
								},
								"session_timeout": {
									Description: "The maximum duration of the authentication session, in seconds.",
									Type:        schema.TypeInt,
									Optional:    true,
									Default:     604800,
								},
								"token_endpoint": {
									Description: "The token endpoint of the IdP.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"user_info_endpoint": {
									Description: "The user info endpoint of the IdP.",
									Type:        schema.TypeString,
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
		"condition": {
			Description: "A Condition block. Multiple condition blocks of different types can be set and all must be satisfied for the rule to match.",
			Type:        schema.TypeSet,
			Required:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"host_header": {
						Type:     schema.TypeList,
						MaxItems: 1,
						Optional: true,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"values": {
									Description: "Contains a single `values` item which is a list of host header patterns to match. The maximum size of each pattern is 128 characters.",
									Type:        schema.TypeSet,
									Required:    true,
									MinItems:    1,
									Elem: &schema.Schema{
										Type:         schema.TypeString,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									Set: schema.HashString,
								},
							},
						},
					},
					"http_header": {
						Description: "HTTP headers to match.",
						Type:        schema.TypeList,
						MaxItems:    1,
						Optional:    true,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"http_header_name": {
									Description:  "Name of HTTP header to search. The maximum size is 40 characters.",
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validation.StringMatch(regexp.MustCompile("^[A-Za-z0-9!#$%&'*+-.^_`|~]{1,40}$"), ""),
								},
								"values": {
									Description: "List of header value patterns to match. Maximum size of each pattern is 128 characters.",
									Type:        schema.TypeSet,
									Elem: &schema.Schema{
										Type:         schema.TypeString,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									Required: true,
									Set:      schema.HashString,
								},
							},
						},
					},
					"http_request_method": {
						Description: "Contains a single `values` item which is a list of HTTP request methods or verbs to match. Maximum size is 40 characters.",
						Type:        schema.TypeList,
						MaxItems:    1,
						Optional:    true,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"values": {
									Type: schema.TypeSet,
									Elem: &schema.Schema{
										Type:         schema.TypeString,
										ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Za-z-_]{1,40}$`), ""),
									},
									Required: true,
									Set:      schema.HashString,
								},
							},
						},
					},
					"path_pattern": {
						Description: "Contains a single `values` item which is a list of path patterns to match against the request URL. Maximum size of each pattern is 128 characters.",
						Type:        schema.TypeList,
						MaxItems:    1,
						Optional:    true,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"values": {
									Type:     schema.TypeSet,
									Required: true,
									MinItems: 1,
									Elem: &schema.Schema{
										Type:         schema.TypeString,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									Set: schema.HashString,
								},
							},
						},
					},
					"query_string": {
						Description: "Query strings to match.",
						Type:        schema.TypeSet,
						Optional:    true,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"key": {
									Description: "Query string key pattern to match.",
									Type:        schema.TypeString,
									Optional:    true,
									Computed:    true,
								},
								"value": {
									Description: "Query string value pattern to match.",
									Type:        schema.TypeString,
									Required:    true,
								},
							},
						},
					},
					"source_ip": {
						Description: "Contains a single `values` item which is a list of source IP CIDR notations to match.",
						Type:        schema.TypeList,
						MaxItems:    1,
						Optional:    true,
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"values": {
									Type: schema.TypeSet,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
									Required: true,
									Set:      schema.HashString,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Resource for managing an AWS Listener Rule
func resourceAwsLbListenerRule() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_lb_listener_rule` manages an AWS Listener Rule in Duplo.",

		ReadContext:   resourceAwsLbListenerRuleRead,
		CreateContext: resourceAwsLbListenerRuleCreate,
		UpdateContext: resourceAwsLbListenerRuleUpdate,
		DeleteContext: resourceAwsLbListenerRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsLbListenerRuleSchema(),
	}
}

// READ resource
func resourceAwsLbListenerRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, ruleArn, err := parseAwsLbListenerRuleIdParts(id)
	listenerArn := getListenerArn(ruleArn)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsLbListenerRuleRead(%s, %s, %s): start", tenantID, listenerArn, ruleArn)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, clientErr := c.DuploAwsLbListenerRuleGet(tenantID, listenerArn, ruleArn)
	if clientErr != nil {
		if clientErr.Status() == 404 || (clientErr.Status() == 400 && clientErr.Response()["Message"] == "One or more listeners not found") { // TODO : remove second condition after backend API fixes.
			d.SetId("") // object missing
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s listener rule '%s': %s", tenantID, ruleArn, clientErr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("listener_arn", listenerArn)
	flattenAwsLbListenerRule(d, duplo)

	log.Printf("[TRACE] resourceAwsLbListenerRuleRead(%s, %s, %s): end", tenantID, listenerArn, ruleArn)
	return nil
}

// CREATE resource
func resourceAwsLbListenerRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	tenantID := d.Get("tenant_id").(string)
	listenerArn := d.Get("listener_arn").(string)
	log.Printf("[TRACE] resourceAwsLbListenerRuleCreate(%s, %s): start", tenantID, listenerArn)

	params := &duplosdk.DuploAwsLbListenerRuleCreateReq{
		ListenerArn: listenerArn,
		Priority:    d.Get("priority").(int),
	}

	var err error

	params.Actions, err = expandLbListenerActions(d.Get("action").([]interface{}))
	if err != nil {
		return diag.FromErr(fmt.Errorf("creating LB Listener Rule for Listener (%s): %w", listenerArn, err))
	}

	params.Conditions, err = expandLbListenerRuleConditions(d.Get("condition").(*schema.Set).List())
	if err != nil {
		return diag.FromErr(fmt.Errorf("creating LB Listener Rule for Listener (%s): %w", listenerArn, err))
	}

	params.Tags = keyValueFromState("tags", d)

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	resp, err := c.DuploAwsLbListenerRuleCreate(tenantID, params)
	if err != nil {
		return diag.Errorf("Error creating tenant %s listener rule '%s': %s", tenantID, listenerArn, err)
	}

	// Wait for Duplo to be able to return the cluster's details.
	id := fmt.Sprintf("%s/%s", tenantID, resp.RuleArn)
	log.Printf("[TRACE] Resource-Id : (%s)", id)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "listener rule", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploAwsLbListenerRuleGet(tenantID, listenerArn, resp.RuleArn)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsLbListenerRuleRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsLbListenerRuleCreate(%s, %s): end", tenantID, resp.RuleArn)
	return diags
}

func resourceAwsLbListenerRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, ruleArn, err := parseAwsLbListenerRuleIdParts(id)
	listenerArn := getListenerArn(ruleArn)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsLbListenerRuleUpdate(%s, %s, %s): start", tenantID, listenerArn, ruleArn)

	params := &duplosdk.DuploAwsLbListenerRule{
		RuleArn:     ruleArn,
		ListenerArn: listenerArn,
	}
	requestUpdate := false
	if d.HasChange("action") {
		var err error
		params.Actions, err = expandLbListenerActions(d.Get("action").([]interface{}))
		if err != nil {
			return diag.FromErr(fmt.Errorf("modifying LB Listener Rule (%s) action: %w", d.Id(), err))
		}
		requestUpdate = true
	}

	if d.HasChange("condition") {
		var err error
		params.Conditions, err = expandLbListenerRuleConditions(d.Get("condition").(*schema.Set).List())
		if err != nil {
			return diag.FromErr(fmt.Errorf("modifying LB Listener Rule (%s) condition: %w", d.Id(), err))
		}
		requestUpdate = true
	}

	if requestUpdate {
		c := m.(*duplosdk.Client)
		_, err := c.DuploAwsLbListenerRuleUpdate(tenantID, params)
		if err != nil {
			return diag.Errorf("Error updating tenant %s listener rule '%s': %s", tenantID, ruleArn, err)
		}
	}
	// Read the latest state.
	diags := resourceAwsLbListenerRuleRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsLbListenerRuleUpdate(%s, %s, %s): end", tenantID, listenerArn, ruleArn)

	return diags
}

// DELETE resource
func resourceAwsLbListenerRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, ruleArn, err := parseAwsLbListenerRuleIdParts(id)
	listenerArn := getListenerArn(ruleArn)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsLbListenerRuleDelete(%s, %s, %s): start", tenantID, listenerArn, ruleArn)

	c := m.(*duplosdk.Client)
	clientErr := c.DuploAwsLbListenerRuleDelete(tenantID, listenerArn, ruleArn)
	if clientErr != nil {
		if clientErr.Status() == 404 || (clientErr.Status() == 400 && clientErr.Response()["Message"] == "One or more listeners not found") { // TODO : remove second condition after backend API fixes.
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s listener rule '%s': %s", tenantID, ruleArn, clientErr)
	}

	// Wait up to 60 seconds for Duplo to delete the function.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "listener rule", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploAwsLbListenerRuleGet(tenantID, listenerArn, ruleArn)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsLbListenerRuleDelete(%s, %s, %s): end", tenantID, listenerArn, ruleArn)
	return nil
}

func parseAwsLbListenerRuleIdParts(id string) (tenantID, ruleArn string, err error) {
	idParts := strings.SplitN(id, "/", 6)
	log.Printf("[TRACE] idParts(%s)", idParts)
	if len(idParts) == 6 {
		tenantID = idParts[0]
		ruleArn = strings.Join([]string{idParts[1], idParts[2], idParts[3], idParts[4], idParts[5]}, "/")
		log.Printf("[TRACE] ruleArn(%s)", ruleArn)
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAwsLbListenerRule(d *schema.ResourceData, duplo *duplosdk.DuploAwsLbListenerRule) {
	d.Set("arn", duplo.RuleArn)
	if len(duplo.Priority) > 0 {
		p, _ := strconv.Atoi(duplo.Priority)
		d.Set("priority", p)
	}
	d.Set("tags", keyValueToState("tags", duplo.Tags))
	sort.Slice(*duplo.Actions, func(i, j int) bool {
		return (*duplo.Actions)[i].Order < (*duplo.Actions)[j].Order
	})
	actions := make([]interface{}, len(*duplo.Actions))
	for i, action := range *duplo.Actions {
		actionMap := make(map[string]interface{})
		actionMap["type"] = action.Type.Value
		actionMap["order"] = action.Order

		switch actionMap["type"] {
		case "forward":
			if action.TargetGroupArn != "" {
				actionMap["target_group_arn"] = action.TargetGroupArn
			} else {
				targetGroups := make([]map[string]interface{}, 0, len(*action.ForwardConfig.TargetGroups))
				for _, targetGroup := range *action.ForwardConfig.TargetGroups {
					targetGroups = append(targetGroups,
						map[string]interface{}{
							"arn":    targetGroup.TargetGroupArn,
							"weight": targetGroup.Weight,
						},
					)
				}
				actionMap["forward"] = []map[string]interface{}{
					{
						"target_group": targetGroups,
						"stickiness": []map[string]interface{}{
							{
								"enabled":  action.ForwardConfig.TargetGroupStickinessConfig.Enabled,
								"duration": action.ForwardConfig.TargetGroupStickinessConfig.DurationSeconds,
							},
						},
					},
				}
			}

		case "redirect":
			actionMap["redirect"] = []map[string]interface{}{
				{
					"host":        action.RedirectConfig.Host,
					"path":        action.RedirectConfig.Path,
					"port":        action.RedirectConfig.Port,
					"protocol":    action.RedirectConfig.Protocol,
					"query":       action.RedirectConfig.Query,
					"status_code": action.RedirectConfig.StatusCode.Value,
				},
			}

		case "fixed-response":
			actionMap["fixed_response"] = []map[string]interface{}{
				{
					"content_type": action.FixedResponseConfig.ContentType,
					"message_body": action.FixedResponseConfig.MessageBody,
					"status_code":  action.FixedResponseConfig.StatusCode,
				},
			}

		case "authenticate-cognito":
			authenticationRequestExtraParams := make(map[string]interface{})
			for key, value := range action.AuthenticateCognitoConfig.AuthenticationRequestExtraParams {
				authenticationRequestExtraParams[key] = value
			}

			actionMap["authenticate_cognito"] = []map[string]interface{}{
				{
					"authentication_request_extra_params": authenticationRequestExtraParams,
					"on_unauthenticated_request":          action.AuthenticateCognitoConfig.OnUnauthenticatedRequest.Value,
					"scope":                               action.AuthenticateCognitoConfig.Scope,
					"session_cookie_name":                 action.AuthenticateCognitoConfig.SessionCookieName,
					"session_timeout":                     action.AuthenticateCognitoConfig.SessionTimeout,
					"user_pool_arn":                       action.AuthenticateCognitoConfig.UserPoolArn,
					"user_pool_client_id":                 action.AuthenticateCognitoConfig.UserPoolClientId,
					"user_pool_domain":                    action.AuthenticateCognitoConfig.UserPoolDomain,
				},
			}

		case "authenticate-oidc":
			authenticationRequestExtraParams := make(map[string]interface{})
			for key, value := range action.AuthenticateOidcConfig.AuthenticationRequestExtraParams {
				authenticationRequestExtraParams[key] = value
			}

			// The LB API currently provides no way to read the ClientSecret
			// Instead we passthrough the configuration value into the state
			clientSecret := d.Get("action." + strconv.Itoa(i) + ".authenticate_oidc.0.client_secret").(string)

			actionMap["authenticate_oidc"] = []map[string]interface{}{
				{
					"authentication_request_extra_params": authenticationRequestExtraParams,
					"authorization_endpoint":              action.AuthenticateOidcConfig.AuthorizationEndpoint,
					"client_id":                           action.AuthenticateOidcConfig.ClientId,
					"client_secret":                       clientSecret,
					"issuer":                              action.AuthenticateOidcConfig.Issuer,
					"on_unauthenticated_request":          action.AuthenticateOidcConfig.OnUnauthenticatedRequest.Value,
					"scope":                               action.AuthenticateOidcConfig.Scope,
					"session_cookie_name":                 action.AuthenticateOidcConfig.SessionCookieName,
					"session_timeout":                     action.AuthenticateOidcConfig.SessionTimeout,
					"token_endpoint":                      action.AuthenticateOidcConfig.TokenEndpoint,
					"user_info_endpoint":                  action.AuthenticateOidcConfig.UserInfoEndpoint,
				},
			}
		}

		actions[i] = actionMap
	}
	d.Set("action", actions)

	conditions := make([]interface{}, len(*duplo.Conditions))
	for i, condition := range *duplo.Conditions {
		conditionMap := make(map[string]interface{})

		switch condition.Field {
		case "host-header":
			conditionMap["host_header"] = []interface{}{
				map[string]interface{}{
					"values": flattenStringSet(condition.HostHeaderConfig.Values),
				},
			}

		case "http-header":
			conditionMap["http_header"] = []interface{}{
				map[string]interface{}{
					"http_header_name": condition.HttpHeaderConfig.HttpHeaderName,
					"values":           flattenStringSet(condition.HttpHeaderConfig.Values),
				},
			}

		case "http-request-method":
			conditionMap["http_request_method"] = []interface{}{
				map[string]interface{}{
					"values": flattenStringSet(condition.HttpRequestMethodConfig.Values),
				},
			}

		case "path-pattern":
			conditionMap["path_pattern"] = []interface{}{
				map[string]interface{}{
					"values": flattenStringSet(condition.PathPatternConfig.Values),
				},
			}

		case "query-string":
			values := make([]interface{}, len(*condition.QueryStringConfig.Values))
			for k, value := range *condition.QueryStringConfig.Values {
				values[k] = map[string]interface{}{
					"key":   value.Key,
					"value": value.Value,
				}
			}
			conditionMap["query_string"] = values

		case "source-ip":
			conditionMap["source_ip"] = []interface{}{
				map[string]interface{}{
					"values": flattenStringSet(condition.SourceIpConfig.Values),
				},
			}
		}

		conditions[i] = conditionMap
	}
	d.Set("condition", conditions)
}

func expandLbListenerActions(l []interface{}) (*[]duplosdk.DuploAwsLbListenerRuleAction, error) {
	if len(l) == 0 {
		return nil, nil
	}

	var actions []duplosdk.DuploAwsLbListenerRuleAction
	var err error

	for i, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		action := duplosdk.DuploAwsLbListenerRuleAction{
			Order: i + 1,
			Type:  &duplosdk.DuploStringValue{Value: tfMap["type"].(string)},
		}

		if order, ok := tfMap["order"].(int); ok && order != 0 {
			action.Order = order
		}

		switch tfMap["type"].(string) {
		case "forward":
			if v, ok := tfMap["target_group_arn"].(string); ok && v != "" {
				action.TargetGroupArn = v
			} else if v, ok := tfMap["forward"].([]interface{}); ok {
				action.ForwardConfig = expandLbListenerActionForwardConfig(v)
			} else {
				err = errors.New("for actions of type 'forward', you must specify a 'forward' block or 'target_group_arn'")
			}

		case "redirect":
			if v, ok := tfMap["redirect"].([]interface{}); ok {
				action.RedirectConfig = expandLbListenerRedirectActionConfig(v)
			} else {
				err = errors.New("for actions of type 'redirect', you must specify a 'redirect' block")
			}

		case "fixed-response":
			if v, ok := tfMap["fixed_response"].([]interface{}); ok {
				action.FixedResponseConfig = expandLbListenerFixedResponseConfig(v)
			} else {
				err = errors.New("for actions of type 'fixed-response', you must specify a 'fixed_response' block")
			}

		case "authenticate-cognito":
			if v, ok := tfMap["authenticate_cognito"].([]interface{}); ok {
				action.AuthenticateCognitoConfig = expandLbListenerAuthenticateCognitoConfig(v)
			} else {
				err = errors.New("for actions of type 'authenticate-cognito', you must specify a 'authenticate_cognito' block")
			}

		case "authenticate-oidc":
			if v, ok := tfMap["authenticate_oidc"].([]interface{}); ok {
				action.AuthenticateOidcConfig = expandAuthenticateOIDCConfig(v)
			} else {
				err = errors.New("for actions of type 'authenticate-oidc', you must specify a 'authenticate_oidc' block")
			}
		}

		actions = append(actions, action)
	}

	return &actions, err
}
func expandLbListenerActionForwardConfig(l []interface{}) *duplosdk.DuploAwsLbListenerRuleActionForwardConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &duplosdk.DuploAwsLbListenerRuleActionForwardConfig{}

	if v, ok := tfMap["target_group"].(*schema.Set); ok && v.Len() > 0 {
		config.TargetGroups = expandLbListenerActionForwardConfigTargetGroups(v.List())
	}

	if v, ok := tfMap["stickiness"].([]interface{}); ok && len(v) > 0 {
		config.TargetGroupStickinessConfig = expandLbListenerActionForwardConfigTargetGroupStickinessConfig(v)
	}

	return config
}

func expandLbListenerActionForwardConfigTargetGroups(l []interface{}) *[]duplosdk.DuploTargetGroupTuple {
	if len(l) == 0 {
		return nil
	}

	groups := []duplosdk.DuploTargetGroupTuple{}

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue

		}

		group := duplosdk.DuploTargetGroupTuple{
			TargetGroupArn: tfMap["arn"].(string),
			Weight:         tfMap["weight"].(int),
		}

		groups = append(groups, group)
	}

	return &groups
}

func expandLbListenerActionForwardConfigTargetGroupStickinessConfig(l []interface{}) *duplosdk.DuploTargetGroupStickinessConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return &duplosdk.DuploTargetGroupStickinessConfig{
		Enabled:         tfMap["enabled"].(bool),
		DurationSeconds: tfMap["duration"].(int),
	}
}

func expandLbListenerRedirectActionConfig(l []interface{}) *duplosdk.DuploAwsLbListenerRuleActionRedirectConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	return &duplosdk.DuploAwsLbListenerRuleActionRedirectConfig{
		Host:       tfMap["host"].(string),
		Path:       tfMap["path"].(string),
		Port:       tfMap["port"].(string),
		Protocol:   tfMap["protocol"].(string),
		Query:      tfMap["query"].(string),
		StatusCode: &duplosdk.DuploStringValue{Value: tfMap["status_code"].(string)},
	}
}

func expandLbListenerFixedResponseConfig(l []interface{}) *duplosdk.DuploAwsLbListenerRuleActionFixedResponseConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	return &duplosdk.DuploAwsLbListenerRuleActionFixedResponseConfig{
		ContentType: tfMap["content_type"].(string),
		MessageBody: tfMap["message_body"].(string),
		StatusCode:  tfMap["status_code"].(string),
	}
}

func expandLbListenerAuthenticateCognitoConfig(l []interface{}) *duplosdk.DuploAwsLbListenerRuleActionAuthenticateCognitoConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	config := &duplosdk.DuploAwsLbListenerRuleActionAuthenticateCognitoConfig{
		AuthenticationRequestExtraParams: fieldToStringMap("authentication_request_extra_params", tfMap),
		UserPoolArn:                      tfMap["user_pool_arn"].(string),
		UserPoolClientId:                 tfMap["user_pool_client_id"].(string),
		UserPoolDomain:                   tfMap["user_pool_domain"].(string),
	}

	if v, ok := tfMap["on_unauthenticated_request"].(string); ok && v != "" {
		config.OnUnauthenticatedRequest = &duplosdk.DuploStringValue{Value: v}
	}

	if v, ok := tfMap["scope"].(string); ok && v != "" {
		config.Scope = v
	}

	if v, ok := tfMap["session_cookie_name"].(string); ok && v != "" {
		config.SessionCookieName = v
	}

	if v, ok := tfMap["session_timeout"].(int); ok && v != 0 {
		config.SessionTimeout = v
	}

	return config
}

func expandAuthenticateOIDCConfig(l []interface{}) *duplosdk.DuploAwsLbListenerRuleActionAuthenticateOidcConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	config := &duplosdk.DuploAwsLbListenerRuleActionAuthenticateOidcConfig{
		AuthenticationRequestExtraParams: fieldToStringMap("authentication_request_extra_params", tfMap),
		AuthorizationEndpoint:            tfMap["authorization_endpoint"].(string),
		ClientId:                         tfMap["client_id"].(string),
		ClientSecret:                     tfMap["client_secret"].(string),
		Issuer:                           tfMap["issuer"].(string),
		TokenEndpoint:                    tfMap["token_endpoint"].(string),
		UserInfoEndpoint:                 tfMap["user_info_endpoint"].(string),
	}

	if v, ok := tfMap["on_unauthenticated_request"].(string); ok && v != "" {
		config.OnUnauthenticatedRequest = &duplosdk.DuploStringValue{Value: v}
	}

	if v, ok := tfMap["scope"].(string); ok && v != "" {
		config.Scope = v
	}

	if v, ok := tfMap["session_cookie_name"].(string); ok && v != "" {
		config.SessionCookieName = v
	}

	if v, ok := tfMap["session_timeout"].(int); ok && v != 0 {
		config.SessionTimeout = v
	}

	return config
}

func expandLbListenerRuleConditions(conditions []interface{}) (*[]duplosdk.DuploAwsLbListenerRuleCondition, error) {
	elbConditions := make([]duplosdk.DuploAwsLbListenerRuleCondition, len(conditions))
	for i, condition := range conditions {
		elbConditions[i] = duplosdk.DuploAwsLbListenerRuleCondition{}
		conditionMap := condition.(map[string]interface{})
		var field string
		var attrs int

		if hostHeader, ok := conditionMap["host_header"].([]interface{}); ok && len(hostHeader) > 0 {
			field = "host-header"
			attrs += 1
			values := hostHeader[0].(map[string]interface{})["values"].(*schema.Set)

			elbConditions[i].HostHeaderConfig = &duplosdk.DuploStringValues{
				Values: expandStringSet(values),
			}
		}

		if httpHeader, ok := conditionMap["http_header"].([]interface{}); ok && len(httpHeader) > 0 {
			field = "http-header"
			attrs += 1
			httpHeaderMap := httpHeader[0].(map[string]interface{})
			values := httpHeaderMap["values"].(*schema.Set)

			elbConditions[i].HttpHeaderConfig = &duplosdk.DuploAwsLbListenerRuleConditionHttpRequestMethodConfig{
				HttpHeaderName: httpHeaderMap["http_header_name"].(string),
				Values:         expandStringSet(values),
			}
		}

		if httpRequestMethod, ok := conditionMap["http_request_method"].([]interface{}); ok && len(httpRequestMethod) > 0 {
			field = "http-request-method"
			attrs += 1
			values := httpRequestMethod[0].(map[string]interface{})["values"].(*schema.Set)

			elbConditions[i].HttpRequestMethodConfig = &duplosdk.DuploStringValues{
				Values: expandStringSet(values),
			}
		}

		if pathPattern, ok := conditionMap["path_pattern"].([]interface{}); ok && len(pathPattern) > 0 {
			field = "path-pattern"
			attrs += 1
			values := pathPattern[0].(map[string]interface{})["values"].(*schema.Set)

			elbConditions[i].PathPatternConfig = &duplosdk.DuploStringValues{
				Values: expandStringSet(values),
			}
		}

		if queryString, ok := conditionMap["query_string"].(*schema.Set); ok && queryString.Len() > 0 {
			field = "query-string"
			attrs += 1
			values := queryString.List()
			kvs := make([]duplosdk.DuploKeyStringValue, len(values))
			elbConditions[i].QueryStringConfig = &duplosdk.DuploAwsLbListenerRuleConditionQueryStringConfig{
				Values: &kvs,
			}

			for j, p := range values {
				valuePair := p.(map[string]interface{})
				elbValuePair := duplosdk.DuploKeyStringValue{
					Value: valuePair["value"].(string),
				}
				if valuePair["key"].(string) != "" {
					elbValuePair.Key = valuePair["key"].(string)
				}
				kvs[j] = elbValuePair
			}
			elbConditions[i].QueryStringConfig.Values = &kvs
		}

		if sourceIp, ok := conditionMap["source_ip"].([]interface{}); ok && len(sourceIp) > 0 {
			field = "source-ip"
			attrs += 1
			values := sourceIp[0].(map[string]interface{})["values"].(*schema.Set)

			elbConditions[i].SourceIpConfig = &duplosdk.DuploStringValues{
				Values: expandStringSet(values),
			}
		}

		// FIXME Rework this and use ConflictsWith when it finally works with collections:
		// https://github.com/hashicorp/terraform/issues/13016
		// Still need to ensure that one of the condition attributes is set.
		if attrs == 0 {
			return nil, errors.New("One of host_header, http_header, http_request_method, path_pattern, query_string or source_ip must be set in a condition block")
		} else if attrs > 1 {
			return nil, errors.New("Only one of host_header, http_header, http_request_method, path_pattern, query_string or source_ip can be set in a condition block")
		}

		elbConditions[i].Field = field
	}
	return &elbConditions, nil
}

func validListenerRulePriority(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)
	if value < listenerRulePriorityMin || (value > listenerRulePriorityMax && value != listenerRulePriorityDefault) {
		errors = append(errors, fmt.Errorf("%q must be in the range %d-%d for normal rule or %d for the default rule", k, listenerRulePriorityMin, listenerRulePriorityMax, listenerRulePriorityDefault))
	}
	return
}

func suppressIfActionTypeNot(t string) schema.SchemaDiffSuppressFunc {
	return func(k, old, new string, d *schema.ResourceData) bool {
		take := 2
		i := strings.IndexFunc(k, func(r rune) bool {
			if r == '.' {
				take -= 1
				return take == 0
			}
			return false
		})
		at := k[:i+1] + "type"
		return d.Get(at).(string) != t
	}
}

func getListenerArn(ruleArn string) string {
	listenerArn := strings.Replace(ruleArn, "listener-rule", "listener", -1)
	parts := strings.Split(listenerArn, "/")
	parts = parts[:len(parts)-1]
	return strings.Join(parts, "/")
}
