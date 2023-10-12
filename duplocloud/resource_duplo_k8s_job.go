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

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
		Schema: resourceKubernetesJobV1Schema(),
	}
}

func resourceKubernetesJobV1Schema() map[string]*schema.Schema {
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
			Required:    true,
			MaxItems:    1,
			ForceNew:    false,
			Elem: &schema.Resource{
				Schema: jobSpecFields(false),
			},
		},
		"wait_for_completion": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
	}
}

func resourceKubernetesJobV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//conn, err := meta.(KubeClientsets).MainClientset()
	//if err != nil {
	//	return diag.FromErr(err)
	//}

	//metadata := expandMetadata(d.Get("metadata").([]interface{}))
	//spec, err := expandJobV1Spec(d.Get("spec").([]interface{}))
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	//
	//job := batchv1.Job{
	//	ObjectMeta: metadata,
	//	Spec:       spec,
	//}
	//
	//log.Printf("[INFO] Creating new Job: %#v", job)
	//
	//out, err := conn.BatchV1().Jobs(metadata.Namespace).Create(ctx, &job, metav1.CreateOptions{})
	//if err != nil {
	//	return diag.Errorf("Failed to create Job! API error: %s", err)
	//}
	//log.Printf("[INFO] Submitted new job: %#v", out)
	//
	//d.SetId(buildId(out.ObjectMeta))
	//
	//namespace, name, err := idParts(d.Id())
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	//if d.Get("wait_for_completion").(bool) {
	//	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate),
	//		retryUntilJobV1IsFinished(ctx, conn, namespace, name))
	//	if err != nil {
	//		return diag.FromErr(err)
	//	}
	//	return diag.Diagnostics{}
	//}

	return resourceKubernetesJobV1Read(ctx, d, meta)
}

func resourceKubernetesJobV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)

	jobName, err := getK8sJobName(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the object from Duplo, detecting a missing object
	c := meta.(*duplosdk.Client)
	job, err := c.K8sJobGet(tenantID, jobName)
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return diag.Errorf("Failed to read Job! API error: %s", err)
	}
	log.Printf("[INFO] Received job: %#v", job)

	// Remove server-generated labels unless using manual selector
	if _, ok := d.GetOk("spec.0.manual_selector"); !ok {
		labels := job.Metadata.Labels

		if _, ok := labels["controller-uid"]; ok {
			delete(labels, "controller-uid")
		}

		if _, ok := labels["job-name"]; ok {
			delete(labels, "job-name")
		}

		labels = job.Spec.Selector.MatchLabels

		if _, ok := labels["controller-uid"]; ok {
			delete(labels, "controller-uid")
		}
	}

	err = d.Set("metadata", flattenMetadata(job.Metadata, d, meta))
	if err != nil {
		return diag.FromErr(err)
	}

	jobSpec, err := flattenJobV1Spec(job.Spec, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("spec", jobSpec)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func resourceKubernetesJobV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//conn, err := meta.(KubeClientsets).MainClientset()
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	//
	//namespace, name, err := idParts(d.Id())
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	//
	//ops := patchMetadata("metadata.0.", "/metadata/", d)
	//
	//if d.HasChange("spec") {
	//	specOps, err := patchJobV1Spec("/spec", "spec.0.", d)
	//	if err != nil {
	//		return diag.FromErr(err)
	//	}
	//	ops = append(ops, specOps...)
	//}
	//
	//data, err := ops.MarshalJSON()
	//if err != nil {
	//	return diag.Errorf("Failed to marshal update operations: %s", err)
	//}
	//
	//log.Printf("[INFO] Updating job %s: %#v", d.Id(), ops)
	//
	//out, err := conn.BatchV1().Jobs(namespace).Patch(ctx, name, pkgApi.JSONPatchType, data, metav1.PatchOptions{})
	//if err != nil {
	//	return diag.Errorf("Failed to update Job! API error: %s", err)
	//}
	//log.Printf("[INFO] Submitted updated job: %#v", out)
	//
	//d.SetId(buildId(out.ObjectMeta))
	//
	//if d.Get("wait_for_completion").(bool) {
	//	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate),
	//		retryUntilJobV1IsFinished(ctx, conn, namespace, name))
	//	if err != nil {
	//		return diag.FromErr(err)
	//	}
	//}
	return resourceKubernetesJobV1Read(ctx, d, meta)
}

func resourceKubernetesJobV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	//conn, err := meta.(KubeClientsets).MainClientset()
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	//
	//namespace, name, err := idParts(d.Id())
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	//
	//log.Printf("[INFO] Deleting job: %#v", name)
	//err = conn.BatchV1().Jobs(namespace).Delete(ctx, name, deleteOptions)
	//if err != nil {
	//	if statusErr, ok := err.(*errors.StatusError); ok && errors.IsNotFound(statusErr) {
	//		return nil
	//	}
	//	return diag.Errorf("Failed to delete Job! API error: %s", err)
	//}
	//
	//err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
	//	_, err := conn.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	//	if err != nil {
	//		if statusErr, ok := err.(*errors.StatusError); ok && errors.IsNotFound(statusErr) {
	//			return nil
	//		}
	//		return resource.NonRetryableError(err)
	//	}
	//
	//	e := fmt.Errorf("Job %s still exists", name)
	//	return resource.RetryableError(e)
	//})
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	//
	//log.Printf("[INFO] Job %s deleted", name)

	d.SetId("")
	return nil
}

// retryUntilJobV1IsFinished checks if a given job has finished its execution in either a Complete or Failed state
func retryUntilJobV1IsFinished(ctx context.Context, conn *kubernetes.Clientset, ns, name string) resource.RetryFunc {
	return func() *resource.RetryError {
		job, err := conn.BatchV1().Jobs(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if statusErr, ok := err.(*errors.StatusError); ok && errors.IsNotFound(statusErr) {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		for _, c := range job.Status.Conditions {
			if c.Status == corev1.ConditionTrue {
				log.Printf("[DEBUG] Current condition of job: %s/%s: %s\n", ns, name, c.Type)
				switch c.Type {
				case batchv1.JobComplete:
					return nil
				case batchv1.JobFailed:
					return resource.NonRetryableError(fmt.Errorf("job: %s/%s is in failed state", ns, name))
				}
			}
		}

		return resource.RetryableError(fmt.Errorf("job: %s/%s is not in complete state", ns, name))
	}
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
