package duplocloud

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/#taint-based-evictions
var builtInTolerations = map[string]string{
	v1.TaintNodeNotReady:           "",
	v1.TaintNodeUnreachable:        "",
	v1.TaintNodeUnschedulable:      "",
	v1.TaintNodeMemoryPressure:     "",
	v1.TaintNodeDiskPressure:       "",
	v1.TaintNodeNetworkUnavailable: "",
	v1.TaintNodePIDPressure:        "",
}

// Flatteners

func flattenPodSpec(in v1.PodSpec) ([]interface{}, error) {
	att := make(map[string]interface{})
	if in.ActiveDeadlineSeconds != nil {
		att["active_deadline_seconds"] = *in.ActiveDeadlineSeconds
	}

	if in.Affinity != nil {
		att["affinity"] = flattenAffinity(in.Affinity)
	}

	if in.AutomountServiceAccountToken != nil {
		att["automount_service_account_token"] = *in.AutomountServiceAccountToken
	}

	// To avoid perpetual diff, remove the service account token volume from PodSpec.
	serviceAccountName := "default"
	if in.ServiceAccountName != "" {
		serviceAccountName = in.ServiceAccountName
	}
	serviceAccountRegex := fmt.Sprintf("%s-token-([a-z0-9]{5})", serviceAccountName)

	containers, err := flattenContainers(in.Containers, serviceAccountRegex)
	if err != nil {
		return nil, err
	}
	att["container"] = containers

	gates, err := flattenReadinessGates(in.ReadinessGates)
	if err != nil {
		return nil, err
	}
	att["readiness_gate"] = gates

	initContainers, err := flattenContainers(in.InitContainers, serviceAccountRegex)
	if err != nil {
		return nil, err
	}
	att["init_container"] = initContainers

	att["dns_policy"] = in.DNSPolicy
	if in.DNSConfig != nil {
		v, err := flattenPodDNSConfig(in.DNSConfig)
		if err != nil {
			return []interface{}{att}, err
		}
		att["dns_config"] = v
	}

	if in.EnableServiceLinks != nil {
		att["enable_service_links"] = *in.EnableServiceLinks
	}

	att["host_aliases"] = flattenHostaliases(in.HostAliases)

	att["host_ipc"] = in.HostIPC
	att["host_network"] = in.HostNetwork
	att["host_pid"] = in.HostPID

	if in.Hostname != "" {
		att["hostname"] = in.Hostname
	}
	att["image_pull_secrets"] = flattenLocalObjectReferenceArray(in.ImagePullSecrets)

	if in.NodeName != "" {
		att["node_name"] = in.NodeName
	}
	if len(in.NodeSelector) > 0 {
		att["node_selector"] = in.NodeSelector
	}
	if in.RuntimeClassName != nil {
		att["runtime_class_name"] = *in.RuntimeClassName
	}
	if in.PriorityClassName != "" {
		att["priority_class_name"] = in.PriorityClassName
	}
	if in.RestartPolicy != "" {
		att["restart_policy"] = in.RestartPolicy
	}

	if in.SecurityContext != nil {
		att["security_context"] = flattenPodSecurityContext(in.SecurityContext)
	}

	if in.SchedulerName != "" {
		att["scheduler_name"] = in.SchedulerName
	}

	if in.ServiceAccountName != "" {
		att["service_account_name"] = in.ServiceAccountName
	}
	if in.ShareProcessNamespace != nil {
		att["share_process_namespace"] = *in.ShareProcessNamespace
	}

	if in.Subdomain != "" {
		att["subdomain"] = in.Subdomain
	}

	if in.TerminationGracePeriodSeconds != nil {
		att["termination_grace_period_seconds"] = *in.TerminationGracePeriodSeconds
	}

	if len(in.Tolerations) > 0 {
		att["toleration"] = flattenTolerations(in.Tolerations)
	}

	if len(in.TopologySpreadConstraints) > 0 {
		att["topology_spread_constraint"] = flattenTopologySpreadConstraints(in.TopologySpreadConstraints)
	}

	if len(in.Volumes) > 0 {
		for i, volume := range in.Volumes {
			// To avoid perpetual diff, remove the service account token volume from PodSpec.
			nameMatchesDefaultToken, err := regexp.MatchString(serviceAccountRegex, volume.Name)
			if err != nil {
				return []interface{}{att}, err
			}
			if nameMatchesDefaultToken || strings.HasPrefix(volume.Name, "kube-api-access") {
				in.Volumes = removeVolumeFromPodSpec(i, in.Volumes)
				break
			}
		}

		v, err := flattenVolumes(in.Volumes)
		if err != nil {
			return []interface{}{att}, err
		}
		att["volumes"] = v
		att["volume"] = v
	}
	return []interface{}{att}, nil
}

// removeVolumeFromPodSpec removes the specified Volume index (i) from the given list of Volumes.
func removeVolumeFromPodSpec(i int, v []v1.Volume) []v1.Volume {
	return append(v[:i], v[i+1:]...)
}

