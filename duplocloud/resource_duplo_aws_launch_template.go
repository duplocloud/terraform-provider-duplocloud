package duplocloud

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func awsLaunchTemplateSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the launch template will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The fullname of the asg group",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"version": {
			Description: "Any of the existing version of the launch template, if not provided, the latest version will be used",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"default_version": {
			Description: "The current default version of the launch template.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"latest_version": {
			Description: "The latest launch template version",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"version_description": {
			Description: "The version of the launch template",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},

		"instance_type": {
			Description: "Asg instance type to be used to update the version from the current version",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			//	ConflictsWith: []string{"instance_requirements"},
		},
		"ami": {
			Description: "Asg ami to be used to update the version from the current version",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"version_metadata": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"block_device_mapping": {
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Description: "Configure additional volumes of the instance besides specified by the AMI",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"device_name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The name of the device to mount",
					},
					"ebs": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "Configure EBS volume properties.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"delete_on_termination": {
									Type:        schema.TypeBool,
									Optional:    true,
									Default:     true,
									Description: "Whether the volume should be destroyed on instance termination",
								},
								"encrypted": {
									Type:        schema.TypeBool,
									Optional:    true,
									Default:     false,
									Description: "Enables EBS encryption on the volume. Cannot be used with snapshot_id",
								},
								"iops": {
									Type:        schema.TypeInt,
									Optional:    true,
									Description: "The amount of provisioned IOPS. This must be set with a volume_type of 'io1/io2/gp3'",
								},
								"snapshot_id": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "The Snapshot ID to mount. Should not be used if encrypted is true",
								},
								"volume_size": {
									Type:     schema.TypeInt,
									Optional: true,
									Description: `The size of the volume in gigabytes.\n
									gp2 and gp3: 1 - 16,384 GiB\n+
									io1: 4 - 16,384 GiB
									io2: 4 - 65,536 GiB
									st1 and sc1: 125 - 16,384 GiB
									standard: 1 - 1024 GiB`,
								},
								"throughput": {
									Type:         schema.TypeInt,
									Optional:     true,
									Description:  "The throughput to provision for a 'gp3' volume in MiB/s. Minumum value of 125 and maximum of 1000.",
									ValidateFunc: validation.IntBetween(125, 1000),
								},
								"volume_type": {
									Type:         schema.TypeString,
									Optional:     true,
									Description:  "The volume type. Can be one of standard, gp2, gp3, io1, io2, sc1 or st1",
									ValidateFunc: validation.StringInSlice([]string{"standard", "gp2", "gp3", "io1", "io2", "sc1", "st1"}, false),
								},
								"volume_initialization_rate": {
									Type:         schema.TypeInt,
									Optional:     true,
									Description:  "The volume initialization rate in MiB/s, with a minimum of 100 MiB/s and maximum of 300 MiB/s.",
									ValidateFunc: validation.IntBetween(100, 300),
								},
								"kms_key_id": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "The ARN of the KMS Key to use when encrypting the volume (if encrypted is true).",
								},
							},
						},
					},
					"no_device": {
						Type:        schema.TypeString,
						Optional:    true,
						ForceNew:    true,
						Description: "Suppresses the specified device included in the AMI's block device mapping.",
					},
					"virtual_name": {
						Type:        schema.TypeString,
						Optional:    true,
						ForceNew:    true,
						Description: "The virtual device name (ephemeralN). Instance store volumes are numbered starting from 0. An instance type with 2 available instance store volumes can specify mappings for ephemeral0 and ephemeral1. The number of available instance store volumes depends on the instance type. After you connect to the instance, you must mount the volume.",
					},
				},
			},
		},
		//"instance_requirements": {
		//	Description:   "Whether to manage instance requirements instead of a specific instance type",
		//	Type:          schema.TypeList,
		//	MaxItems:      1,
		//	Optional:      true,
		//	ConflictsWith: []string{"instance_type"},
		//	Computed:      true,
		//
		//	Elem: &schema.Resource{
		//		Schema: map[string]*schema.Schema{
		//			"allowed_instance_types": {
		//				Type:     schema.TypeList,
		//				Optional: true,
		//				Computed: true,
		//				Elem: &schema.Schema{
		//					Type: schema.TypeString,
		//				},
		//			},
		//			"vcpu_count": {
		//				Type:         schema.TypeList,
		//				MaxItems:     1,
		//				RequiredWith: []string{"instance_requirements.0.allowed_instance_types"},
		//				Optional:     true,
		//				Computed:     true,
		//				Description:  "Block describing the minimum and maximum number of vCPUs. It is a required field when allowed_instance_types is set ",
		//				Elem: &schema.Resource{
		//					Schema: map[string]*schema.Schema{
		//						"min": {
		//							Type:     schema.TypeInt,
		//							Required: true,
		//						},
		//						"max": {
		//							Type:     schema.TypeInt,
		//							Optional: true,
		//							Computed: true,
		//						},
		//					},
		//				},
		//			},
		//			"memory_mib": {
		//				Type:         schema.TypeList,
		//				MaxItems:     1,
		//				RequiredWith: []string{"instance_requirements.0.allowed_instance_types"},
		//				Optional:     true,
		//				Description:  "Block describing the minimum and maximum amount of memory (MiB). It is a required field when allowed_instance_types is set",
		//				Elem: &schema.Resource{
		//					Schema: map[string]*schema.Schema{
		//						"min": {
		//							Type:     schema.TypeInt,
		//							Required: true,
		//						},
		//						"max": {
		//							Type:     schema.TypeInt,
		//							Optional: true,
		//							Computed: true,
		//						},
		//					},
		//				},
		//			},
		//		},
		//	},
		//},
	}
}
func resourceAwsLaunchTemplate() *schema.Resource {
	return &schema.Resource{
		Description:   "duplocloud_aws_launch_template creates the new version over current launch template version",
		ReadContext:   resourceAwsLaunchTemplateRead,
		CreateContext: resourceAwsLaunchTemplateCreate,
		UpdateContext: resourceAwsLaunchTemplateUpdate,
		DeleteContext: resourceAwsLaunchTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: awsLaunchTemplateSchema(),
	}
}

func resourceAwsLaunchTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	ver := ""
	tenantId, asgName := idParts[0], idParts[2]
	if len(idParts) == 4 {
		ver = idParts[3]
	}

	c := m.(*duplosdk.Client)
	fullName := asgName
	var err1 error
	prefix, err := c.GetResourcePrefixWithoutTenant("duploservices")
	if err != nil {
		return diag.FromErr(err)
	}
	if !strings.Contains(asgName, prefix) {
		fullName, err1 = c.GetResourceName(prefix, tenantId, asgName, false)
		if err1 != nil {
			diag.FromErr(err1)
		}
	}
	rp, err := c.GetAwsLaunchTemplate(tenantId, fullName)
	if err != nil {
		if err.Status() == 404 {
			log.Printf("[TRACE] resourceAwsLaunchTemplateRead(%s, %s): object missing", tenantId, fullName)
			d.SetId("")
		}
		return diag.Errorf("%s", err.Error())

	}
	if rp == nil {
		log.Printf("[TRACE] resourceAwsLaunchTemplateRead(%s, %s): object missing", tenantId, fullName)
		d.SetId("")
		return nil
	}
	d.Set("tenant_id", tenantId)
	fErr := flattenLaunchTemplate(d, rp, ver)
	if fErr != nil {
		return diag.Errorf("%s", fErr.Error())
	}
	return nil
}
func resourceAwsLaunchTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)

	rq, cerr := expandLaunchTemplate(d, c, tenantId, "")
	if cerr != nil {
		return diag.Errorf("%s", cerr.Error())
	}
	name := rq.LaunchTemplateName
	var err duplosdk.ClientError
	if !strings.Contains(name, "duploservices") {
		rq.LaunchTemplateName, err = c.GetResourceName("duploservices", tenantId, name, false)
		if err != nil {
			diag.FromErr(err)
		}
	}
	err = c.CreateAwsLaunchTemplate(tenantId, rq)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	d.SetId(tenantId + "/launch-template/" + rq.LaunchTemplateName)
	diag := resourceAwsLaunchTemplateRead(ctx, d, m)
	return diag

}

func resourceAwsLaunchTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	token := strings.Split(d.Id(), "/")
	tenantId, name := token[0], token[2]
	c := m.(*duplosdk.Client)

	rq, cerr := expandLaunchTemplate(d, c, tenantId, name)
	if cerr != nil {
		return diag.Errorf("%s", cerr.Error())
	}
	var err duplosdk.ClientError
	if !strings.Contains(name, "duploservices") {
		rq.LaunchTemplateName, err = c.GetResourceName("duploservices", tenantId, name, false)
		if err != nil {
			diag.FromErr(err)
		}
	}
	err = c.CreateAwsLaunchTemplate(tenantId, rq)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	diag := resourceAwsLaunchTemplateRead(ctx, d, m)
	return diag

}

func resourceAwsLaunchTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil

}

