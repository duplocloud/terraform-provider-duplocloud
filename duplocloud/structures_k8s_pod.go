package duplocloud

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v1 "k8s.io/api/core/v1"
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

	return obj, nil
}
