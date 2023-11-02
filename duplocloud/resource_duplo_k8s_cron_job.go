package duplocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Schema:        resourceKubernetesCronJobSchemaV1Beta1(),
	}
}

func resourceKubernetesCronJobSchemaV1Beta1() map[string]*schema.Schema {
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
			Required:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: cronJobSpecFieldsV1Beta1(),
			},
		},
	}
}

func resourceKubernetesCronJobV1Beta1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceKubernetesJobV1Create(%s): start", tenantId)

	// TODO: validate the getK8sJobName() function works for cron jobs
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
	log.Printf("[TRACE] resourceKubernetesCronJobV1Update(%s): end", tenantId)

	name, err := getK8sJobName(d)
	if err != nil {
		return diag.FromErr(err)
	}

	var rq duplosdk.DuploK8sCronJob

	if d.HasChange("spec") {
		spec, err := expandCronJobSpecV1Beta1(d.Get("spec").([]interface{}))
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
	err = c.K8sCronJobUpdate(tenantId, name, &rq)
	if err != nil {
		return diag.Errorf("Failed to update CronJob. API error: %s", err)
	}
	log.Printf("[INFO] Submitted updated cron job")

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	spec, err := expandCronJobSpecV1Beta1(d.Get("spec").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}
	spec.JobTemplate.ObjectMeta.Annotations = metadata.Annotations

	cronjob := &v1beta1.CronJob{
		ObjectMeta: metadata,
		Spec:       spec,
	}

	log.Printf("[INFO] Updating cron job %s: %s", d.Id(), cronjob)

	out, err := conn.BatchV1beta1().CronJobs(namespace).Update(ctx, cronjob, metav1.UpdateOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Submitted updated cron job: %#v", out)

	d.SetId(buildId(out.ObjectMeta))
	return resourceKubernetesCronJobV1Beta1Read(ctx, d, meta)
}

func resourceKubernetesCronJobV1Beta1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId, jobName, err := parseK8sJobIdParts(d.Id())
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

	// Remove server-generated labels unless using manual selector
	if _, ok := d.GetOk("spec.0.manual_selector"); !ok {
		labels := job.Metadata.Labels

		if _, ok := labels["controller-uid"]; ok {
			delete(labels, "controller-uid")
		}

		if _, ok := labels["cron-job-name"]; ok {
			delete(labels, "cron-job-name")
		}

		if job.Spec.JobTemplate.Spec.Selector != nil &&
			job.Spec.JobTemplate.Spec.Selector.MatchLabels != nil {
			labels = job.Spec.JobTemplate.Spec.Selector.MatchLabels
		}

		if _, ok := labels["controller-uid"]; ok {
			delete(labels, "controller-uid")
		}
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

	return nil
}

func resourceKubernetesCronJobV1Beta1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn, err := meta.(KubeClientsets).MainClientset()
	if err != nil {
		return diag.FromErr(err)
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleting cron job: %#v", name)
	err = conn.BatchV1beta1().CronJobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.BatchV1beta1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if statusErr, ok := err.(*errors.StatusError); ok && errors.IsNotFound(statusErr) {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		e := fmt.Errorf("Cron Job %s still exists", name)
		return resource.RetryableError(e)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Cron Job %s deleted", name)

	d.SetId("")
	return nil
}

func resourceKubernetesCronJobV1Beta1Exists(ctx context.Context, d *schema.ResourceData, meta interface{}) (bool, error) {
	conn, err := meta.(KubeClientsets).MainClientset()
	if err != nil {
		return false, err
	}

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return false, err
	}

	log.Printf("[INFO] Checking cron job %s", name)
	_, err = conn.BatchV1beta1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && errors.IsNotFound(statusErr) {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, err
}
