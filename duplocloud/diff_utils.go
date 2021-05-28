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
//nolint:deadcode,unused // utility function
func diffSuppressWhenNew(k, old, new string, d *schema.ResourceData) bool {
	return d.IsNewResource()
}

// suppresses a diff when a resource exists
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
