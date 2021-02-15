package duplosdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"encoding/json"
	"log"
)

// DuploValue is a generic value holder
type DuploValue struct {
	Value string `json:"Value,omitempty"`
}

// DuploName is a generic name holder
type DuploName struct {
	Name string `json:"Name,omitempty"`
}

// DuploKeyValue is a generic key value holder
type DuploKeyValue struct {
	Key   string `json:"Key"`
	Value string `json:"Value,omitempty"`
}

// DuploNameValue is a generic name value holder
type DuploNameValue struct {
	Name  string `json:"Name"`
	Value string `json:"Value,omitempty"`
}

func duploKeyValueFromState(fieldName string, d *schema.ResourceData) *[]DuploKeyValue {
	var ary []DuploKeyValue

	kvs := d.Get(fieldName).([]interface{})
	if len(kvs) > 0 {
		log.Printf("[TRACE] duploKeyValueFromState ********: found %s", fieldName)
		ary = make([]DuploKeyValue, 0, len(kvs))
		for _, raw := range kvs {
			kv := raw.(map[string]interface{})
			ary = append(ary, DuploKeyValue{
				Key:   kv["key"].(string),
				Value: kv["value"].(string),
			})
		}
	}

	return &ary
}

func duploKeyValueToState(fieldName string, duploObjects *[]DuploKeyValue) []interface{} {
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
