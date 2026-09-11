package duplocloud

import (
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v1 "k8s.io/api/core/v1"
)

// TestExpandPodSpecVolumesRoundTrip walks every volume source the schema
// accepts through expand -> flatten -> expand and asserts the source survives.
//
// The regression this guards (DUPLO-44511) is a volume that reaches the API
// with no source set at all: Kubernetes defaults that to `emptyDir: {}`, so a
// host_path volume silently became an emptyDir one.
func TestExpandPodSpecVolumesRoundTrip(t *testing.T) {
	stringSet := func(values ...string) *schema.Set {
		items := make([]interface{}, 0, len(values))
		for _, v := range values {
			items = append(items, v)
		}
		return schema.NewSet(schema.HashString, items)
	}

	cases := []struct {
		name   string
		volume map[string]interface{}
		// assert reports the offending detail, or "" when the source survived.
		assert func(v v1.Volume) string
	}{
		{
			name: "host_path",
			volume: map[string]interface{}{
				"name": "var-log",
				"host_path": []interface{}{map[string]interface{}{
					"path": "/var/log",
					"type": "DirectoryOrCreate",
				}},
			},
			assert: func(v v1.Volume) string {
				if v.HostPath == nil {
					return "HostPath is nil"
				}
				if v.HostPath.Path != "/var/log" {
					return "path = " + v.HostPath.Path
				}
				if v.HostPath.Type == nil || *v.HostPath.Type != v1.HostPathDirectoryOrCreate {
					return "unexpected type"
				}
				return ""
			},
		},
		{
			name: "host_path without a type",
			volume: map[string]interface{}{
				"name": "docker-containers",
				"host_path": []interface{}{map[string]interface{}{
					"path": "/var/lib/docker/containers",
					"type": "",
				}},
			},
			assert: func(v v1.Volume) string {
				if v.HostPath == nil {
					return "HostPath is nil"
				}
				if v.HostPath.Type != nil {
					return "type should stay unset when the config omits it"
				}
				return ""
			},
		},
		{
			name: "empty_dir without a size_limit",
			volume: map[string]interface{}{
				"name": "config-vol",
				"empty_dir": []interface{}{map[string]interface{}{
					"medium":     "Memory",
					"size_limit": "",
				}},
			},
			assert: func(v v1.Volume) string {
				if v.EmptyDir == nil {
					return "EmptyDir is nil"
				}
				if v.EmptyDir.Medium != v1.StorageMediumMemory {
					return "medium not carried over"
				}
				if v.EmptyDir.SizeLimit != nil {
					return "size_limit should stay unset when the config omits it"
				}
				return ""
			},
		},
		{
			name: "aws_elastic_block_store",
			volume: map[string]interface{}{
				"name": "ebs",
				"aws_elastic_block_store": []interface{}{map[string]interface{}{
					"volume_id": "vol-0123456789abcdef0",
					"fs_type":   "ext4",
					"partition": 1,
					"read_only": true,
				}},
			},
			assert: func(v v1.Volume) string {
				if v.AWSElasticBlockStore == nil {
					return "AWSElasticBlockStore is nil"
				}
				if v.AWSElasticBlockStore.VolumeID != "vol-0123456789abcdef0" {
					return "volume_id = " + v.AWSElasticBlockStore.VolumeID
				}
				if v.AWSElasticBlockStore.Partition != 1 || !v.AWSElasticBlockStore.ReadOnly {
					return "partition/read_only not carried over"
				}
				return ""
			},
		},
		{
			name: "azure_disk",
			volume: map[string]interface{}{
				"name": "azdisk",
				"azure_disk": []interface{}{map[string]interface{}{
					"disk_name":     "disk1",
					"data_disk_uri": "https://example.blob.core.windows.net/vhds/disk1.vhd",
					"caching_mode":  "ReadWrite",
					"kind":          "Managed",
					"fs_type":       "ext4",
					"read_only":     false,
				}},
			},
			assert: func(v v1.Volume) string {
				if v.AzureDisk == nil {
					return "AzureDisk is nil"
				}
				if v.AzureDisk.DiskName != "disk1" {
					return "disk_name = " + v.AzureDisk.DiskName
				}
				if v.AzureDisk.CachingMode == nil || *v.AzureDisk.CachingMode != v1.AzureDataDiskCachingReadWrite {
					return "caching_mode not carried over"
				}
				if v.AzureDisk.Kind == nil || *v.AzureDisk.Kind != v1.AzureManagedDisk {
					return "kind not carried over"
				}
				return ""
			},
		},
		{
			name: "azure_file",
			volume: map[string]interface{}{
				"name": "azfile",
				"azure_file": []interface{}{map[string]interface{}{
					"secret_name": "azure-secret",
					"share_name":  "share1",
					"read_only":   true,
				}},
			},
			assert: func(v v1.Volume) string {
				if v.AzureFile == nil {
					return "AzureFile is nil"
				}
				if v.AzureFile.ShareName != "share1" || !v.AzureFile.ReadOnly {
					return "share_name/read_only not carried over"
				}
				return ""
			},
		},
		{
			name: "ceph_fs",
			volume: map[string]interface{}{
				"name": "cephfs",
				"ceph_fs": []interface{}{map[string]interface{}{
					"monitors":    stringSet("10.16.154.78:6789", "10.16.154.82:6789"),
					"path":        "/data",
					"user":        "admin",
					"secret_file": "/etc/ceph/admin.secret",
					"secret_ref":  []interface{}{map[string]interface{}{"name": "ceph-secret"}},
					"read_only":   true,
				}},
			},
			assert: func(v v1.Volume) string {
				if v.CephFS == nil {
					return "CephFS is nil"
				}
				if len(v.CephFS.Monitors) != 2 {
					return "monitors not carried over"
				}
				if v.CephFS.SecretRef == nil || v.CephFS.SecretRef.Name != "ceph-secret" {
					return "secret_ref not carried over"
				}
				return ""
			},
		},
		{
			name: "cinder",
			volume: map[string]interface{}{
				"name": "cinder",
				"cinder": []interface{}{map[string]interface{}{
					"volume_id": "cinder-vol-1",
					"fs_type":   "ext4",
					"read_only": false,
				}},
			},
			assert: func(v v1.Volume) string {
				if v.Cinder == nil {
					return "Cinder is nil"
				}
				if v.Cinder.VolumeID != "cinder-vol-1" {
					return "volume_id = " + v.Cinder.VolumeID
				}
				return ""
			},
		},
		{
			name: "fc",
			volume: map[string]interface{}{
				"name": "fc",
				"fc": []interface{}{map[string]interface{}{
					"target_ww_ns": stringSet("500a0982991b8dc5"),
					"lun":          2,
					"fs_type":      "ext4",
					"read_only":    false,
				}},
			},
			assert: func(v v1.Volume) string {
				if v.FC == nil {
					return "FC is nil"
				}
				if len(v.FC.TargetWWNs) != 1 {
					return "target_ww_ns not carried over"
				}
				if v.FC.Lun == nil || *v.FC.Lun != 2 {
					return "lun not carried over"
				}
				return ""
			},
		},
		{
			name: "flex_volume",
			volume: map[string]interface{}{
				"name": "flex",
				"flex_volume": []interface{}{map[string]interface{}{
					"driver":     "kubernetes.io/lvm",
					"fs_type":    "ext4",
					"options":    map[string]interface{}{"volumeID": "vol1"},
					"secret_ref": []interface{}{map[string]interface{}{"name": "flex-secret"}},
					"read_only":  false,
				}},
			},
			assert: func(v v1.Volume) string {
				if v.FlexVolume == nil {
					return "FlexVolume is nil"
				}
				if v.FlexVolume.Driver != "kubernetes.io/lvm" {
					return "driver = " + v.FlexVolume.Driver
				}
				if v.FlexVolume.Options["volumeID"] != "vol1" {
					return "options not carried over"
				}
				if v.FlexVolume.SecretRef == nil || v.FlexVolume.SecretRef.Name != "flex-secret" {
					return "secret_ref not carried over"
				}
				return ""
			},
		},
		{
			name: "flocker",
			volume: map[string]interface{}{
				"name": "flocker",
				"flocker": []interface{}{map[string]interface{}{
					"dataset_name": "ds1",
					"dataset_uuid": "uuid-1",
				}},
			},
			assert: func(v v1.Volume) string {
				if v.Flocker == nil {
					return "Flocker is nil"
				}
				if v.Flocker.DatasetName != "ds1" || v.Flocker.DatasetUUID != "uuid-1" {
					return "dataset fields not carried over"
				}
				return ""
			},
		},
		{
			name: "gce_persistent_disk",
			volume: map[string]interface{}{
				"name": "gcepd",
				"gce_persistent_disk": []interface{}{map[string]interface{}{
					"pd_name":   "pd1",
					"fs_type":   "ext4",
					"partition": 2,
					"read_only": true,
				}},
			},
			assert: func(v v1.Volume) string {
				if v.GCEPersistentDisk == nil {
					return "GCEPersistentDisk is nil"
				}
				if v.GCEPersistentDisk.PDName != "pd1" || v.GCEPersistentDisk.Partition != 2 {
					return "pd_name/partition not carried over"
				}
				return ""
			},
		},
		{
			name: "glusterfs",
			volume: map[string]interface{}{
				"name": "gluster",
				"glusterfs": []interface{}{map[string]interface{}{
					"endpoints_name": "glusterfs-cluster",
					"path":           "vol1",
					"read_only":      false,
				}},
			},
			assert: func(v v1.Volume) string {
				if v.Glusterfs == nil {
					return "Glusterfs is nil"
				}
				if v.Glusterfs.EndpointsName != "glusterfs-cluster" || v.Glusterfs.Path != "vol1" {
					return "endpoints_name/path not carried over"
				}
				return ""
			},
		},
		{
			name: "iscsi",
			volume: map[string]interface{}{
				"name": "iscsi",
				"iscsi": []interface{}{map[string]interface{}{
					"target_portal":   "10.0.2.15:3260",
					"iqn":             "iqn.2001-04.com.example:storage.kube.sys1.xyz",
					"lun":             1,
					"iscsi_interface": "default",
					"fs_type":         "ext4",
					"read_only":       true,
				}},
			},
			assert: func(v v1.Volume) string {
				if v.ISCSI == nil {
					return "ISCSI is nil"
				}
				if v.ISCSI.TargetPortal != "10.0.2.15:3260" || v.ISCSI.Lun != 1 {
					return "target_portal/lun not carried over"
				}
				return ""
			},
		},
		{
			name: "nfs",
			volume: map[string]interface{}{
				"name": "nfs",
				"nfs": []interface{}{map[string]interface{}{
					"server":    "nfs.example.com",
					"path":      "/exports",
					"read_only": false,
				}},
			},
			assert: func(v v1.Volume) string {
				if v.NFS == nil {
					return "NFS is nil"
				}
				if v.NFS.Server != "nfs.example.com" || v.NFS.Path != "/exports" {
					return "server/path not carried over"
				}
				return ""
			},
		},
		{
			name: "photon_persistent_disk",
			volume: map[string]interface{}{
				"name": "photon",
				"photon_persistent_disk": []interface{}{map[string]interface{}{
					"pd_id":   "pd-1",
					"fs_type": "ext4",
				}},
			},
			assert: func(v v1.Volume) string {
				if v.PhotonPersistentDisk == nil {
					return "PhotonPersistentDisk is nil"
				}
				if v.PhotonPersistentDisk.PdID != "pd-1" {
					return "pd_id = " + v.PhotonPersistentDisk.PdID
				}
				return ""
			},
		},
		{
			name: "quobyte",
			volume: map[string]interface{}{
				"name": "quobyte",
				"quobyte": []interface{}{map[string]interface{}{
					"registry":  "registry:7861",
					"volume":    "vol1",
					"user":      "root",
					"group":     "root",
					"read_only": false,
				}},
			},
			assert: func(v v1.Volume) string {
				if v.Quobyte == nil {
					return "Quobyte is nil"
				}
				if v.Quobyte.Registry != "registry:7861" || v.Quobyte.Volume != "vol1" {
					return "registry/volume not carried over"
				}
				return ""
			},
		},
		{
			name: "rbd",
			volume: map[string]interface{}{
				"name": "rbd",
				"rbd": []interface{}{map[string]interface{}{
					"ceph_monitors": stringSet("10.16.154.78:6789"),
					"rbd_image":     "foo",
					"fs_type":       "ext4",
					"rbd_pool":      "kube",
					"rados_user":    "admin",
					"keyring":       "/etc/ceph/keyring",
					"secret_ref":    []interface{}{map[string]interface{}{"name": "ceph-secret"}},
					"read_only":     false,
				}},
			},
			assert: func(v v1.Volume) string {
				if v.RBD == nil {
					return "RBD is nil"
				}
				if len(v.RBD.CephMonitors) != 1 || v.RBD.RBDImage != "foo" {
					return "ceph_monitors/rbd_image not carried over"
				}
				if v.RBD.SecretRef == nil || v.RBD.SecretRef.Name != "ceph-secret" {
					return "secret_ref not carried over"
				}
				return ""
			},
		},
		{
			name: "vsphere_volume",
			volume: map[string]interface{}{
				"name": "vsphere",
				"vsphere_volume": []interface{}{map[string]interface{}{
					"volume_path": "[datastore1] volumes/myDisk",
					"fs_type":     "ext4",
				}},
			},
			assert: func(v v1.Volume) string {
				if v.VsphereVolume == nil {
					return "VsphereVolume is nil"
				}
				if v.VsphereVolume.VolumePath != "[datastore1] volumes/myDisk" {
					return "volume_path = " + v.VsphereVolume.VolumePath
				}
				return ""
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expanded, err := expandPodSpecVolumes([]interface{}{tc.volume})
			if err != nil {
				t.Fatalf("expandPodSpecVolumes: %s", err)
			}
			if len(expanded) != 1 {
				t.Fatalf("expected 1 volume, got %d", len(expanded))
			}

			// The bug being guarded: a volume with no source at all, which the
			// Kubernetes API server rewrites into an emptyDir.
			if reflect.DeepEqual(expanded[0].VolumeSource, v1.VolumeSource{}) {
				t.Fatal("volume reached the API with no source set - Kubernetes will default it to emptyDir")
			}
			if problem := tc.assert(expanded[0]); problem != "" {
				t.Fatalf("after expand: %s", problem)
			}

			// Reading the resource back and re-planning must produce the same
			// request, otherwise the config drifts on every apply.
			flattened, err := flattenVolumes(expanded)
			if err != nil {
				t.Fatalf("flattenVolumes: %s", err)
			}
			reExpanded, err := expandPodSpecVolumes(flattened)
			if err != nil {
				t.Fatalf("expandPodSpecVolumes (round trip): %s", err)
			}
			if problem := tc.assert(reExpanded[0]); problem != "" {
				t.Fatalf("after round trip: %s", problem)
			}
			if !reflect.DeepEqual(expanded[0], reExpanded[0]) {
				t.Fatalf("round trip changed the volume:\n before: %#v\n after:  %#v", expanded[0], reExpanded[0])
			}
		})
	}
}

