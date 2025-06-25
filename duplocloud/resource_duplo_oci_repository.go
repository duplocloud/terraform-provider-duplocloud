package duplocloud

import (
	"context"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Resource for managing an infrastructure's settings.
func resourceOCIRepository() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_k8_oci_repository` manages oci repository (Flux) in duplocloud",

		ReadContext:   resourceOCIRepositoryRead,
		CreateContext: resourceOCIRepositoryCreate,
		UpdateContext: resourceOCIRepositoryUpdate,
		DeleteContext: resourceOCIRepositoryDelete,
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
				Description:  "The GUID of the tenant that the oci repository will be created in.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Description:  "The identifier name for the oci repository in duplocloud",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]{0,62}[a-zA-Z0-9]$`), "Invalid name format, name can be 64 character long and start with an alphabet or digit and can contain hypen or periods"),
			},
			"spec": {
				Type:        schema.TypeList,
				Description: "The spec block of the oci repository",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interval": {
							Description:  "The interval associated to oci repository",
							Type:         schema.TypeString,
							Default:      "5m0s",
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([0-5]?\d)m([0-5]?\d)s$`), "invalid minute second format, valid format 0m0s or 00m00s m[0-59] s[0-59]"),
						},
						"url": {
							Description: "The url of oci repository to be attached",
							Type:        schema.TypeString,
							Required:    true,
						},
						"ref": {
							Description: "The ref of oci repository to be attached",
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"tag": {
										Description: "The tag of oci repository to be attached",
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "latest",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceOCIRepositoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID := idParts[0]
	name := idParts[2]
	log.Printf("[TRACE] resourceOCIRepositoryRead(%s,%s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	duplo, err := c.DuploOCIRepositoryGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to retrieve helm repository details for '%s': %s", name, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	flattenOCI(d, *duplo)
	d.Set("tenant_id", tenantID)
	log.Printf("[TRACE] resourceOCIRepositoryRead(%s,%s): end", tenantID, name)
	return nil
}

func resourceOCIRepositoryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	rq := expandOCI(d)
	log.Printf("[TRACE] resourceOCIRepositoryCreate(%s,%s): start", tenantID, rq.Metadata.Name)

	c := m.(*duplosdk.Client)

	err := c.DuploOCIRepositoryCreate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("resourceOCIRepositoryCreate cannot create helm repository %s for tenant %s error: %s", rq.Metadata.Name, tenantID, err.Error())
	}
	d.SetId(tenantID + "/oci-repository/" + rq.Metadata.Name)

	diags := resourceOCIRepositoryRead(ctx, d, m)
	log.Printf("[TRACE] resourceOCIRepositoryCreate(%s,%s): end", tenantID, rq.Metadata.Name)
	return diags
}

func resourceOCIRepositoryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID := idParts[0]
	name := idParts[2]
	rq := expandOCI(d)
	log.Printf("[TRACE] resourceOCIRepositoryUpdate(%s,%s): start", tenantID, name)

	c := m.(*duplosdk.Client)

	err := c.DuploOCIRepositoryUpdate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("resourceOCIRepositoryUpdate cannot update helm repository %s for tenant %s error: %s", rq.Metadata.Name, tenantID, err.Error())
	}

	diags := resourceOCIRepositoryRead(ctx, d, m)
	log.Printf("[TRACE] resourceOCIRepositoryUpdate(%s,%s): end", tenantID, rq.Metadata.Name)
	return diags
}

func resourceOCIRepositoryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID := idParts[0]
	name := idParts[2]
	log.Printf("[TRACE] resourceOCIRepositoryUpdate(%s,%s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	err := c.DuploOCIRepositoryDelete(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to delete helm repository %s for '%s': %s", name, tenantID, err)
	}

	log.Printf("[TRACE] resourceOCIRepositoryUpdate(%s,%s): end", tenantID, name)
	return nil
}

func expandOCI(d *schema.ResourceData) duplosdk.DuploOCIRepository {
	obj := duplosdk.DuploOCIRepository{}
	obj.Metadata = duplosdk.DuploOCIMetadata{
		Name: d.Get("name").(string),
	}
	obj.Spec = &duplosdk.DuploOCISpec{
		Interval: d.Get("spec.0.interval").(string),
		URL:      d.Get("spec.0.url").(string),
	}
	if ref, ok := d.GetOk("spec.0.ref"); ok {
		obj.Spec.Ref = &duplosdk.DuploOCISpecRef{
			Tag: ref.([]interface{})[0].(map[string]interface{})["tag"].(string),
		}
	}
	return obj
}

func flattenOCI(d *schema.ResourceData, rb duplosdk.DuploOCIRepository) {
	d.Set("name", rb.Metadata.Name)
	m := map[string]interface{}{"interval": rb.Spec.Interval, "url": rb.Spec.URL}
	if rb.Spec.Ref != nil {
		ref := []interface{}{}
		ref = append(ref, map[string]interface{}{"tag": rb.Spec.Ref.Tag})
		m["ref"] = ref
	} else {
		m["ref"] = nil
	}
	s := []interface{}{}
	s = append(s, m)
	d.Set("spec", s)

}
