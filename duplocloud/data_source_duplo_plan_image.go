package duplocloud

import (
	"context"
	"fmt"

	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func planImageDataSourceSchema(single bool) map[string]*schema.Schema {

	// Create a fully computed schema.
	img := planImageSchema()
	for k := range img {
		img[k].Required = false
		img[k].Computed = true
	}

	// For a single image, the name is required, not computed.
	var result map[string]*schema.Schema
	if single {
		result = img
		result["name"].Computed = false
		result["name"].Required = true

		// For a list of images, move the list under the result key.
	} else {
		result = map[string]*schema.Schema{
			"images": {
				Description: "The list of images for this plan.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: img,
				},
			},
		}

	}

	// Always require the plan ID.
	result["plan_id"] = &schema.Schema{
		Description: "The plan ID",
		Type:        schema.TypeString,
		Required:    true,
	}

	return result
}

func dataSourcePlanImages() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_images` retrieves a list of images for a given plan.",

		ReadContext: dataSourcePlanImagesRead,
		Schema:      planImageDataSourceSchema(false),
	}
}

func dataSourcePlanImage() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan_image` retrieves details of a specific image for a given plan.",

		ReadContext: dataSourcePlanImageRead,
		Schema:      planImageDataSourceSchema(true),
	}
}

func dataSourcePlanImagesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	log.Printf("[TRACE] dataSourcePlanImagesRead(%s): start", planID)

	// Get all of the plan images from duplo.
	c := m.(*duplosdk.Client)
	all, diags := getPlanImages(c, planID)
	if diags != nil {
		return diags
	}

	// Populate the results from the list.
	d.Set("images", flattenPlanImages(all))

	d.SetId(planID)

	log.Printf("[TRACE] dataSourcePlanImagesRead(%s): end", planID)
	return nil
}

/// READ/SEARCH resources
func dataSourcePlanImageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	planID := d.Get("plan_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] dataSourcePlanImageRead(%s, %s): start", planID, name)

	// Get the plan image from Duplo.
	c := m.(*duplosdk.Client)
	image, diags := getPlanImage(c, planID, name)
	if diags != nil {
		return diags
	}
	d.SetId(fmt.Sprintf("%s/%s", planID, name))

	// Populate the results
	d.Set("image_id", image.ImageId)
	d.Set("os", image.OS)
	d.Set("username", image.Username)
	d.Set("tags", keyValueToState("images[].tags", image.Tags))

	log.Printf("[TRACE] dataSourcePlanImageRead(): end")
	return nil
}

func getPlanImage(c *duplosdk.Client, planID, name string) (*duplosdk.DuploPlanImage, diag.Diagnostics) {

	// First, try the newer method of getting the plan images.
	duplo, err := c.PlanImageGet(planID, name)
	if err != nil && !err.PossibleMissingAPI() {
		return nil, diag.Errorf("failed to retrieve plan image for '%s/%s': %s", planID, name, err)
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

		if plan.Images != nil {
			for _, v := range *plan.Images {
				if v.Name == name {
					duplo = &v
				}
			}
		}
	}

	return duplo, nil
}
