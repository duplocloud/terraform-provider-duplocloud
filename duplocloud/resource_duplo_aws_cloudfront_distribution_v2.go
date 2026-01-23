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

func duploAwsCloudfrontDistributionSchemaV2() map[string]*schema.Schema {
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
		"name": {
			Description:  "Name of the cloudfront distribution",
			Type:         schema.TypeString,
			ValidateFunc: validation.StringLenBetween(0, 128),
			Required:     true,
			ForceNew:     true,
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
			Optional:    true,
			Default:     true,
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
		"use_origin_access_control": {
			Description: `Duplo will create an origin access control (OAC) and restrict the S3 origin access. On false it will be public</br>For migration from OAI to OAC can be done from duplo cloud portal.`,
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
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
						Description: `<br />						
| Policy name                                                                                                                                                                                  | Policy Id                            |
|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------|
| [Amplify](#https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/using-managed-cache-policies.html#managed-cache-policy-amplify)                                                | 2e54312d-136d-493c-8eb9-b001f22f67d2 |
| [CachingDisabled](#https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/using-managed-cache-policies.html#managed-cache-policy-caching-disabled)                               | 4135ea2d-6df8-44a3-9df3-4b5a84be39ad |
| [CachingOptimized](#https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/using-managed-cache-policies.html#managed-cache-caching-optimized)                                    | 658327ea-f89d-4fab-a63d-7e88639e58f6 |
| [CachingOptimizedForUncompressedObjects](#https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/using-managed-cache-policies.html#managed-cache-caching-optimized-uncompressed) | b2884449-e4de-46a7-ac36-70bc7f1ddd6d |
| [Elemental-MediaPackage](#https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/using-managed-cache-policies.html#managed-cache-policy-mediapackage)                            | 08627262-05a9-4f76-9ded-b50ca2e3a84f |
<br />`,
					},
					"compress": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
					"default_ttl": {
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
						Description: "default time to live: Not required when cache_policy_id is set",
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
									Type:        schema.TypeSet,
									Optional:    true,
									Computed:    true,
									Elem:        &schema.Schema{Type: schema.TypeString},
									Description: "headers: Not required when cache_policy_id is set",
								},
								"query_string": {
									Type:     schema.TypeBool,
									Required: true,
								},
								"query_string_cache_keys": {
									Type:        schema.TypeList,
									Optional:    true,
									Computed:    true,
									Elem:        &schema.Schema{Type: schema.TypeString},
									Description: "query_string_cache_keys: Not required when cache_policy_id is set",
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
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
						Description: "Maximum time to live: Not required when cache_policy_id is set",
					},
					"min_ttl": {
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     0,
						Description: "Minimum time to live: Not required when cache_policy_id is set",
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

func resourceAwsCloudfrontDistributionV2() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_cloudfront_distribution_v2` manages an aws cloudfront distribution in Duplo.",

		ReadContext:   resourceAwsCloudfrontDistributionV2Read,
		CreateContext: resourceAwsCloudfrontDistributionV2Create,
		UpdateContext: resourceAwsCloudfrontDistributionV2Update,
		DeleteContext: resourceAwsCloudfrontDistributionV2Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema:        duploAwsCloudfrontDistributionSchemaV2(),
		CustomizeDiff: validateCloudDistributionParameters,
	}
}

func resourceAwsCloudfrontDistributionV2Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, cfdId, name, err := parseAwsCloudfrontDistributionIdPartsV2(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsCloudfrontDistributionRead(%s, %s): start", tenantID, cfdId)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.AwsCloudfrontDistributionGet(tenantID, cfdId)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[TRACE] resourceAwsCloudfrontDistributionRead(%s, %s): end - distribution not found", tenantID, cfdId)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s aws cloudfront distribution%s : %s", tenantID, cfdId, clientErr)
	}
	if duplo == nil {
		log.Printf("[TRACE] resourceAwsCloudfrontDistributionRead(%s, %s): end - distribution response empty", tenantID, cfdId)
		d.SetId("")
		return nil
	}
	if len(duplo.Distribution.DistributionConfig.CorsAllowedHostNames) == 0 && d.Get("cors_allowed_host_names") != nil {
		corsAllowedHostNames := []string{}

		for _, v := range d.Get("cors_allowed_host_names").([]interface{}) {
			corsAllowedHostNames = append(corsAllowedHostNames, v.(string))
		}
		duplo.Distribution.DistributionConfig.CorsAllowedHostNames = corsAllowedHostNames
	}
	flattenAwsCloudfrontDistributionV2(d, duplo.Distribution.DistributionConfig)
	d.Set("status", duplo.Distribution.Status)
	d.Set("domain_name", duplo.Distribution.DomainName)
	d.Set("etag", duplo.ETag)
	d.Set("arn", duplo.Distribution.ARN)
	d.Set("tenant_id", tenantID)
	d.Set("hosted_zone_id", duplosdk.CloudFrontRoute53ZoneID)
	d.Set("name", name)
	log.Printf("[TRACE] resourceAwsCloudfrontDistributionRead(%s, %s): end", tenantID, cfdId)
	return nil
}

func resourceAwsCloudfrontDistributionV2Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsCloudfrontDistributionCreate(%s): start", tenantID)
	c := m.(*duplosdk.Client)

	rq := expandAwsCloudfrontDistributionV2Config(d, false)
	resp, err := c.AwsCloudfrontDistributionCreate(tenantID, &duplosdk.DuploAwsCloudfrontDistributionCreate{
		DistributionConfig:   rq,
		UseOAIIdentity:       d.Get("use_origin_access_control").(bool),
		CorsAllowedHostNames: expandStringList(d.Get("cors_allowed_host_names").([]interface{})),
	})
	if err != nil {
		return diag.Errorf("Error creating tenant %s aws cloudfront distribution.: %s", tenantID, err)
	}

	id := fmt.Sprintf("%s/%s/%s", tenantID, resp.Id, name)
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

	diags = resourceAwsCloudfrontDistributionV2Read(ctx, d, m)
	log.Printf("[TRACE] resourceAwsCloudfrontDistributionCreate(%s, %s): end", tenantID, resp.Id)
	return diags
}

func resourceAwsCloudfrontDistributionV2Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	log.Printf("[TRACE] resourceAwsCloudfrontDistributionUpdate(%s): start", id)
	tenantID, cfdId, _, err := parseAwsCloudfrontDistributionIdPartsV2(id)
	if err != nil {
		return diag.FromErr(err)
	}

	c := m.(*duplosdk.Client)

	//	duplo, clientErr := c.AwsCloudfrontDistributionGet(tenantID, cfdId)
	//	if clientErr != nil {
	//		if clientErr.Status() == 404 {
	//			d.SetId("")
	//			return nil
	//		}
	//		return diag.Errorf("Unable to retrieve tenant %s aws cloudfront distribution%s : %s", tenantID, cfdId, clientErr)
	//	}

	rq := expandAwsCloudfrontDistributionV2Config(d, true)
	// Update OAI which is generated at backend
	//updateS3OAI(duplo.Distribution.DistributionConfig, rq)

	resp, err := c.AwsCloudfrontDistributionUpdate(tenantID, &duplosdk.DuploAwsCloudfrontDistributionCreate{
		Id:                   cfdId,
		DistributionConfig:   rq,
		IfMatch:              d.Get("etag").(string),
		UseOAIIdentity:       d.Get("use_origin_access_control").(bool),
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

	diags = resourceAwsCloudfrontDistributionV2Read(ctx, d, m)
	log.Printf("[TRACE] resourceAwsCloudfrontDistributionUpdate(%s, %s): end", tenantID, resp.Id)
	return diags
}

func resourceAwsCloudfrontDistributionV2Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	enabled := d.Get("enabled").(bool)
	tenantID, cfdId, name, err := parseAwsCloudfrontDistributionIdPartsV2(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsCloudfrontDistributionDelete(%s, %s,%s): start", tenantID, cfdId, name)

	c := m.(*duplosdk.Client)
	log.Printf("[TRACE] Cloudfront distribution enabled : (%s, %v)", cfdId, enabled)

	if enabled {
		log.Printf("[TRACE] Disable cloudfront distribution before delete. (%s, %s): start", tenantID, cfdId)

		clientErr := c.AwsCloudfrontDistributionDisable(tenantID, cfdId)
		if clientErr != nil {
			if clientErr.Status() == 404 {
				log.Printf("[TRACE] Cloudfront distribution disabled before delete. (%s, %s): nd", tenantID, cfdId)
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
			log.Printf("[TRACE] resourceAwsCloudfrontDistributionDelete(%s, %s): end", tenantID, cfdId)
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

	log.Printf("[TRACE] resourceAwsCloudfrontDistributionDelete(%s, %s,%s): end", tenantID, cfdId, name)
	return nil
}

func expandAwsCloudfrontDistributionV2Config(d *schema.ResourceData, update bool) *duplosdk.DuploAwsCloudfrontDistributionConfig {
	distributionConfig := &duplosdk.DuploAwsCloudfrontDistributionConfig{
		DefaultCacheBehavior: expandAwsCloudfrontDistributionDefaultCacheBehavior(d.Get("default_cache_behavior").([]interface{})[0].(map[string]interface{})),

		CacheBehaviors: expandAwsCloudfrontDistributionCacheBehaviorAll(d.Get("ordered_cache_behavior").([]interface{})),
		Comment:        d.Get("name").(string),

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

func flattenAwsCloudfrontDistributionV2(d *schema.ResourceData, duplo *duplosdk.DuploAwsCloudfrontDistributionConfig) {
	d.Set("enabled", duplo.Enabled)
	d.Set("is_ipv6_enabled", duplo.IsIPV6Enabled)
	d.Set("price_class", duplo.PriceClass.Value)

	d.Set("default_cache_behavior", flattenDefaultCacheBehavior(duplo.DefaultCacheBehavior))

	d.Set("viewer_certificate", flattenViewerCertificate(duplo.ViewerCertificate))

	if duplo.Comment != "" {
		d.Set("name", duplo.Comment)
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
		errorSet := schema.NewSet(schema.HashResource(customErrorHashResource()), nil)
		for _, v := range *duplo.CustomErrorResponses.Items {
			m := flattenCustomErrorResponse(v)
			errorSet.Add(m)
		}
		d.Set("custom_error_response", errorSet)

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

		originSet := schema.NewSet(schema.HashResource(originHashResource()), nil)

		for _, origin := range *duplo.Origins.Items {
			mp := flattenOriginV2(origin)
			originSet.Add(mp)
		}
		d.Set("origin", originSet)
	}

	if duplo.OriginGroups.Quantity > 0 {
		d.Set("origin_group", flattenOriginGroups(duplo.OriginGroups))
	}

	if len(duplo.CorsAllowedHostNames) > 0 {
		corsList := make([]interface{}, len(duplo.CorsAllowedHostNames))
		for i, v := range duplo.CorsAllowedHostNames {
			corsList[i] = v
		}

		d.Set("cors_allowed_host_names", corsList)
	} else {
		d.Set("cors_allowed_host_names", nil)
	}

}

func flattenOriginV2(or duplosdk.DuploAwsCloudfrontOrigin) map[string]interface{} {
	m := make(map[string]interface{})
	m["origin_id"] = or.Id
	m["domain_name"] = or.DomainName
	m["origin_path"] = or.OriginPath
	m["connection_attempts"] = or.ConnectionAttempts
	m["connection_timeout"] = or.ConnectionTimeout

	if or.CustomHeaders != nil && or.CustomHeaders.Quantity > 0 {
		headers := flattenCustomHeaders(or.CustomHeaders)
		m["custom_header"] = schema.NewSet(
			schema.HashResource(&schema.Resource{
				Schema: map[string]*schema.Schema{
					"name":  {Type: schema.TypeString, Required: true},
					"value": {Type: schema.TypeString, Required: true},
				},
			}),
			castToInterfaceSlice(headers),
		)
	} else {
		//sm := map[string]interface{}{
		//	//"name":  "",
		//	//"value": "",
		//}
		//inf := make([]interface{}, 1)
		//inf[0] = m
		s := schema.NewSet(
			schema.HashResource(&schema.Resource{
				Schema: map[string]*schema.Schema{
					"name":  {Type: schema.TypeString, Required: true},
					"value": {Type: schema.TypeString, Required: true},
				},
			}),
			nil)
		m["custom_header"] = s
	}

	if or.CustomOriginConfig != nil {
		m["custom_origin_config"] = []interface{}{flattenCustomOriginConfig(or.CustomOriginConfig)}
	}

	if or.OriginShield != nil && or.OriginShield.Enabled {
		m["origin_shield"] = []interface{}{flattenOriginShield(or.OriginShield)}
	}

	return m
}

func parseAwsCloudfrontDistributionIdPartsV2(id string) (tenantID, cfdId, name string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		tenantID, cfdId, name = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
