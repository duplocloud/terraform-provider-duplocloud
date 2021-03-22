package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

func getAsStringArray(data *schema.ResourceData, key string) (*[]string, bool) {
	var ok bool
	var result []string
	var v interface{}

	if v, ok = data.GetOk(key); ok && v != nil {
		list := v.([]interface{})
		result = make([]string, len(list), len(list))
		for i, el := range list {
			result[i] = el.(string)
		}
	}

	return &result, ok
}

func waitForResourceToBeMissingAfterDelete(ctx context.Context, d *schema.ResourceData, kind string, id string, get func() (interface{}, error)) diag.Diagnostics {
	err := resource.RetryContext(ctx, d.Timeout("delete"), func() *resource.RetryError {
		resp, errget := get()

		if errget != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting %s '%s': %s", kind, id, errget))
		}

		if !isInterfaceNil(resp) {
			return resource.RetryableError(fmt.Errorf("Expected %s '%s' to be missing, but it still exists", kind, id))
		}

		return nil
	})
	if err != nil {
		return diag.Errorf("Error deleting %s '%s': %s", kind, id, err)
	}
	return nil
}

func waitForResourceToBePresentAfterCreate(ctx context.Context, d *schema.ResourceData, kind string, id string, get func() (interface{}, error)) diag.Diagnostics {
	err := resource.RetryContext(ctx, d.Timeout("create"), func() *resource.RetryError {
		resp, errget := get()

		if errget != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting %s '%s': %s", kind, id, errget))
		}

		if isInterfaceNil(resp) {
			return resource.RetryableError(fmt.Errorf("Expected %s '%s' to be retrieved, but got: nil", kind, id))
		}

		return nil
	})
	if err != nil {
		return diag.Errorf("Error creating %s '%s': %s", kind, id, err)
	}
	return nil
}

func isInterfaceNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

func isInterfaceEmptySlice(v interface{}) bool {
	slice := reflect.ValueOf(v)

	return slice.Kind() == reflect.Slice && slice.IsValid() && !slice.IsNil() && slice.Len() == 0
}

// Internal function to check if a given encoded JSON value represents a valid JSON object array.
func validateJsonObjectArray(key string, value string) (ws []string, errors []error) {
	result := []map[string]interface{}{}
	err := json.Unmarshal([]byte(value), &result)
	if err != nil {
		errors = append(errors, fmt.Errorf("%s is invalid: %s", key, err))
	}
	return
}