// TestExpandPodSpecVolumesRejectsLocal pins the one volume source that has no
// pod-level equivalent. LocalVolumeSource only exists on a PersistentVolume, so
// there is nothing to expand a local block into - left alone it would reach the
// API with an empty VolumeSource and come back as an emptyDir, which is the
// DUPLO-44511 regression wearing a different hat.
func TestExpandPodSpecVolumesRejectsLocal(t *testing.T) {
	_, err := expandPodSpecVolumes([]interface{}{map[string]interface{}{
		"name":  "ssd",
		"local": []interface{}{map[string]interface{}{"path": "/mnt/disks/ssd1"}},
	}})
	if err == nil {
		t.Fatal("expected a local volume to be rejected, got no error")
	}
	if !strings.Contains(err.Error(), "persistent_volume_claim") {
		t.Fatalf("error should point at the supported alternatives, got: %s", err)
	}

	// An empty local block carries no configuration, so it must not trip the
	// check - Terraform hands one over for an unset optional block.
	if _, err := expandPodSpecVolumes([]interface{}{map[string]interface{}{
		"name":  "cache",
		"local": []interface{}{},
		"empty_dir": []interface{}{map[string]interface{}{
			"medium": "Memory",
		}},
	}}); err != nil {
		t.Fatalf("an unset local block should be ignored, got: %s", err)
	}
}
