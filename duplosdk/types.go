package duplosdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"encoding/json"
	"log"
)

// DuploEnabled is a generic flag holder
type DuploEnabled struct {
	Enabled bool `json:"Enabled,omitempty"`
}

// DuploStringValue is a generic value holder
type DuploStringValue struct {
	Value string `json:"Value,omitempty"`
}

// DuploName is a generic name holder
type DuploName struct {
	Name string `json:"Name,omitempty"`
}

// DuploKeyStringValue is a generic key value holder
type DuploKeyStringValue struct {
	Key   string `json:"Key"`
	Value string `json:"Value,omitempty"`
}

// DuploNameStringValue is a generic name value holder
type DuploNameStringValue struct {
	Name  string `json:"Name"`
	Value string `json:"Value,omitempty"`
}

func duploKeyValueSchema() *schema.Resource {
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

func duploKeyValueFromState(fieldName string, d *schema.ResourceData) *[]DuploKeyStringValue {
	var ary []DuploKeyStringValue

	kvs := d.Get(fieldName).([]interface{})
	if len(kvs) > 0 {
		log.Printf("[TRACE] duploKeyValueFromState ********: found %s", fieldName)
		ary = make([]DuploKeyStringValue, 0, len(kvs))
		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, DuploKeyStringValue{
				Key:   kv["key"].(string),
				Value: kv["value"].(string),
			})
		}
	}

	return &ary
}

// KeyValueToState converts a DuploKeyValue array into terraform resource data.
func KeyValueToState(fieldName string, duploObjects *[]DuploKeyStringValue) []interface{} {
	if duploObjects != nil {
		input, _ := json.Marshal(&duploObjects)
		log.Printf("[TRACE] duploKeyValueToState[%s] ******** INPUT <= %s", fieldName, input)

		output := make([]interface{}, len(*duploObjects), len(*duploObjects))
		for i, duploObject := range *duploObjects {
			jo := make(map[string]interface{})
			jo["key"] = duploObject.Key
			jo["value"] = duploObject.Value
			output[i] = jo
		}
		dump, _ := json.Marshal(output)
		log.Printf("[TRACE] duploKeyValueToState[%s] ******** OUTPUT => %s", fieldName, dump)
		return output
	}

	log.Printf("[TRACE] duploKeyValueToState[%s] ******** EMPTY INPUT", fieldName)
	return make([]interface{}, 0)
}
