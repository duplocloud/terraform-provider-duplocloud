package duplocloud

import (
	"context"
	"log"
	"strings"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

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
			Elem:     KeyValueSchema(),
		},
	}
}

func nativeHostImageDataSourceSchema(single bool) map[string]*schema.Schema {
	img := nativeHostImageSchema()

	// For a single image, the name is required, not computed.
	var result map[string]*schema.Schema
	if single {
		result = img
		result["name"].Computed = false
		result["arch"].Computed = false
		result["os"].Computed = false
		result["k8s_version"].Computed = false

		result["is_kubernetes"].Optional = true
		result["is_kubernetes"].Deprecated = "This field is not in use. Use k8s_version for precise filtering"

		result["name"].Required = true
		result["arch"].Required = true
		result["os"].Required = true
		result["k8s_version"].Optional = true

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
	k8Ver := d.Get("k8s_version").(string)
	isKube := d.Get("is_kubernetes").(bool)
	os := d.Get("os").(string)
	log.Printf("[TRACE] dataSourceNativeHostImageRead(%s): start", tenantID)

	// Get the plan image from Duplo.
	c := m.(*duplosdk.Client)
	image, diags := getNativeHostImage(c, tenantID, name, arch, os, k8Ver, isKube)
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
	d.Set("tags", keyValueToState("images[].tags", image.Tags))

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

func getNativeHostImage(c *duplosdk.Client, tenantID, name, arch, os, k8Ver string, isKube bool) (*duplosdk.DuploNativeHostImage, diag.Diagnostics) {

	// First, validate parameters.
	// Normalize arch
	arch = strings.ToLower(arch)
	if arch == "x86_64" {
		arch = "amd64"
	}

	// Get full list
	list, err := getNativeHostImages(c, tenantID)
	if err != nil {
		return nil, err
	}

	var matches []duplosdk.DuploNativeHostImage

	for _, v := range *list {
		if isKube && v.K8sVersion == "" {
			continue // ignore images without k8sVersion for kube case
		}

		// Match each filter only if provided
		if (k8Ver == "" || strings.EqualFold(v.K8sVersion, k8Ver)) &&
			(arch == "" || strings.EqualFold(v.Arch, arch)) &&
			(os == "" || strings.EqualFold(v.OS, os)) &&
			(name == "" || strings.EqualFold(v.Name, name)) {

			matches = append(matches, v)
		}
	}

	switch len(matches) {
	case 0:
		return nil, diag.Errorf("failed to retrieve the native host image for '%s': no matching image found", tenantID)

	case 1:
		return &matches[0], nil

	default:
		return nil, diag.Errorf(
			"failed to retrieve a valid native host image for '%s' due to multiple matches (%d): add more filters to narrow the result",
			tenantID,
			len(matches),
		)
	}

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
			"tags":          keyValueToState("images[].tags", image.Tags),
		})
	}

	return result
}