func flattenPodDNSConfig(in *v1.PodDNSConfig) ([]interface{}, error) {
	att := make(map[string]interface{})

	if len(in.Nameservers) > 0 {
		att["nameservers"] = in.Nameservers
	}
	if len(in.Searches) > 0 {
		att["searches"] = in.Searches
	}
	if len(in.Options) > 0 {
		v, err := flattenPodDNSConfigOptions(in.Options)
		if err != nil {
			return []interface{}{att}, err
		}
		att["option"] = v
	}

	if len(att) > 0 {
		return []interface{}{att}, nil
	}
	return []interface{}{}, nil
}

func flattenPodDNSConfigOptions(options []v1.PodDNSConfigOption) ([]interface{}, error) {
	att := make([]interface{}, len(options))
	for i, v := range options {
		obj := map[string]interface{}{}

		if v.Name != "" {
			obj["name"] = v.Name
		}
		if v.Value != nil {
			obj["value"] = *v.Value
		}
		att[i] = obj
	}
	return att, nil
}

func flattenPodSecurityContext(in *v1.PodSecurityContext) []interface{} {
	att := make(map[string]interface{})

	if in.FSGroup != nil {
		att["fs_group"] = strconv.Itoa(int(*in.FSGroup))
	}
	if in.RunAsGroup != nil {
		att["run_as_group"] = strconv.Itoa(int(*in.RunAsGroup))
	}
	if in.RunAsNonRoot != nil {
		att["run_as_non_root"] = *in.RunAsNonRoot
	}
	if in.RunAsUser != nil {
		att["run_as_user"] = strconv.Itoa(int(*in.RunAsUser))
	}
	if in.SeccompProfile != nil {
		att["seccomp_profile"] = flattenSeccompProfile(in.SeccompProfile)
	}
	if in.FSGroupChangePolicy != nil {
		att["fs_group_change_policy"] = *in.FSGroupChangePolicy
	}
	if len(in.SupplementalGroups) > 0 {
		att["supplemental_groups"] = newInt64Set(schema.HashSchema(&schema.Schema{
			Type: schema.TypeInt,
		}), in.SupplementalGroups)
	}
	if in.SELinuxOptions != nil {
		att["se_linux_options"] = flattenSeLinuxOptions(in.SELinuxOptions)
	}
	if in.Sysctls != nil {
		att["sysctl"] = flattenSysctls(in.Sysctls)
	}

	if len(att) > 0 {
		return []interface{}{att}
	}
	return []interface{}{}
}

func flattenSeccompProfile(in *v1.SeccompProfile) []interface{} {
	att := make(map[string]interface{})
	if in.Type != "" {
		att["type"] = in.Type
		if in.Type == "Localhost" {
			att["localhost_profile"] = in.LocalhostProfile
		}
	}
	return []interface{}{att}
}

func flattenSeLinuxOptions(in *v1.SELinuxOptions) []interface{} {
	att := make(map[string]interface{})
	if in.User != "" {
		att["user"] = in.User
	}
	if in.Role != "" {
		att["role"] = in.Role
	}
	if in.Type != "" {
		att["type"] = in.Type
	}
	if in.Level != "" {
		att["level"] = in.Level
	}
	return []interface{}{att}
}

func flattenSysctls(sysctls []v1.Sysctl) []interface{} {
	att := []interface{}{}
	for _, v := range sysctls {
		obj := map[string]interface{}{}

		if v.Name != "" {
			obj["name"] = v.Name
		}
		if v.Value != "" {
			obj["value"] = v.Value
		}
		att = append(att, obj)
	}
	return att
}

func flattenTolerations(tolerations []v1.Toleration) []interface{} {
	att := []interface{}{}
	for _, v := range tolerations {
		// The API Server may automatically add several Tolerations to pods, strip these to avoid TF diff.
		if _, ok := builtInTolerations[v.Key]; ok {
			log.Printf("[INFO] ignoring toleration with key: %s", v.Key)
			continue
		}
		obj := map[string]interface{}{}

		if v.Effect != "" {
			obj["effect"] = string(v.Effect)
		}
		if v.Key != "" {
			obj["key"] = v.Key
		}
		if v.Operator != "" {
			obj["operator"] = string(v.Operator)
		}
		if v.TolerationSeconds != nil {
			obj["toleration_seconds"] = strconv.FormatInt(*v.TolerationSeconds, 10)
		}
		if v.Value != "" {
			obj["value"] = v.Value
		}
		att = append(att, obj)
	}
	return att
}

func flattenTopologySpreadConstraints(tsc []v1.TopologySpreadConstraint) []interface{} {
	att := []interface{}{}
	for _, v := range tsc {
		obj := map[string]interface{}{}

		if v.TopologyKey != "" {
			obj["topology_key"] = v.TopologyKey
		}
		if v.MaxSkew != 0 {
			obj["max_skew"] = v.MaxSkew
		}
		if v.WhenUnsatisfiable != "" {
			obj["when_unsatisfiable"] = string(v.WhenUnsatisfiable)
		}
		if v.LabelSelector != nil {
			obj["label_selector"] = flattenLabelSelector(v.LabelSelector)
		}
		att = append(att, obj)
	}
	return att
}

