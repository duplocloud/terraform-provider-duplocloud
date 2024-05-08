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

func gcpSchedulerJobSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the scheduler job will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the scheduler job.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the scheduler job.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"self_link": {
			Description: "The SelfLink of the scheduler job.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"schedule": {
			Description: "The desired schedule, in cron format.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"timezone": {
			Description: "The timezone used to determine the schedule, in UNIX format",
			Type:        schema.TypeString,
			Required:    true,
		},
		"attempt_deadline": {
			Description: "The attempt deadline for the scheduler job.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"description": {
			Description: "The description of the scheduler job.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"pubsub_target": {
			Description: "Specifies a pubsub target for the scheduler job.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"topic_name": {
						Description: "The name of the topic to target",
						Type:        schema.TypeString,
						Required:    true,
					},
					"data": {
						Description: "The data to send to the pubsub topic.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						//AtLeastOneOf: []string{"data", "attributes"},
					},
					"attributes": {
						Description: "The attributes to send to the pubsub target.",
						Type:        schema.TypeMap,
						Optional:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						//AtLeastOneOf: []string{"data", "attributes"},
					},
				},
			},
			ConflictsWith: []string{"http_target", "app_engine_target"},
		},
		"http_target": {
			Description: "Specifies an HTTP target for the scheduler job.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"method": {
						Description:  "The HTTP method to use.",
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"POST", "GET", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"}, false),
					},
					"headers": {
						Description: "The HTTP headers to send.",
						Type:        schema.TypeMap,
						Optional:    true,
						Computed:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"body": {
						Description: "The HTTP request body to send.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"uri": {
						Description: "The request URI.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"oidc_token": {
						Description: "Specifies OIDC authentication.",
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"enabled": {
									Description: "Must be set to `true`.",
									Type:        schema.TypeBool,
									Optional:    true,
									Default:     true,
								},
								"audience": {
									Description: "The OIDC token audience.",
									Type:        schema.TypeString,
									Optional:    true,
									Computed:    true,
								},
								"service_account_email": {
									Description: "The OIDC token service account email.",
									Type:        schema.TypeString,
									Computed:    true,
								},
							},
						},
						//ConflictsWith: []string{"oauth_token"},
					},
					"oauth_token": {
						Description: "Specifies OAuth authentication.",
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"enabled": {
									Description: "Must be set to `true`.",
									Type:        schema.TypeBool,
									Optional:    true,
									Default:     true,
								},
								"scope": {
									Description: "The OAuth token scope.",
									Type:        schema.TypeString,
									Optional:    true,
									Computed:    true,
								},
								"service_account_email": {
									Description: "The OAuth token service account email.",
									Type:        schema.TypeString,
									Computed:    true,
								},
							},
						},
						//ConflictsWith: []string{"oidc_token"},
					},
				},
			},
			ConflictsWith: []string{"pubsub_target", "app_engine_target"},
		},
		"app_engine_target": {
			Description: "Specifies an App Engine target for the scheduler job.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"method": {
						Description:  "The HTTP method to use.",
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"POST", "GET", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"}, false),
					},
					"headers": {
						Description: "The HTTP headers to send.",
						Type:        schema.TypeMap,
						Optional:    true,
						Computed:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"body": {
						Description: "The HTTP request body to send.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"relative_uri": {
						Description: "The relative URI.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"routing": {
						Description: "Specifies App Engine routing.",
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"service": {
									Description: "The App Engine service.",
									Type:        schema.TypeString,
									Optional:    true,
									Computed:    true,
								},
								"version": {
									Description: "The App Engine service version.",
									Type:        schema.TypeString,
									Optional:    true,
									Computed:    true,
								},
								"host": {
									Description: "The App Engine host.",
									Type:        schema.TypeString,
									Optional:    true,
									Computed:    true,
								},
								"instance": {
									Description: "The App Engine instance.",
									Type:        schema.TypeString,
									Optional:    true,
									Computed:    true,
								},
							},
						},
					},
				},
			},
			ConflictsWith: []string{"pubsub_target", "http_target"},
		},
	}
}

