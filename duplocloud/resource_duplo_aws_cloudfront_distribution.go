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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAwsCloudfrontDistributionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the aws cloudfront distribution will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"domain_name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"etag": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"aliases": {
			Description: "Extra CNAMEs (alternate domain names), if any, for this distribution.",
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"comment": {
			Description:  "Any comments you want to include about the distribution.",
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringLenBetween(0, 128),
		},
		"default_root_object": {
			Description: "The object that you want CloudFront to return (for example, index.html) when an end user requests the root URL.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"enabled": {
			Description: "Whether the distribution is enabled to accept end user requests for content.",
			Type:        schema.TypeBool,
			Required:    true,
		},
		"http_version": {
			Description: "The maximum HTTP version to support on the distribution. Allowed values are `http1.1` and `http2`",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "http2",
			ValidateFunc: validation.StringInSlice([]string{
				"http1.1",
				"http2",
			}, false),
		},
		"price_class": {
			Description: "The price class for this distribution. One of `PriceClass_All`, `PriceClass_200`, `PriceClass_100`",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "PriceClass_All",
			ValidateFunc: validation.StringInSlice([]string{
				"PriceClass_All",
				"PriceClass_200",
				"PriceClass_100",
			}, false),
		},
		"is_ipv6_enabled": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"web_acl_id": {
			Description: "A unique identifier that specifies the AWS WAF web ACL, if any, to associate with this distribution.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"status": {
			Description: "The current status of the distribution. `Deployed` if the distribution's information is fully propagated throughout the Amazon CloudFront system.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"hosted_zone_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"wait_for_deployment": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"use_origin_access_identity": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"cors_allowed_host_names": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"custom_error_response": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"error_caching_min_ttl": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
					"error_code": {
						Type:     schema.TypeInt,
						Required: true,
					},
					"response_code": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
					"response_page_path": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
				},
			},
		},
		"viewer_certificate": {
			Type:     schema.TypeList,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"acm_certificate_arn": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"cloudfront_default_certificate": {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
					},
					"iam_certificate_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"minimum_protocol_version": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "TLSv1.2_2021",
					},
					"ssl_support_method": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "sni-only",
						ValidateFunc: validation.StringInSlice([]string{
							"vip",
							"sni-only",
						}, false),
					},
				},
			},
		},
		"restrictions": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"geo_restriction": {
						Type:     schema.TypeList,
						Required: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"locations": {
									Type:     schema.TypeSet,
									Optional: true,
									Computed: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"restriction_type": {
									Type:     schema.TypeString,
									Required: true,
									ValidateFunc: validation.StringInSlice([]string{
										"none",
										"whitelist",
										"blacklist",
									}, false),
								},
							},
						},
					},
				},
			},
		},
		"logging_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"bucket": {
						Type:     schema.TypeString,
						Required: true,
					},
					"include_cookies": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
					"prefix": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "",
					},
				},
			},
		},
		"origin_group": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"origin_id": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					"failover_criteria": {
						Type:     schema.TypeList,
						Required: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"status_codes": {
									Type:     schema.TypeSet,
									Required: true,
									Elem:     &schema.Schema{Type: schema.TypeInt},
								},
							},
						},
					},
					"member": {
						Type:     schema.TypeList,
						Required: true,
						MinItems: 2,
						MaxItems: 2,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"origin_id": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
				},
			},
		},
		"origin": {
			Type:     schema.TypeSet,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"connection_attempts": {
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      3,
						ValidateFunc: validation.IntBetween(1, 3),
					},
					"connection_timeout": {
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      10,
						ValidateFunc: validation.IntBetween(1, 10),
					},
					"custom_origin_config": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"http_port": {
									Type:     schema.TypeInt,
									Optional: true,
									Default:  80,
								},
								"https_port": {
									Type:     schema.TypeInt,
									Optional: true,
									Default:  443,
								},
								"origin_keepalive_timeout": {
									Type:         schema.TypeInt,
									Optional:     true,
									Default:      5,
									ValidateFunc: validation.IntBetween(1, 60),
								},
								"origin_read_timeout": {
									Type:         schema.TypeInt,
									Optional:     true,
									Default:      30,
									ValidateFunc: validation.IntBetween(1, 180),
								},
								"origin_protocol_policy": {
									Type:     schema.TypeString,
									Required: true,
									ValidateFunc: validation.StringInSlice([]string{
										"http-only",
										"https-only",
										"match-viewer",
									}, false),
								},
								"origin_ssl_protocols": {
									Type:     schema.TypeSet,
									Required: true,
									Elem: &schema.Schema{
										Type:    schema.TypeString,
										Default: "TLSv1.2",
										ValidateFunc: validation.StringInSlice([]string{
											"SSLv3",
											"TLSv1",
											"TLSv1.1",
											"TLSv1.2",
										}, false),
									},
								},
							},
						},
					},
					"domain_name": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					"custom_header": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:     schema.TypeString,
									Required: true,
								},
								"value": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
					"origin_id": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					"origin_path": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "",
					},
					"origin_shield": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"enabled": {
									Type:     schema.TypeBool,
									Required: true,
								},
								"origin_shield_region": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
					"s3_origin_config": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"origin_access_identity": {
									Type:     schema.TypeString,
									Optional: true,
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
		"default_cache_behavior": {
			Type:     schema.TypeList,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"allowed_methods": {
						Type:     schema.TypeSet,
						Required: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"cached_methods": {
						Type:     schema.TypeSet,
						Required: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"cache_policy_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"compress": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
					"default_ttl": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
					"field_level_encryption_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"forwarded_values": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"cookies": {
									Type:     schema.TypeList,
									Required: true,
									MaxItems: 1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"forward": {
												Type:     schema.TypeString,
												Required: true,
												ValidateFunc: validation.StringInSlice([]string{
													"all",
													"none",
													"whitelist",
												}, false),
											},
											"whitelisted_names": {
												Type:     schema.TypeSet,
												Optional: true,
												Computed: true,
												Elem:     &schema.Schema{Type: schema.TypeString},
											},
										},
									},
								},
								"headers": {
									Type:     schema.TypeSet,
									Optional: true,
									Computed: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"query_string": {
									Type:     schema.TypeBool,
									Required: true,
								},
								"query_string_cache_keys": {
									Type:     schema.TypeList,
									Optional: true,
									Computed: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
					"lambda_function_association": {
						Type:     schema.TypeSet,
						Optional: true,
						MaxItems: 4,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"event_type": {
									Type:     schema.TypeString,
									Required: true,
									ValidateFunc: validation.StringInSlice([]string{
										"viewer-request",
										"origin-request",
										"viewer-response",
										"origin-response",
									}, false),
								},
								"lambda_arn": {
									Type:     schema.TypeString,
									Required: true,
								},
								"include_body": {
									Type:     schema.TypeBool,
									Optional: true,
									Default:  false,
								},
							},
						},
					},
					"function_association": {
						Type:     schema.TypeSet,
						Optional: true,
						MaxItems: 2,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"event_type": {
									Type:     schema.TypeString,
									Required: true,
									ValidateFunc: validation.StringInSlice([]string{
										"viewer-request",
										"viewer-response",
									}, false),
								},
								"function_arn": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
					"max_ttl": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
					"min_ttl": {
						Type:     schema.TypeInt,
						Optional: true,
						Default:  0,
					},
					"origin_request_policy_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"realtime_log_config_arn": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"response_headers_policy_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"smooth_streaming": {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
					},
					"target_origin_id": {
						Type:     schema.TypeString,
						Required: true,
					},
					"trusted_key_groups": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"trusted_signers": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"viewer_protocol_policy": {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice([]string{
							"allow-all",
							"https-only",
							"redirect-to-https",
						}, false),
					},
				},
			},
		},
		"ordered_cache_behavior": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"allowed_methods": {
						Type:     schema.TypeSet,
						Required: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"cached_methods": {
						Type:     schema.TypeSet,
						Required: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"cache_policy_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"compress": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
					"default_ttl": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
					"field_level_encryption_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"forwarded_values": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"cookies": {
									Type:     schema.TypeList,
									Required: true,
									MaxItems: 1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"forward": {
												Type:     schema.TypeString,
												Required: true,
												ValidateFunc: validation.StringInSlice([]string{
													"all",
													"none",
													"whitelist",
												}, false),
											},
											"whitelisted_names": {
												Type:     schema.TypeSet,
												Optional: true,
												Elem:     &schema.Schema{Type: schema.TypeString},
												Computed: true,
											},
										},
									},
								},
								"headers": {
									Type:     schema.TypeSet,
									Optional: true,
									Computed: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"query_string": {
									Type:     schema.TypeBool,
									Required: true,
								},
								"query_string_cache_keys": {
									Type:     schema.TypeList,
									Optional: true,
									Computed: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
					"lambda_function_association": {
						Type:     schema.TypeSet,
						Optional: true,
						MaxItems: 4,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"event_type": {
									Type:     schema.TypeString,
									Required: true,
									ValidateFunc: validation.StringInSlice([]string{
										"viewer-request",
										"origin-request",
										"viewer-response",
										"origin-response",
									}, false),
								},
								"lambda_arn": {
									Type:     schema.TypeString,
									Required: true,
								},
								"include_body": {
									Type:     schema.TypeBool,
									Optional: true,
									Default:  false,
								},
							},
						},
					},
					"function_association": {
						Type:     schema.TypeSet,
						Optional: true,
						MaxItems: 2,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"event_type": {
									Type:     schema.TypeString,
									Required: true,
									ValidateFunc: validation.StringInSlice([]string{
										"viewer-request",
										"viewer-response",
									}, false),
								},
								"function_arn": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
					"max_ttl": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
					"min_ttl": {
						Type:     schema.TypeInt,
						Optional: true,
						Default:  0,
					},
					"origin_request_policy_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"path_pattern": {
						Type:     schema.TypeString,
						Required: true,
					},
					"realtime_log_config_arn": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"response_headers_policy_id": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
					},
					"smooth_streaming": {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
					},
					"target_origin_id": {
						Type:     schema.TypeString,
						Required: true,
					},
					"trusted_key_groups": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"trusted_signers": {
						Type:     schema.TypeList,
						Optional: true,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"viewer_protocol_policy": {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice([]string{
							"allow-all",
							"https-only",
							"redirect-to-https",
						}, false),
					},
				},
			},
		},
	}
}

