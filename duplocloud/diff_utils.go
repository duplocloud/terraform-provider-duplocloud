package duplocloud

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//nolint:deadcode,unused // utility function
func suppressEquivalentTypeStringBoolean(k, old, new string, d *schema.ResourceData) bool {
	if old == "false" && new == "0" {
		return true
	}
	if old == "true" && new == "1" {
		return true
	}
	return false
}

//nolint:deadcode,unused // utility function
func suppressEquivalentJSONDiffs(k, old, new string, d *schema.ResourceData) bool {
	ob := bytes.NewBufferString("")
	if err := json.Compact(ob, []byte(old)); err != nil {
		return false
	}

	nb := bytes.NewBufferString("")
	if err := json.Compact(nb, []byte(new)); err != nil {
		return false
	}

	return jsonBytesEqual(ob.Bytes(), nb.Bytes())
}

//nolint:deadcode,unused // utility function
func base64Encode(data []byte) string {
	if isBase64Encoded(data) {
		return string(data)
	}
	return base64.StdEncoding.EncodeToString(data)
}

//nolint:deadcode,unused // utility function
func isBase64Encoded(data []byte) bool {
	_, err := base64.StdEncoding.DecodeString(string(data))
	return err == nil
}

//nolint:deadcode,unused // utility function
func looksLikeJSONString(s interface{}) bool {
	return regexp.MustCompile(`^\s*{`).MatchString(s.(string))
}

//nolint:deadcode,unused // utility function
func jsonBytesEqual(b1, b2 []byte) bool {
	var o1 interface{}
	if err := json.Unmarshal(b1, &o1); err != nil {
		return false
	}

	var o2 interface{}
	if err := json.Unmarshal(b2, &o2); err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}

//nolint:deadcode,unused // utility function
func suppressMissingOptionalConfigurationBlock(k, old, new string, d *schema.ResourceData) bool {
	return old == "1" && new == "0"
}

// suppresses a diff when not (re)creating a resource.
func diffSuppressWhenNotCreating(k, old, new string, d *schema.ResourceData) bool {
	return d.Id() != ""
}

// suppresses a diff when a resource is brand new
//
//nolint:deadcode,unused // utility function
func diffSuppressWhenNew(k, old, new string, d *schema.ResourceData) bool {
	return d.IsNewResource()
}

// suppresses a diff when a resource exists
//
//nolint:deadcode,unused // utility function
func diffSuppressWhenExisting(k, old, new string, d *schema.ResourceData) bool {
	return !d.IsNewResource()
}

// suppresses a diff at all times
func diffSuppressFuncIgnore(k, old, new string, d *schema.ResourceData) bool {
	return true
}

//func diffIgnoreIfAlreadySet(k, old, new string, d *schema.ResourceData) bool {
//	return new !="" || old !=""
//}

func diffIgnoreIfAlreadySet(k, old, new string, d *schema.ResourceData) bool {
	return old != ""
}

//nolint:deadcode,unused // utility function
func diffIgnoreIfSameHash(k, old, new string, d *schema.ResourceData) bool {
	if old == "" {
		return false
	}
	newHash := hashForData(new)
	return old == newHash
}

func hashForData(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	var apiStr = fmt.Sprintf("%d==", h.Sum32())
	return apiStr
}

