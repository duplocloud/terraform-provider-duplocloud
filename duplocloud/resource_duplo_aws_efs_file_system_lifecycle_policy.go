package duplocloud

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	// TransitionToIARulesAfter7Days is a TransitionToIARules enum value
	TransitionToIARulesAfter7Days = "AFTER_7_DAYS"

	// TransitionToIARulesAfter14Days is a TransitionToIARules enum value
	TransitionToIARulesAfter14Days = "AFTER_14_DAYS"

	// TransitionToIARulesAfter30Days is a TransitionToIARules enum value
	TransitionToIARulesAfter30Days = "AFTER_30_DAYS"

	// TransitionToIARulesAfter60Days is a TransitionToIARules enum value
	TransitionToIARulesAfter60Days = "AFTER_60_DAYS"

	// TransitionToIARulesAfter90Days is a TransitionToIARules enum value
	TransitionToIARulesAfter90Days = "AFTER_90_DAYS"

	// TransitionToIARulesAfter1Day is a TransitionToIARules enum value
	TransitionToIARulesAfter1Day = "AFTER_1_DAY"

	// TransitionToIARulesAfter180Days is a TransitionToIARules enum value
	TransitionToIARulesAfter180Days = "AFTER_180_DAYS"

	// TransitionToIARulesAfter270Days is a TransitionToIARules enum value
	TransitionToIARulesAfter270Days = "AFTER_270_DAYS"

	// TransitionToIARulesAfter365Days is a TransitionToIARules enum value
	TransitionToIARulesAfter365Days = "AFTER_365_DAYS"
)

// TransitionToIARules_Values returns all elements of the TransitionToIARules enum
func TransitionToIARules_Values() []string {
	return []string{
		TransitionToIARulesAfter7Days,
		TransitionToIARulesAfter14Days,
		TransitionToIARulesAfter30Days,
		TransitionToIARulesAfter60Days,
		TransitionToIARulesAfter90Days,
		TransitionToIARulesAfter1Day,
		TransitionToIARulesAfter180Days,
		TransitionToIARulesAfter270Days,
		TransitionToIARulesAfter365Days,
	}
}

const (
	// TransitionToPrimaryStorageClassRulesAfter1Access is a TransitionToPrimaryStorageClassRules enum value
	TransitionToPrimaryStorageClassRulesAfter1Access = "AFTER_1_ACCESS"
)

// TransitionToPrimaryStorageClassRules_Values returns all elements of the TransitionToPrimaryStorageClassRules enum
func TransitionToPrimaryStorageClassRules_Values() []string {
	return []string{
		TransitionToPrimaryStorageClassRulesAfter1Access,
	}
}

func awsEFSLifecyclePolicy() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"file_system_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"lifecycle_policy": {
			Type:     schema.TypeList,
			Required: true,
			MaxItems: 2,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"transition_to_ia": {
						Description:  "Indicates how long it takes to transition files to the IA storage class. Valid values: `AFTER_1_DAY`, `AFTER_7_DAYS`, `AFTER_14_DAYS`, `AFTER_30_DAYS`, `AFTER_60_DAYS`, or `AFTER_90_DAYS`",
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice(TransitionToIARules_Values(), false),
					},
					"transition_to_primary_storage_class": {
						Description:  "Describes the policy used to transition a file from infequent access storage to primary storage. Valid values: `AFTER_1_ACCESS`",
						Type:         schema.TypeString,
						Optional:     true,
						ValidateFunc: validation.StringInSlice(TransitionToPrimaryStorageClassRules_Values(), false),
					},
				},
			},
		},
	}
}

func resourceAwsEFSLifecyclePolicy() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_efs_lifecycle_policy` Provides an Elastic File System (EFS) File System Lifecycle Policy resource in DuploCloud.",

		ReadContext:   resourceFileSystemPolicyRead,
		CreateContext: resourceFileSystemPolicyPut,
		UpdateContext: resourceFileSystemPolicyPut,
		DeleteContext: resourceFileSystemPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsEFSLifecyclePolicy(),
	}
}

func resourceFileSystemPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*duplosdk.Client)

	input := &duplosdk.PutLifecycleConfigurationInput{
		FileSystemId:      d.Get("file_system_id").(string),
		LifecyclePolicies: expandFileSystemLifecyclePolicies(d.Get("lifecycle_policy").([]interface{})),
	}

	tenantID := d.Get("tenant_id").(string)
	_, err := c.DuploAwsLifecyclePolicyUpdate(tenantID, input)

	if err != nil {
		return diag.Errorf("putting EFS File System Policy (%s): %s", input.FileSystemId, err)
	}
	id := fmt.Sprintf("%s/%s", tenantID, input.FileSystemId)
	if d.IsNewResource() {
		d.SetId(id)
	}

	return nil
}

func resourceFileSystemPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*duplosdk.Client)

	id := d.Id()
	tenantID, fsID, err := parseAwsEFSLifecyclesIdParts(id)
	if err != nil {
		return diag.Errorf("reading EFS File System Lifecycle Policy (%s): %s", d.Id(), err)
	}

	output, err := c.DuploAWsLifecyclePolicyGet(tenantID, fsID)

	if err != nil {
		return diag.Errorf("reading EFS File System Lifecycle Policy (%s): %s", d.Id(), err)
	}

	d.Set("lifecycle_policy", output)

	return diags
}

func resourceFileSystemPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func expandFileSystemLifecyclePolicies(tfList []interface{}) []*duplosdk.LifecyclePolicy {
	var apiObjects []*duplosdk.LifecyclePolicy

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := &duplosdk.LifecyclePolicy{}

		if v, ok := tfMap["transition_to_ia"].(string); ok && v != "" {
			apiObject.TransitionToIA = &duplosdk.DuploStringValue{Value: v}
		}

		if v, ok := tfMap["transition_to_primary_storage_class"].(string); ok && v != "" {
			apiObject.TransitionToPrimaryStorageClass = &duplosdk.DuploStringValue{Value: v}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func parseAwsEFSLifecyclesIdParts(id string) (tenantID, efsId string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, efsId = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