func flattenVolumes(volumes []v1.Volume) ([]interface{}, error) {
	att := make([]interface{}, len(volumes))
	for i, v := range volumes {
		obj := map[string]interface{}{}

		if v.Name != "" {
			obj["name"] = v.Name
		}
		if v.ConfigMap != nil {
			obj["config_map"] = flattenConfigMapVolumeSource(v.ConfigMap)
		}
		if v.GitRepo != nil {
			obj["git_repo"] = flattenGitRepoVolumeSource(v.GitRepo)
		}
		if v.EmptyDir != nil {
			obj["empty_dir"] = flattenEmptyDirVolumeSource(v.EmptyDir)
		}
		if v.DownwardAPI != nil {
			obj["downward_api"] = flattenDownwardAPIVolumeSource(v.DownwardAPI)
		}
		if v.PersistentVolumeClaim != nil {
			obj["persistent_volume_claim"] = flattenPersistentVolumeClaimVolumeSource(v.PersistentVolumeClaim)
		}
		if v.Secret != nil {
			obj["secret"] = flattenSecretVolumeSource(v.Secret)
		}
		if v.Projected != nil {
			obj["projected"] = flattenProjectedVolumeSource(v.Projected)
		}
		if v.GCEPersistentDisk != nil {
			obj["gce_persistent_disk"] = flattenGCEPersistentDiskVolumeSource(v.GCEPersistentDisk)
		}
		if v.AWSElasticBlockStore != nil {
			obj["aws_elastic_block_store"] = flattenAWSElasticBlockStoreVolumeSource(v.AWSElasticBlockStore)
		}
		if v.HostPath != nil {
			obj["host_path"] = flattenHostPathVolumeSource(v.HostPath)
		}
		if v.Glusterfs != nil {
			obj["glusterfs"] = flattenGlusterfsVolumeSource(v.Glusterfs)
		}
		if v.NFS != nil {
			obj["nfs"] = flattenNFSVolumeSource(v.NFS)
		}
		if v.RBD != nil {
			obj["rbd"] = flattenRBDVolumeSource(v.RBD)
		}
		if v.ISCSI != nil {
			obj["iscsi"] = flattenISCSIVolumeSource(v.ISCSI)
		}
		if v.Cinder != nil {
			obj["cinder"] = flattenCinderVolumeSource(v.Cinder)
		}
		if v.CephFS != nil {
			obj["ceph_fs"] = flattenCephFSVolumeSource(v.CephFS)
		}
		if v.CSI != nil {
			obj["csi"] = flattenCSIVolumeSource(v.CSI)
		}
		if v.FC != nil {
			obj["fc"] = flattenFCVolumeSource(v.FC)
		}
		if v.Flocker != nil {
			obj["flocker"] = flattenFlockerVolumeSource(v.Flocker)
		}
		if v.FlexVolume != nil {
			obj["flex_volume"] = flattenFlexVolumeSource(v.FlexVolume)
		}
		if v.AzureFile != nil {
			obj["azure_file"] = flattenAzureFileVolumeSource(v.AzureFile)
		}
		if v.VsphereVolume != nil {
			obj["vsphere_volume"] = flattenVsphereVirtualDiskVolumeSource(v.VsphereVolume)
		}
		if v.Quobyte != nil {
			obj["quobyte"] = flattenQuobyteVolumeSource(v.Quobyte)
		}
		if v.AzureDisk != nil {
			obj["azure_disk"] = flattenAzureDiskVolumeSource(v.AzureDisk)
		}
		if v.PhotonPersistentDisk != nil {
			obj["photon_persistent_disk"] = flattenPhotonPersistentDiskVolumeSource(v.PhotonPersistentDisk)
		}
		att[i] = obj
	}
	return att, nil
}

func flattenPersistentVolumeClaimVolumeSource(in *v1.PersistentVolumeClaimVolumeSource) []interface{} {
	att := make(map[string]interface{})
	if in.ClaimName != "" {
		att["claim_name"] = in.ClaimName
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}

	return []interface{}{att}
}
func flattenGitRepoVolumeSource(in *v1.GitRepoVolumeSource) []interface{} {
	att := make(map[string]interface{})
	if in.Directory != "" {
		att["directory"] = in.Directory
	}

	att["repository"] = in.Repository

	if in.Revision != "" {
		att["revision"] = in.Revision
	}
	return []interface{}{att}
}

func flattenDownwardAPIVolumeSource(in *v1.DownwardAPIVolumeSource) []interface{} {
	att := make(map[string]interface{})
	if in.DefaultMode != nil {
		att["default_mode"] = "0" + strconv.FormatInt(int64(*in.DefaultMode), 8)
	}
	if len(in.Items) > 0 {
		att["items"] = flattenDownwardAPIVolumeFile(in.Items)
	}
	return []interface{}{att}
}

