package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/ucarion/jcs"
)

func duploAwsBatchJobDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the aws batch Job Definition will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Specifies the name of the Job Definition.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the Job Definition.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status": {
			Description: "The status of the Job Definition.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"container_properties": {
			Description: "A valid container properties provided as a single valid JSON document. This parameter is required if the type parameter is `container`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			StateFunc: func(v interface{}) string {
				log.Printf("[TRACE] duplocloud_aws_batch_job_definition.container_properties.StateFunc: <= %v", v)
				props, _ := expandJobContainerProperties(v.(string))
				reorderJobContainerPropertiesEnvironmentVariables(props)
				json, err := jcs.Format(props)
				if json == "{}" {
					json = ""
				}
				log.Printf("[TRACE] duplocloud_aws_batch_job_definition.container_properties.StateFunc: => %s (error: %s)", json, err)
				return json
			},
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				equal, _ := otherJobContainerPropertiesAreEquivalent(old, new)
				return equal
			},
			ValidateFunc: validJobContainerProperties,
		},
		"parameters": {
			Description: "Specifies the parameter substitution placeholders to set in the job definition.",
			Type:        schema.TypeMap,
			Optional:    true,
			ForceNew:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"platform_capabilities": {
			Description: "The platform capabilities required by the job definition. If no value is specified, it defaults to `EC2`. To run the job on Fargate resources, specify `FARGATE`.",
			Type:        schema.TypeSet,
			Optional:    true,
			ForceNew:    true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice([]string{"EC2", "FARGATE"}, false),
			},
		},
		"retry_strategy": {
			Description: "Specifies the retry strategy to use for failed jobs that are submitted with this job definition. Maximum number of `retry_strategy` is `1`.",
			Type:        schema.TypeList,
			Optional:    true,
			ForceNew:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"attempts": {
						Description:  "The number of times to move a job to the RUNNABLE status. You may specify between `1` and `10` attempts.",
						Type:         schema.TypeInt,
						Optional:     true,
						ForceNew:     true,
						ValidateFunc: validation.IntBetween(1, 10),
					},
					"evaluate_on_exit": {
						Description: "The evaluate on exit conditions under which the job should be retried or failed. If this parameter is specified, then the attempts parameter must also be specified. You may specify up to 5 configuration blocks.",
						Type:        schema.TypeList,
						Optional:    true,
						ForceNew:    true,
						MinItems:    0,
						MaxItems:    5,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"action": {
									Description: "Specifies the action to take if all of the specified conditions are met. The values are not case sensitive. Valid values: `RETRY`, `EXIT`.",
									Type:        schema.TypeString,
									Required:    true,
									ForceNew:    true,
									StateFunc: func(v interface{}) string {
										return strings.ToLower(v.(string))
									},
									ValidateFunc: validation.StringInSlice([]string{"RETRY", "EXIT"}, true),
								},
								"on_exit_code": {
									Description: "A glob pattern to match against the decimal representation of the exit code returned for a job.",
									Type:        schema.TypeString,
									Optional:    true,
									ForceNew:    true,
									ValidateFunc: validation.All(
										validation.StringLenBetween(1, 512),
										validation.StringMatch(regexp.MustCompile(`^[0-9]*\*?$`), "must contain only numbers, and can optionally end with an asterisk"),
									),
								},
								"on_reason": {
									Description: "A glob pattern to match against the reason returned for a job.",
									Type:        schema.TypeString,
									Optional:    true,
									ForceNew:    true,
									ValidateFunc: validation.All(
										validation.StringLenBetween(1, 512),
										validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\.:\s]*\*?$`), "must contain letters, numbers, periods, colons, and white space, and can optionally end with an asterisk"),
									),
								},
								"on_status_reason": {
									Description: "A glob pattern to match against the status reason returned for a job.",
									Type:        schema.TypeString,
									Optional:    true,
									ForceNew:    true,
									ValidateFunc: validation.All(
										validation.StringLenBetween(1, 512),
										validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\.:\s]*\*?$`), "must contain letters, numbers, periods, colons, and white space, and can optionally end with an asterisk"),
									),
								},
							},
						},
					},
				},
			},
		},
		"arn": {
			Description: "The Amazon Resource Name of the Job Definition.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"timeout": {
			Description: "Specifies the timeout for jobs so that if a job runs longer, AWS Batch terminates the job. Maximum number of `timeout` is `1`.",
			Type:        schema.TypeList,
			Optional:    true,
			ForceNew:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"attempt_duration_seconds": {
						Description:  "The time duration in seconds after which AWS Batch terminates your jobs if they have not finished. The minimum value for the timeout is `60`seconds.",
						Type:         schema.TypeInt,
						Optional:     true,
						ForceNew:     true,
						ValidateFunc: validation.IntAtLeast(60),
					},
				},
			},
		},
		"type": {
			Description:  "The `type` of job definition. Must be `container`.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"container"}, true),
		},
		"revision": {
			Description: "The revision of the job definition.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"tags": {
			Description: "Key-value map of resource tags.",
			Type:        schema.TypeMap,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Computed:    true,
		},
	}
}

