package duplocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v1 "k8s.io/api/core/v1"
)

// Expanders for the volume sources declared by commonVolumeSources().
//
// Each function mirrors the matching flatten<X>VolumeSource in
// structure_persistent_volume_spec.go key for key. Without them the source is
// silently dropped on the way to the API, and the Kubernetes API server
// defaults a volume with no source at all to `emptyDir: {}`.
//
// The `local` block in commonVolumeSources() has no expander on purpose:
// LocalVolumeSource only exists on a PersistentVolume, not on a pod volume.

func expandHostPathVolumeSource(l []interface{}) *v1.HostPathVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.HostPathVolumeSource{}
	if v, ok := in["path"].(string); ok {
		obj.Path = v
	}
	// An empty type is the documented default, and sending it back as an
	// explicit "" would make the API echo a value the config never set.
	if v, ok := in["type"].(string); ok && v != "" {
		hostPathType := v1.HostPathType(v)
		obj.Type = &hostPathType
	}
	return &obj
}

func expandAWSElasticBlockStoreVolumeSource(l []interface{}) *v1.AWSElasticBlockStoreVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.AWSElasticBlockStoreVolumeSource{}
	if v, ok := in["volume_id"].(string); ok {
		obj.VolumeID = v
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["partition"].(int); ok {
		obj.Partition = int32(v)
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return &obj
}

func expandAzureDiskVolumeSource(l []interface{}) *v1.AzureDiskVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.AzureDiskVolumeSource{}
	if v, ok := in["disk_name"].(string); ok {
		obj.DiskName = v
	}
	if v, ok := in["data_disk_uri"].(string); ok {
		obj.DataDiskURI = v
	}
	if v, ok := in["caching_mode"].(string); ok && v != "" {
		cachingMode := v1.AzureDataDiskCachingMode(v)
		obj.CachingMode = &cachingMode
	}
	if v, ok := in["fs_type"].(string); ok && v != "" {
		obj.FSType = ptrToString(v)
	}
	if v, ok := in["kind"].(string); ok && v != "" {
		kind := v1.AzureDataDiskKind(v)
		obj.Kind = &kind
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = ptrToBool(v)
	}
	return &obj
}

func expandAzureFileVolumeSource(l []interface{}) *v1.AzureFileVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	// NOTE: the schema also exposes secret_namespace, which only exists on
	// AzureFilePersistentVolumeSource - a pod volume cannot carry it.
	obj := v1.AzureFileVolumeSource{}
	if v, ok := in["secret_name"].(string); ok {
		obj.SecretName = v
	}
	if v, ok := in["share_name"].(string); ok {
		obj.ShareName = v
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return &obj
}

func expandCephFSVolumeSource(l []interface{}) *v1.CephFSVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.CephFSVolumeSource{}
	if v, ok := in["monitors"].(*schema.Set); ok {
		obj.Monitors = expandStringSet(v)
	}
	if v, ok := in["path"].(string); ok {
		obj.Path = v
	}
	if v, ok := in["user"].(string); ok {
		obj.User = v
	}
	if v, ok := in["secret_file"].(string); ok {
		obj.SecretFile = v
	}
	if v, ok := in["secret_ref"].([]interface{}); ok {
		obj.SecretRef = expandCommonVolumeSourcesSecretRef(v)
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return &obj
}

func expandCinderVolumeSource(l []interface{}) *v1.CinderVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.CinderVolumeSource{}
	if v, ok := in["volume_id"].(string); ok {
		obj.VolumeID = v
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return &obj
}

func expandFCVolumeSource(l []interface{}) *v1.FCVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.FCVolumeSource{}
	if v, ok := in["target_ww_ns"].(*schema.Set); ok {
		obj.TargetWWNs = expandStringSet(v)
	}
	if v, ok := in["lun"].(int); ok {
		obj.Lun = ptrToInt32(int32(v))
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return &obj
}

