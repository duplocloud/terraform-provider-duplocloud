package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	// ScaleDownBehaviorTerminateAtInstanceHour is a ScaleDownBehavior enum value
	ScaleDownBehaviorTerminateAtInstanceHour = "TERMINATE_AT_INSTANCE_HOUR"

	// ScaleDownBehaviorTerminateAtTaskCompletion is a ScaleDownBehavior enum value
	ScaleDownBehaviorTerminateAtTaskCompletion = "TERMINATE_AT_TASK_COMPLETION"
)

// Resource for managing an AWS emrCluster
func resourceAwsEmrCluster() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_emr_cluster` manages an AWS emrCluster in Duplo.",

		ReadContext:   resourceAwsEmrClusterRead,
		CreateContext: resourceAwsEmrClusterCreate,
		DeleteContext: resourceAwsEmrClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsEmrClusterSchema(),
	}
}

func awsEmrClusterSchema() map[string]*schema.Schema {
	// todo:
	return map[string]*schema.Schema{

		//local only
		"tenant_id": {
			Description:  "The GUID of the tenant that the emrCluster will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the emrCluster.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		//computed
		"job_flow_id": {
			Description: "job flow id.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
		"full_name": {
			Description: "full_name - Duplo will add a prefix to the name.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
		"arn": {
			Description: "The ARN of the emrCluster.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
		"status": {
			Description: "The status of the emrCluster.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},

		//attr
		"release_label": {
			Description: "EMR ReleaseLabel.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		//not tested this
		"custom_ami_id": {
			Description: "EMR CustomAmiId.",
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
		},
		"log_uri": {
			Description: "S3 bucket path for logs.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Optional:    true,
		},
		"termination_protection": {
			Description: "Emr termination protection setting.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
		},
		"keep_job_flow_alive_when_no_steps": {
			Description: "Keep Job Flow Alive When No Steps. Emr Cluster will be terminated if true.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
			Default:     true,
		},
		"visible_to_all_users": {
			Description: "Emr Cluster visible to all users settings.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
		},
		//not tested this
		"ebs_root_volume_size": {
			Description: "Emr Cluster Ec2 ebs_root_volume_size settings.",
			Type:        schema.TypeInt,
			ForceNew:    true,
			Optional:    true,
		},
		//not tested this
		"step_concurrency_level": {
			Description:  "Emr Cluster step_concurrency_level settings.",
			Type:         schema.TypeInt,
			Optional:     true,
			ForceNew:     true,
			Default:      1,
			ValidateFunc: validation.IntBetween(1, 256),
		},

		//ec2 attributes  hide and use these custom until new awsdk-api
		//ConflictsWith "instance_groups", "instance_fleets"
		"master_instance_type": {
			Description:   "Emr MasterInstanceType. Supported InstanceTypes e.g. m4.large",
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			ConflictsWith: []string{"instance_groups", "instance_fleets"},
		},
		"slave_instance_type": {
			Description:   "Emr SlaveInstanceType. Supported InstanceTypes e.g. m4.large",
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			ConflictsWith: []string{"instance_groups", "instance_fleets"},
		},
		"instance_count": {
			Description:   "Emr Instance Count.",
			Type:          schema.TypeInt,
			Optional:      true,
			ForceNew:      true,
			ConflictsWith: []string{"instance_groups", "instance_fleets"},
		},
		//scale
		"scale_down_behavior": {
			Description: "Emr scale_down_behavior. Specifies the way that individual Amazon EC2 instances terminate when an automatic scale-in activity occurs or an instance group is resized.",
			Type:        schema.TypeString,
			ForceNew:    true,
			Optional:    true,
			ValidateFunc: validation.StringInSlice([]string{
				ScaleDownBehaviorTerminateAtInstanceHour,
				ScaleDownBehaviorTerminateAtTaskCompletion,
			}, false),
		},

		//jsonstr: Not using nested attrbutes as they are absolete for new apis
		// == wait until to creating nested attribute until we move to new AWSSDK.EMR apis to 3.7
		// == 1- the cluster create api does not have update equivalent in current c# client api
		// == 2- any updates needs to handled by sub-api..
		// == 3- As the nested attributes are way different in new api from older api AWSSDK.EMR and official emr tf
		"applications": {
			Description:      "Emr - list of applications to be installed.",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			DiffSuppressFunc: diffIgnoreForEmrApplications,
		},
		"bootstrap_actions": {
			Description:      "Emr - list of bootstrap_actions to be installed.",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			DiffSuppressFunc: diffIgnoreForEmrBootstrap,
		},
		"steps": {
			Description:      "Emr - list of steps to be run after cluster is ready.",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			DiffSuppressFunc: diffIgnoreForEmrSteps,
		},
		"configurations": {
			Description:      "Emr - list of application configurations to be updated.",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			DiffSuppressFunc: diffIgnoreForEmrConfigurations,
		},
		"additional_info": {
			Description:      "Emr - additional_info.",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			DiffSuppressFunc: diffIgnoreForEmrAdditionalInfo,
		},
		"managed_scaling_policy": {
			Description:      "Emr - managed_scaling_policy.",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			DiffSuppressFunc: diffIgnoreForEmrManagedScalingPolicy,
			ConflictsWith:    []string{"instance_fleets"},
		},
		"instance_fleets": {
			Description:      "Emr - instance_fleets.",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			DiffSuppressFunc: diffIgnoreForEmrInstanceFleets,
			ConflictsWith:    []string{"instance_count", "instance_groups", "managed_scaling_policy"},
		},
		"instance_groups": {
			Description:      "Emr - instance_groups.",
			Type:             schema.TypeString,
			Optional:         true,
			ForceNew:         true,
			DiffSuppressFunc: diffIgnoreForEmrInstanceGroups,
			ConflictsWith:    []string{"instance_count", "instance_fleets"},
		},

		// hash to detect changes
		"applications_hash": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"configurations_hash": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"bootstrap_actions_hash": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"steps_hash": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"additional_info_hash": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"managed_scaling_policy_hash": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"instance_fleets_hash": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"instance_groups_hash": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// READ resource
func resourceAwsEmrClusterRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	id := d.Id()
	name := d.Get("name").(string)
	tenantID, jobFlowId, err := parseAwsEmrClusterIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsEmrClusterRead(%s, %s, %s): start", tenantID, jobFlowId, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, clientErr := c.DuploEmrClusterGet(tenantID, jobFlowId)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("") // object missing
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s emrCluster '%s': %s %s", tenantID, jobFlowId, name, clientErr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("full_name", duplo.Name)
	d.Set("job_flow_id", duplo.JobFlowID)
	d.Set("arn", duplo.Arn)
	d.Set("status", duplo.Status)
	d.Set("release_label", duplo.ReleaseLabel)
	d.Set("custom_ami_id", duplo.CustomAmiId)
	d.Set("visible_to_all_users", duplo.VisibleToAllUsers)

	log.Printf("[TRACE] resourceAwsEmrClusterRead(%s, %s, %s): end", tenantID, jobFlowId, name)
	return nil
}

// CREATE resource
func resourceAwsEmrClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error
	var resp *duplosdk.DuploEmrClusterRequest

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsEmrClusterCreate(%s, %s ): start", tenantID, name)

	// track hash
	setEmrHashForKey("applications", d)
	setEmrHashForKey("configurations", d)
	setEmrHashForKey("bootstrap_actions", d)
	setEmrHashForKey("steps", d)
	setEmrHashForKey("additional_info", d)
	setEmrHashForKey("managed_scaling_policy", d)
	setEmrHashForKey("instance_fleets", d)
	setEmrHashForKey("instance_groups", d)

	// Create the request object.
	rq := duplosdk.DuploEmrClusterRequest{
		Name:                        name, // local only
		ReleaseLabel:                d.Get("release_label").(string),
		CustomAmiId:                 d.Get("custom_ami_id").(string),
		LogUri:                      d.Get("log_uri").(string),
		VisibleToAllUsers:           d.Get("visible_to_all_users").(bool),
		KeepJobFlowAliveWhenNoSteps: d.Get("keep_job_flow_alive_when_no_steps").(bool),
		TerminationProtection:       d.Get("termination_protection").(bool),
		EbsRootVolumeSize:           d.Get("ebs_root_volume_size").(int),
		StepConcurrencyLevel:        d.Get("step_concurrency_level").(int),

		//jsonstr
		Applications:         d.Get("applications").(string),
		Configurations:       d.Get("configurations").(string),
		BootstrapActions:     d.Get("bootstrap_actions").(string),
		Steps:                d.Get("steps").(string),
		AdditionalInfo:       d.Get("additional_info").(string),
		ManagedScalingPolicy: d.Get("managed_scaling_policy").(string),
		InstanceFleets:       d.Get("instance_fleets").(string),
		InstanceGroups:       d.Get("instance_groups").(string),
	}

	if d.Get("instance_count").(int) > 1 {
		rq.MasterInstanceType = d.Get("master_instance_type").(string)
		rq.SlaveInstanceType = d.Get("slave_instance_type").(string)
		rq.InstanceCount = d.Get("instance_count").(int)
	}

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	resp, err = c.DuploEmrClusterCreate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s emrCluster '%s': %s", tenantID, name, err)
	}
	if resp == nil {
		return diag.Errorf("Error creating tenant %s emrCluster '%s': %s", tenantID, name, err)
	}

	// Wait for Duplo to be able to return the emr cluster details.
	jobFlowId := resp.JobFlowId
	id := fmt.Sprintf("%s/%s", tenantID, resp.JobFlowId)

	diags := waitForResourceWithStatusDone(ctx, d, "emrCluster-create", id, func() (bool, duplosdk.ClientError) {
		rp, errget := c.DuploEmrClusterGet(tenantID, jobFlowId)
		if rp != nil && rp.Arn != "" && rp.Status != "" {
			if checkEmrStartStatus(rp.Status) {
				return true, errget
			} else {
				return false, errget
			}
		}
		return true, errget
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)
	diags = resourceAwsEmrClusterRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsEmrClusterCreate(%s, %s, %s): end", tenantID, name, jobFlowId)
	return diags
}

// DELETE resource
func resourceAwsEmrClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Parse the identifying attributes
	id := d.Id()
	name := d.Get("name").(string)
	tenantID, jobFlowId, err := parseAwsEmrClusterIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsEmrClusterDelete(%s, %s, %s): start", tenantID, jobFlowId, name)

	// Delete the function.
	c := m.(*duplosdk.Client)
	clientErr := c.DuploEmrClusterDelete(tenantID, jobFlowId)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s emrCluster '%s': %s %s", tenantID, jobFlowId, name, clientErr)
	}

	// Wait up to 60 seconds for Duplo to delete the cluster.
	diagDiagnostics := waitForResourceWithStatusDone(ctx, d, "emrCluster-delete", id, func() (bool, duplosdk.ClientError) {
		rp, errget := c.DuploEmrClusterGet(tenantID, jobFlowId)
		if rp != nil && rp.Arn != "" && rp.Status != "" {
			if checkEmrTerminateStatus(rp.Status) {
				return true, errget
			} else {
				return false, errget
			}
		}
		return true, errget
	})

	if diagDiagnostics != nil {
		return diagDiagnostics
	}

	log.Printf("[TRACE] resourceAwsEmrClusterDelete(%s, %s, %s): end", tenantID, jobFlowId, name)
	return nil
}

func parseAwsEmrClusterIdParts(id string) (tenantID, jobFlowId string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, jobFlowId = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func checkEmrStartStatus(statusStr string) bool {
	//STARTING | BOOTSTRAPPING | RUNNING | WAITING | TERMINATING | TERMINATED | TERMINATED_WITH_ERRORS
	// ok   => BOOTSTRAPPING | RUNNING | WAITING  and  also TERMINATING | TERMINATED | TERMINATED_WITH_ERRORS
	//just wait for to move out of STARTING?
	status := strings.ToLower(statusStr)
	log.Printf("[TRACE] checkEMRStatus create?  %t %s ", status != "starting" && strings.Contains(status, "termin"), status)
	return status == "running" || status == "waiting" || status == "bootstrapping" || strings.Contains(status, "wait") || strings.Contains(status, "termin")
}

func checkEmrTerminateStatus(statusStr string) bool {
	log.Printf("[TRACE] checkEMRStatus delete?  %t %s ", strings.Contains(strings.ToLower(statusStr), "termin"), statusStr)
	return strings.Contains(strings.ToLower(statusStr), "termin")
}

func diffIgnoreForEmrSteps(_, _, _ string, d *schema.ResourceData) bool {
	return diffIgnoreForJsonNonLocalChanges("steps", d)
}

func diffIgnoreForEmrBootstrap(_, _, _ string, d *schema.ResourceData) bool {
	return diffIgnoreForJsonNonLocalChanges("bootstrap_actions", d)
}

func diffIgnoreForEmrApplications(_, _, _ string, d *schema.ResourceData) bool {
	return diffIgnoreForJsonNonLocalChanges("applications", d)
}

func diffIgnoreForEmrConfigurations(_, _, _ string, d *schema.ResourceData) bool {
	return diffIgnoreForJsonNonLocalChanges("configurations", d)
}

func diffIgnoreForEmrAdditionalInfo(_, _, _ string, d *schema.ResourceData) bool {
	return diffIgnoreForJsonNonLocalChanges("additional_info", d)
}

func diffIgnoreForEmrManagedScalingPolicy(_, _, _ string, d *schema.ResourceData) bool {
	return diffIgnoreForJsonNonLocalChanges("managed_scaling_policy", d)
}

func diffIgnoreForEmrInstanceFleets(_, _, _ string, d *schema.ResourceData) bool {
	return diffIgnoreForJsonNonLocalChanges("instance_fleets", d)
}

func diffIgnoreForEmrInstanceGroups(_, _, _ string, d *schema.ResourceData) bool {
	return diffIgnoreForJsonNonLocalChanges("instance_groups", d)
}

func diffIgnoreForJsonNonLocalChanges(key string, d *schema.ResourceData) bool {
	mapFieldName := fmt.Sprintf("%s_hash", key)
	hashFieldName := key
	_, dataNew := d.GetChange(hashFieldName)
	hashOld := d.Get(mapFieldName).(string)
	hashNew := hashForData(dataNew.(string))
	log.Printf("[TRACE] diffIgnoreForJsonNonLocalChanges emr-cluster  %s  ******** 1: hash old vs new %s=%s %t?", key, hashOld, hashNew, hashOld == hashNew)
	return hashOld == hashNew
}

func setEmrHashForKey(key string, d *schema.ResourceData) {
	keyHash := fmt.Sprintf("%s_hash", key)
	value := d.Get(key).(string)
	if value == "" {
		d.Set(keyHash, "0")
	} else {
		hash := hashForData(value)
		d.Set(keyHash, hash)
	}
	log.Printf("[TRACE] diffIgnoreForJsonNonLocalChanges emr-cluster  %s  ******** 1: hash %s", keyHash, value)

}
