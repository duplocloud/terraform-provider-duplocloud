package duplocloud

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"log"
	"regexp"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKubernetesJobV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKubernetesJobV1Create,
		ReadContext:   resourceKubernetesJobV1Read,
		UpdateContext: resourceKubernetesJobV1Update,
		DeleteContext: resourceKubernetesJobV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: resourceKubernetesJobV1Schema(false),
	}
}

func resourceKubernetesJobV1Schema(readonly bool) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the job will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"metadata": jobMetadataSchema(),
		"spec": {
			Type:        schema.TypeList,
			Description: "Spec of the job owned by the cluster",
			Optional:    !readonly,
			Computed:    readonly,
			ForceNew:    false,
			Elem: &schema.Resource{
				Schema: jobSpecFields(false),
			},
		},
		"wait_for_completion": {
			Type:     schema.TypeBool,
			Optional: !readonly,
			Computed: readonly,
		},
	}
}

func resourceKubernetesJobV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceKubernetesJobV1Create(%s): start", tenantId)

	name, err := getK8sJobName(d)
	if err != nil {
		return diag.FromErr(err)
	}

	metadata := expandMetadata(d.Get("metadata").([]interface{}))

	spec, err := expandJobV1Spec(d.Get("spec").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	var rq duplosdk.DuploK8sJob
	rq.Metadata = metadata
	rq.Spec = spec
	rq.TenantId = tenantId

	// initiate create Job
	c := meta.(*duplosdk.Client)
	err = c.K8sJobCreate(&rq)
	if err != nil {
		return diag.Errorf("Failed to create Job. API error: %s", err)
	}
	log.Printf("[INFO] Submitted updated Job")

	id := fmt.Sprintf("v3/subscriptions/%s/k8s/job/%s", tenantId, name)
	d.SetId(id)

	if d.Get("wait_for_completion").(bool) {
		// wait for completion
		diags := waitForResourceToBePresentAfterCreate(ctx, d, "k8s job", id, func() (interface{}, duplosdk.ClientError) {
			return c.K8sJobGet(tenantId, name)
		})
		if diags != nil {
			return diags
		}
	}

	if d.Get("wait_for_completion").(bool) {
		err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate),
			retryUntilJobV1IsFinished(ctx, c, tenantId, name))
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.Diagnostics{}
	}

	log.Printf("[TRACE] resourceKubernetesJobV1Create(%s): end", tenantId)
	return resourceKubernetesJobV1Read(ctx, d, meta)
}

func resourceKubernetesJobV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId, jobName, err := parseK8sJobIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Reading job %s/%s", tenantId, jobName)

	// Get the object from Duplo, detecting a missing object
	c := meta.(*duplosdk.Client)
	job, err := c.K8sJobGet(tenantId, jobName)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return diag.Errorf("Failed to read Job. API error: %s", err)
	}
	log.Printf("[INFO] Received Job: %#v", job)

	// Remove server-generated labels unless using manual selector
	if _, ok := d.GetOk("spec.0.manual_selector"); !ok {
		labels := job.Metadata.Labels

		delete(labels, "job-name")

		labels = job.Spec.Selector.MatchLabels

		delete(labels, "controller-uid")
	}

	metaError := d.Set("metadata", flattenMetadata(job.Metadata, d, meta))
	if metaError != nil {
		return diag.FromErr(metaError)
	}

	jobSpec, jobError := flattenJobV1Spec(job.Spec, d, meta)
	if jobError != nil {
		return diag.FromErr(jobError)
	}

	specError := d.Set("spec", jobSpec)
	if specError != nil {
		return diag.FromErr(specError)
	}

	return diag.Diagnostics{}
}