func flattenDownwardAPIVolumeFile(in []v1.DownwardAPIVolumeFile) []interface{} {
	att := make([]interface{}, len(in))
	for i, v := range in {
		m := map[string]interface{}{}
		if v.FieldRef != nil {
			m["field_ref"] = flattenObjectFieldSelector(v.FieldRef)
		}
		if v.Mode != nil {
			m["mode"] = "0" + strconv.FormatInt(int64(*v.Mode), 8)
		}
		if v.Path != "" {
			m["path"] = v.Path
		}
		if v.ResourceFieldRef != nil {
			m["resource_field_ref"] = flattenResourceFieldSelector(v.ResourceFieldRef)
		}
		att[i] = m
	}
	return att
}

func flattenConfigMapVolumeSource(in *v1.ConfigMapVolumeSource) []interface{} {
	att := make(map[string]interface{})
	if in.DefaultMode != nil {
		att["default_mode"] = "0" + strconv.FormatInt(int64(*in.DefaultMode), 8)
	}
	att["name"] = in.Name
	if len(in.Items) > 0 {
		items := make([]interface{}, len(in.Items))
		for i, v := range in.Items {
			m := map[string]interface{}{}
			if v.Key != "" {
				m["key"] = v.Key
			}
			if v.Mode != nil {
				m["mode"] = "0" + strconv.FormatInt(int64(*v.Mode), 8)
			}
			if v.Path != "" {
				m["path"] = v.Path
			}
			items[i] = m
		}
		att["items"] = items
	}
	if in.Optional != nil {
		att["optional"] = *in.Optional
	}
	return []interface{}{att}
}

func flattenEmptyDirVolumeSource(in *v1.EmptyDirVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["medium"] = string(in.Medium)
	if in.SizeLimit != nil {
		att["size_limit"] = in.SizeLimit.String()
	}
	return []interface{}{att}
}

func flattenSecretVolumeSource(in *v1.SecretVolumeSource) []interface{} {
	att := make(map[string]interface{})
	if in.DefaultMode != nil {
		att["default_mode"] = "0" + strconv.FormatInt(int64(*in.DefaultMode), 8)
	}
	if in.SecretName != "" {
		att["secret_name"] = in.SecretName
	}
	if len(in.Items) > 0 {
		items := make([]interface{}, len(in.Items))
		for i, v := range in.Items {
			m := map[string]interface{}{}
			m["key"] = v.Key
			if v.Mode != nil {
				m["mode"] = "0" + strconv.FormatInt(int64(*v.Mode), 8)
			}
			m["path"] = v.Path
			items[i] = m
		}
		att["items"] = items
	}
	if in.Optional != nil {
		att["optional"] = *in.Optional
	}
	return []interface{}{att}
}

func flattenProjectedVolumeSource(in *v1.ProjectedVolumeSource) []interface{} {
	att := make(map[string]interface{})
	if in.DefaultMode != nil {
		att["default_mode"] = "0" + strconv.FormatInt(int64(*in.DefaultMode), 8)
	}
	if len(in.Sources) > 0 {
		sources := make([]interface{}, 0, len(in.Sources))
		for _, src := range in.Sources {
			s := make(map[string]interface{})
			if src.Secret != nil {
				s["secret"] = flattenSecretProjection(src.Secret)
			}
			if src.ConfigMap != nil {
				s["config_map"] = flattenConfigMapProjection(src.ConfigMap)
			}
			if src.DownwardAPI != nil {
				s["downward_api"] = flattenDownwardAPIProjection(src.DownwardAPI)
			}
			if src.ServiceAccountToken != nil {
				s["service_account_token"] = flattenServiceAccountTokenProjection(src.ServiceAccountToken)
			}
			sources = append(sources, s)
		}
		att["sources"] = sources
	}
	return []interface{}{att}
}

func flattenSecretProjection(in *v1.SecretProjection) []interface{} {
	att := make(map[string]interface{})
	if in.Name != "" {
		att["name"] = in.Name
	}
	if len(in.Items) > 0 {
		items := make([]interface{}, len(in.Items))
		for i, v := range in.Items {
			m := map[string]interface{}{}
			m["key"] = v.Key
			if v.Mode != nil {
				m["mode"] = "0" + strconv.FormatInt(int64(*v.Mode), 8)
			}
			m["path"] = v.Path
			items[i] = m
		}
		att["items"] = items
	}
	if in.Optional != nil {
		att["optional"] = *in.Optional
	}
	return []interface{}{att}
}

func flattenConfigMapProjection(in *v1.ConfigMapProjection) []interface{} {
	att := make(map[string]interface{})
	att["name"] = in.Name
	if len(in.Items) > 0 {
		items := make([]interface{}, len(in.Items))
		for i, v := range in.Items {
			m := map[string]interface{}{}
			if v.Key != "" {
				m["key"] = v.Key
			}
			if v.Mode != nil {
				m["mode"] = "0" + strconv.FormatInt(int64(*v.Mode), 8)
			}
			if v.Path != "" {
				m["path"] = v.Path
			}
			items[i] = m
		}
		att["items"] = items
	}
	return []interface{}{att}
}

func flattenDownwardAPIProjection(in *v1.DownwardAPIProjection) []interface{} {
	att := make(map[string]interface{})
	if len(in.Items) > 0 {
		att["items"] = flattenDownwardAPIVolumeFile(in.Items)
	}
	return []interface{}{att}
}

