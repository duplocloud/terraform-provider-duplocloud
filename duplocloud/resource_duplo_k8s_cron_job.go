package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceKubernetesCronJobV1Beta1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKubernetesCronJobV1Beta1Create,
		ReadContext:   resourceKubernetesCronJobV1Beta1Read,
		UpdateContext: resourceKubernetesCronJobV1Beta1Update,
		DeleteContext: resourceKubernetesCronJobV1Beta1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		SchemaVersion: 1,
		Schema:        resourceKubernetesCronJobSchemaV1Beta1(false),
	}
}

func resourceKubernetesCronJobSchemaV1Beta1(readonly bool) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the job will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"metadata": namespacedMetadataSchema("cronjob", true),
		"spec": {
			Type:        schema.TypeList,
			Description: "Spec of the cron job owned by the cluster",
			Optional:    !readonly,
			Computed:    readonly,
			Elem: &schema.Resource{
				Schema: cronJobSpecFieldsV1Beta1(),
			},
		},
		"is_any_host_allowed": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
	}
}

func resourceKubernetesCronJobV1Beta1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	isAnyHostAllowed := d.Get("is_any_host_allowed").(bool)
	log.Printf("[TRACE] resourceKubernetesJobV1Create(%s): start", tenantId)

	name, err := getK8sJobName(d)
	if err != nil {
		return diag.FromErr(err)
	}

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	spec, err := expandCronJobSpecV1Beta1(d.Get("spec").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	var rq duplosdk.DuploK8sCronJob
	rq.Metadata = metadata
	rq.Spec = spec
	rq.TenantId = tenantId
	rq.IsAnyHostAllowed = isAnyHostAllowed

	c := meta.(*duplosdk.Client)
	err = c.K8sCronJobCreate(&rq)
	if err != nil {
		return diag.Errorf("Failed to create CronJob. API error: %s", err)
	}
	log.Printf("[INFO] Submitted new cron job")

	id := fmt.Sprintf("v3/subscriptions/%s/k8s/cronjob/%s", tenantId, name)
	d.SetId(id)

	return resourceKubernetesCronJobV1Beta1Read(ctx, d, meta)
}

func resourceKubernetesCronJobV1Beta1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	isAnyHostAllowed := d.Get("is_any_host_allowed").(bool)
	log.Printf("[TRACE] resourceKubernetesCronJobV1Update(%s): end", tenantId)

	name, err := getK8sJobName(d)
	if err != nil {
		return diag.FromErr(err)
	}

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	spec, err := expandCronJobSpecV1Beta1(d.Get("spec").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}
	spec.JobTemplate.ObjectMeta.Annotations = metadata.Annotations

	rq := duplosdk.DuploK8sCronJob{
		Metadata:         metadata,
		Spec:             spec,
		IsAnyHostAllowed: isAnyHostAllowed,
	}

	// initiate update Job
	c := meta.(*duplosdk.Client)
	err = c.K8sCronJobUpdate(tenantId, name, &rq)
	if err != nil {
		return diag.Errorf("Failed to update CronJob. API error: %s", err)
	}
	log.Printf("[INFO] Submitted updated CronJob")

	// wait for completion
	id := fmt.Sprintf("v3/subscriptions/%s/k8s/cronjob/%s", tenantId, name)
	d.SetId(id)

	diags := resourceKubernetesCronJobV1Beta1Read(ctx, d, meta)
	log.Printf("[TRACE] resourceKubernetesCronJobV1Update(%s): end", tenantId)
	return diags
}

func resourceKubernetesCronJobV1Beta1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId, jobName, err := parseK8sCronJobIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Reading cron job %s/%s", tenantId, jobName)

	c := meta.(*duplosdk.Client)
	job, err := c.K8sCronJobGet(tenantId, jobName)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return diag.Errorf("Failed to read CronJob. API error: %s", err)
	}
	log.Printf("[INFO] Received CronJob: %#v", job)

	isAnyHostAllowed := GetIsAnyHostAllowed(job.Metadata.Annotations)
	job.IsAnyHostAllowed = isAnyHostAllowed

	// Remove server-generated labels unless using manual selector
	if _, ok := d.GetOk("spec.0.manual_selector"); !ok {
		labels := job.Metadata.Labels

		delete(labels, "cron-job-name")

		if job.Spec.JobTemplate.Spec.Selector != nil &&
			job.Spec.JobTemplate.Spec.Selector.MatchLabels != nil {
			labels = job.Spec.JobTemplate.Spec.Selector.MatchLabels
		}

		delete(labels, "controller-uid")
	}

	metaError := d.Set("metadata", flattenMetadata(job.Metadata, d, meta))
	if metaError != nil {
		return diag.FromErr(metaError)
	}

	jobSpec, jobError := flattenCronJobSpecV1Beta1(job.Spec, d, meta)
	if jobError != nil {
		return diag.FromErr(jobError)
	}

	err = d.Set("spec", jobSpec)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("tenant_id", tenantId)
	return diag.Diagnostics{}
}

func GetIsAnyHostAllowed(annotations map[string]string) bool {
	if val, ok := annotations["duplocloud.net/is-any-host-allowed"]; ok {
		boolValue, err := strconv.ParseBool(val)
		if err != nil {
			log.Printf("[DEBUG] Received error: %#v", err)
			boolValue = false
		}
		return boolValue
	}
	return false
}

func resourceKubernetesCronJobV1Beta1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	name, err := getK8sJobName(d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceKubernetesCronJobV1Beta1Delete(%s, %s): start", tenantId, name)

	// Get the object from Duplo, detecting a missing object
	c := meta.(*duplosdk.Client)
	rp, err := c.K8sCronJobGet(tenantId, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if rp != nil || rp.Metadata.Name != "" {
		clientError := c.K8sCronJobDelete(tenantId, name)
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

	log.Printf("[TRACE] resourceKubernetesCronJobV1Beta1Delete(%s, %s): end", tenantId, name)
	return nil
}

func parseK8sCronJobIdParts(id string) (tenantId, name string, err error) {
	// Compile a regular expression that matches a GUID and a job name in your specific URL format.
	r := regexp.MustCompile(`^v3/subscriptions/([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12})/k8s/cronjob/([^/]+)$`)
	matches := r.FindStringSubmatch(id)

	if len(matches) == 3 {
		// The first element of matches is the entire string, the second is the first capture group (tenantId), and the third is the second capture group (name).
		tenantId, name = matches[1], matches[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