func resourceKubernetesJobV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceKubernetesJobV1Update(%s): end", tenantId)

	name, err := getK8sJobName(d)
	if err != nil {
		return diag.FromErr(err)
	}

	var rq duplosdk.DuploK8sJob
	rq.TenantId = tenantId

	if d.HasChange("spec") {
		spec, err := expandJobV1Spec(d.Get("spec").([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		rq.Spec = spec
	}

	if d.HasChange("metadata") {
		metadata := expandMetadata(d.Get("metadata").([]interface{}))
		rq.Metadata = metadata
	}

	// initiate update Job
	c := meta.(*duplosdk.Client)
	err = c.K8sJobUpdate(&rq)
	if err != nil {
		return diag.Errorf("Failed to update Job. API error: %s", err)
	}
	log.Printf("[INFO] Submitted updated Job")

	// wait for completion
	id := fmt.Sprintf("v3/subscriptions/%s/k8s/job/%s", tenantId, name)
	d.SetId(id)

	diags := resourceKubernetesJobV1Read(ctx, d, meta)
	log.Printf("[TRACE] resourceKubernetesJobV1Update(%s): end", tenantId)
	return diags
}

func resourceKubernetesJobV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	name, err := getK8sJobName(d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceKubernetesJobV1Delete(%s, %s): start", tenantId, name)

	// Get the object from Duplo, detecting a missing object
	c := meta.(*duplosdk.Client)
	rp, err := c.K8sJobGet(tenantId, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if rp != nil || rp.Metadata.Name != "" {
		clientError := c.K8sJobDelete(tenantId, name)
		if clientError != nil {
			if clientError.Status() == 404 {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}

	// resource deleted
	d.SetId("")

	log.Printf("[TRACE] resourceKubernetesJobV1Delete(%s, %s): end", tenantId, name)
	return nil
}

func getK8sJobName(d *schema.ResourceData) (string, error) {
	// Retrieve the metadata, checking for its existence.
	metadata, exists := d.GetOk("metadata")
	if !exists || len(metadata.([]interface{})) < 1 {
		return "", fmt.Errorf("metadata must be specified")
	}

	// Cast the metadata entry to a map and retrieve the job name.
	metadataMap := metadata.([]interface{})[0].(map[string]interface{})
	jobNameRaw, exists := metadataMap["name"]
	if !exists || jobNameRaw == "" {
		return "", fmt.Errorf("name must be specified inside the metadata block")
	}

	// Convert the job name to a string.
	return jobNameRaw.(string), nil
}

func parseK8sJobIdParts(id string) (tenantId, name string, err error) {
	// Compile a regular expression that matches a GUID and a job name in your specific URL format.
	r := regexp.MustCompile(`^v3/subscriptions/([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12})/k8s/job/([^/]+)$`)
	matches := r.FindStringSubmatch(id)

	if len(matches) == 3 {
		// The first element of matches is the entire string, the second is the first capture group (tenantId), and the third is the second capture group (name).
		tenantId, name = matches[1], matches[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func diffIgnoreDuploCreatedLabels(k, old, new string, d *schema.ResourceData) bool {
	// List of labels created by the backend API
	backendLabels := map[string]bool{
		"metadata.0.labels.duplocloud.net/owner":      true,
		"metadata.0.labels.duplocloud.net/tenantid":   true,
		"metadata.0.labels.duplocloud.net/tenantname": true,
	}

	// Check if the label key is in the list of backend-created labels
	if backendLabels[k] {
		// Ignore the difference for backend-created labels
		return true
	}

	// For other labels, compare old and new values
	return old == new
}

// retryUntilJobV1IsFinished checks if a given job has finished its execution in either a Complete or Failed state
func retryUntilJobV1IsFinished(ctx context.Context, client *duplosdk.Client, tenantId string, name string) retry.RetryFunc {
	return func() *retry.RetryError {
		job, err := client.K8sJobGet(tenantId, name)
		if err != nil {
			if statusErr := err.Status(); statusErr == 404 {
				return nil
			}
			return retry.NonRetryableError(err)
		}

		for _, c := range job.Status.Conditions {
			if c.Status == corev1.ConditionTrue {
				log.Printf("[DEBUG] Current condition of job: %s: %s\n", name, c.Type)
				switch c.Type {
				case batchv1.JobComplete:
					return nil
				case batchv1.JobFailed:
					return retry.NonRetryableError(fmt.Errorf("job: %s is in failed state", name))
				}
			}
		}

		return retry.RetryableError(fmt.Errorf("job: %s is not in complete state", name))
	}
}
