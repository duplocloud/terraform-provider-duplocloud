package duplocloud

import "k8s.io/api/core/v1"

func flattenHostaliases(in []v1.HostAlias) []interface{} {
	att := make([]interface{}, len(in))
	for i, v := range in {
		ha := make(map[string]interface{})
		ha["ip"] = v.IP
		if len(v.Hostnames) > 0 {
			ha["hostnames"] = v.Hostnames
		}
		att[i] = ha
	}
	return att
}