func resourceAwsCloudfrontDistribution() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_cloudfront_distribution` manages an aws cloudfront distribution in Duplo.",

		ReadContext:   resourceAwsCloudfrontDistributionRead,
		CreateContext: resourceAwsCloudfrontDistributionCreate,
		UpdateContext: resourceAwsCloudfrontDistributionUpdate,
		DeleteContext: resourceAwsCloudfrontDistributionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsCloudfrontDistributionSchema(),
	}
}

func resourceAwsCloudfrontDistributionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, cfdId, err := parseAwsCloudfrontDistributionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsCloudfrontDistributionRead(%s, %s): start", tenantID, cfdId)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.AwsCloudfrontDistributionGet(tenantID, cfdId)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s aws cloudfront distribution%s : %s", tenantID, cfdId, clientErr)
	}

	flattenAwsCloudfrontDistribution(d, duplo.Distribution.DistributionConfig)
	d.Set("status", duplo.Distribution.Status)
	d.Set("domain_name", duplo.Distribution.DomainName)
	d.Set("etag", duplo.ETag)
	d.Set("arn", duplo.Distribution.ARN)
	d.Set("tenant_id", tenantID)
	d.Set("hosted_zone_id", duplosdk.CloudFrontRoute53ZoneID)

	log.Printf("[TRACE] resourceAwsCloudfrontDistributionRead(%s, %s): end", tenantID, cfdId)
	return nil
}

func resourceAwsCloudfrontDistributionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceAwsCloudfrontDistributionCreate(%s): start", tenantID)
	c := m.(*duplosdk.Client)

	rq := expandAwsCloudfrontDistributionConfig(d)
	resp, err := c.AwsCloudfrontDistributionCreate(tenantID, &duplosdk.DuploAwsCloudfrontDistributionCreate{
		DistributionConfig:   rq,
		UseOAIIdentity:       d.Get("use_origin_access_identity").(bool),
		CorsAllowedHostNames: expandStringList(d.Get("cors_allowed_host_names").([]interface{})),
	})
	if err != nil {
		return diag.Errorf("Error creating tenant %s aws cloudfront distribution.: %s", tenantID, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, resp.Id)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws cloudfront distribution", id, func() (interface{}, duplosdk.ClientError) {
		return c.AwsCloudfrontDistributionGet(tenantID, resp.Id)
	})

	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_for_deployment") == nil || d.Get("wait_for_deployment").(bool) {
		err = cloudfrontDistributionWaitUntilReady(ctx, c, tenantID, resp.Id, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAwsCloudfrontDistributionRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsCloudfrontDistributionCreate(%s, %s): end", tenantID, resp.Id)
	return diags
}

func resourceAwsCloudfrontDistributionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	log.Printf("[TRACE] resourceAwsCloudfrontDistributionUpdate(%s): start", id)
	tenantID, cfdId, err := parseAwsCloudfrontDistributionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}

	c := m.(*duplosdk.Client)

	duplo, clientErr := c.AwsCloudfrontDistributionGet(tenantID, cfdId)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s aws cloudfront distribution%s : %s", tenantID, cfdId, clientErr)
	}

	rq := expandAwsCloudfrontDistributionConfig(d)
	// Update OAI which is generated at backend
	updateS3OAI(duplo.Distribution.DistributionConfig, rq)

	resp, err := c.AwsCloudfrontDistributionUpdate(tenantID, &duplosdk.DuploAwsCloudfrontDistributionCreate{
		Id:                   cfdId,
		DistributionConfig:   rq,
		IfMatch:              d.Get("etag").(string),
		UseOAIIdentity:       d.Get("use_origin_access_identity").(bool),
		CorsAllowedHostNames: expandStringList(d.Get("cors_allowed_host_names").([]interface{})),
	})
	if err != nil {
		return diag.Errorf("Error updating tenant %s aws cloudfront distribution.: %s", tenantID, err)
	}

	if d.Get("wait_for_deployment") == nil || d.Get("wait_for_deployment").(bool) {
		err = cloudfrontDistributionWaitUntilReady(ctx, c, tenantID, resp.Id, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws cloudfront distribution", id, func() (interface{}, duplosdk.ClientError) {
		return c.AwsCloudfrontDistributionGet(tenantID, resp.Id)
	})

	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsCloudfrontDistributionRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsCloudfrontDistributionUpdate(%s, %s): end", tenantID, resp.Id)
	return diags
}

func resourceAwsCloudfrontDistributionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	enabled := d.Get("enabled").(bool)
	tenantID, cfdId, err := parseAwsCloudfrontDistributionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsCloudfrontDistributionDelete(%s, %s): start", tenantID, cfdId)

	c := m.(*duplosdk.Client)
	log.Printf("[TRACE] Cloudfront distribution enabled : (%s, %v)", cfdId, enabled)

	if enabled {
		log.Printf("[TRACE] Disable cloudfront distribution before delete. (%s, %s): start", tenantID, cfdId)

		clientErr := c.AwsCloudfrontDistributionDisable(tenantID, cfdId)
		if clientErr != nil {
			if clientErr.Status() == 404 {
				return nil
			}
			return diag.Errorf("Unable to disable tenant %s aws cloudfront distribution '%s': %s", tenantID, cfdId, clientErr)
		}
		err = cloudfrontDistributionWaitUntilDisabled(ctx, c, tenantID, cfdId, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	log.Printf("[TRACE] Cloudfront distribution disabled before delete. (%s, %s): nd", tenantID, cfdId)

	clientErr := c.AwsCloudfrontDistributionDelete(tenantID, cfdId)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s aws cloudfront distribution '%s': %s", tenantID, cfdId, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "aws cloudfront distribution", id, func() (interface{}, duplosdk.ClientError) {
		if rp, err := c.AwsCloudfrontDistributionExists(tenantID, cfdId); rp || err != nil {
			return rp, err
		}
		return nil, nil
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsCloudfrontDistributionDelete(%s, %s): end", tenantID, cfdId)
	return nil
}

func expandAwsCloudfrontDistributionConfig(d *schema.ResourceData) *duplosdk.DuploAwsCloudfrontDistributionConfig {
	distributionConfig := &duplosdk.DuploAwsCloudfrontDistributionConfig{
		DefaultCacheBehavior: expandAwsCloudfrontDistributionDefaultCacheBehavior(d.Get("default_cache_behavior").([]interface{})[0].(map[string]interface{})),

		CacheBehaviors:       expandAwsCloudfrontDistributionCacheBehaviorAll(d.Get("ordered_cache_behavior").([]interface{})),
		Comment:              d.Get("comment").(string),
		CustomErrorResponses: expandCustomErrorResponses(d.Get("custom_error_response").(*schema.Set)),
		DefaultRootObject:    d.Get("default_root_object").(string),
		Enabled:              d.Get("enabled").(bool),
		IsIPV6Enabled:        d.Get("is_ipv6_enabled").(bool),
		HttpVersion:          &duplosdk.DuploStringValue{Value: d.Get("http_version").(string)},
		Origins:              expandAwsCloudfrontDistributionOrigins(d.Get("origin").(*schema.Set)),
		PriceClass:           &duplosdk.DuploStringValue{Value: d.Get("price_class").(string)},
		WebACLId:             d.Get("web_acl_id").(string),
	}

	if v, ok := d.GetOk("logging_config"); ok {
		distributionConfig.Logging = expandLoggingConfig(v.([]interface{})[0].(map[string]interface{}))
	} else {
		distributionConfig.Logging = expandLoggingConfig(nil)
	}

	if v, ok := d.GetOk("aliases"); ok {
		distributionConfig.Aliases = expandAliases(v.(*schema.Set))
	}
	if v, ok := d.GetOk("restrictions"); ok {
		distributionConfig.Restrictions = expandRestrictions(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("viewer_certificate"); ok {
		distributionConfig.ViewerCertificate = expandViewerCertificate(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("origin_group"); ok {
		distributionConfig.OriginGroups = expandOriginGroups(v.(*schema.Set))
	}
	return distributionConfig
}

func expandAwsCloudfrontDistributionDefaultCacheBehavior(m map[string]interface{}) *duplosdk.DuploAwsCloudfrontDefaultCacheBehavior {
	dcb := &duplosdk.DuploAwsCloudfrontDefaultCacheBehavior{
		CachePolicyId:          m["cache_policy_id"].(string),
		Compress:               m["compress"].(bool),
		FieldLevelEncryptionId: m["field_level_encryption_id"].(string),
		OriginRequestPolicyId:  m["origin_request_policy_id"].(string),
		//TODO Handle response_headers_policy_id
		TargetOriginId:       m["target_origin_id"].(string),
		ViewerProtocolPolicy: &duplosdk.DuploStringValue{Value: m["viewer_protocol_policy"].(string)},
	}

	if forwardedValuesFlat, ok := m["forwarded_values"].([]interface{}); ok && len(forwardedValuesFlat) == 1 {
		dcb.ForwardedValues = expandForwardedValues(m["forwarded_values"].([]interface{})[0].(map[string]interface{}))
	}

	if m["cache_policy_id"].(string) == "" {
		dcb.MinTTL = m["min_ttl"].(int)
		dcb.MaxTTL = m["max_ttl"].(int)
		dcb.DefaultTTL = m["default_ttl"].(int)
	}

	// TODO Handle "trusted_key_groups"

	if v, ok := m["trusted_signers"]; ok {
		dcb.TrustedSigners = expandTrustedSigners(v.([]interface{}))
	} else {
		dcb.TrustedSigners = expandTrustedSigners([]interface{}{})
	}

	if v, ok := m["lambda_function_association"]; ok {
		dcb.LambdaFunctionAssociations = expandLambdaFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := m["function_association"]; ok {
		dcb.FunctionAssociations = expandFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := m["smooth_streaming"]; ok {
		dcb.SmoothStreaming = v.(bool)
	}

	if v, ok := m["allowed_methods"]; ok {
		dcb.AllowedMethods = expandAllowedMethods(v.(*schema.Set))
	}

	if v, ok := m["cached_methods"]; ok {
		dcb.AllowedMethods.CachedMethods = expandCachedMethods(v.(*schema.Set))
	}

	// TODO Handle realtime_log_config_arn
	return dcb
}

func expandAwsCloudfrontDistributionCacheBehavior(m map[string]interface{}) duplosdk.DuploAwsCloudfrontCacheBehavior {
	cb := duplosdk.DuploAwsCloudfrontCacheBehavior{
		CachePolicyId:          m["cache_policy_id"].(string),
		Compress:               m["compress"].(bool),
		FieldLevelEncryptionId: m["field_level_encryption_id"].(string),
		OriginRequestPolicyId:  m["origin_request_policy_id"].(string),
		//TODO Handle response_headers_policy_id
		TargetOriginId:       m["target_origin_id"].(string),
		ViewerProtocolPolicy: &duplosdk.DuploStringValue{Value: m["viewer_protocol_policy"].(string)},
	}

	if forwardedValuesFlat, ok := m["forwarded_values"].([]interface{}); ok && len(forwardedValuesFlat) == 1 {
		cb.ForwardedValues = expandForwardedValues(m["forwarded_values"].([]interface{})[0].(map[string]interface{}))
	}

	if m["cache_policy_id"].(string) == "" {
		cb.MinTTL = m["min_ttl"].(int)
		cb.MaxTTL = m["max_ttl"].(int)
		cb.DefaultTTL = m["default_ttl"].(int)
	}

	// TODO Handle "trusted_key_groups"

	if v, ok := m["trusted_signers"]; ok {
		cb.TrustedSigners = expandTrustedSigners(v.([]interface{}))
	} else {
		cb.TrustedSigners = expandTrustedSigners([]interface{}{})
	}

	if v, ok := m["function_association"]; ok {
		cb.FunctionAssociations = expandFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := m["lambda_function_association"]; ok {
		cb.LambdaFunctionAssociations = expandLambdaFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := m["smooth_streaming"]; ok {
		cb.SmoothStreaming = v.(bool)
	}

	if v, ok := m["allowed_methods"]; ok {
		cb.AllowedMethods = expandAllowedMethods(v.(*schema.Set))
	}

	if v, ok := m["cached_methods"]; ok {
		cb.AllowedMethods.CachedMethods = expandCachedMethods(v.(*schema.Set))
	}

	if v, ok := m["path_pattern"]; ok {
		cb.PathPattern = v.(string)
	}
	return cb
}

func expandAwsCloudfrontDistributionCacheBehaviorAll(lst []interface{}) *duplosdk.DuploAwsCloudfrontCacheBehaviors {
	var qty int
	items := make([]duplosdk.DuploAwsCloudfrontCacheBehavior, 0, len(lst))
	//var items []duplosdk.DuploAwsCloudfrontCacheBehavior
	for _, v := range lst {
		items = append(items, expandAwsCloudfrontDistributionCacheBehavior(v.(map[string]interface{})))
		qty++
	}
	return &duplosdk.DuploAwsCloudfrontCacheBehaviors{
		Quantity: qty,
		Items:    &items,
	}
}

func expandForwardedValues(m map[string]interface{}) *duplosdk.DuploCFDForwardedValues {
	if len(m) < 1 {
		return nil
	}

	fv := &duplosdk.DuploCFDForwardedValues{
		QueryString: m["query_string"].(bool),
	}
	if v, ok := m["cookies"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		fv.Cookies = expandCookiePreference(v.([]interface{})[0].(map[string]interface{}))
	}
	if v, ok := m["headers"]; ok {
		fv.Headers = expandHeaders(v.(*schema.Set).List())
	}
	if v, ok := m["query_string_cache_keys"]; ok {
		fv.QueryStringCacheKeys = expandQueryStringCacheKeys(v.([]interface{}))
	}
	return fv
}

func expandCookiePreference(m map[string]interface{}) *duplosdk.DuploCFDCookiePreference {
	cp := &duplosdk.DuploCFDCookiePreference{
		Forward: duplosdk.DuploStringValue{Value: m["forward"].(string)},
	}
	if v, ok := m["whitelisted_names"]; ok {
		cp.WhitelistedNames = expandCookieNames(v.(*schema.Set).List())
	}
	return cp
}

func expandCookieNames(d []interface{}) *duplosdk.DuploCFDStringItems {
	return &duplosdk.DuploCFDStringItems{
		Quantity: len(d),
		Items:    expandStringList(d),
	}
}

func expandHeaders(d []interface{}) *duplosdk.DuploCFDStringItems {
	return &duplosdk.DuploCFDStringItems{
		Quantity: len(d),
		Items:    expandStringList(d),
	}
}

func expandQueryStringCacheKeys(d []interface{}) *duplosdk.DuploCFDStringItems {
	return &duplosdk.DuploCFDStringItems{
		Quantity: len(d),
		Items:    expandStringList(d),
	}
}

func expandTrustedSigners(s []interface{}) *duplosdk.DuploCFDTrustedSigners {
	var ts duplosdk.DuploCFDTrustedSigners
	if len(s) > 0 {
		ts.Quantity = len(s)
		ts.Items = expandStringList(s)
		ts.Enabled = true
	} else {
		ts.Quantity = 0
		ts.Enabled = false
	}
	return &ts
}

func expandAllowedMethods(s *schema.Set) *duplosdk.DuploCFDAllowedMethods {
	return &duplosdk.DuploCFDAllowedMethods{
		Quantity: s.Len(),
		Items:    expandStringSet(s),
	}
}

func expandCachedMethods(s *schema.Set) *duplosdk.DuploCFDStringItems {
	return &duplosdk.DuploCFDStringItems{
		Quantity: s.Len(),
		Items:    expandStringSet(s),
	}
}

func parseAwsCloudfrontDistributionIdParts(id string) (tenantID, cfdId string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, cfdId = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func expandAwsCloudfrontDistributionOrigins(s *schema.Set) *duplosdk.DuploAwsCloudfrontOrigins {
	qty := 0
	// items := []*duplosdk.DuploAwsCloudfrontOrigin{}
	items := make([]duplosdk.DuploAwsCloudfrontOrigin, 0, s.Len())
	for _, v := range s.List() {
		items = append(items, expandAwsCloudfrontDistributionOrigin(v.(map[string]interface{})))
		qty++
	}
	return &duplosdk.DuploAwsCloudfrontOrigins{
		Quantity: qty,
		Items:    &items,
	}
}

func expandAwsCloudfrontDistributionOrigin(m map[string]interface{}) duplosdk.DuploAwsCloudfrontOrigin {
	origin := duplosdk.DuploAwsCloudfrontOrigin{
		Id:         m["origin_id"].(string),
		DomainName: m["domain_name"].(string),
	}

	if v, ok := m["connection_attempts"]; ok {
		origin.ConnectionAttempts = v.(int)
	}
	if v, ok := m["connection_timeout"]; ok {
		origin.ConnectionTimeout = v.(int)
	}
	if v, ok := m["custom_header"]; ok {
		origin.CustomHeaders = expandOriginCustomHeaders(v.(*schema.Set))
	}
	if v, ok := m["custom_origin_config"]; ok {
		if s := v.([]interface{}); len(s) > 0 {
			origin.CustomOriginConfig = expandCustomOriginConfig(s[0].(map[string]interface{}))
		}
	}
	if v, ok := m["origin_path"]; ok {
		origin.OriginPath = v.(string)
	} else {
		origin.OriginPath = ""
	}

	if v, ok := m["s3_origin_config"]; ok {
		if s := v.([]interface{}); len(s) > 0 {
			origin.S3OriginConfig = expandS3OriginConfig(s[0].(map[string]interface{}))
		}
	}

	if v, ok := m["origin_shield"]; ok {
		if s := v.([]interface{}); len(s) > 0 {
			origin.OriginShield = expandOriginShield(s[0].(map[string]interface{}))
		}
	}

	// if both custom and s3 origin are missing, add an empty s3 origin
	// One or the other must be specified, but the S3 origin can be "empty"
	if origin.S3OriginConfig == nil && origin.CustomOriginConfig == nil {
		origin.S3OriginConfig = &duplosdk.DuploAwsCloudfrontOriginS3Config{
			OriginAccessIdentity: "",
		}
	}

	return origin
}

func expandOriginCustomHeaders(s *schema.Set) *duplosdk.DuploAwsCloudfrontOriginCustomHeaders {
	qty := 0
	// items := []*duplosdk.DuploAwsCloudfrontOriginCustomHeader{}
	items := make([]duplosdk.DuploAwsCloudfrontOriginCustomHeader, 0, s.Len())
	for _, v := range s.List() {
		items = append(items, expandOriginCustomHeader(v.(map[string]interface{})))
		qty++
	}
	return &duplosdk.DuploAwsCloudfrontOriginCustomHeaders{
		Quantity: qty,
		Items:    &items,
	}
}

func expandOriginCustomHeader(m map[string]interface{}) duplosdk.DuploAwsCloudfrontOriginCustomHeader {
	return duplosdk.DuploAwsCloudfrontOriginCustomHeader{
		HeaderName:  m["name"].(string),
		HeaderValue: m["value"].(string),
	}
}

func expandCustomOriginConfig(m map[string]interface{}) *duplosdk.DuploAwsCloudfrontCustomOriginConfig {
	customOrigin := &duplosdk.DuploAwsCloudfrontCustomOriginConfig{
		HTTPPort:               m["http_port"].(int),
		HTTPSPort:              m["https_port"].(int),
		OriginSslProtocols:     expandCustomOriginConfigSSL(m["origin_ssl_protocols"].(*schema.Set).List()),
		OriginReadTimeout:      m["origin_read_timeout"].(int),
		OriginKeepaliveTimeout: m["origin_keepalive_timeout"].(int),
		OriginProtocolPolicy:   &duplosdk.DuploStringValue{Value: m["origin_protocol_policy"].(string)},
	}

	return customOrigin
}

func expandCustomOriginConfigSSL(s []interface{}) *duplosdk.DuploCFDStringItems {
	items := expandStringList(s)
	return &duplosdk.DuploCFDStringItems{
		Quantity: len(items),
		Items:    items,
	}
}

func expandS3OriginConfig(m map[string]interface{}) *duplosdk.DuploAwsCloudfrontOriginS3Config {
	return &duplosdk.DuploAwsCloudfrontOriginS3Config{
		OriginAccessIdentity: m["origin_access_identity"].(string),
	}
}

func expandAliases(s *schema.Set) *duplosdk.DuploCFDStringItems {
	aliases := duplosdk.DuploCFDStringItems{
		Quantity: s.Len(),
	}
	if s.Len() > 0 {
		aliases.Items = expandStringSet(s)
	}
	return &aliases
}

func expandRestrictions(m map[string]interface{}) *duplosdk.DuploAwsCloudfrontDistributionRestrictions {
	return &duplosdk.DuploAwsCloudfrontDistributionRestrictions{
		GeoRestriction: expandGeoRestriction(m["geo_restriction"].([]interface{})[0].(map[string]interface{})),
	}
}
func expandGeoRestriction(m map[string]interface{}) *duplosdk.DuploAwsCloudfrontDistributionGeoRestriction {
	gr := &duplosdk.DuploAwsCloudfrontDistributionGeoRestriction{
		Quantity:        0,
		RestrictionType: &duplosdk.DuploStringValue{Value: m["restriction_type"].(string)},
	}

	if v, ok := m["locations"]; ok {
		gr.Items = expandStringSet(v.(*schema.Set))
		gr.Quantity = v.(*schema.Set).Len()
	}

	return gr
}

func expandViewerCertificate(m map[string]interface{}) *duplosdk.DuploAwsCloudfrontDistributionViewerCertificate {
	var vc duplosdk.DuploAwsCloudfrontDistributionViewerCertificate
	if v, ok := m["iam_certificate_id"]; ok && v != "" {
		vc.IAMCertificateId = v.(string)
		vc.SSLSupportMethod = &duplosdk.DuploStringValue{Value: m["ssl_support_method"].(string)}
	} else if v, ok := m["acm_certificate_arn"]; ok && v != "" {
		vc.ACMCertificateArn = v.(string)
		vc.SSLSupportMethod = &duplosdk.DuploStringValue{Value: m["ssl_support_method"].(string)}
	} else {
		vc.CloudFrontDefaultCertificate = m["cloudfront_default_certificate"].(bool)
	}
	if v, ok := m["minimum_protocol_version"]; ok && v != "" {
		vc.MinimumProtocolVersion = &duplosdk.DuploStringValue{Value: v.(string)}
	}
	return &vc
}

func expandOriginGroups(s *schema.Set) *duplosdk.DuploAwsCloudfrontDistributionOriginGroups {
	qty := 0
	//items := []*duplosdk.DuploAwsCloudfrontDistributionOriginGroup{}
	items := make([]duplosdk.DuploAwsCloudfrontDistributionOriginGroup, 0, s.Len())
	for _, v := range s.List() {
		items = append(items, expandOriginGroup(v.(map[string]interface{})))
		qty++
	}
	return &duplosdk.DuploAwsCloudfrontDistributionOriginGroups{
		Quantity: qty,
		Items:    &items,
	}
}

func expandOriginGroup(m map[string]interface{}) duplosdk.DuploAwsCloudfrontDistributionOriginGroup {
	failoverCriteria := m["failover_criteria"].([]interface{})[0].(map[string]interface{})
	members := m["member"].([]interface{})
	originGroup := duplosdk.DuploAwsCloudfrontDistributionOriginGroup{
		Id:               m["origin_id"].(string),
		FailoverCriteria: expandOriginGroupFailoverCriteria(failoverCriteria),
		Members:          expandMembers(members),
	}
	return originGroup
}

func expandMembers(l []interface{}) *duplosdk.DuploAwsCloudfrontDistributionOriginGroupMembers {
	qty := 0
	// items := []*duplosdk.DuploAwsCloudfrontDistributionOriginGroupMember{}
	items := make([]duplosdk.DuploAwsCloudfrontDistributionOriginGroupMember, 0, len(l))
	for _, m := range l {
		ogm := duplosdk.DuploAwsCloudfrontDistributionOriginGroupMember{
			OriginId: m.(map[string]interface{})["origin_id"].(string),
		}
		items = append(items, ogm)
		qty++
	}
	return &duplosdk.DuploAwsCloudfrontDistributionOriginGroupMembers{
		Quantity: qty,
		Items:    &items,
	}
}

func expandOriginGroupFailoverCriteria(m map[string]interface{}) *duplosdk.DuploOriginGroupFailoverCriteria {
	failoverCriteria := &duplosdk.DuploOriginGroupFailoverCriteria{}
	if v, ok := m["status_codes"]; ok {
		codes := []int{}
		for _, code := range v.(*schema.Set).List() {
			codes = append(codes, code.(int))
		}
		failoverCriteria.StatusCodes = &duplosdk.DuploOriginGroupFailoverCriteriaStatusCodes{
			Items:    codes,
			Quantity: len(codes),
		}
	}
	return failoverCriteria
}

func expandLoggingConfig(m map[string]interface{}) *duplosdk.DuploAwsCloudfrontDistributionLoggingConfig {
	var lc duplosdk.DuploAwsCloudfrontDistributionLoggingConfig
	if m != nil {
		lc.Prefix = m["prefix"].(string)
		lc.Bucket = m["bucket"].(string)
		lc.IncludeCookies = m["include_cookies"].(bool)
		lc.Enabled = true
	} else {
		lc.Prefix = ""
		lc.Bucket = ""
		lc.IncludeCookies = false
		lc.Enabled = false
	}
	return &lc
}

func expandCustomErrorResponses(s *schema.Set) *duplosdk.DuploAwsCloudfrontDistributionCustomErrorResponses {
	qty := 0
	// items := []*duplosdk.DuploAwsCloudfrontDistributionCustomErrorResponse{}
	items := make([]duplosdk.DuploAwsCloudfrontDistributionCustomErrorResponse, 0, s.Len())
	for _, v := range s.List() {
		items = append(items, expandCustomErrorResponse(v.(map[string]interface{})))
		qty++
	}
	return &duplosdk.DuploAwsCloudfrontDistributionCustomErrorResponses{
		Quantity: qty,
		Items:    &items,
	}
}

func expandCustomErrorResponse(m map[string]interface{}) duplosdk.DuploAwsCloudfrontDistributionCustomErrorResponse {
	er := duplosdk.DuploAwsCloudfrontDistributionCustomErrorResponse{
		ErrorCode: m["error_code"].(int),
	}
	if v, ok := m["error_caching_min_ttl"]; ok {
		er.ErrorCachingMinTTL = v.(int)
	}
	if v, ok := m["response_code"]; ok && v.(int) != 0 {
		er.ResponseCode = strconv.Itoa(v.(int))
	} else {
		er.ResponseCode = ""
	}
	if v, ok := m["response_page_path"]; ok {
		er.ResponsePagePath = v.(string)
	}

	return er
}

func expandLambdaFunctionAssociations(v interface{}) *duplosdk.DuploAwsCloudfrontLambdaFunctionAssociations {
	if v == nil {
		return &duplosdk.DuploAwsCloudfrontLambdaFunctionAssociations{
			Quantity: 0,
		}
	}
	s := v.([]interface{})
	qty := 0
	items := make([]duplosdk.DuploAwsCloudfrontLambdaFunctionAssociation, 0, len(s))
	for _, i := range s {
		items = append(items, expandLambdaFunctionAssociation(i.(map[string]interface{})))
		qty++
	}
	return &duplosdk.DuploAwsCloudfrontLambdaFunctionAssociations{
		Quantity: qty,
		Items:    &items,
	}
}

func expandLambdaFunctionAssociation(lf map[string]interface{}) duplosdk.DuploAwsCloudfrontLambdaFunctionAssociation {
	var lfa duplosdk.DuploAwsCloudfrontLambdaFunctionAssociation
	if v, ok := lf["event_type"]; ok {
		lfa.EventType = &duplosdk.DuploStringValue{Value: v.(string)}
	}
	if v, ok := lf["lambda_arn"]; ok {
		lfa.LambdaFunctionARN = v.(string)
	}
	if v, ok := lf["include_body"]; ok {
		lfa.IncludeBody = v.(bool)
	}
	return lfa
}

func expandFunctionAssociations(v interface{}) *duplosdk.DuploAwsCloudfrontFunctionAssociations {
	if v == nil {
		return &duplosdk.DuploAwsCloudfrontFunctionAssociations{
			Quantity: 0,
		}
	}
	s := v.([]interface{})
	qty := 0
	items := make([]duplosdk.DuploAwsCloudfrontFunctionAssociation, 0, len(s))
	for _, i := range s {
		items = append(items, expandFunctionAssociation(i.(map[string]interface{})))
		qty++
	}
	return &duplosdk.DuploAwsCloudfrontFunctionAssociations{
		Quantity: qty,
		Items:    &items,
	}
}

func expandFunctionAssociation(lf map[string]interface{}) duplosdk.DuploAwsCloudfrontFunctionAssociation {
	var fa duplosdk.DuploAwsCloudfrontFunctionAssociation
	if v, ok := lf["event_type"]; ok {
		fa.EventType = &duplosdk.DuploStringValue{Value: v.(string)}
	}
	if v, ok := lf["function_arn"]; ok {
		fa.FunctionARN = v.(string)
	}
	return fa
}

func flattenAwsCloudfrontDistribution(d *schema.ResourceData, duplo *duplosdk.DuploAwsCloudfrontDistributionConfig) {
	d.Set("enabled", duplo.Enabled)
	d.Set("is_ipv6_enabled", duplo.IsIPV6Enabled)
	d.Set("price_class", duplo.PriceClass.Value)

	d.Set("default_cache_behavior", flattenDefaultCacheBehavior(duplo.DefaultCacheBehavior))

	d.Set("viewer_certificate", flattenViewerCertificate(duplo.ViewerCertificate))

	if duplo.Comment != "" {
		d.Set("comment", duplo.Comment)
	}

	if duplo.DefaultRootObject != "" {
		d.Set("default_root_object", duplo.DefaultRootObject)
	}
	if duplo.HttpVersion != nil {
		d.Set("http_version", duplo.HttpVersion.Value)
	}
	if duplo.WebACLId != "" {
		d.Set("web_acl_id", duplo.WebACLId)
	}
	if duplo.CustomErrorResponses != nil {
		d.Set("custom_error_response", flattenCustomErrorResponses(duplo.CustomErrorResponses))
	}
	if duplo.CacheBehaviors != nil {
		d.Set("ordered_cache_behavior", flattenCacheBehaviors(duplo.CacheBehaviors))
	}

	if duplo.Logging != nil && duplo.Logging.Enabled {
		d.Set("logging_config", flattenLoggingConfig(duplo.Logging))
	} else {
		d.Set("logging_config", []interface{}{})
	}

	if duplo.Aliases != nil {
		d.Set("aliases", flattenAliases(duplo.Aliases))
	}

	if duplo.Restrictions != nil {
		d.Set("restrictions", flattenRestrictions(duplo.Restrictions))
	}

	if duplo.Origins.Quantity > 0 {
		d.Set("origin", flattenOrigins(duplo.Origins, d.Get("use_origin_access_identity").(bool)))
	}

	if duplo.OriginGroups.Quantity > 0 {
		d.Set("origin_group", flattenOriginGroups(duplo.OriginGroups))
	}

}

func flattenDefaultCacheBehavior(dcb *duplosdk.DuploAwsCloudfrontDefaultCacheBehavior) []interface{} {
	return []interface{}{flattenCloudFrontDefaultCacheBehavior(dcb)}
}

func flattenCloudFrontDefaultCacheBehavior(dcb *duplosdk.DuploAwsCloudfrontDefaultCacheBehavior) map[string]interface{} {
	m := map[string]interface{}{
		"cache_policy_id":           dcb.CachePolicyId,
		"compress":                  dcb.Compress,
		"field_level_encryption_id": dcb.FieldLevelEncryptionId,
		"viewer_protocol_policy":    dcb.ViewerProtocolPolicy.Value,
		"target_origin_id":          dcb.TargetOriginId,
		"min_ttl":                   dcb.MinTTL,
		"max_ttl":                   dcb.MaxTTL,
		"default_ttl":               dcb.DefaultTTL,
		"smooth_streaming":          dcb.SmoothStreaming,
		"origin_request_policy_id":  dcb.OriginRequestPolicyId,
		//"realtime_log_config_arn":    dcb.RealtimeLogConfigArn,
		//"response_headers_policy_id": dcb.ResponseHeadersPolicyId,
	}

	if dcb.ForwardedValues != nil {
		m["forwarded_values"] = []interface{}{flattenForwardedValues(dcb.ForwardedValues)}
	}

	if len(dcb.TrustedSigners.Items) > 0 {
		m["trusted_signers"] = flattenTrustedSigners(dcb.TrustedSigners)
	}
	if dcb.AllowedMethods != nil {
		m["allowed_methods"] = flattenAllowedMethods(dcb.AllowedMethods)
	}
	if dcb.AllowedMethods.CachedMethods != nil {
		m["cached_methods"] = flattenCachedMethods(dcb.AllowedMethods.CachedMethods)
	}
	lfaItems := dcb.LambdaFunctionAssociations.Items
	if dcb.LambdaFunctionAssociations != nil && len(*lfaItems) > 0 {
		m["lambda_function_association"] = flattenLambdaFunctionAssociations(dcb.LambdaFunctionAssociations)
	}

	if dcb.FunctionAssociations != nil && dcb.FunctionAssociations.Items != nil && len(*dcb.FunctionAssociations.Items) > 0 {
		m["function_association"] = flattenFunctionAssociations(dcb.FunctionAssociations)
	}

	return m
}

func flattenForwardedValues(fv *duplosdk.DuploCFDForwardedValues) map[string]interface{} {
	m := make(map[string]interface{})
	m["query_string"] = fv.QueryString
	if fv.Cookies != nil {
		m["cookies"] = []interface{}{flattenCookiePreference(fv.Cookies)}
	}
	if fv.Headers != nil {
		m["headers"] = schema.NewSet(schema.HashString, flattenHeaders(fv.Headers))
	}
	if fv.QueryStringCacheKeys != nil {
		m["query_string_cache_keys"] = flattenQueryStringCacheKeys(fv.QueryStringCacheKeys)
	}
	return m
}

func flattenHeaders(h *duplosdk.DuploCFDStringItems) []interface{} {
	if h.Items != nil {
		return flattenStringList(h.Items)
	}
	return []interface{}{}
}

func flattenQueryStringCacheKeys(k *duplosdk.DuploCFDStringItems) []interface{} {
	if k.Items != nil {
		return flattenStringList(k.Items)
	}
	return []interface{}{}
}

func flattenCookiePreference(cp *duplosdk.DuploCFDCookiePreference) map[string]interface{} {
	m := make(map[string]interface{})
	m["forward"] = cp.Forward.Value
	if cp.WhitelistedNames != nil {
		m["whitelisted_names"] = schema.NewSet(schema.HashString, flattenCookieNames(cp.WhitelistedNames))
	}
	return m
}

func flattenCookieNames(cn *duplosdk.DuploCFDStringItems) []interface{} {
	if cn.Items != nil {
		return flattenStringList(cn.Items)
	}
	return []interface{}{}
}

func flattenTrustedSigners(ts *duplosdk.DuploCFDTrustedSigners) []interface{} {
	if ts.Items != nil {
		return flattenStringList(ts.Items)
	}
	return []interface{}{}
}

func flattenAllowedMethods(am *duplosdk.DuploCFDAllowedMethods) *schema.Set {
	if am.Items != nil {
		return flattenStringSet(am.Items)
	}
	return nil
}

func flattenCachedMethods(cm *duplosdk.DuploCFDStringItems) *schema.Set {
	if cm.Items != nil {
		return flattenStringSet(cm.Items)
	}
	return nil
}

func flattenViewerCertificate(vc *duplosdk.DuploAwsCloudfrontDistributionViewerCertificate) []interface{} {
	m := make(map[string]interface{})

	if vc.IAMCertificateId != "" {
		m["iam_certificate_id"] = vc.IAMCertificateId
		m["ssl_support_method"] = vc.SSLSupportMethod.Value
	}
	if vc.ACMCertificateArn != "" {
		m["acm_certificate_arn"] = vc.ACMCertificateArn
		m["ssl_support_method"] = vc.SSLSupportMethod.Value
	}
	m["cloudfront_default_certificate"] = vc.CloudFrontDefaultCertificate
	if vc.MinimumProtocolVersion != nil {
		m["minimum_protocol_version"] = vc.MinimumProtocolVersion.Value
	}
	return []interface{}{m}
}

func flattenCustomErrorResponses(ers *duplosdk.DuploAwsCloudfrontDistributionCustomErrorResponses) []map[string]interface{} {
	s := []map[string]interface{}{}
	for _, v := range *ers.Items {
		s = append(s, flattenCustomErrorResponse(v))
	}
	// return schema.NewSet(CustomErrorResponseHash, s)
	return s
}

func flattenCustomErrorResponse(er duplosdk.DuploAwsCloudfrontDistributionCustomErrorResponse) map[string]interface{} {
	m := make(map[string]interface{})
	m["error_code"] = er.ErrorCode
	m["error_caching_min_ttl"] = er.ErrorCachingMinTTL
	m["response_code"] = er.ResponseCode
	m["response_page_path"] = er.ResponsePagePath
	return m
}

func flattenCacheBehaviors(cbs *duplosdk.DuploAwsCloudfrontCacheBehaviors) []interface{} {
	lst := []interface{}{}
	for _, v := range *cbs.Items {
		lst = append(lst, flattenCacheBehavior(v))
	}
	return lst
}

func flattenCacheBehavior(cb duplosdk.DuploAwsCloudfrontCacheBehavior) map[string]interface{} {
	m := map[string]interface{}{
		"cache_policy_id":           cb.CachePolicyId,
		"compress":                  cb.Compress,
		"field_level_encryption_id": cb.FieldLevelEncryptionId,
		"viewer_protocol_policy":    cb.ViewerProtocolPolicy.Value,
		"target_origin_id":          cb.TargetOriginId,
		"min_ttl":                   cb.MinTTL,
		"max_ttl":                   cb.MaxTTL,
		"default_ttl":               cb.DefaultTTL,
		"smooth_streaming":          cb.SmoothStreaming,
		"origin_request_policy_id":  cb.OriginRequestPolicyId,
		"path_pattern":              cb.PathPattern,
		//"realtime_log_config_arn":    dcb.RealtimeLogConfigArn,
		//"response_headers_policy_id": dcb.ResponseHeadersPolicyId,
	}

	if cb.ForwardedValues != nil {
		m["forwarded_values"] = []interface{}{flattenForwardedValues(cb.ForwardedValues)}
	}

	if len(cb.TrustedSigners.Items) > 0 {
		m["trusted_signers"] = flattenTrustedSigners(cb.TrustedSigners)
	}
	if cb.AllowedMethods != nil {
		m["allowed_methods"] = flattenAllowedMethods(cb.AllowedMethods)
	}
	if cb.AllowedMethods.CachedMethods != nil {
		m["cached_methods"] = flattenCachedMethods(cb.AllowedMethods.CachedMethods)
	}
	lfaItems := cb.LambdaFunctionAssociations.Items
	if cb.LambdaFunctionAssociations != nil && len(*lfaItems) > 0 {
		m["lambda_function_association"] = flattenLambdaFunctionAssociations(cb.LambdaFunctionAssociations)
	}
	if cb.FunctionAssociations != nil && cb.FunctionAssociations.Items != nil && len(*cb.FunctionAssociations.Items) > 0 {
		m["function_association"] = flattenFunctionAssociations(cb.FunctionAssociations)
	}
	return m
}

func flattenLoggingConfig(lc *duplosdk.DuploAwsCloudfrontDistributionLoggingConfig) []interface{} {
	m := map[string]interface{}{
		"bucket":          lc.Bucket,
		"include_cookies": lc.IncludeCookies,
		"prefix":          lc.Prefix,
	}

	return []interface{}{m}
}

func flattenRestrictions(r *duplosdk.DuploAwsCloudfrontDistributionRestrictions) []interface{} {
	m := map[string]interface{}{
		"geo_restriction": []interface{}{flattenGeoRestriction(r.GeoRestriction)},
	}

	return []interface{}{m}
}

func flattenGeoRestriction(gr *duplosdk.DuploAwsCloudfrontDistributionGeoRestriction) map[string]interface{} {
	m := make(map[string]interface{})

	m["restriction_type"] = gr.RestrictionType.Value
	if gr.Items != nil {
		m["locations"] = flattenStringSet(gr.Items)
	}
	return m
}

func flattenAliases(aliases *duplosdk.DuploCFDStringItems) *schema.Set {
	if aliases.Items != nil {
		return flattenStringSet(aliases.Items)
	}
	return nil
}

func flattenOrigins(ors *duplosdk.DuploAwsCloudfrontOrigins, useOAI bool) []map[string]interface{} {
	s := []map[string]interface{}{}
	for _, v := range *ors.Items {
		s = append(s, flattenOrigin(v, useOAI))
	}
	return s
}

func flattenOrigin(or duplosdk.DuploAwsCloudfrontOrigin, useOAI bool) map[string]interface{} {
	m := make(map[string]interface{})
	m["origin_id"] = or.Id
	m["domain_name"] = or.DomainName
	m["origin_path"] = or.OriginPath
	m["connection_attempts"] = or.ConnectionAttempts
	m["connection_timeout"] = or.ConnectionTimeout
	if or.CustomHeaders != nil {
		m["custom_header"] = flattenCustomHeaders(or.CustomHeaders)
	}
	if or.CustomOriginConfig != nil {
		m["custom_origin_config"] = []interface{}{flattenCustomOriginConfig(or.CustomOriginConfig)}
	}

	if or.S3OriginConfig != nil && or.S3OriginConfig.OriginAccessIdentity != "" {
		m["s3_origin_config"] = []interface{}{flattenS3OriginConfig(or.S3OriginConfig)}
	}
	if or.OriginShield != nil && or.OriginShield.Enabled {
		m["origin_shield"] = []interface{}{flattenOriginShield(or.OriginShield)}
	}

	return m
}

func flattenCustomHeaders(chs *duplosdk.DuploAwsCloudfrontOriginCustomHeaders) []map[string]interface{} {
	s := []map[string]interface{}{}
	for _, v := range *chs.Items {
		s = append(s, flattenOriginCustomHeader(v))
	}
	return s
}

func flattenOriginCustomHeader(och duplosdk.DuploAwsCloudfrontOriginCustomHeader) map[string]interface{} {
	return map[string]interface{}{
		"name":  och.HeaderName,
		"value": och.HeaderValue,
	}
}

func flattenCustomOriginConfig(cor *duplosdk.DuploAwsCloudfrontCustomOriginConfig) map[string]interface{} {

	customOrigin := map[string]interface{}{
		"origin_protocol_policy":   cor.OriginProtocolPolicy.Value,
		"http_port":                cor.HTTPPort,
		"https_port":               cor.HTTPSPort,
		"origin_ssl_protocols":     flattenCustomOriginConfigSSL(cor.OriginSslProtocols),
		"origin_read_timeout":      cor.OriginReadTimeout,
		"origin_keepalive_timeout": cor.OriginKeepaliveTimeout,
	}

	return customOrigin
}

func flattenCustomOriginConfigSSL(osp *duplosdk.DuploCFDStringItems) *schema.Set {
	return flattenStringSet(osp.Items)
}

func flattenS3OriginConfig(s3o *duplosdk.DuploAwsCloudfrontOriginS3Config) map[string]interface{} {
	return map[string]interface{}{
		"origin_access_identity": s3o.OriginAccessIdentity,
	}
}

func flattenOriginGroups(ogs *duplosdk.DuploAwsCloudfrontDistributionOriginGroups) []map[string]interface{} {
	s := []map[string]interface{}{}
	for _, v := range *ogs.Items {
		s = append(s, flattenOriginGroup(v))
	}
	return s
}

func flattenOriginGroup(og duplosdk.DuploAwsCloudfrontDistributionOriginGroup) map[string]interface{} {
	m := make(map[string]interface{})
	m["origin_id"] = og.Id
	if og.FailoverCriteria != nil {
		m["failover_criteria"] = flattenOriginGroupFailoverCriteria(og.FailoverCriteria)
	}
	if og.Members != nil {
		m["member"] = flattenOriginGroupMembers(og.Members)
	}
	return m
}

func flattenOriginGroupFailoverCriteria(ogfc *duplosdk.DuploOriginGroupFailoverCriteria) []interface{} {
	m := make(map[string]interface{})
	if ogfc.StatusCodes.Items != nil {
		l := []interface{}{}
		for _, i := range ogfc.StatusCodes.Items {
			l = append(l, i)
		}
		m["status_codes"] = schema.NewSet(schema.HashInt, l)
	}
	return []interface{}{m}
}

func flattenOriginGroupMembers(ogm *duplosdk.DuploAwsCloudfrontDistributionOriginGroupMembers) []interface{} {
	s := []interface{}{}
	for _, i := range *ogm.Items {
		m := map[string]interface{}{
			"origin_id": i.OriginId,
		}
		s = append(s, m)
	}
	return s
}

func flattenLambdaFunctionAssociations(lfa *duplosdk.DuploAwsCloudfrontLambdaFunctionAssociations) []map[string]interface{} {
	s := []map[string]interface{}{}
	for _, v := range *lfa.Items {
		s = append(s, flattenLambdaFunctionAssociation(v))
	}
	return s
}

func flattenLambdaFunctionAssociation(lfa duplosdk.DuploAwsCloudfrontLambdaFunctionAssociation) map[string]interface{} {
	m := map[string]interface{}{}
	m["event_type"] = lfa.EventType.Value
	m["lambda_arn"] = lfa.LambdaFunctionARN
	m["include_body"] = lfa.IncludeBody
	return m
}

func flattenFunctionAssociations(lfa *duplosdk.DuploAwsCloudfrontFunctionAssociations) []map[string]interface{} {
	s := []map[string]interface{}{}
	for _, v := range *lfa.Items {
		s = append(s, flattenFunctionAssociation(v))
	}
	return s
}

func flattenFunctionAssociation(fa duplosdk.DuploAwsCloudfrontFunctionAssociation) map[string]interface{} {
	m := map[string]interface{}{}
	m["event_type"] = fa.EventType.Value
	m["function_arn"] = fa.FunctionARN
	return m
}
func expandOriginShield(m map[string]interface{}) *duplosdk.DuploAwsCloudfrontOriginShield {
	return &duplosdk.DuploAwsCloudfrontOriginShield{
		Enabled:            m["enabled"].(bool),
		OriginShieldRegion: m["origin_shield_region"].(string),
	}
}

func flattenOriginShield(o *duplosdk.DuploAwsCloudfrontOriginShield) map[string]interface{} {
	return map[string]interface{}{
		"origin_shield_region": o.OriginShieldRegion,
		"enabled":              o.Enabled,
	}
}

func cloudfrontDistributionWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, cfdId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AwsCloudfrontDistributionGet(tenantID, cfdId)
			log.Printf("[TRACE] Cloudfront distribution status is (%s).", rp.Distribution.Status)
			status := "pending"
			if err == nil {
				if rp.Distribution.Status == "Deployed" {
					status = "ready"
				} else {
					status = "pending"
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] cloudfrontDistributionWaitUntilReady(%s, %s)", tenantID, cfdId)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func cloudfrontDistributionWaitUntilDisabled(ctx context.Context, c *duplosdk.Client, tenantID string, cfdId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AwsCloudfrontDistributionGet(tenantID, cfdId)
			log.Printf("[TRACE] Cloudfront distribution status is (%s).", rp.Distribution.Status)
			status := "pending"
			if err == nil {
				if !rp.Distribution.DistributionConfig.Enabled && rp.Distribution.Status == "Deployed" {
					status = "ready"
				} else {
					status = "pending"
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] cloudfrontDistributionWaitUntilDisabled(%s, %s)", tenantID, cfdId)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func updateS3OAI(existingCfd, updatedCfd *duplosdk.DuploAwsCloudfrontDistributionConfig) {
	for _, eo := range *existingCfd.Origins.Items {
		for _, uo := range *updatedCfd.Origins.Items {
			if eo.Id == uo.Id {
				uo.S3OriginConfig.OriginAccessIdentity = eo.S3OriginConfig.OriginAccessIdentity
				break
			}
		}
	}
}