func flattenServiceAccountTokenProjection(in *v1.ServiceAccountTokenProjection) []interface{} {
	att := make(map[string]interface{})
	if in.Audience != "" {
		att["audience"] = in.Audience
	}
	if in.ExpirationSeconds != nil {
		att["expiration_seconds"] = in.ExpirationSeconds
	}
	if in.Path != "" {
		att["path"] = in.Path
	}
	return []interface{}{att}
}

func flattenReadinessGates(in []v1.PodReadinessGate) ([]interface{}, error) {
	att := make([]interface{}, len(in))
	for i, v := range in {
		c := make(map[string]interface{})
		c["condition_type"] = v.ConditionType
		att[i] = c
	}
	return att, nil
}

// Expanders
func expandPodSpec(p []interface{}) (*v1.PodSpec, error) {
	obj := &v1.PodSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, nil
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["active_deadline_seconds"].(int); ok && v > 0 {
		obj.ActiveDeadlineSeconds = ptrToInt64(int64(v))
	}

	if v, ok := in["image_pull_secrets"].([]interface{}); ok && len(v) > 0 {
		obj.ImagePullSecrets = expandImagePullSecrets(v)
	}
	if v, ok := in["affinity"].([]interface{}); ok && len(v) > 0 {
		a, err := expandAffinity(v)
		if err != nil {
			return obj, err
		}
		obj.Affinity = a
	}

	if v, ok := in["automount_service_account_token"].(bool); ok {
		obj.AutomountServiceAccountToken = ptrToBool(v)
	}

	if v, ok := in["container"].([]interface{}); ok && len(v) > 0 {
		cs, err := expandContainers(v)
		if err != nil {
			return obj, err
		}
		obj.Containers = cs
	}

	if v, ok := in["init_container"].([]interface{}); ok && len(v) > 0 {
		cs, err := expandContainers(v)
		if err != nil {
			return obj, err
		}
		obj.InitContainers = cs
	}

	if v, ok := in["dns_policy"].(string); ok {
		obj.DNSPolicy = v1.DNSPolicy(v)
	}

	if v, ok := in["enable_service_links"].(bool); ok {
		obj.EnableServiceLinks = ptrToBool(v)
	}

	if v, ok := in["host_ipc"]; ok {
		obj.HostIPC = v.(bool)
	}

	if v, ok := in["host_network"]; ok {
		obj.HostNetwork = v.(bool)
	}

	if v, ok := in["host_pid"]; ok {
		obj.HostPID = v.(bool)
	}

	if v, ok := in["hostname"]; ok {
		obj.Hostname = v.(string)
	}

	if v, ok := in["node_name"]; ok {
		obj.NodeName = v.(string)
	}

	if v, ok := in["node_selector"].(map[string]interface{}); ok {
		nodeSelectors := make(map[string]string)
		for k, v := range v {
			if val, ok := v.(string); ok {
				nodeSelectors[k] = val
			}
		}
		obj.NodeSelector = nodeSelectors
	}

	if v, ok := in["os"].(map[string]interface{}); ok {
		if n, ok := v["name"].(string); ok && n != "" {
			obj.OS.Name = v1.OSName(n)
		}
	}

	if v, ok := in["runtime_class_name"].(string); ok && v != "" {
		obj.RuntimeClassName = ptrToString(v)
	}

	if v, ok := in["priority_class_name"].(string); ok {
		obj.PriorityClassName = v
	}

	if v, ok := in["restart_policy"].(string); ok {
		obj.RestartPolicy = v1.RestartPolicy(v)
	}

	if v, ok := in["scheduler_name"].(string); ok {
		obj.SchedulerName = v
	}

	if v, ok := in["service_account_name"].(string); ok {
		obj.ServiceAccountName = v
	}

	if v, ok := in["share_process_namespace"]; ok {
		obj.ShareProcessNamespace = ptrToBool(v.(bool))
	}

	if v, ok := in["subdomain"].(string); ok {
		obj.Subdomain = v
	}

	if v, ok := in["termination_grace_period_seconds"].(int); ok {
		obj.TerminationGracePeriodSeconds = ptrToInt64(int64(v))
	}
	if v, ok := in["volumes"]; ok {
		r, err := expandPodSpecVolumes(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		obj.Volumes = r
	} else if v, ok := in["volume"]; ok { // Support for deprecated attribute
		r, err := expandPodSpecVolumes(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		obj.Volumes = r
	}
	return obj, nil
}

func expandPodSpecVolumes(volumes []interface{}) ([]v1.Volume, error) {
	if len(volumes) == 0 {
		return []v1.Volume{}, nil
	}
	vols := make([]v1.Volume, 0, len(volumes))
	for _, v := range volumes {
		vol := v1.Volume{}
		var err error
		mp := v.(map[string]interface{})
		if n, ok := mp["name"]; ok {
			vol.Name = n.(string)
		}
		if vmp, ok := mp["config_map"].([]interface{}); ok {
			vol.ConfigMap, err = expandConfigMap(vmp)
			if err != nil {
				return nil, err
			}

		}
		if vmp, ok := mp["secret"].([]interface{}); ok {
			vol.Secret, err = expandSecret(vmp)
			if err != nil {
				return nil, err
			}
		}
		if vmp, ok := mp["git_repo"].([]interface{}); ok {
			vol.GitRepo, err = expandGitRepo(vmp)
			if err != nil {
				return nil, err
			}
		}
		if vmp, ok := mp["downward_api"].([]interface{}); ok {
			api, err := expandDownwardAPI(vmp)
			if err != nil {
				return nil, err
			}
			vol.DownwardAPI = api
		}

		if vmp, ok := mp["csi"].([]interface{}); ok {
			vol.CSI = expandCSI(vmp)

		}
		if vmp, ok := mp["empty_dir"].([]interface{}); ok {
			emptyDir, err := expandEmptyDir(vmp)
			if err != nil {
				return nil, err
			}
			vol.EmptyDir = emptyDir
		}
		if vmp, ok := mp["ephemeral"].([]interface{}); ok {
			emph := expandEmphemeral(vmp)
			if emph != nil {
				vol.Ephemeral = emph
			}
		}
		if pvc, ok := mp["persistent_volume_claim"].([]interface{}); ok {
			vol.PersistentVolumeClaim = expandPVC(pvc)
		}

		if p, ok := mp["projected"].([]interface{}); ok {
			obj, err := expandProjected(p)
			if err != nil {
				return nil, err
			}
			if obj != nil {
				vol.Projected = obj
			}
		}
		vols = append(vols, vol)
	}
	return vols, nil
}

func expandConfigMap(configMap []interface{}) (*v1.ConfigMapVolumeSource, error) {
	if len(configMap) == 0 || configMap[0] == nil {
		return nil, nil
	}
	cmap := v1.ConfigMapVolumeSource{}
	mp := configMap[0].(map[string]interface{})
	if v, ok := mp["default_mode"]; ok {
		num, err := OctalToNumericInt32(v.(string))
		if err != nil {
			return nil, err
		}
		cmap.DefaultMode = ptrToInt32(num)
	}
	if v, ok := mp["optional"]; ok {
		val := v.(bool)
		cmap.Optional = &val
	}
	if v, ok := mp["name"]; ok {
		cmap.Name = v.(string)
	}
	if v, ok := mp["items"].([]interface{}); ok {
		items, err := expandItems(v)
		if err != nil {
			return nil, err
		}
		cmap.Items = items
	}
	return &cmap, nil
}

func expandSecret(secrets []interface{}) (*v1.SecretVolumeSource, error) {
	if len(secrets) == 0 || secrets[0] == nil {
		return nil, nil
	}
	secret := v1.SecretVolumeSource{}
	mp := secrets[0].(map[string]interface{})
	if v, ok := mp["default_mode"]; ok {
		num, err := OctalToNumericInt32(v.(string))
		if err != nil {
			return nil, err
		}
		secret.DefaultMode = ptrToInt32(num)
	}
	if v, ok := mp["optional"]; ok {
		val := v.(bool)
		secret.Optional = &val
	}
	if v, ok := mp["secret_name"]; ok {
		secret.SecretName = v.(string)
	}
	if v, ok := mp["items"].([]interface{}); ok {
		items, err := expandItems(v)
		if err != nil {
			return nil, err
		}
		secret.Items = items
	}
	return &secret, nil
}

func expandItems(items []interface{}) ([]v1.KeyToPath, error) {
	if len(items) == 0 {
		return nil, nil
	}
	mapItems := make([]v1.KeyToPath, 0, len(items))
	for _, item := range items {
		val := item.(map[string]interface{})
		i := v1.KeyToPath{}
		if v, ok := val["key"]; ok {
			i.Key = v.(string)
		}
		if v, ok := val["mode"]; ok {
			val, err := strconv.Atoi(v.(string))
			if err != nil {
				return nil, err
			}
			i.Mode = ptrToInt32(int32(val))
		}
		if v, ok := val["path"]; ok {
			i.Path = v.(string)
		}
		mapItems = append(mapItems, i)
	}
	return mapItems, nil
}

func expandGitRepo(gitRepo []interface{}) (*v1.GitRepoVolumeSource, error) {
	if len(gitRepo) == 0 || gitRepo[0] == nil {
		return nil, nil
	}
	gitBody := v1.GitRepoVolumeSource{}
	gitMap := gitRepo[0].(map[string]interface{})
	if v, ok := gitMap["directory"]; ok {
		gitBody.Directory = v.(string)
	}
	if v, ok := gitMap["repository"]; ok {
		gitBody.Repository = v.(string)
	}
	if v, ok := gitMap["revision"]; ok {
		gitBody.Revision = v.(string)
	}
	return &gitBody, nil
}

func expandCSI(csi []interface{}) *v1.CSIVolumeSource {
	if len(csi) == 0 || csi[0] == nil {
		return nil
	}
	csiBody := v1.CSIVolumeSource{}
	csiMap := csi[0].(map[string]interface{})
	if v, ok := csiMap["driver"]; ok {
		csiBody.Driver = v.(string)
	}
	if v, ok := csiMap["volume_attributes"]; ok {
		csiBody.VolumeAttributes = v.(map[string]string)
	}
	if v, ok := csiMap["fs_type"]; ok {
		str := v.(string)
		csiBody.FSType = &str
	}
	if v, ok := csiMap["read_only"]; ok {
		flag := v.(bool)
		csiBody.ReadOnly = &flag
	}
	if v, ok := csiMap["node_publish_secret_ref"].([]interface{}); ok {
		csiBody.NodePublishSecretRef = expandNodePublishSecretRef(v)
	}
	return &csiBody
}

func expandNodePublishSecretRef(npsr []interface{}) *v1.LocalObjectReference {
	if len(npsr) == 0 || npsr[0] == nil {
		return nil
	}
	npsrBody := v1.LocalObjectReference{}
	npsrMap := npsr[0].(map[string]interface{})
	if v, ok := npsrMap["name"]; ok {
		npsrBody.Name = v.(string)
	}
	return &npsrBody
}

func expandDownwardAPI(downwardApi []interface{}) (*v1.DownwardAPIVolumeSource, error) {
	if len(downwardApi) == 0 || downwardApi[0] == nil {
		return nil, nil
	}
	downwardAPIBody := v1.DownwardAPIVolumeSource{}
	apiMap := downwardApi[0].(map[string]interface{})
	if v, ok := apiMap["default_mode"]; ok {
		num, err := OctalToNumericInt32(v.(string))
		if err != nil {
			return nil, err
		}
		downwardAPIBody.DefaultMode = ptrToInt32(num)
	}

	if v, ok := apiMap["items"]; ok {
		items, err := expandDownwardAPIItems(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		downwardAPIBody.Items = items
	}
	return &downwardAPIBody, nil
}

func expandDownwardAPIItems(items []interface{}) ([]v1.DownwardAPIVolumeFile, error) {

	if len(items) == 0 {
		return nil, nil
	}
	mapItems := make([]v1.DownwardAPIVolumeFile, 0, len(items))
	for _, item := range items {
		val := item.(map[string]interface{})
		i := v1.DownwardAPIVolumeFile{}
		if v, ok := val["field_ref"]; ok {
			ref, err := expandFieldRef(v.([]interface{}))
			if err != nil {
				return nil, err
			}
			i.FieldRef = ref
		}
		if v, ok := val["mode"]; ok {
			val, err := strconv.Atoi(v.(string))
			if err != nil {
				return nil, err
			}
			i.Mode = ptrToInt32(int32(val))
		}
		if v, ok := val["path"]; ok {
			i.Path = v.(string)
		}
		if v, ok := val["resource_field_ref"]; ok {
			ref, err := expandResourceFieldRef(v.([]interface{}))
			if err != nil {
				return nil, err
			}
			i.ResourceFieldRef = ref
		}
		mapItems = append(mapItems, i)

	}
	return mapItems, nil
}

func expandEmptyDir(dir []interface{}) (*v1.EmptyDirVolumeSource, error) {
	if len(dir) == 0 || dir[0] == nil {
		return nil, nil
	}
	dirBody := v1.EmptyDirVolumeSource{}
	dirMap := dir[0].(map[string]interface{})
	if v, ok := dirMap["medium"]; ok {
		dirBody.Medium = v.(v1.StorageMedium)
	}
	if v, ok := dirMap["size_limit"]; ok {

		qty, err := resource.ParseQuantity(v.(string))
		if err != nil {
			return nil, err
		}
		dirBody.SizeLimit = &qty
	}
	return &dirBody, nil
}

func expandEmphemeral(emp []interface{}) *v1.EphemeralVolumeSource {
	if len(emp) == 0 || emp[0] == nil {
		return nil
	}
	empBody := v1.EphemeralVolumeSource{}
	empMap := emp[0].(map[string]interface{})
	if v, ok := empMap["volume_claim_template"]; ok {
		empBody.VolumeClaimTemplate = expandVolumeClaimTemplate(v.([]interface{}))
	}
	return &empBody

}

func expandVolumeClaimTemplate(vct []interface{}) *v1.PersistentVolumeClaimTemplate {
	if len(vct) == 0 || vct[0] == nil {
		return nil
	}
	vctBody := v1.PersistentVolumeClaimTemplate{}
	vctMap := vct[0].(map[string]interface{})
	if v, ok := vctMap["metadata"]; ok {
		vctBody.ObjectMeta = expandMetadata(v.([]interface{}))
	}
	if v, ok := vctMap["spec"]; ok {
		r := expandVolumeClaimTemplateSpec(v.([]interface{}))
		if r != nil {
			vctBody.Spec = *r
		}
	}
	return &vctBody //, nil
}

func expandVolumeClaimTemplateSpec(s []interface{}) *v1.PersistentVolumeClaimSpec {
	if len(s) == 0 || s[0] == nil {
		return nil
	}
	specBody := v1.PersistentVolumeClaimSpec{}
	specMap := s[0].(map[string]interface{})
	if v, ok := specMap["access_modes"]; ok {
		specBody.AccessModes = v.([]v1.PersistentVolumeAccessMode)
	}
	if v, ok := specMap["resources"]; ok {
		r := expandSpecResource(v.([]interface{}))
		if r != nil {
			specBody.Resources = *r
		}
	}
	if v, ok := specMap["volume_name"]; ok {
		specBody.VolumeName = v.(string)
	}
	if v, ok := specMap["volume_mode"]; ok {
		specBody.VolumeMode = v.(*v1.PersistentVolumeMode)

	}
	if v, ok := specMap["storage_class_name"]; ok {
		val := v.(string)
		specBody.StorageClassName = &val

	}
	return &specBody
}

func expandSpecResource(r []interface{}) *v1.ResourceRequirements {
	if len(r) == 0 || r[0] == nil {
		return nil
	}
	resource := v1.ResourceRequirements{}
	rsrcMap := r[0].(map[string]interface{})

	if v, ok := rsrcMap["limits"]; ok {
		resource.Limits = v.(v1.ResourceList)

	}
	if v, ok := rsrcMap["requests"]; ok {
		resource.Requests = v.(v1.ResourceList)
	}
	return &resource
}

func expandPVC(p []interface{}) *v1.PersistentVolumeClaimVolumeSource {
	if len(p) == 0 || p[0] != nil {
		return nil
	}
	pvcBody := v1.PersistentVolumeClaimVolumeSource{}
	pvcMap := p[0].(map[string]interface{})

	if v, ok := pvcMap["claim_name"]; ok {
		pvcBody.ClaimName = v.(string)
	}
	if v, ok := pvcMap["read_only"]; ok {
		pvcBody.ReadOnly = v.(bool)
	}
	return &pvcBody
}

func expandProjected(p []interface{}) (*v1.ProjectedVolumeSource, error) {
	if len(p) == 0 || p[0] != nil {
		return nil, nil
	}
	projBody := v1.ProjectedVolumeSource{}

	mp := p[0].(map[string]interface{})

	if v, ok := mp["default_mode"]; ok {
		num, err := OctalToNumericInt32(v.(string))
		if err != nil {
			return nil, err
		}
		projBody.DefaultMode = ptrToInt32(num)
	}

	if m, ok := mp["sources"]; ok {
		obj, err := expandSources(m.([]interface{}))
		if err != nil {
			return nil, err
		}
		projBody.Sources = obj
	}

	return &projBody, nil
}

func expandSources(s []interface{}) ([]v1.VolumeProjection, error) {
	if len(s) == 0 {
		return nil, nil
	}
	sourceBody := make([]v1.VolumeProjection, 0, len(s))
	for _, v := range s {
		sourceObj := v1.VolumeProjection{}
		mp := v.(map[string]interface{})
		if m, ok := mp["secret"]; ok {
			obj, err := expandSecret(m.([]interface{}))
			if err != nil {
				return nil, err
			}
			sourceObj.Secret = &v1.SecretProjection{
				LocalObjectReference: v1.LocalObjectReference{
					Name: obj.SecretName},
				Items:    obj.Items,
				Optional: obj.Optional,
			}
		}
		if m, ok := mp["config_map"]; ok {
			obj, err := expandConfigMap(m.([]interface{}))
			if err != nil {
				return nil, err
			}
			sourceObj.ConfigMap = &v1.ConfigMapProjection{
				LocalObjectReference: v1.LocalObjectReference{
					Name: obj.Name,
				},
				Items:    obj.Items,
				Optional: obj.Optional,
			}
		}

		if m, ok := mp["downward_api"]; ok {
			obj, err := expandDownwardAPI(m.([]interface{}))
			if err != nil {
				return nil, err
			}
			sourceObj.DownwardAPI = &v1.DownwardAPIProjection{
				Items: obj.Items,
			}
		}
		if m, ok := mp["service_account_token"]; ok {
			sourceObj.ServiceAccountToken = expandServiceAccountToken(m.([]interface{}))
		}
		sourceBody = append(sourceBody, sourceObj)
	}
	return sourceBody, nil
}

func expandServiceAccountToken(sat []interface{}) *v1.ServiceAccountTokenProjection {
	if len(sat) == 0 || sat[0] != nil {
		return nil
	}
	tokenBody := v1.ServiceAccountTokenProjection{}
	tokenMap := sat[0].(map[string]interface{})

	if v, ok := tokenMap["audience"]; ok {
		tokenBody.Audience = v.(string)

	}
	if v, ok := tokenMap["expiration_seconds"]; ok {
		i := int64(v.(int))
		tokenBody.ExpirationSeconds = &i
	}
	if v, ok := tokenMap["path"]; ok {
		tokenBody.Path = v.(string)
	}
	return &tokenBody
}

func expandImagePullSecrets(val []interface{}) []v1.LocalObjectReference {
	sec := []v1.LocalObjectReference{}
	for _, v := range val {
		m := v.(map[string]interface{})

		se := v1.LocalObjectReference{
			Name: m["name"].(string),
		}
		sec = append(sec, se)
	}
	return sec
}
