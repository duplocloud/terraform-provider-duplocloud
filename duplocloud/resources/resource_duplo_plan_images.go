package resources

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplocloud"
	"terraform-provider-duplocloud/duplocloud/data_sources"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// planImageSchema returns a Terraform schema to represent a plan image
func planImageSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"image_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"os": {
			Type:     schema.TypeString,
			Required: true,
		},
		"username": {
			Type:     schema.TypeString,
			Required: true,
		},
		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     duplocloud.KeyValueSchema(),
		},
	}
}

// Resource for managing an AWS ElasticSearch instance
func resourcePlanImages() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_images` manages the list of images avaialble to a plan in Duplo.\n\n" +
			"This resource allows you take control of individual images for a specific plan.",

		ReadContext:   resourcePlanImagesRead,
		CreateContext: resourcePlanImagesCreateOrUpdate,
		UpdateContext: resourcePlanImagesCreateOrUpdate,
		DeleteContext: resourcePlanImagesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"plan_id": {
				Description: "The ID of the plan to configure.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"image": {
				Description: "A list of images to manage.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: planImageSchema(),
				},
			},
			"delete_unspecified_images": {
				Description: "Whether or not this resource should delete any images not specified by this resource. " +
					"**WARNING:**  It is not recommended to change the default value of `false`.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"images": {
				Description: "A complete list of images for this plan, even ones not being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: planImageSchema(),
				},
			},
			"specified_images": {
				Description: "A list of image names being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourcePlanImagesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Id()
	log.Printf("[TRACE] resourcePlanImagesRead(%s): start", planID)

	c := m.(*duplosdk.Client)

	// First, try the newer method of getting the plan images.
	duplo, err := c.PlanImageGetList(planID)
	if err != nil && !err.PossibleMissingAPI() {
		return diag.Errorf("failed to retrieve plan images for '%s': %s", planID, err)
	}

	// If it failed, try the fallback method.
	if duplo == nil {
		plan, err := c.PlanGet(planID)
		if err != nil {
			return diag.Errorf("failed to read plan images: %s", err)
		}
		if plan == nil {
			return diag.Errorf("failed to read plan: %s", planID)
		}

		duplo = plan.Images
	}

	// Set the simple fields first.
	d.Set("images", data_sources.flattenPlanImages(duplo))

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := duplocloud.getAsStringArray(d, "specified_images"); ok && v != nil {
		d.Set("image", data_sources.flattenPlanImages(selectPlanImages(duplo, *v)))
	}

	log.Printf("[TRACE] resourcePlanImagesRead(%s): end", planID)
	return nil
}

func resourcePlanImagesCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] resourcePlanImagesCreateOrUpdate(%s): start", planID)

	// Get all of the plan images from duplo.
	c := m.(*duplosdk.Client)
	all, diags := getPlanImages(c, planID)
	if diags != nil {
		return diags
	}

	// Get the previous and desired plan images
	previous, desired := getPlanImagesChange(all, d)

	// Apply the changes via Duplo
	var err duplosdk.ClientError
	if d.Get("delete_unspecified_images").(bool) {
		err = c.PlanReplaceImages(planID, desired)
	} else {
		err = c.PlanChangeImages(planID, previous, desired)
	}
	if err != nil {
		return diag.Errorf("Error updating plan images for '%s': %s", planID, err)
	}
	d.SetId(planID)

	diags = resourcePlanImagesRead(ctx, d, m)
	log.Printf("[TRACE] resourcePlanImagesCreateOrUpdate(%s): end", planID)
	return diags
}

func resourcePlanImagesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Id()
	log.Printf("[TRACE] resourcePlanImagesDelete(%s): start", planID)

	// Get all of the plan images from duplo.
	c := m.(*duplosdk.Client)
	all, diags := getPlanImages(c, planID)
	if diags != nil {
		return diags
	}

	// Get the previous and desired plan images
	previous, _ := getPlanImagesChange(all, d)
	desired := &[]duplosdk.DuploPlanImage{}

	// Apply the changes via Duplo
	var err duplosdk.ClientError
	if d.Get("delete_unspecified_images").(bool) {
		err = c.PlanReplaceImages(planID, desired)
	} else {
		err = c.PlanChangeImages(planID, previous, desired)
	}
	if err != nil {
		return diag.Errorf("Error updating plan images for '%s': %s", planID, err)
	}

	log.Printf("[TRACE] resourcePlanImagesDelete(%s): end", planID)
	return nil
}

// Utiliy function to return a filtered list of plan images, given the selected keys.
func selectPlanImagesFromMap(all *[]duplosdk.DuploPlanImage, keys map[string]interface{}) *[]duplosdk.DuploPlanImage {
	certs := make([]duplosdk.DuploPlanImage, 0, len(keys))
	for _, pc := range *all {
		if _, ok := keys[pc.Name]; ok {
			certs = append(certs, pc)
		}
	}

	return &certs
}

// Utiliy function to return a filtered list of tenant metadata, given the selected keys.
func selectPlanImages(all *[]duplosdk.DuploPlanImage, keys []string) *[]duplosdk.DuploPlanImage {
	specified := map[string]interface{}{}
	for _, k := range keys {
		specified[k] = struct{}{}
	}

	return selectPlanImagesFromMap(all, specified)
}

func expandPlanImages(fieldName string, d *schema.ResourceData) *[]duplosdk.DuploPlanImage {
	var ary []duplosdk.DuploPlanImage

	if v, ok := d.GetOk(fieldName); ok && v != nil && len(v.([]interface{})) > 0 {
		kvs := v.([]interface{})
		log.Printf("[TRACE] expandPlanImages ********: found %s", fieldName)
		ary = make([]duplosdk.DuploPlanImage, 0, len(kvs))
		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, duplosdk.DuploPlanImage{
				Name:     kv["name"].(string),
				ImageId:  kv["image_id"].(string),
				Username: kv["username"].(string),
				OS:       kv["os"].(string),
			})
		}
	}

	return &ary
}

func getPlanImages(c *duplosdk.Client, planID string) (*[]duplosdk.DuploPlanImage, diag.Diagnostics) {

	// First, try the newer method of getting the plan images.
	duplo, err := c.PlanImageGetList(planID)
	if err != nil && !err.PossibleMissingAPI() {
		return nil, diag.Errorf("failed to retrieve plan images for '%s': %s", planID, err)
	}

	// If it failed, try the fallback method.
	if duplo == nil {
		plan, err := c.PlanGet(planID)
		if err != nil {
			return nil, diag.Errorf("failed to read plan images: %s", err)
		}
		if plan == nil {
			return nil, diag.Errorf("failed to read plan: %s", planID)
		}

		duplo = plan.Images
	}

	return duplo, nil
}

func getPlanImagesChange(all *[]duplosdk.DuploPlanImage, d *schema.ResourceData) (previous, desired *[]duplosdk.DuploPlanImage) {
	if v, ok := duplocloud.getAsStringArray(d, "specified_images"); ok && v != nil {
		previous = selectPlanImages(all, *v)
	} else {
		previous = &[]duplosdk.DuploPlanImage{}
	}

	// Collect the desired state of settings specified by the user.
	desired = expandPlanImages("image", d)
	specified := make([]string, len(*desired))
	for i, pc := range *desired {
		specified[i] = pc.Name
	}

	// Track the change
	d.Set("specified_images", specified)

	return
}