//nolint:deadcode,unused // utility function
func stringHash(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

//nolint:deadcode,unused // utility function
func ptrStringValue(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

// diffStringMaps returns the set of keys and values that must be created,
// and the set of keys and values that must be destroyed.
// Equivalent to 'diffTagsGeneric'.
//
//nolint:deadcode,unused // utility function
func diffStringMaps(oldMap, newMap map[string]interface{}) (map[string]*string, map[string]*string) {
	// First, we're creating everything we have
	create := map[string]*string{}
	for k, v := range newMap {
		str := v.(string)
		create[k] = &str
	}

	// Build the map of what to remove
	remove := map[string]*string{}
	for k, v := range oldMap {
		old, ok := create[k]
		if !ok || ptrStringValue(old) != v.(string) {
			// Delete it!
			str := v.(string)
			remove[k] = &str
		} else if ok {
			// already present so remove from new
			delete(create, k)
		}
	}

	return create, remove
}

func suppressAzureManagedTags(k, old, new string, d *schema.ResourceData) bool {
	_, tags := d.GetChange("tags")
	suppressTagKeys := DuploManagedAzureTags()

	for _, mp := range tags.([]interface{}) {
		keyval := mp.(map[string]interface{})
		for key := range keyval {
			if Contains(suppressTagKeys, key) {
				return true
			}

		}
	}

	return false
}

func diffIgnoreIfCaseSensitive(k, old, new string, d *schema.ResourceData) bool {
	newVar, oldVar := d.GetChange(k)
	return strings.EqualFold(newVar.(string), oldVar.(string))
}

func diffSuppressOnComputedDataOnMetadataBlock(_, _, _ string, d *schema.ResourceData) bool {

	n, o := d.GetChange("metadata")
	m := n.(map[string]interface{})
	mo := o.(map[string]interface{})
	_, ok := m["CreatedBy"]
	_, ok1 := m["CreatedOn"]
	for k, vl := range m {
		if mov, ok := mo[k]; ok || (mov != nil && mov.(string) != vl) {
			return false
		}
	}
	if ok && ok1 {
		return true
	}
	return false
}

func diffSuppressOnComputedDataOnLabelBlock(_, _, _ string, d *schema.ResourceData) bool {

	n, _ := d.GetChange("labels")
	m := n.(map[string]interface{})
	_, ok := m["image-id"]
	return ok
}

func diffSuppressGCPHostImageIdIfSame(_, _, _ string, d *schema.ResourceData) bool {

	n, o := d.GetChange("image_id")
	ok := false
	if n != "" {
		ok = strings.Contains(o.(string), n.(string))
	}
	return ok
}

func diffSuppressListKeyValueOrdering(k, old, new string, d *schema.ResourceData) bool {
	// Extract attribute name to get the old and new values
	attributeName := strings.Split(k, ".")[0]
	o, n := d.GetChange(attributeName)

	// Type assertion to ensure both old and new values are lists
	oList, ok1 := o.([]interface{})
	nList, ok2 := n.([]interface{})
	if !ok1 || !ok2 {
		return false // If not lists, don't suppress
	}

	// Helper function to sort and convert list to a map of key-value pairs
	toSortedKeyValueList := func(list []interface{}) []map[string]string {
		result := make([]map[string]string, 0, len(list))
		for _, item := range list {
			if entry, ok := item.(map[string]interface{}); ok {
				entryMap := make(map[string]string)
				if key, keyOk := entry["key"].(string); keyOk {
					if value, valueOk := entry["value"].(string); valueOk {
						entryMap["key"] = key
						entryMap["value"] = value
						result = append(result, entryMap)
					}
				}
			}
		}
		// Sort list by key for consistency
		sort.Slice(result, func(i, j int) bool {
			return result[i]["key"] < result[j]["key"]
		})
		return result
	}

	// Sort and transform both old and new lists
	oldSortedList := toSortedKeyValueList(oList)
	newSortedList := toSortedKeyValueList(nList)

	// Compare the lists for equality
	if len(oldSortedList) != len(newSortedList) {
		return false // Length mismatch indicates a change
	}
	for i := range oldSortedList {
		oldEntry := oldSortedList[i]
		newEntry := newSortedList[i]
		// Compare key-value pairs; suppress if identical
		if oldEntry["key"] != newEntry["key"] || oldEntry["value"] != newEntry["value"] {
			return false
		}
	}

	// No meaningful difference, suppress the diff
	return true
}

func diffSuppressDynamodbTTLHandler(k, old, new string, d *schema.ResourceData) bool {
	ostate, nstate := d.GetChange("ttl")
	if len(ostate.([]interface{})) == 0 && len(nstate.([]interface{})) > 0 { //if ttl already disable ignnoring change
		l := nstate.([]interface{})
		mp := l[0].(map[string]interface{})
		if !mp["enabled"].(bool) {
			return true
		}
	} else if len(ostate.([]interface{})) > 0 && len(nstate.([]interface{})) == 0 {
		l := ostate.([]interface{})

		mp := l[0].(map[string]interface{})
		if !mp["enabled"].(bool) {
			return true
		}
	}
	return false
}

/*
func diffSuppressListOrderingAsWhole(k, old, new string, d *schema.ResourceData) bool {
	// Extract attribute name to get the old and new values
	attributeName := strings.Split(k, ".")[0]
	o, n := d.GetChange(attributeName)

	// Type assertion to ensure both old and new values are lists
	oList, ok1 := o.([]interface{})
	nList, ok2 := n.([]interface{})
	if !ok1 || !ok2 {
		return false // If not lists, don't suppress
	}

	// Helper function to sort and convert list to a map of key-value pairs
	toSortedList := func(list []interface{}, fields []string) []map[string]string {
		result := make([]map[string]string, 0, len(list))
		for _, item := range list {
			if entry, ok := item.(map[string]interface{}); ok {
				sortedEntry := make(map[string]string)
				for _, field := range fields {
					if value, valueOk := entry[field].(string); valueOk {
						sortedEntry[field] = value
					}
				}
				result = append(result, sortedEntry)
			}
		}
		// Sort list by concatenated field values for consistency
		sort.Slice(result, func(i, j int) bool {
			for _, field := range fields {
				if result[i][field] != result[j][field] {
					return result[i][field] < result[j][field]
				}
			}
			return false
		})
		return result
	}

	// Dynamically determine the fields to compare based on schema
	schemaMap := d.Get(attributeName).([]interface{})
	fields := []string{}
	if len(schemaMap) > 0 {
		if firstItem, ok := schemaMap[0].(map[string]interface{}); ok {
			for key := range firstItem {
				fields = append(fields, key)
			}
		}
	}

	// Sort and transform both old and new lists
	oldSortedList := toSortedList(oList, fields)
	newSortedList := toSortedList(nList, fields)

	// Compare the lists for equality
	if len(oldSortedList) != len(newSortedList) {
		return false // Length mismatch indicates a change
	}
	for i := range oldSortedList {
		for _, field := range fields {
			if oldSortedList[i][field] != newSortedList[i][field] {
				return false // Field mismatch indicates a change
			}
		}
	}

	// No meaningful difference, suppress the diff
	return true
}*/

func diffSuppressListOrderingOnNestedField(k, old, new string, d *schema.ResourceData) bool {
	// Extract attribute name to get the old and new values
	token := strings.Split(k, ".")
	attributeName := token[0]
	child := token[2]
	o, _ := d.GetChange(attributeName)
	for _, om := range o.([]interface{}) {
		m := om.(map[string]interface{})
		//for _, m := range mi {
		if new == m[child] {
			return true
		}
		//}
	}

	// No meaningful difference, suppress the diff
	return false
}