func expandLaunchTemplate(d *schema.ResourceData, c *duplosdk.Client, tenantId, name string) (*duplosdk.DuploAwsLaunchTemplateRequest, error) {
	sv := d.Get("version").(string)
	if name == "" {
		name = d.Get("name").(string)
	}
	if sv == "" {
		rp, err := c.GetAwsLaunchTemplate(tenantId, name)
		if err != nil {
			return nil, err
		}
		m := extractASGTemplateDetails(rp)
		sv = m["latest_version"].(string)
		log.Printf("Setting the version to latest version %s since source version not provided ", sv)
	}

	obj := &duplosdk.DuploAwsLaunchTemplateRequest{
		LaunchTemplateName: name,
		SourceVersion:      sv,
		VersionDescription: d.Get("version_description").(string),
		LaunchTemplateData: &duplosdk.DuploLaunchTemplateData{
			ImageId:             d.Get("ami").(string),
			BlockDeviceMappings: expandBlockDeviceMappings(d),
		},
	}
	if instanceType, ok := d.GetOk("instance_type"); ok && instanceType.(string) != "" {
		obj.LaunchTemplateData.InstanceType = &duplosdk.DuploStringValue{
			Value: instanceType.(string),
		}
	}
	/*if mir, ok := d.GetOk("instance_requirements"); ok && mir != nil {
		mirMap := mir.([]interface{})[0].(map[string]interface{})
		if ait, ok := mirMap["allowed_instance_types"]; ok && len(ait.([]interface{})) > 0 {
			allowedInstanceList := []string{}

			for _, it := range ait.([]interface{}) {
				allowedInstanceList = append(allowedInstanceList, it.(string))
			}
			obj.LaunchTemplateData.InstanceRequirementsRequest = &duplosdk.InstanceRequirementsRequest{
				AllowedInstanceTypes: allowedInstanceList,
			}

		}
		if vcpu, ok := mirMap["vcpu_count"]; ok && vcpu != nil {
			vcpuMap := vcpu.([]interface{})[0].(map[string]interface{})
			min := vcpuMap["min"].(int)
			max := vcpuMap["max"].(int)
			obj.LaunchTemplateData.InstanceRequirementsRequest.VCpuCount = &duplosdk.DuploLaunchTemplateVCpuCountRequest{
				Min: min,
				Max: max,
			}
		}
		if memMap, ok := mirMap["memory_mib"]; ok && memMap != nil {
			mMap := memMap.([]interface{})[0].(map[string]interface{})
			min := mMap["min"].(int)
			max := mMap["max"].(int)
			obj.LaunchTemplateData.InstanceRequirementsRequest.MemoryMiB = &duplosdk.DuploLaunchTemplateMemoryMiB{
				Min: min,
				Max: max,
			}
		}
	}*/
	return obj, nil

}

func expandBlockDeviceMappings(d *schema.ResourceData) []duplosdk.DuploLaunchTemplateBlockDeviceMappingRequest {
	var blockDeviceMappings []duplosdk.DuploLaunchTemplateBlockDeviceMappingRequest
	if v, ok := d.GetOk("block_device_mapping"); ok {
		for _, bdm := range v.([]interface{}) {
			bdmMap := bdm.(map[string]interface{})
			bdmObj := duplosdk.DuploLaunchTemplateBlockDeviceMappingRequest{
				DeviceName: bdmMap["device_name"].(string),
			}
			if ebsList, ok := bdmMap["ebs"].([]interface{}); ok && len(ebsList) > 0 {
				ebsMap := ebsList[0].(map[string]interface{})
				ebsObj := duplosdk.DuploLaunchTemplateEbsBlockDeviceRequest{
					DeleteOnTermination: ebsMap["delete_on_termination"].(bool),
					Encrypted:           ebsMap["encrypted"].(bool),
					VolumeSize:          ebsMap["volume_size"].(int),
					Throughput:          ebsMap["throughput"].(int),
				}
				if v, ok := ebsMap["iops"].(int); ok && v != 0 {
					ebsObj.Iops = v
				}
				if v, ok := ebsMap["snapshot_id"].(string); ok && v != "" {
					ebsObj.SnapshotId = v
				}
				if v, ok := ebsMap["volume_type"].(string); ok && v != "" {
					ebsObj.VolumeType = v
				}
				if v, ok := ebsMap["kms_key_id"].(string); ok && v != "" {
					ebsObj.KmsKeyId = v
				}
				if v, ok := bdmMap["volume_initialization_rate"].(int); ok && v != 0 {
					ebsObj.VolumeInitializationRate = v
				}
				bdmObj.Ebs = &ebsObj
			}
			if v, ok := bdmMap["no_device"].(string); ok && v != "" {
				bdmObj.NoDevice = v
			}
			if v, ok := bdmMap["virtual_name"].(string); ok && v != "" {
				bdmObj.VirtualName = v
			}

			blockDeviceMappings = append(blockDeviceMappings, bdmObj)
		}
	}
	return blockDeviceMappings
}
func flattenLaunchTemplate(d *schema.ResourceData, rp *[]duplosdk.DuploLaunchTemplateResponse, ver string) error {

	b, err := json.Marshal(rp)
	if err != nil {
		return err
	}
	m := extractASGTemplateDetails(rp)
	d.Set("version_metadata", string(b))
	d.Set("instance_type", m["instance_type"])
	d.Set("version_description", m["ver_desc"])
	n := d.Get("name").(string)
	d.Set("name", m["name"])
	if n != "" && !strings.Contains(n, "duploservices") {
		d.Set("name", n)
	}

	if v, ok := d.GetOk("version"); ok && v.(string) != "" {
		d.Set("version", v.(string))
	} else {
		d.Set("version", m["latest_version"])
	}
	d.Set("latest_version", m["latest_version"])
	d.Set("default_version", m["default_version"])
	d.Set("ami", m["image_id"])
	d.Set("block_device_mapping", m["block_device_mapping"])
	//d.Set("instance_requirements", m["instance_requirements"])
	return nil
}

