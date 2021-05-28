package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"unicode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

// Utility function to convert the `from` interface to a JSON encoded string `field` in the `to` map.
func toJsonStringField(field string, from interface{}, to map[string]interface{}) {
	if json, err := json.Marshal(from); err == nil {
		to[field] = string(json)
	} else {
		log.Printf("[DEBUG] mapToJsonStringField: failed to serialize %s to JSON: %s", field, err)
	}
}

// Utility function to convert the `from` interface to a JSON encoded string `field`.
func toJsonStringState(field string, from interface{}, to *schema.ResourceData) {
	var err error
	var encoded []byte

	if encoded, err = json.Marshal(from); err == nil {
		err = to.Set(field, string(encoded))
	}

	if err != nil {
		log.Printf("[DEBUG] toJsonStringState: failed to serialize %s to JSON: %s", field, err)
	}
}

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

func tagsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
}

//nolint:deadcode,unused // utility function
func tagsSchemaForceNew() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		ForceNew: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
}

func tagsSchemaComputed() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
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

func keyValueToState(fieldName string, duploObjects *[]duplosdk.DuploKeyStringValue) []interface{} {
	if duploObjects != nil {
		input, _ := json.Marshal(&duploObjects)
		log.Printf("[TRACE] duplokeyValueToState[%s] ******** INPUT <= %s", fieldName, input)

		output := make([]interface{}, len(*duploObjects))
		for i, duploObject := range *duploObjects {
			jo := make(map[string]interface{})
			jo["key"] = duploObject.Key
			jo["value"] = duploObject.Value
			output[i] = jo
		}
		dump, _ := json.Marshal(output)
		log.Printf("[TRACE] duplokeyValueToState[%s] ******** OUTPUT => %s", fieldName, dump)
		return output
	}

	log.Printf("[TRACE] duplokeyValueToState[%s] ******** EMPTY INPUT", fieldName)
	return make([]interface{}, 0)
}

func keyValueFromState(fieldName string, d *schema.ResourceData) *[]duplosdk.DuploKeyStringValue {
	var ary []duplosdk.DuploKeyStringValue

	if v, ok := d.GetOk(fieldName); ok && v != nil && len(v.([]interface{})) > 0 {
		kvs := v.([]interface{})
		log.Printf("[TRACE] duploKeyValueFromState ********: found %s", fieldName)
		ary = make([]duplosdk.DuploKeyStringValue, 0, len(kvs))
		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, duplosdk.DuploKeyStringValue{
				Key:   kv["key"].(string),
				Value: kv["value"].(string),
			})
		}
	}

	return &ary
}

func keyValueToMap(list *[]duplosdk.DuploKeyStringValue) map[string]interface{} {
	result := map[string]interface{}{}
	if list != nil {
		for _, item := range *list {
			result[item.Key] = item.Value
		}
	}
	return result
}

func keyValueFromMap(d map[string]interface{}) *[]duplosdk.DuploKeyStringValue {
	list := make([]duplosdk.DuploKeyStringValue, 0, len(d))

	for k, v := range d {
		list = append(list, duplosdk.DuploKeyStringValue{
			Key:   k,
			Value: v.(string),
		})
	}

	return &list
}

func keyValueFromStateList(fieldName string, d map[string]interface{}) *[]duplosdk.DuploKeyStringValue {
	var ary []duplosdk.DuploKeyStringValue

	if v, ok := d[fieldName]; ok && v != nil && len(v.([]interface{})) > 0 {
		kvs := v.([]interface{})
		log.Printf("[TRACE] duploKeyValueFromMap ********: found %s", fieldName)
		ary = make([]duplosdk.DuploKeyStringValue, 0, len(kvs))
		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, duplosdk.DuploKeyStringValue{
				Key:   kv["key"].(string),
				Value: kv["value"].(string),
			})
		}
	}

	return &ary
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

