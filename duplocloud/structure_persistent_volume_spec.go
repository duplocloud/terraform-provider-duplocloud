package duplocloud

import (
	v1 "k8s.io/api/core/v1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Flatteners

func flattenAWSElasticBlockStoreVolumeSource(in *v1.AWSElasticBlockStoreVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["volume_id"] = in.VolumeID
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.Partition != 0 {
		att["partition"] = in.Partition
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenAzureDiskVolumeSource(in *v1.AzureDiskVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["disk_name"] = in.DiskName
	att["data_disk_uri"] = in.DataDiskURI
	if in.Kind != nil {
		att["kind"] = string(*in.Kind)
	}
	att["caching_mode"] = string(*in.CachingMode)
	if in.FSType != nil {
		att["fs_type"] = *in.FSType
	}
	if in.ReadOnly != nil {
		att["read_only"] = *in.ReadOnly
	}
	return []interface{}{att}
}

func flattenAzureFileVolumeSource(in *v1.AzureFileVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["secret_name"] = in.SecretName
	att["share_name"] = in.ShareName
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

//func flattenAzureFilePersistentVolumeSource(in *v1.AzureFilePersistentVolumeSource) []interface{} {
//	att := make(map[string]interface{})
//	att["secret_name"] = in.SecretName
//	att["share_name"] = in.ShareName
//	if in.ReadOnly != false {
//		att["read_only"] = in.ReadOnly
//	}
//	if in.SecretNamespace != nil {
//		att["secret_namespace"] = *in.SecretNamespace
//	}
//	return []interface{}{att}
//}

func flattenCephFSVolumeSource(in *v1.CephFSVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["monitors"] = newStringSet(schema.HashString, in.Monitors)
	if in.Path != "" {
		att["path"] = in.Path
	}
	if in.User != "" {
		att["user"] = in.User
	}
	if in.SecretFile != "" {
		att["secret_file"] = in.SecretFile
	}
	if in.SecretRef != nil {
		att["secret_ref"] = flattenLocalObjectReference(in.SecretRef)
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

//func flattenCephFSPersistentVolumeSource(in *v1.CephFSPersistentVolumeSource) []interface{} {
//	att := make(map[string]interface{})
//	att["monitors"] = newStringSet(schema.HashString, in.Monitors)
//	if in.Path != "" {
//		att["path"] = in.Path
//	}
//	if in.User != "" {
//		att["user"] = in.User
//	}
//	if in.SecretFile != "" {
//		att["secret_file"] = in.SecretFile
//	}
//	if in.SecretRef != nil {
//		att["secret_ref"] = flattenSecretReference(in.SecretRef)
//	}
//	if in.ReadOnly != false {
//		att["read_only"] = in.ReadOnly
//	}
//	return []interface{}{att}
//}

//func flattenCinderPersistentVolumeSource(in *v1.CinderPersistentVolumeSource) []interface{} {
//	att := make(map[string]interface{})
//	att["volume_id"] = in.VolumeID
//	if in.FSType != "" {
//		att["fs_type"] = in.FSType
//	}
//	if in.ReadOnly != false {
//		att["read_only"] = in.ReadOnly
//	}
//	return []interface{}{att}
//}

func flattenCinderVolumeSource(in *v1.CinderVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["volume_id"] = in.VolumeID
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenFCVolumeSource(in *v1.FCVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["target_ww_ns"] = newStringSet(schema.HashString, in.TargetWWNs)
	att["lun"] = *in.Lun
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

//func flattenFlexPersistentVolumeSource(in *v1.FlexPersistentVolumeSource) []interface{} {
//	att := make(map[string]interface{})
//	att["driver"] = in.Driver
//	if in.FSType != "" {
//		att["fs_type"] = in.FSType
//	}
//	if in.SecretRef != nil {
//		att["secret_ref"] = flattenSecretReference(in.SecretRef)
//	}
//	if in.ReadOnly != false {
//		att["read_only"] = in.ReadOnly
//	}
//	if len(in.Options) > 0 {
//		att["options"] = in.Options
//	}
//	return []interface{}{att}
//}

func flattenFlexVolumeSource(in *v1.FlexVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["driver"] = in.Driver
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.SecretRef != nil {
		att["secret_ref"] = flattenLocalObjectReference(in.SecretRef)
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	if len(in.Options) > 0 {
		att["options"] = in.Options
	}
	return []interface{}{att}
}

func flattenFlockerVolumeSource(in *v1.FlockerVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["dataset_name"] = in.DatasetName
	att["dataset_uuid"] = in.DatasetUUID
	return []interface{}{att}
}

func flattenGCEPersistentDiskVolumeSource(in *v1.GCEPersistentDiskVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["pd_name"] = in.PDName
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.Partition != 0 {
		att["partition"] = in.Partition
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

//func flattenGlusterfsPersistentVolumeSource(in *v1.GlusterfsPersistentVolumeSource) []interface{} {
//	att := make(map[string]interface{})
//	att["endpoints_name"] = in.EndpointsName
//	att["path"] = in.Path
//	if in.ReadOnly != false {
//		att["read_only"] = in.ReadOnly
//	}
//	return []interface{}{att}
//}

func flattenGlusterfsVolumeSource(in *v1.GlusterfsVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["endpoints_name"] = in.EndpointsName
	att["path"] = in.Path
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenHostPathVolumeSource(in *v1.HostPathVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["path"] = in.Path
	if in.Type != nil {
		att["type"] = string(*in.Type)
	}
	return []interface{}{att}
}

//func flattenLocalVolumeSource(in *v1.LocalVolumeSource) []interface{} {
//	att := make(map[string]interface{})
//	att["path"] = in.Path
//	return []interface{}{att}
//}

func flattenISCSIVolumeSource(in *v1.ISCSIVolumeSource) []interface{} {
	att := make(map[string]interface{})
	if in.TargetPortal != "" {
		att["target_portal"] = in.TargetPortal
	}
	if in.IQN != "" {
		att["iqn"] = in.IQN
	}
	if in.Lun != 0 {
		att["lun"] = in.Lun
	}
	if in.ISCSIInterface != "" {
		att["iscsi_interface"] = in.ISCSIInterface
	}
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

//func flattenISCSIPersistentVolumeSource(in *v1.ISCSIPersistentVolumeSource) []interface{} {
//	att := make(map[string]interface{})
//	if in.TargetPortal != "" {
//		att["target_portal"] = in.TargetPortal
//	}
//	if in.IQN != "" {
//		att["iqn"] = in.IQN
//	}
//	if in.Lun != 0 {
//		att["lun"] = in.Lun
//	}
//	if in.ISCSIInterface != "" {
//		att["iscsi_interface"] = in.ISCSIInterface
//	}
//	if in.FSType != "" {
//		att["fs_type"] = in.FSType
//	}
//	if in.ReadOnly != false {
//		att["read_only"] = in.ReadOnly
//	}
//	return []interface{}{att}
//}

func flattenLocalObjectReference(in *v1.LocalObjectReference) []interface{} {
	att := make(map[string]interface{})
	if in.Name != "" {
		att["name"] = in.Name
	}
	return []interface{}{att}
}

//func flattenSecretReference(in *v1.SecretReference) []interface{} {
//	att := make(map[string]interface{})
//	if in.Name != "" {
//		att["name"] = in.Name
//	}
//	if in.Namespace != "" {
//		att["namespace"] = in.Namespace
//	}
//	return []interface{}{att}
//}

func flattenNFSVolumeSource(in *v1.NFSVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["server"] = in.Server
	att["path"] = in.Path
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

//func flattenPersistentVolumeSource(in v1.PersistentVolumeSource) []interface{} {
//	att := make(map[string]interface{})
//	if in.GCEPersistentDisk != nil {
//		att["gce_persistent_disk"] = flattenGCEPersistentDiskVolumeSource(in.GCEPersistentDisk)
//	}
//	if in.AWSElasticBlockStore != nil {
//		att["aws_elastic_block_store"] = flattenAWSElasticBlockStoreVolumeSource(in.AWSElasticBlockStore)
//	}
//	if in.HostPath != nil {
//		att["host_path"] = flattenHostPathVolumeSource(in.HostPath)
//	}
//	if in.Local != nil {
//		att["local"] = flattenLocalVolumeSource(in.Local)
//	}
//	if in.Glusterfs != nil {
//		att["glusterfs"] = flattenGlusterfsPersistentVolumeSource(in.Glusterfs)
//	}
//	if in.NFS != nil {
//		att["nfs"] = flattenNFSVolumeSource(in.NFS)
//	}
//	if in.RBD != nil {
//		att["rbd"] = flattenRBDPersistentVolumeSource(in.RBD)
//	}
//	if in.ISCSI != nil {
//		att["iscsi"] = flattenISCSIPersistentVolumeSource(in.ISCSI)
//	}
//	if in.Cinder != nil {
//		att["cinder"] = flattenCinderPersistentVolumeSource(in.Cinder)
//	}
//	if in.CephFS != nil {
//		att["ceph_fs"] = flattenCephFSPersistentVolumeSource(in.CephFS)
//	}
//	if in.FC != nil {
//		att["fc"] = flattenFCVolumeSource(in.FC)
//	}
//	if in.Flocker != nil {
//		att["flocker"] = flattenFlockerVolumeSource(in.Flocker)
//	}
//	if in.FlexVolume != nil {
//		att["flex_volume"] = flattenFlexPersistentVolumeSource(in.FlexVolume)
//	}
//	if in.AzureFile != nil {
//		att["azure_file"] = flattenAzureFilePersistentVolumeSource(in.AzureFile)
//	}
//	if in.VsphereVolume != nil {
//		att["vsphere_volume"] = flattenVsphereVirtualDiskVolumeSource(in.VsphereVolume)
//	}
//	if in.Quobyte != nil {
//		att["quobyte"] = flattenQuobyteVolumeSource(in.Quobyte)
//	}
//	if in.AzureDisk != nil {
//		att["azure_disk"] = flattenAzureDiskVolumeSource(in.AzureDisk)
//	}
//	if in.PhotonPersistentDisk != nil {
//		att["photon_persistent_disk"] = flattenPhotonPersistentDiskVolumeSource(in.PhotonPersistentDisk)
//	}
//	if in.CSI != nil {
//		att["csi"] = flattenCSIPersistentVolumeSource(in.CSI)
//	}
//	return []interface{}{att}
//}

//func flattenPersistentVolumeSpec(in v1.PersistentVolumeSpec) []interface{} {
//	att := make(map[string]interface{})
//	if len(in.Capacity) > 0 {
//		att["capacity"] = flattenResourceList(in.Capacity)
//	}
//
//	att["persistent_volume_source"] = flattenPersistentVolumeSource(in.PersistentVolumeSource)
//	if len(in.AccessModes) > 0 {
//		att["access_modes"] = flattenPersistentVolumeAccessModes(in.AccessModes)
//	}
//	if in.PersistentVolumeReclaimPolicy != "" {
//		att["persistent_volume_reclaim_policy"] = in.PersistentVolumeReclaimPolicy
//	}
//	if in.StorageClassName != "" {
//		att["storage_class_name"] = in.StorageClassName
//	}
//	if in.NodeAffinity != nil {
//		att["node_affinity"] = flattenVolumeNodeAffinity(in.NodeAffinity)
//	}
//	if in.MountOptions != nil {
//		att["mount_options"] = flattenPersistentVolumeMountOptions(in.MountOptions)
//	}
//	if in.VolumeMode != nil {
//		att["volume_mode"] = in.VolumeMode
//	}
//	if in.ClaimRef != nil {
//		att["claim_ref"] = flattenObjectRef(in.ClaimRef)
//	}
//	return []interface{}{att}
//}

//func flattenObjectRef(in *v1.ObjectReference) []interface{} {
//	att := make(map[string]interface{})
//	if in.Name != "" {
//		att["name"] = in.Name
//	}
//	if in.Namespace != "" {
//		att["namespace"] = in.Namespace
//	}
//	return []interface{}{att}
//}

func flattenPhotonPersistentDiskVolumeSource(in *v1.PhotonPersistentDiskVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["pd_id"] = in.PdID
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	return []interface{}{att}
}

func flattenCSIVolumeSource(in *v1.CSIVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["driver"] = in.Driver
	if in.ReadOnly != nil {
		att["read_only"] = *in.ReadOnly
	}
	if in.FSType != nil {
		att["fs_type"] = *in.FSType
	}
	if len(in.VolumeAttributes) > 0 {
		att["volume_attributes"] = in.VolumeAttributes
	}
	if in.NodePublishSecretRef != nil {
		att["node_publish_secret_ref"] = flattenLocalObjectReference(in.NodePublishSecretRef)
	}
	return []interface{}{att}
}

func flattenQuobyteVolumeSource(in *v1.QuobyteVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["registry"] = in.Registry
	att["volume"] = in.Volume
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	if in.User != "" {
		att["user"] = in.User
	}
	if in.Group != "" {
		att["group"] = in.Group
	}
	return []interface{}{att}
}

func flattenRBDVolumeSource(in *v1.RBDVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["ceph_monitors"] = newStringSet(schema.HashString, in.CephMonitors)
	att["rbd_image"] = in.RBDImage
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	if in.RBDPool != "" {
		att["rbd_pool"] = in.RBDPool
	}
	if in.RadosUser != "" {
		att["rados_user"] = in.RadosUser
	}
	if in.Keyring != "" {
		att["keyring"] = in.Keyring
	}
	if in.SecretRef != nil {
		att["secret_ref"] = flattenLocalObjectReference(in.SecretRef)
	}
	if in.ReadOnly {
		att["read_only"] = in.ReadOnly
	}
	return []interface{}{att}
}

func flattenVsphereVirtualDiskVolumeSource(in *v1.VsphereVirtualDiskVolumeSource) []interface{} {
	att := make(map[string]interface{})
	att["volume_path"] = in.VolumePath
	if in.FSType != "" {
		att["fs_type"] = in.FSType
	}
	return []interface{}{att}
}