// Resource for managing a GCP scheduler job
func resourceGcpSchedulerJob() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_scheduler_job` manages a GCP scheduler job in Duplo.",

		ReadContext:   resourceGcpSchedulerJobRead,
		CreateContext: resourceGcpSchedulerJobCreate,
		UpdateContext: resourceGcpSchedulerJobUpdate,
		DeleteContext: resourceGcpSchedulerJobDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
		Schema: gcpSchedulerJobSchema(),
	}
}

// READ resource
func resourceGcpSchedulerJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpSchedulerJobRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.GcpSchedulerJobGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s scheduler job '%s': %s", tenantID, name, err)
	}
	name = d.Get("name").(string)
	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/%s", duplo.TenantID, name))
	resourceGcpSchedulerJobSetData(d, tenantID, name, duplo)
	log.Printf("[TRACE] resourceGcpSchedulerJobRead ******** end")
	return nil
}

// CREATE resource
func resourceGcpSchedulerJobCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpSchedulerJobCreate ******** start")

	// Create the request object.
	rq := expandGcpSchedulerJob(d)

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	rp, err := c.GcpSchedulerJobCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s scheduler job '%s': %s", tenantID, rq.Name, err)
	}

	// Wait for Duplo to be able to return the scheduler job's details.
	id := fmt.Sprintf("%s/%s", tenantID, rp.Name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "scheduler job", id, func() (interface{}, duplosdk.ClientError) {
		return c.GcpSchedulerJobGet(tenantID, rp.Name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)
	diags = resourceGcpSchedulerJobRead(ctx, d, m)
	//resourceGcpSchedulerJobSetData(d, tenantID, rq.Name, rp)
	log.Printf("[TRACE] resourceGcpSchedulerJobCreate ******** end")
	return diags
}

// UPDATE resource
func resourceGcpSchedulerJobUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpSchedulerJobUpdate ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	// Create the request object.
	rq := expandGcpSchedulerJob(d)
	rq.Name = d.Get("fullname").(string)

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	rp, err := c.GcpSchedulerJobUpdate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error updating tenant %s scheduler job '%s': %s", tenantID, rq.Name, err)
	}
	resourceGcpSchedulerJobSetData(d, tenantID, name, rp)

	log.Printf("[TRACE] resourceGcpSchedulerJobUpdate ******** end")
	return nil
}

// DELETE resource
func resourceGcpSchedulerJobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpSchedulerJobDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	err := c.GcpSchedulerJobDelete(idParts[0], idParts[1])
	if err != nil {
		return diag.Errorf("Error deleting scheduler job '%s': %s", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the scheduler job.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "scheduler job", id, func() (interface{}, duplosdk.ClientError) {
		return c.GcpSchedulerJobGet(idParts[0], idParts[1])
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceGcpSchedulerJobDelete ******** end")
	return nil
}

func resourceGcpSchedulerJobSetData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploGcpSchedulerJob) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("self_link", duplo.SelfLink)
	d.Set("description", duplo.Description)
	d.Set("schedule", duplo.Schedule)
	d.Set("timezone", duplo.TimeZone)
	d.Set("attempt_deadline", duplo.AttemptDeadline)

	if duplo.TargetType == duplosdk.GcpSchedulerJob_TargetType_PubsubTarget {
		d.Set("http_target", []interface{}{})
		d.Set("app_engine_target", []interface{}{})
		d.Set("pubsub_target", []interface{}{
			map[string]interface{}{
				"topic_name": duplo.PubsubTarget.TopicName,
				"data":       duplo.PubsubTargetData,
				"attributes": flattenStringMap(duplo.PubsubTarget.Attributes),
			},
		})
	} else if duplo.TargetType == duplosdk.GcpSchedulerJob_TargetType_HttpTarget {
		m := map[string]interface{}{
			"method":  flattenGcpSchedulerJobHttpMethod(duplo.AppEngineTarget.HTTPMethod),
			"headers": flattenStringMap(duplo.AppEngineTarget.Headers),
			"body":    duplo.AnyHTTPTargetBody,
			"uri":     duplo.HTTPTarget.URI,
		}

		if duplo.HTTPTarget.AuthorizationHeaderCase == duplosdk.GcpSchedulerJob_AuthorizationHeader_OauthToken {
			m["oauth_token"] = []interface{}{
				map[string]interface{}{
					"enabled":               true,
					"scope":                 duplo.HTTPTarget.OAuthToken.Scope,
					"service_account_email": duplo.HTTPTarget.OAuthToken.ServiceAccountEmail,
				},
			}
		} else if duplo.HTTPTarget.AuthorizationHeaderCase == duplosdk.GcpSchedulerJob_AuthorizationHeader_OidcToken {
			m["oidc_token"] = []interface{}{
				map[string]interface{}{
					"enabled":               true,
					"audience":              duplo.HTTPTarget.OidcToken.Audience,
					"service_account_email": duplo.HTTPTarget.OidcToken.ServiceAccountEmail,
				},
			}
		}

		d.Set("pubsub_target", []interface{}{})
		d.Set("app_engine_target", []interface{}{})
		d.Set("http_target", []interface{}{m})
	} else if duplo.TargetType == duplosdk.GcpSchedulerJob_TargetType_AppEngineHttpTarget {
		m := map[string]interface{}{
			"method":       flattenGcpSchedulerJobHttpMethod(duplo.AppEngineTarget.HTTPMethod),
			"headers":      flattenStringMap(duplo.AppEngineTarget.Headers),
			"body":         duplo.AnyHTTPTargetBody,
			"relative_uri": duplo.AppEngineTarget.RelativeURI,
		}

		d.Set("pubsub_target", []interface{}{})
		d.Set("http_target", []interface{}{})
		d.Set("app_engine_target", []interface{}{m})
	}
}

func expandGcpSchedulerJob(d *schema.ResourceData) *duplosdk.DuploGcpSchedulerJob {
	duplo := duplosdk.DuploGcpSchedulerJob{
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		Schedule:        d.Get("schedule").(string),
		TimeZone:        d.Get("timezone").(string),
		AttemptDeadline: d.Get("attempt_deadline").(string),
	}

	if pubsub, err := getOptionalBlockAsMap(d, "pubsub_target"); err == nil && len(pubsub) > 0 {
		duplo.PubsubTarget = &duplosdk.DuploGcpSchedulerJobPubsubTarget{TopicName: pubsub["topic_name"].(string)}

		if v, ok := pubsub["data"]; ok && v != nil && v.(string) != "" {
			duplo.PubsubTargetData = v.(string)
		}
		if v := fieldToStringMap("attributes", pubsub); len(v) > 0 {
			duplo.PubsubTargetAttributes = v
		}

	} else if http, err := getOptionalBlockAsMap(d, "http_target"); err == nil && len(http) > 0 {
		duplo.HTTPTarget = &duplosdk.DuploGcpSchedulerJobHTTPTarget{
			HTTPMethod: expandGcpSchedulerJobHttpMethod(http["method"].(string)),
			URI:        http["uri"].(string),
		}

		if v, ok := http["body"]; ok && v != nil && v.(string) != "" {
			duplo.AnyHTTPTargetBody = v.(string)
		}
		if v := fieldToStringMap("headers", http); len(v) > 0 {
			duplo.AnyHTTPTargetHeaders = v
		}

		if oidc, err := getOptionalNestedBlockAsMap(http, "oidc_token"); err == nil && len(oidc) > 0 {
			duplo.HTTPTarget.OidcToken = &duplosdk.DuploGcpSchedulerJobOidcToken{}
			if v, ok := oidc["audience"]; ok && v != nil && v.(string) != "" {
				duplo.HTTPTarget.OidcToken.Audience = v.(string)
			} else {
				duplo.HTTPTarget.OidcToken.Audience = duplo.HTTPTarget.URI
			}
		} else if oauth, err := getOptionalNestedBlockAsMap(http, "oauth_token"); err == nil && len(oauth) > 0 {
			duplo.HTTPTarget.OAuthToken = &duplosdk.DuploGcpSchedulerJobOAuthToken{}
			if v, ok := oauth["scope"]; ok && v != nil && v.(string) != "" {
				duplo.HTTPTarget.OAuthToken.Scope = v.(string)
			} else {
				duplo.HTTPTarget.OAuthToken.Scope = "https://www.googleapis.com/auth/cloud-platform"
			}
		}

	} else if appeng, err := getOptionalBlockAsMap(d, "app_engine_target"); err == nil && len(appeng) > 0 {
		duplo.AppEngineTarget = &duplosdk.DuploGcpSchedulerJobAppEngineTarget{
			HTTPMethod:  expandGcpSchedulerJobHttpMethod(appeng["method"].(string)),
			RelativeURI: appeng["relative_uri"].(string),
		}

		if v, ok := appeng["body"]; ok && v != nil && v.(string) != "" {
			duplo.AnyHTTPTargetBody = v.(string)
		}
		if v := fieldToStringMap("headers", appeng); len(v) > 0 {
			duplo.AnyHTTPTargetHeaders = v
		}

		if routing, err := getOptionalNestedBlockAsMap(appeng, "routing"); err == nil && len(routing) > 0 {
			duplo.AppEngineTarget.AppEngineRouting = &duplosdk.DuploGcpSchedulerJobAppEngineRouting{}

			if v, ok := routing["service"]; ok && v != nil && v.(string) != "" {
				duplo.AppEngineTarget.AppEngineRouting.Service = v.(string)
			}
			if v, ok := routing["version"]; ok && v != nil && v.(string) != "" {
				duplo.AppEngineTarget.AppEngineRouting.Version = v.(string)
			}
			if v, ok := routing["host"]; ok && v != nil && v.(string) != "" {
				duplo.AppEngineTarget.AppEngineRouting.Host = v.(string)
			}
			if v, ok := routing["instance"]; ok && v != nil && v.(string) != "" {
				duplo.AppEngineTarget.AppEngineRouting.Instance = v.(string)
			}
		}
	}

	return &duplo
}

func flattenGcpSchedulerJobHttpMethod(httpMethod int) string {
	switch httpMethod {
	case duplosdk.GcpSchedulerJob_HttpMethod_Post:
		return "POST"
	case duplosdk.GcpSchedulerJob_HttpMethod_Get:
		return "GET"
	case duplosdk.GcpSchedulerJob_HttpMethod_Head:
		return "HEAD"
	case duplosdk.GcpSchedulerJob_HttpMethod_Put:
		return "PUT"
	case duplosdk.GcpSchedulerJob_HttpMethod_Delete:
		return "DELETE"
	case duplosdk.GcpSchedulerJob_HttpMethod_Patch:
		return "PATCH"
	case duplosdk.GcpSchedulerJob_HttpMethod_Options:
		return "OPTIONS"
	default:
		return "POST"
	}
}

func expandGcpSchedulerJobHttpMethod(httpMethod string) int {
	switch httpMethod {
	case "POST":
		return duplosdk.GcpSchedulerJob_HttpMethod_Post
	case "GET":
		return duplosdk.GcpSchedulerJob_HttpMethod_Get
	case "HEAD":
		return duplosdk.GcpSchedulerJob_HttpMethod_Head
	case "PUT":
		return duplosdk.GcpSchedulerJob_HttpMethod_Put
	case "DELETE":
		return duplosdk.GcpSchedulerJob_HttpMethod_Delete
	case "PATCH":
		return duplosdk.GcpSchedulerJob_HttpMethod_Patch
	case "OPTIONS":
		return duplosdk.GcpSchedulerJob_HttpMethod_Options
	default:
		return duplosdk.GcpSchedulerJob_HttpMethod_Post
	}
}
