package duplocloud

import (
	"context"
	"log"
	"regexp"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Resource for managing an infrastructure's settings.
func resourceHelmRepository() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_helm_repository` manages helm repository in duplocloud",

		ReadContext:   resourceHelmRepositoryRead,
		CreateContext: resourceHelmRepositoryCreate,
		UpdateContext: resourceHelmRepositoryUpdate,
		DeleteContext: resourceHelmRepositoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The GUID of the tenant that the storage bucket will be created in.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Description:  "The identifier name for the helm repository in duplocloud",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]{0,62}[a-zA-Z0-9]$`), "Invalid name format, name can be 64 character long and start with an alphabet or digit and can contain hypen or periods"),
			},
			"interval": {
				Description:  "The interval associated to helm repository",
				Type:         schema.TypeString,
				Default:      "5m0s",
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([0-5]?\d)m([0-5]?\d)s$`), "invalid minute second format, valid format 0m0s or 00m00s m[0-59] s[0-59]"),
			},

			"url": {
				Description: "The url of helm repository to be attached",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceHelmRepositoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID := idParts[0]
	name := idParts[2]
	log.Printf("[TRACE] resourceHelmRepositoryRead(%s,%s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	duplo, err := c.DuploHelmRepositoryGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to retrieve helm repository details for '%s': %s", name, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	flattenHelm(d, *duplo)

	log.Printf("[TRACE] resourceHelmRepositoryRead(%s,%s): end", tenantID, name)
	return nil
}

func resourceHelmRepositoryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	rq := expandHelm(d)
	log.Printf("[TRACE] resourceHelmRepositoryCreate(%s,%s): start", tenantID, rq.Metadata.Name)

	c := m.(*duplosdk.Client)

	err := c.DuploHelmRepositoryCreate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("resourceHelmRepositoryCreate cannot create helm repository %s for tenant %s error: %s", rq.Metadata.Name, tenantID, err.Error())
	}
	d.SetId(tenantID + "/helm-repository/" + rq.Metadata.Name)

	diags := resourceHelmRepositoryRead(ctx, d, m)
	log.Printf("[TRACE] resourceHelmRepositoryCreate(%s,%s): end", tenantID, rq.Metadata.Name)
	return diags
}

func resourceHelmRepositoryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID := idParts[0]
	name := idParts[2]
	rq := expandHelm(d)
	log.Printf("[TRACE] resourceHelmRepositoryUpdate(%s,%s): start", tenantID, name)

	c := m.(*duplosdk.Client)

	err := c.DuploHelmRepositoryUpdate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("resourceHelmRepositoryUpdate cannot update helm repository %s for tenant %s error: %s", rq.Metadata.Name, tenantID, err.Error())
	}

	diags := resourceHelmRepositoryRead(ctx, d, m)
	log.Printf("[TRACE] resourceHelmRepositoryUpdate(%s,%s): end", tenantID, rq.Metadata.Name)
	return diags
}

func resourceHelmRepositoryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID := idParts[0]
	name := idParts[2]
	log.Printf("[TRACE] resourceHelmRepositoryDelete(%s,%s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	err := c.DuploHelmRepositoryDelete(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to delete helm repository %s for '%s': %s", name, tenantID, err)
	}

	log.Printf("[TRACE] resourceHelmRepositoryDelete(%s,%s): end", tenantID, name)
	return nil
}

func expandHelm(d *schema.ResourceData) duplosdk.DuploHelmRepository {
	obj := duplosdk.DuploHelmRepository{}
	obj.Metadata = duplosdk.DuploHelmMetadata{
		Name: d.Get("name").(string),
	}
	obj.Spec = duplosdk.DuploHelmSpec{
		Interval: d.Get("interval").(string),
		URL:      d.Get("url").(string),
	}

	return obj
}

func flattenHelm(d *schema.ResourceData, rb duplosdk.DuploHelmRepository) {
	d.Set("name", rb.Metadata.Name)
	d.Set("interval", rb.Spec.Interval)
	d.Set("url", rb.Spec.URL)
}