// Utility function to return a pointer to a single valid (but optional) resource data block.
func getOptionalBlock(data *schema.ResourceData, key string) (*interface{}, error) {
	var value *interface{}

	if v, ok := data.GetOk(key); ok {
		x := v.([]interface{})

		if len(x) == 1 {
			if x[0] == nil {
				return nil, fmt.Errorf("at least one field is expected inside %s", key)
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
		result = make([]string, len(list))
		for i, el := range list {
			result[i] = el.(string)
		}
	}

	return &result, ok
}

// Utiliy function to return a filtered list of tenant metadata, given the selected keys.
func selectKeyValuesFromMap(metadata *[]duplosdk.DuploKeyStringValue, keys map[string]interface{}) *[]duplosdk.DuploKeyStringValue {
	settings := make([]duplosdk.DuploKeyStringValue, 0, len(keys))
	for _, kv := range *metadata {
		if _, ok := keys[kv.Key]; ok {
			settings = append(settings, kv)
		}
	}

	return &settings
}

// Utiliy function to return a filtered list of tenant metadata, given the selected keys.
func selectKeyValues(metadata *[]duplosdk.DuploKeyStringValue, keys []string) *[]duplosdk.DuploKeyStringValue {
	specified := map[string]interface{}{}
	for _, k := range keys {
		specified[k] = struct{}{}
	}

	return selectKeyValuesFromMap(metadata, specified)
}

// Internal function used to re-order key value pairs
//nolint:deadcode,unused // utility function
func reorderKeyValues(pairs []interface{}) {

	// Re-order environment variables to a canonical order.
	sort.Slice(pairs, func(i, j int) bool {
		mi := pairs[i].(map[string]interface{})
		mj := pairs[j].(map[string]interface{})

		// Get both name keys, fall back on an empty string.
		si := ""
		sj := ""
		if v, ok := mi["key"]; ok && !isInterfaceNil(v) {
			si = v.(string)
		}
		if v, ok := mj["key"]; ok && !isInterfaceNil(v) {
			sj = v.(string)
		}

		// Compare the two.
		return si < sj
	})
}

func getStringArray(data map[string]interface{}, key string) (*[]string, bool) {
	var ok bool
	var result []string
	var v interface{}

	if v, ok = data[key]; ok && v != nil {
		list := v.([]interface{})
		result = make([]string, len(list))
		for i, el := range list {
			result[i] = el.(string)
		}
	}

	return &result, ok
}

func waitForResourceToBeMissingAfterDelete(ctx context.Context, d *schema.ResourceData, kind string, id string, get func() (interface{}, duplosdk.ClientError)) diag.Diagnostics {
	err := resource.RetryContext(ctx, d.Timeout("delete"), func() *resource.RetryError {
		resp, errget := get()

		if errget != nil {
			if errget.Status() == 404 {
				return nil
			}

			return resource.NonRetryableError(fmt.Errorf("error getting %s '%s': %s", kind, id, errget))
		}

		if !isInterfaceNil(resp) {
			return resource.RetryableError(fmt.Errorf("expected %s '%s' to be missing, but it still exists", kind, id))
		}

		return nil
	})
	if err != nil {
		return diag.Errorf("error deleting %s '%s': %s", kind, id, err)
	}
	return nil
}

func waitForResourceToBePresentAfterCreate(ctx context.Context, d *schema.ResourceData, kind string, id string, get func() (interface{}, duplosdk.ClientError)) diag.Diagnostics {
	err := resource.RetryContext(ctx, d.Timeout("create"), func() *resource.RetryError {
		resp, errget := get()

		if errget != nil {
			if errget.Status() == 404 {
				return resource.RetryableError(fmt.Errorf("expected %s '%s' to be retrieved, but got a 404", kind, id))
			}

			return resource.NonRetryableError(fmt.Errorf("error getting %s '%s': %s", kind, id, errget))
		}

		if isInterfaceNil(resp) {
			return resource.RetryableError(fmt.Errorf("expected %s '%s' to be retrieved, but got: nil", kind, id))
		}

		return nil
	})
	if err != nil {
		return diag.Errorf("error creating %s '%s': %s", kind, id, err)
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

// Internal function to convert map keys from lower camel-case to upper camel-case.
//  - Adds an upper camel-case entry for each lower camel-case entry, unless the upper exists already.
//  - Removes any lower camel-case entry.
//  - Never overwrites any existing upper camel-case keys.
func makeMapUpperCamelCase(m map[string]interface{}) {
	for k := range m {

		// Only convert lowercase entries.
		if unicode.IsLower([]rune(k)[0]) {
			upper := strings.Title(k)

			// Add the upper camel-case entry, if it doesn't exist.
			if _, ok := m[upper]; !ok {
				m[upper] = m[k]
			}

			// Remove the lower camel-case entry.
			delete(m, k)
		}
	}
}

func waitForResourceWithStatusDone(ctx context.Context, d *schema.ResourceData, kind string, id string, get func() (bool, duplosdk.ClientError)) diag.Diagnostics {
	err := resource.RetryContext(ctx, d.Timeout(kind), func() *resource.RetryError {
		status, errget := get()

		if errget != nil {
			if errget.Status() == 404 {
				return nil
			}

			return resource.NonRetryableError(fmt.Errorf("error getting %s '%s': %s", kind, id, errget))
		}
		// return nil if want to complete wait
		if !status {
			return resource.RetryableError(fmt.Errorf("expected %s '%s' to be missing, but it still exists", kind, id))
		}

		return nil
	})
	if err != nil {
		return diag.Errorf("error deleting %s '%s': %s", kind, id, err)
	}
	return nil
}
