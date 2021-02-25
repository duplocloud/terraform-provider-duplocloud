package duplocloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

// ValidateJSONString performs validation of a string that is supposed to be JSON.
func ValidateJSONString(v interface{}, k string) (ws []string, errors []error) {
	// IAM Policy documents need to be valid JSON, and pass legacy parsing
	value := v.(string)
	if len(value) < 1 {
		errors = append(errors, fmt.Errorf("%q contains an invalid JSON policy", k))
		return
	}
	if value[:1] != "{" {
		errors = append(errors, fmt.Errorf("%q contains an invalid JSON policy", k))
		return
	}
	if _, err := structure.NormalizeJsonString(v); err != nil {
		errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %s", k, err))
	}
	return
}

// KeyValueSchema returns a Terraform schema to represent a key value pair
func KeyValueSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

// FilterSchema returns a Terraform schema to represent a filter
func FilterSchema() *schema.Schema {
	return &schema.Schema{
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
	}
}

// FiltersSchema returns a Terraform schema to represent a set of filters
func FiltersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"values": {
					Type:     schema.TypeString,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func suppressMissingOptionalConfigurationBlock(k, old, new string, d *schema.ResourceData) bool {
	return old == "1" && new == "0"
}

// suppresses a diff when not (re)creating a resource.
func diffSuppressWhenNotCreating(k, old, new string, d *schema.ResourceData) bool {
	return d.Id() != ""
}

// suppresses a diff when a resource is brand new
func diffSuppressWhenNew(k, old, new string, d *schema.ResourceData) bool {
	return d.IsNewResource()
}

// suppresses a diff when a resource exists
func diffSuppressWhenExisting(k, old, new string, d *schema.ResourceData) bool {
	return !d.IsNewResource()
}

// suppresses a diff at all times
func diffSuppressFuncIgnore(k, old, new string, d *schema.ResourceData) bool {
	return true
}

//func diffIgnoreIfAlreadySet(k, old, new string, d *schema.ResourceData) bool {
//
//	if new !="" || old !="" {
//		return true
//	}
//
//	return false
//}
func diffIgnoreIfAlreadySet(k, old, new string, d *schema.ResourceData) bool {

	if old != "" {
		return true
	}

	return false
}

// Utility function to return a pointer to a single valid (but optional) resource data block.
func getOptionalBlock(data *schema.ResourceData, key string) (*interface{}, error) {
	var value *interface{}

	if v, ok := data.GetOk(key); ok {
		x := v.([]interface{})

		if len(x) == 1 {
			if x[0] == nil {
				return nil, fmt.Errorf("At least one field is expected inside %s", key)
			}
			value = &x[0]
		}
	}

	return value, nil
}

// Utility function to return a pointer to a single valid (but optional) resource data block as a map.
func getOptionalBlockAsMap(data *schema.ResourceData, key string) (map[string]interface{}, error) {
	block, err := getOptionalBlock(data, key)
	if block == nil || err != nil {
		return nil, err
	}
	return (*block).(map[string]interface{}), nil
}
