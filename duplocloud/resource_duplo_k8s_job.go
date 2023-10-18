package duplocloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"regexp"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		"name": {
			Description:  "The short name of the job.  Duplo will add a prefix to the name.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotWhiteSpace,
		},
		"metadata": jobMetadataSchema(),
		"spec": {
			Type:        schema.TypeList,
			Description: "Spec of the job owned by the cluster",
			Required:    true,
			MaxItems:    1,
			ForceNew:    false,
			Elem: &schema.Resource{
				Schema: jobSpecFields(true),
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
	tenantID := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceKubernetesJobV1Create(%s): end", tenantID)

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	// try to get name from root level, if not present, use metadata name
	name := d.Get("name").(string)
	if name == "" {
		name = metadata.Name
	}

	spec, err := expandJobV1Spec(d.Get("spec").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	var rq duplosdk.DuploK8sJob
	rq.Metadata = metadata
	rq.Spec = spec

	// initiate create Job
	c := meta.(*duplosdk.Client)
	err = c.K8sJobCreate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("Failed to create Job! API error: %s", err)
	}

	// wait for completion
	id := fmt.Sprintf("v3/subscriptions/%s/k8s/job/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "k8s job", id, func() (interface{}, duplosdk.ClientError) {
		return c.K8sJobGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceKubernetesJobV1Read(ctx, d, meta)
	log.Printf("[TRACE] resourceKubernetesJobV1Create(%s): end", tenantID)
	return diags
}

func resourceKubernetesJobV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantID, jobName, err := parseK8sJobIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Reading job %s/%s", tenantID, jobName)

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

		delete(labels, "controller-uid")

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
	log.Printf("[TRACE] resourceKubernetesJobV1Create(%s): end", tenantId)

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	// try to get name from root level, if not present, use metadata name
	name := d.Get("name").(string)
	if name == "" {
		name = metadata.Name
	}

	var rq duplosdk.DuploK8sJob

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
	err := c.K8sJobUpdate(tenantId, name, &rq)
	if err != nil {
		return diag.Errorf("Failed to update Job! API error: %s", err)
	}

	// wait for completion
	id := fmt.Sprintf("v3/subscriptions/%s/k8s/job/%s", tenantId, name)
	d.SetId(id)

	diags := resourceKubernetesJobV1Read(ctx, d, meta)
	log.Printf("[TRACE] resourceKubernetesJobV1Create(%s): end", tenantId)
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
		err := c.K8sJobDelete(tenantId, name)
		if err != nil {
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

func parseK8sJobIdParts(id string) (tenantID, name string, err error) {
	// Compile a regular expression that matches a GUID and a job name in your specific URL format.
	r := regexp.MustCompile(`^v3/subscriptions/([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12})/k8s/job/([^/]+)$`)
	matches := r.FindStringSubmatch(id)

	if len(matches) == 3 {
		// The first element of matches is the entire string, the second is the first capture group (tenantID), and the third is the second capture group (name).
		tenantID, name = matches[1], matches[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
