package data_sources

import (
	"context"
	"terraform-provider-duplocloud/duplocloud"

	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func nativeHostImageSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"image_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"os": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"username": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"region": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"k8s_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"arch": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"is_kubernetes": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     duplocloud.KeyValueSchema(),
		},
	}
}

func nativeHostImageDataSourceSchema(single bool) map[string]*schema.Schema {
	img := nativeHostImageSchema()

	// For a single image, the name is required, not computed.
	var result map[string]*schema.Schema
	if single {
		result = img
		result["name"].Optional = true
		result["arch"].Optional = true
		result["is_kubernetes"].Optional = true

		// For a list of images, move the list under the result key.
	} else {
		result = map[string]*schema.Schema{
			"images": {
				Description: "The list of images for this tenant.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: img,
				},
			},
		}

	}

	// Always require the tenant ID.
	result["tenant_id"] = &schema.Schema{
		Description: "The tenant ID",
		Type:        schema.TypeString,
		Required:    true,
	}

	return result
}

func dataSourceNativeHostImages() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_native_host_images` retrieves a list of applicable images for a given tenant.",

		ReadContext: dataSourceNativeHostImagesRead,
		Schema:      nativeHostImageDataSourceSchema(false),
	}
}

func dataSourceNativeHostImage() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_native_host_image` retrieves details of a specific image for a given tenant.",

		ReadContext: dataSourceNativeHostImageRead,
		Schema:      nativeHostImageDataSourceSchema(true),
	}
}

func dataSourceNativeHostImagesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	log.Printf("[TRACE] dataSourceNativeHostImagesRead(%s): start", tenantID)

	// Get all of the plan images from duplo.
	c := m.(*duplosdk.Client)
	all, diags := getNativeHostImages(c, tenantID)
	if diags != nil {
		return diags
	}

	// Populate the results from the list.
	d.Set("images", flattenNativeHostImages(all))

	d.SetId(tenantID)

	log.Printf("[TRACE] dataSourceNativeHostImagesRead(%s): end", tenantID)
	return nil
}

func dataSourceNativeHostImageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	arch := d.Get("arch").(string)
	isKubernetes := d.Get("is_kubernetes").(bool)

	log.Printf("[TRACE] dataSourceNativeHostImageRead(%s): start", tenantID)

	// Get the plan image from Duplo.
	c := m.(*duplosdk.Client)
	image, diags := getNativeHostImage(c, tenantID, name, arch, isKubernetes)
	if diags != nil {
		return diags
	}
	d.SetId(tenantID)

	// Populate the results
	d.Set("name", image.Name)
	d.Set("image_id", image.ImageId)
	d.Set("os", image.OS)
	d.Set("username", image.Username)
	d.Set("region", image.Region)
	d.Set("k8s_version", image.K8sVersion)
	d.Set("arch", image.Arch)
	d.Set("is_kubernetes", image.K8sVersion != "")
	d.Set("tags", duplocloud.keyValueToState("images[].tags", image.Tags))

	log.Printf("[TRACE] dataSourcePlanImageRead(): end")
	return nil
}

func getNativeHostImages(c *duplosdk.Client, tenantID string) (*[]duplosdk.DuploNativeHostImage, diag.Diagnostics) {

	// First, try the newer method of getting the plan images.
	duplo, err := c.NativeHostImageGetList(tenantID)
	if err != nil && !err.PossibleMissingAPI() {
		return nil, diag.Errorf("failed to retrieve native host images for '%s': %s", tenantID, err)
	}

	// If it failed, try the fallback method.
	if duplo == nil || err != nil {
		duplo, err = c.LegacyNativeHostImageGetList(tenantID)
		if duplo == nil && err == nil {
			return nil, diag.Errorf("no images were returned")
		}
		if err != nil {
			return nil, diag.Errorf("failed to retrieve native host images for '%s': %s", tenantID, err)
		}
	}

	return duplo, nil
}

func getNativeHostImage(c *duplosdk.Client, tenantID, name, arch string, isKubernetes bool) (*duplosdk.DuploNativeHostImage, diag.Diagnostics) {

	// First, validate parameters.
	if arch == "" {
		arch = "amd64"
	}
	if name == "" && !isKubernetes {
		return nil, diag.Errorf("must query by name or is_kubernetes")
	}

	// Then, get the full list.
	list, err := getNativeHostImages(c, tenantID)
	if err != nil {
		return nil, err
	}

	// Finally, return the matching image.
	for _, v := range *list {
		v.Arch = ""
		if isKubernetes {
			if v.K8sVersion != "" && (v.Arch == "" || v.Arch == arch) {
				return &v, nil
			}
		} else if name == v.Name {
			return &v, nil
		}
	}

	return nil, diag.Errorf("failed to retrieve the native host image for '%s': no matching image found", tenantID)
}

func flattenNativeHostImages(list *[]duplosdk.DuploNativeHostImage) []interface{} {
	result := make([]interface{}, 0, len(*list))

	for _, image := range *list {
		result = append(result, map[string]interface{}{
			"name":          image.Name,
			"image_id":      image.ImageId,
			"os":            image.OS,
			"username":      image.Username,
			"region":        image.Region,
			"k8s_version":   image.K8sVersion,
			"arch":          image.Arch,
			"is_kubernetes": image.K8sVersion != "",
			"tags":          duplocloud.keyValueToState("images[].tags", image.Tags),
		})
	}

	return result
}
