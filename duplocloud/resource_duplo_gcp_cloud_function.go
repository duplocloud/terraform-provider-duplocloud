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

func gcpCloudFunctionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description: "The GUID of the tenant that the cloud function will be created in.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"name": {
			Description: "The short name of the cloud function.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the cloud function.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"self_link": {
			Description: "The SelfLink of the cloud function.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"build_id": {
			Description: "The ID of the cloud build that built the cloud function.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"version_id": {
			Description: "The current version of the cloud function.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"runtime": {
			Description: "The runtime of the cloud function.\n" +
				"Should be one of:\n\n" +
				" - `nodejs10` : Node.js 10\n" +
				" - `nodejs12` : Node.js 12\n" +
				" - `nodejs14` : Node.js 14\n" +
				" - `python37` : Python 3.7\n" +
				" - `python38` : Python 3.8\n" +
				" - `python39` : Python 3.9\n" +
				" - `go111` :    Go 1.11\n" +
				" - `go113` :    Go 1.13\n" +
				" - `java11` :   Java 11\n" +
				" - `dotnet3` :  .NET Framework 3\n" +
				" - `ruby26` :   Ruby 2.6\n" +
				" - `ruby27` :   Ruby 2.7\n" +
				" - `nodejs6` :  Node.js 6 (deprecated)\n" +
				" - `nodejs8` :  Node.js 8 (deprecated)\n",
			Type:     schema.TypeString,
			Required: true,
		},
		"entrypoint": {
			Description: "The entry point of the cloud function.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"description": {
			Description: "The description of the cloud function.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"available_memory_mb": {
			Description: "The amount of memory available to the cloud function.",
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     256,
		},
		"build_environment_variables": {
			Description: "The build environment variables for this cloud function.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"environment_variables": {
			Description: "The environment variables for this cloud function.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"require_https": {
			Description: "Whether or not to require HTTPS.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"allow_unauthenticated": {
			Description: "Whether or not to allow unauthenticated invocations.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"ingress_type": {
			Description: "The numerical index of ingress type to use for this cloud function.\n" +
				"Should be one of:\n\n" +
				"   - `1` : Allow all\n" +
				"   - `2` : Allow internal traffic\n" +
				"   - `3` : Allow internal traffic and GCP load balancing\n",
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      1,
			ValidateFunc: validation.IntBetween(1, 3),
		},
		"vpc_networking_type": {
			Description: "The numerical index of the VPC networking type to use for this cloud function.\n" +
				"Should be one of:\n\n" +
				"   - `0` : All traffic through the VPC\n" +
				"   - `1` : Only private traffic through the VPC\n" +
				"   - `2` : No VPC networking\n",
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      0,
			ValidateFunc: validation.IntBetween(0, 2),
		},
		"timeout": {
			Description:  "The execution time limit for the cloud function.",
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      60,
			ValidateFunc: validation.IntBetween(1, 540),
		},
		"https_trigger": {
			Description: "Specifies an HTTPS trigger for the cloud function.",
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
					"security_level": {
						Description: "The security level of the HTTPS trigger",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"url": {
						Description: "The URL of the HTTPS trigger",
						Type:        schema.TypeString,
						Computed:    true,
					},
				},
			},
			ConflictsWith: []string{"event_trigger"},
		},
		"event_trigger": {
			Description: "Specifies an event trigger for the cloud function.",
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"event_type": {
						Description: "The type of event that will trigger the function",
						Type:        schema.TypeString,
						Required:    true,
					},
					"resource": {
						Description: "The resource that will trigger the function",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"service": {
						Description: "The service that will trigger the function",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
				},
			},
			ConflictsWith: []string{"https_trigger"},
		},
		"source_archive_url": {
			Description:  "The cloud storage URL where the cloud function package is located.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsURLWithScheme([]string{"gs"}),
		},
		"labels": {
			Description: "The labels assigned to this cloud function.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

// Resource for managing a GCP cloud function
func resourceGcpCloudFunction() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_cloud_function` manages a GCP cloud function in Duplo.",

		ReadContext:   resourceGcpCloudFunctionRead,
		CreateContext: resourceGcpCloudFunctionCreate,
		UpdateContext: resourceGcpCloudFunctionUpdate,
		DeleteContext: resourceGcpCloudFunctionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},
		Schema: gcpCloudFunctionSchema(),
	}
}

/// READ resource
func resourceGcpCloudFunctionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpCloudFunctionRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.GcpCloudFunctionGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s cloud function '%s': %s", tenantID, name, err)
	}

	// Set simple fields first.
	d.SetId(fmt.Sprintf("%s/%s", duplo.TenantID, name))
	resourceGcpCloudFunctionSetData(d, tenantID, name, duplo)
	log.Printf("[TRACE] resourceGcpCloudFunctionRead ******** end")
	return nil
}

/// CREATE resource
func resourceGcpCloudFunctionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpCloudFunctionCreate ******** start")

	// Create the request object.
	rq := expandGcpCloudFunction(d)

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	rp, err := c.GcpCloudFunctionCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s cloud function '%s': %s", tenantID, rq.Name, err)
	}

	// Wait for Duplo to be able to return the cloud function's details.
	id := fmt.Sprintf("%s/%s", tenantID, rq.Name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "cloud function", id, func() (interface{}, duplosdk.ClientError) {
		return c.GcpCloudFunctionGet(tenantID, rq.Name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	resourceGcpCloudFunctionSetData(d, tenantID, rq.Name, rp)
	log.Printf("[TRACE] resourceGcpCloudFunctionCreate ******** end")
	return diags
}

/// UPDATE resource
func resourceGcpCloudFunctionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpCloudFunctionUpdate ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	// Create the request object.
	rq := expandGcpCloudFunction(d)
	rq.Name = d.Get("fullname").(string)

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	rp, err := c.GcpCloudFunctionUpdate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error updating tenant %s cloud function '%s': %s", tenantID, rq.Name, err)
	}
	resourceGcpCloudFunctionSetData(d, tenantID, name, rp)

	log.Printf("[TRACE] resourceGcpCloudFunctionUpdate ******** end")
	return nil
}

/// DELETE resource
func resourceGcpCloudFunctionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpCloudFunctionDelete ******** start")

	// Delete the object with Duplo
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	err := c.GcpCloudFunctionDelete(idParts[0], idParts[1])
	if err != nil {
		return diag.Errorf("Error deleting cloud function '%s': %s", id, err)
	}

	// Wait up to 60 seconds for Duplo to delete the cloud function.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "cloud function", id, func() (interface{}, duplosdk.ClientError) {
		return c.GcpCloudFunctionGet(idParts[0], idParts[1])
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceGcpCloudFunctionDelete ******** end")
	return nil
}

func resourceGcpCloudFunctionSetData(d *schema.ResourceData, tenantID string, name string, duplo *duplosdk.DuploGcpCloudFunction) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("self_link", duplo.SelfLink)
	d.Set("build_id", duplo.BuildID)
	d.Set("version_id", duplo.VersionID)
	d.Set("entrypoint", duplo.EntryPoint)
	d.Set("runtime", duplo.Runtime)
	d.Set("description", duplo.Description)
	d.Set("available_memory_mb", duplo.AvailableMemoryMB)
	d.Set("timeout", duplo.Timeout)
	d.Set("source_archive_url", duplo.SourceArchiveURL)
	d.Set("allow_unauthenticated", duplo.AllowUnauthenticated)
	d.Set("require_https", duplo.RequireHTTPS)
	d.Set("ingress_type", duplo.IngressType)
	d.Set("vpc_networking_type", duplo.VPCNetworkingType)

	flattenGcpLabels(d, duplo.Labels)
	d.Set("build_environment_variables", flattenStringMap(duplo.BuildEnvironmentVariables))
	d.Set("environment_variables", flattenStringMap(duplo.EnvironmentVariables))

	if duplo.TriggerType == 1 {
		d.Set("https_trigger", []interface{}{
			map[string]interface{}{
				"enabled":        true,
				"security_level": duplo.HTTPSTrigger.SecurityLevel,
				"url":            duplo.HTTPSTrigger.URL,
			},
		})
	} else if duplo.TriggerType == 2 {
		d.Set("event_trigger", []interface{}{
			map[string]interface{}{
				"event_type": duplo.EventTrigger.EventType,
				"resource":   duplo.EventTrigger.Resource,
				"service":    duplo.EventTrigger.Service,
			},
		})
	}
}

func expandGcpCloudFunction(d *schema.ResourceData) *duplosdk.DuploGcpCloudFunction {
	duplo := duplosdk.DuploGcpCloudFunction{
		Name:                      d.Get("name").(string),
		Labels:                    expandAsStringMap("labels", d),
		EntryPoint:                d.Get("entrypoint").(string),
		Runtime:                   d.Get("runtime").(string),
		Description:               d.Get("description").(string),
		AvailableMemoryMB:         d.Get("available_memory_mb").(int),
		BuildEnvironmentVariables: expandAsStringMap("build_environment_variables", d),
		EnvironmentVariables:      expandAsStringMap("environment_variables", d),
		Timeout:                   d.Get("timeout").(int),
		SourceArchiveURL:          d.Get("source_archive_url").(string),
		AllowUnauthenticated:      d.Get("allow_unauthenticated").(bool),
		RequireHTTPS:              d.Get("require_https").(bool),
		IngressType:               d.Get("ingress_type").(int),
		VPCNetworkingType:         d.Get("vpc_networking_type").(int),
	}

	if event, err := getOptionalBlockAsMap(d, "event_trigger"); err == nil && len(event) > 0 {
		duplo.TriggerType = 2
		duplo.EventTrigger = &duplosdk.DuploGcpCloudFunctionEventTrigger{EventType: event["event_type"].(string)}

		if v, ok := event["service"]; ok && v.(string) != "" {
			duplo.EventTrigger.Service = v.(string)
		}
		if v, ok := event["resource"]; ok && v.(string) != "" {
			duplo.EventTrigger.Resource = v.(string)
		}
	} else if https, err := getOptionalBlockAsMap(d, "https_trigger"); err == nil && len(https) > 0 && https["enabled"] == true {
		duplo.TriggerType = 1
	}

	return &duplo
}