func resourceAwsBatchJobDefinition() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_batch_job_definition` manages an aws batch Job Definition in Duplo.",

		ReadContext:   resourceAwsBatchJobDefinitionRead,
		CreateContext: resourceAwsBatchJobDefinitionCreate,
		UpdateContext: resourceAwsBatchJobDefinitionUpdate,
		DeleteContext: resourceAwsBatchJobDefinitionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsBatchJobDefinitionSchema(),
	}
}

func resourceAwsBatchJobDefinitionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsBatchJobDefinitionIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsBatchJobDefinitionRead(%s, %s): start", tenantID, name)

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	jq, clientErr := c.AwsBatchJobDefinitionGet(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(clientErr)
	}
	if jq == nil {
		d.SetId("") // object missing
		return nil
	}
	flattenBatchJobDefinition(d, c, jq, tenantID)

	log.Printf("[TRACE] resourceAwsBatchJobDefinitionRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsBatchJobDefinitionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsBatchJobDefinitionCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	rq := expandAwsBatchJobDefinition(d)
	err = c.AwsBatchJobDefinitionCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s aws batch Job Definition '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws batch Job Definition", id, func() (interface{}, duplosdk.ClientError) {
		return c.AwsBatchJobDefinitionGet(tenantID, fullName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsBatchJobDefinitionRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsBatchJobDefinitionCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsBatchJobDefinitionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAwsBatchJobDefinitionCreate(ctx, d, m)
}

func resourceAwsBatchJobDefinitionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsBatchJobDefinitionIdParts(id)
	//revision := d.Get("revision").(int)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsBatchJobDefinitionDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)

	jq, err := c.AwsBatchJobDefinitionGetAllRevisions(tenantID, fullName)
	if err != nil {
		return diag.FromErr(err)
	}
	if jq == nil {
		return nil
	}
	for _, data := range *jq {
		clientErr := c.AwsBatchJobDefinitionDelete(tenantID, fullName+":"+strconv.Itoa(data.Revision))
		if clientErr != nil {
			if clientErr.Status() == 404 {
				continue
			}
			return diag.Errorf("Unable to delete tenant %s aws batch Job Definition '%s': %s", tenantID, name, clientErr)
		}
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "aws batch Job Definition", id, func() (interface{}, duplosdk.ClientError) {
		return c.AwsBatchJobDefinitionGet(tenantID, fullName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsBatchJobDefinitionDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAwsBatchJobDefinition(d *schema.ResourceData) *duplosdk.DuploAwsBatchJobDefinition {

	input := &duplosdk.DuploAwsBatchJobDefinition{
		JobDefinitionName: d.Get("name").(string),
		Type:              &duplosdk.DuploStringValue{Value: d.Get("type").(string)},
	}

	if v, ok := d.GetOk("container_properties"); ok {
		props, _ := expandJobContainerProperties(v.(string))
		input.ContainerProperties = props
	}

	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandJobDefinitionParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("platform_capabilities"); ok && v.(*schema.Set).Len() > 0 {
		input.PlatformCapabilities = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("retry_strategy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RetryStrategy = expandRetryStrategy(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("tags"); ok && v != nil {
		input.Tags = expandAsStringMap("tags", d)
	}

	if v, ok := d.GetOk("timeout"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Timeout = expandJobTimeout(v.([]interface{})[0].(map[string]interface{}))
	}

	return input
}

func parseAwsBatchJobDefinitionIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenBatchJobDefinition(d *schema.ResourceData, c *duplosdk.Client, duplo *duplosdk.DuploAwsBatchJobDefinitionResp, tenantId string) diag.Diagnostics {
	log.Printf("[TRACE]flattenBatchJobDefinition... Start ")
	prefix, err := c.GetDuploServicesPrefix(tenantId)
	if err != nil {
		return diag.FromErr(err)
	}
	name, _ := duplosdk.UnprefixName(prefix, duplo.JobDefinitionName)
	d.Set("tenant_id", tenantId)
	d.Set("name", name)
	d.Set("arn", duplo.JobDefinitionArn)
	d.Set("fullname", duplo.JobDefinitionName)
	d.Set("status", duplo.Status)
	d.Set("tags", duplo.Tags)
	d.Set("revision", duplo.Revision)
	d.Set("type", duplo.Type)

	flattenContainerProperties("container_properties", duplo.ContainerProperties, d)

	d.Set("parameters", duplo.Parameters)
	d.Set("platform_capabilities", duplo.PlatformCapabilities)

	if duplo.RetryStrategy != nil {
		if err := d.Set("retry_strategy", []interface{}{flattenRetryStrategy(duplo.RetryStrategy)}); err != nil {
			return diag.FromErr(err)
		}
	} else {
		d.Set("retry_strategy", nil)
	}

	if duplo.Timeout != nil {
		if err := d.Set("timeout", []interface{}{flattenJobTimeout(duplo.Timeout)}); err != nil {
			return diag.FromErr(err)
		}
	} else {
		d.Set("timeout", nil)
	}
	log.Printf("[TRACE]flattenBatchJobDefinition... End ")
	return nil
}

func validJobContainerProperties(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := expandJobContainerProperties(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("AWS Batch Job container_properties is invalid: %s", err))
	}
	return
}

func expandJobContainerProperties(rawProps string) (props map[string]interface{}, err error) {
	err = json.Unmarshal([]byte(rawProps), &props)
	log.Printf("[DEBUG] Expanded duplocloud_aws_batch_job_definition.container_properties: %v", props)
	return
}

func expandJobDefinitionParameters(params map[string]interface{}) map[string]string {
	var jobParams = make(map[string]string)
	for k, v := range params {
		jobParams[k] = v.(string)
	}

	return jobParams
}

func expandRetryStrategy(tfMap map[string]interface{}) *duplosdk.DuploAwsBatchJobDefinitionRetryStrategy {
	if tfMap == nil {
		return nil
	}

	apiObject := &duplosdk.DuploAwsBatchJobDefinitionRetryStrategy{}

	if v, ok := tfMap["attempts"].(int); ok && v != 0 {
		apiObject.Attempts = v
	}

	if v, ok := tfMap["evaluate_on_exit"].([]interface{}); ok && len(v) > 0 {
		apiObject.EvaluateOnExit = expandEvaluateOnExits(v)
	}

	return apiObject
}

func expandEvaluateOnExits(tfList []interface{}) *[]duplosdk.DuploAwsBatchJobDefinitionEvaluateOnExit {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []duplosdk.DuploAwsBatchJobDefinitionEvaluateOnExit

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEvaluateOnExit(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return &apiObjects
}

func expandEvaluateOnExit(tfMap map[string]interface{}) duplosdk.DuploAwsBatchJobDefinitionEvaluateOnExit {

	apiObject := duplosdk.DuploAwsBatchJobDefinitionEvaluateOnExit{}

	if v, ok := tfMap["action"].(string); ok && v != "" {
		apiObject.Action = &duplosdk.DuploStringValue{Value: v}
	}

	if v, ok := tfMap["on_exit_code"].(string); ok && v != "" {
		apiObject.OnExitCode = v
	}

	if v, ok := tfMap["on_reason"].(string); ok && v != "" {
		apiObject.OnReason = v
	}

	if v, ok := tfMap["on_status_reason"].(string); ok && v != "" {
		apiObject.OnStatusReason = v
	}

	return apiObject
}

func expandJobTimeout(tfMap map[string]interface{}) *duplosdk.DuploAwsBatchJobDefinitionTimeout {
	if tfMap == nil {
		return nil
	}

	apiObject := &duplosdk.DuploAwsBatchJobDefinitionTimeout{}

	if v, ok := tfMap["attempt_duration_seconds"].(int); ok && v != 0 {
		apiObject.AttemptDurationSeconds = v
	}

	return apiObject
}

func flattenContainerProperties(field string, from interface{}, to *schema.ResourceData) {
	var err error
	var encoded []byte

	if encoded, err = json.Marshal(from); err == nil {
		err = to.Set(field, string(encoded))
	}

	if err != nil {
		log.Printf("[DEBUG] flattenContainerProperties: failed to serialize %s to JSON: %s", field, err)
	}
}

func flattenRetryStrategy(apiObject *duplosdk.DuploAwsBatchJobDefinitionRetryStrategy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Attempts; v != 0 {
		tfMap["attempts"] = v
	}

	if v := apiObject.EvaluateOnExit; v != nil {
		tfMap["evaluate_on_exit"] = flattenEvaluateOnExits(v)
	}

	return tfMap
}

func flattenEvaluateOnExit(apiObject duplosdk.DuploAwsBatchJobDefinitionEvaluateOnExit) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Action; v != nil && v.Value != "" {
		tfMap["action"] = v.Value
	}

	if v := apiObject.OnExitCode; v != "" {
		tfMap["on_exit_code"] = v
	}

	if v := apiObject.OnReason; v != "" {
		tfMap["on_reason"] = v
	}

	if v := apiObject.OnStatusReason; v != "" {
		tfMap["on_status_reason"] = v
	}

	return tfMap
}

func flattenEvaluateOnExits(apiObjects *[]duplosdk.DuploAwsBatchJobDefinitionEvaluateOnExit) []interface{} {
	if len(*apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range *apiObjects {
		tfList = append(tfList, flattenEvaluateOnExit(apiObject))
	}

	return tfList
}

func flattenJobTimeout(apiObject *duplosdk.DuploAwsBatchJobDefinitionTimeout) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AttemptDurationSeconds; v != 0 {
		tfMap["attempt_duration_seconds"] = v
	}

	return tfMap
}

func reduceJobContainerProperties(props map[string]interface{}) error {
	makeMapUpperCamelCase(props)

	reorderJobContainerPropertiesEnvironmentVariables(props)

	reduceNilOrEmptyMapEntries(props)

	// Handle fields that have defaults.
	if v, ok := props["ReadonlyRootFilesystem"]; ok || v != nil && !v.(bool) {
		delete(props, "ReadonlyRootFilesystem")
	}

	if v, ok := props["Privileged"]; ok || v != nil && !v.(bool) {
		delete(props, "Privileged")
	}

	if v, ok := props["Memory"]; ok || v != nil && v.(int) == 0 {
		delete(props, "Memory")
	}
	if v, ok := props["Vcpus"]; ok || v != nil && v.(int) == 0 {
		delete(props, "Vcpus")
	}
	delete(props, "JobRoleArn")

	return nil
}

func canonicalizeJobContainerPropertiesJson(encoded string) (string, error) {
	var props interface{}

	// Unmarshall, reduce, then canonicalize.
	err := json.Unmarshal([]byte(encoded), &props)
	if err != nil {
		return encoded, err
	}
	err = reduceJobContainerProperties(props.(map[string]interface{}))
	if err != nil {
		return encoded, err
	}
	canonical, err := jcs.Format(props)
	if err != nil {
		return encoded, err
	}
	if canonical == "{}" {
		canonical = ""
	}

	return canonical, nil
}

// An internal function that compares two container_properties values to see if they are equivalent.
func otherJobContainerPropertiesAreEquivalent(old, new string) (bool, error) {

	oldCanonical, err := canonicalizeJobContainerPropertiesJson(old)
	if err != nil {
		return false, err
	}
	log.Printf("[TRACE] Canonical Old Container Properties: <= %s", oldCanonical)
	newCanonical, err := canonicalizeJobContainerPropertiesJson(new)
	if err != nil {
		return false, err
	}
	log.Printf("[TRACE] Canonical New Container Properties: <= %s", newCanonical)
	equal := oldCanonical == newCanonical
	if !equal {
		log.Printf("[DEBUG] Canonical container properties are not equal.\nFirst: %s\nSecond: %s\n", oldCanonical, newCanonical)
	}
	return equal, nil
}

func reorderJobContainerPropertiesEnvironmentVariables(defn map[string]interface{}) {

	// Re-order environment variables to a canonical order.
	if v, ok := defn["Environment"]; ok && v != nil {
		if env, ok := v.([]interface{}); ok && env != nil {
			sort.SliceStable(env, func(i, j int) bool {

				// Get both maps, ensure we are using upper camel-case.
				mi := env[i].(map[string]interface{})
				mj := env[j].(map[string]interface{})
				makeMapUpperCamelCase(mi)
				makeMapUpperCamelCase(mj)

				// Get both name keys, fall back on an empty string.
				si := ""
				sj := ""
				if v, ok = mi["Name"]; ok && !isInterfaceNil(v) {
					si = v.(string)
				}
				if v, ok = mj["Name"]; ok && !isInterfaceNil(v) {
					sj = v.(string)
				}

				// Compare the two.
				return si < sj
			})
		}
	}
}
