package duplocloud

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var intOrPercentRegexp = regexp.MustCompile(`^\d+%?$`)

func daemonSetSpecFields() map[string]*schema.Schema {
	podTemplateFields := map[string]*schema.Schema{
		"metadata": metadataSchema("daemonset", true),
		"spec": {
			Type:        schema.TypeList,
			Description: "Spec of the pods managed by the daemonset",
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: podSpecFields(true, false),
			},
		},
	}

	// DaemonSets only allow restart_policy=Always. The shared podSpecFields defaults to
	// OnFailure (which Job/CronJob require) and its validator excludes Always, so override
	// the default and widen the validator to accept Always for this resource only.
	podTemplateSpecSchema := podTemplateFields["spec"].Elem.(*schema.Resource)
	restartPolicy := podTemplateSpecSchema.Schema["restart_policy"]
	restartPolicy.Default = string(corev1.RestartPolicyAlways)
	restartPolicy.ValidateFunc = validation.StringInSlice([]string{
		string(corev1.RestartPolicyAlways),
	}, false)

	return map[string]*schema.Schema{
		"min_ready_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      0,
			ValidateFunc: validateNonNegativeInteger,
			Description:  "The minimum number of seconds for which a newly created pod should be ready without any of its containers crashing, for it to be considered available.",
		},
		"revision_history_limit": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      10,
			ValidateFunc: validateNonNegativeInteger,
			Description:  "The number of old history to retain to allow rollback.",
		},
		"selector": {
			Type:        schema.TypeList,
			Description: "A label query over pods that are managed by the daemon set. Must match in order to be controlled. It must match the pod template's labels.",
			Required:    true,
			ForceNew:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: labelSelectorFields(false),
			},
		},
		"template": {
			Type:        schema.TypeList,
			Description: "An object that describes the pod that will be created. The DaemonSet will create exactly one copy of this pod on every node that matches the template's node selector.",
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: podTemplateFields,
			},
		},
		"update_strategy": {
			Type:        schema.TypeList,
			Description: "An update strategy to replace existing DaemonSet pods with new pods.",
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  string(appsv1.RollingUpdateDaemonSetStrategyType),
						ValidateFunc: validation.StringInSlice([]string{
							string(appsv1.RollingUpdateDaemonSetStrategyType),
							string(appsv1.OnDeleteDaemonSetStrategyType),
						}, false),
						Description: "Type of daemon set update. Can be `RollingUpdate` or `OnDelete`. Default is `RollingUpdate`.",
					},
					"rolling_update": {
						Type:        schema.TypeList,
						Description: "Rolling update config params. Present only if type = RollingUpdate.",
						Optional:    true,
						Computed:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"max_unavailable": {
									Type:         schema.TypeString,
									Optional:     true,
									Default:      "1",
									ValidateFunc: validation.StringMatch(intOrPercentRegexp, "must be a non-negative integer (e.g. 5) or a percentage (e.g. 10%)"),
									Description:  "The maximum number of DaemonSet pods that can be unavailable during the update. Value can be an absolute number (ex: 5) or a percentage of total number of DaemonSet pods at the start of the update (ex: 10%). Default is 1.",
								},
								"max_surge": {
									Type:         schema.TypeString,
									Optional:     true,
									Default:      "0",
									ValidateFunc: validation.StringMatch(intOrPercentRegexp, "must be a non-negative integer (e.g. 5) or a percentage (e.g. 10%)"),
									Description:  "The maximum number of nodes with an existing available DaemonSet pod that can have an updated DaemonSet pod during an update. Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%). Default is 0.",
								},
							},
						},
					},
				},
			},
		},
	}
}