func expandFlexVolumeSource(l []interface{}) *v1.FlexVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.FlexVolumeSource{}
	if v, ok := in["driver"].(string); ok {
		obj.Driver = v
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["options"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Options = expandStringMap(v)
	}
	if v, ok := in["secret_ref"].([]interface{}); ok {
		obj.SecretRef = expandCommonVolumeSourcesSecretRef(v)
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return &obj
}

func expandFlockerVolumeSource(l []interface{}) *v1.FlockerVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.FlockerVolumeSource{}
	if v, ok := in["dataset_name"].(string); ok {
		obj.DatasetName = v
	}
	if v, ok := in["dataset_uuid"].(string); ok {
		obj.DatasetUUID = v
	}
	return &obj
}

func expandGCEPersistentDiskVolumeSource(l []interface{}) *v1.GCEPersistentDiskVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.GCEPersistentDiskVolumeSource{}
	if v, ok := in["pd_name"].(string); ok {
		obj.PDName = v
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["partition"].(int); ok {
		obj.Partition = int32(v)
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return &obj
}

func expandGlusterfsVolumeSource(l []interface{}) *v1.GlusterfsVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.GlusterfsVolumeSource{}
	if v, ok := in["endpoints_name"].(string); ok {
		obj.EndpointsName = v
	}
	if v, ok := in["path"].(string); ok {
		obj.Path = v
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return &obj
}

func expandISCSIVolumeSource(l []interface{}) *v1.ISCSIVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.ISCSIVolumeSource{}
	if v, ok := in["target_portal"].(string); ok {
		obj.TargetPortal = v
	}
	if v, ok := in["iqn"].(string); ok {
		obj.IQN = v
	}
	if v, ok := in["lun"].(int); ok {
		obj.Lun = int32(v)
	}
	if v, ok := in["iscsi_interface"].(string); ok {
		obj.ISCSIInterface = v
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return &obj
}

func expandNFSVolumeSource(l []interface{}) *v1.NFSVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.NFSVolumeSource{}
	if v, ok := in["server"].(string); ok {
		obj.Server = v
	}
	if v, ok := in["path"].(string); ok {
		obj.Path = v
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return &obj
}

func expandPhotonPersistentDiskVolumeSource(l []interface{}) *v1.PhotonPersistentDiskVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.PhotonPersistentDiskVolumeSource{}
	if v, ok := in["pd_id"].(string); ok {
		obj.PdID = v
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	return &obj
}

func expandQuobyteVolumeSource(l []interface{}) *v1.QuobyteVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.QuobyteVolumeSource{}
	if v, ok := in["registry"].(string); ok {
		obj.Registry = v
	}
	if v, ok := in["volume"].(string); ok {
		obj.Volume = v
	}
	if v, ok := in["user"].(string); ok {
		obj.User = v
	}
	if v, ok := in["group"].(string); ok {
		obj.Group = v
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return &obj
}

func expandRBDVolumeSource(l []interface{}) *v1.RBDVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.RBDVolumeSource{}
	if v, ok := in["ceph_monitors"].(*schema.Set); ok {
		obj.CephMonitors = expandStringSet(v)
	}
	if v, ok := in["rbd_image"].(string); ok {
		obj.RBDImage = v
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	if v, ok := in["rbd_pool"].(string); ok {
		obj.RBDPool = v
	}
	if v, ok := in["rados_user"].(string); ok {
		obj.RadosUser = v
	}
	if v, ok := in["keyring"].(string); ok {
		obj.Keyring = v
	}
	if v, ok := in["secret_ref"].([]interface{}); ok {
		obj.SecretRef = expandCommonVolumeSourcesSecretRef(v)
	}
	if v, ok := in["read_only"].(bool); ok {
		obj.ReadOnly = v
	}
	return &obj
}

func expandVsphereVirtualDiskVolumeSource(l []interface{}) *v1.VsphereVirtualDiskVolumeSource {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.VsphereVirtualDiskVolumeSource{}
	if v, ok := in["volume_path"].(string); ok {
		obj.VolumePath = v
	}
	if v, ok := in["fs_type"].(string); ok {
		obj.FSType = v
	}
	return &obj
}

// expandCommonVolumeSourcesSecretRef mirrors commonVolumeSourcesSecretRef().
// Pod-level volume sources reference a secret in the pod's own namespace, so
// only the name is carried over - the schema's namespace is a PersistentVolume
// concept.
func expandCommonVolumeSourcesSecretRef(l []interface{}) *v1.LocalObjectReference {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})
	obj := v1.LocalObjectReference{}
	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}
	return &obj
}