func extractASGTemplateDetails(rp *[]duplosdk.DuploLaunchTemplateResponse) map[string]interface{} {
	max := 0
	lt := map[string]interface{}{}
	for _, v := range *rp {

		if v.DefaultVersion {
			lt["default_version"] = strconv.Itoa(int(v.VersionNumber))
		}
		if max < int(v.VersionNumber) {
			max = int(v.VersionNumber)
			lt["instance_type"] = v.LaunchTemplateData.InstanceType.Value
			lt["ver_desc"] = v.VersionDescription
			lt["image_id"] = v.LaunchTemplateData.ImageId
			lt["name"] = v.LaunchTemplateName
			lt["block_device_mapping"] = flattenBlockDeviceMappings(v.LaunchTemplateData.BlockDeviceMappings)
			//lt["instance_requirements"] = flattenInstanceRequirements(v.LaunchTemplateData.InstanceRequirements)
		}
	}
	lt["latest_version"] = strconv.Itoa(max)
	return lt
}

/*
	func flattenInstanceRequirements(ir *duplosdk.DuploLaunchTemplateInstanceRequirements) []interface{} {
		if ir == nil {
			return []interface{}{}
		}
		irMap := map[string]interface{}{}
		allowedInstanceTypes := []interface{}{}
		for _, ait := range ir.AllowedInstanceTypes {
			allowedInstanceTypes = append(allowedInstanceTypes, ait)
		}
		irMap["allowed_instance_types"] = allowedInstanceTypes
		if ir.VCpuCount != nil {
			vcpuMap := map[string]interface{}{
				"min": ir.VCpuCount.Min,
				"max": ir.VCpuCount.Max,
			}
			irMap["vcpu_count"] = []interface{}{vcpuMap}
		}
		if ir.MemoryMiB != nil {
			memMap := map[string]interface{}{
				"min": ir.MemoryMiB.Min,
				"max": ir.MemoryMiB.Max,
			}
			irMap["memory_mib"] = []interface{}{memMap}
		}
		return []interface{}{irMap}
	}
*/
func flattenBlockDeviceMappings(bdms []duplosdk.DuploLaunchTemplateBlockDeviceMappingResponse) []interface{} {
	bdmI := []interface{}{}
	for _, bdm := range bdms {
		bdmMap := map[string]interface{}{
			"device_name": bdm.DeviceName,
		}
		if bdm.Ebs != nil {
			ebsMap := map[string]interface{}{
				"delete_on_termination":      bdm.Ebs.DeleteOnTermination,
				"encrypted":                  bdm.Ebs.Encrypted,
				"volume_size":                bdm.Ebs.VolumeSize,
				"volume_type":                bdm.Ebs.VolumeType.Value,
				"kms_key_id":                 bdm.Ebs.KmsKeyId,
				"iops":                       bdm.Ebs.Iops,
				"throughput":                 bdm.Ebs.Throughput,
				"volume_initialization_rate": bdm.Ebs.VolumeInitializationRate,
			}
			if bdm.Ebs.SnapshotId != "" {
				ebsMap["snapshot_id"] = bdm.Ebs.SnapshotId
			}
			bdmMap["ebs"] = []interface{}{ebsMap}
		}
		if bdm.NoDevice != "" {
			bdmMap["no_device"] = bdm.NoDevice
		}
		if bdm.VirtualName != "" {
			bdmMap["virtual_name"] = bdm.VirtualName
		}
		bdmI = append(bdmI, bdmMap)
	}
	return bdmI
}
