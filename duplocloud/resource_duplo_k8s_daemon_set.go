package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKubernetesDaemonSetV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKubernetesDaemonSetV1Create,
		ReadContext:   resourceKubernetesDaemonSetV1Read,
		UpdateContext: resourceKubernetesDaemonSetV1Update,
		DeleteContext: resourceKubernetesDaemonSetV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		SchemaVersion: 1,
		Schema:        resourceKubernetesDaemonSetV1Schema(false),
	}
}

func resourceKubernetesDaemonSetV1Schema(readonly bool) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the DaemonSet will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"metadata": namespacedMetadataSchema("daemonset", false),
		"spec": {
			Type:        schema.TypeList,
			Description: "Spec of the DaemonSet.",
			Optional:    !readonly,
			Computed:    readonly,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: daemonSetSpecFields(),
			},
		},
		"is_tenant_local": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "When true, the DaemonSet will be deployed only to the tenant's nodes. When false, it is deployed cluster-wide (requires the tenant to have the CAN_DEPLOY_CLUSTER_WIDE_DAEMONSET metadata enabled).",
		},
	}
}

func resourceKubernetesDaemonSetV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceKubernetesDaemonSetV1Create(%s): start", tenantId)

	name, err := getK8sDaemonSetName(d)
	if err != nil {
		return diag.FromErr(err)
	}

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	spec, err := expandDaemonSetSpec(d.Get("spec").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	rq := duplosdk.DuploK8sDaemonSet{
		TenantId:      tenantId,
		Metadata:      metadata,
		Spec:          spec,
		IsTenantLocal: d.Get("is_tenant_local").(bool),
	}

	c := meta.(*duplosdk.Client)
	if err := c.K8sDaemonSetCreate(&rq); err != nil {
		return diag.Errorf("Failed to create DaemonSet. API error: %s", err)
	}
	log.Printf("[INFO] Submitted new DaemonSet %s/%s", tenantId, name)

	id := fmt.Sprintf("v3/subscriptions/%s/k8s/daemonSet/%s", tenantId, name)
	d.SetId(id)

	diags := waitForResourceToBePresentAfterCreate(ctx, d, "k8s daemonset", id, func() (interface{}, duplosdk.ClientError) {
		return c.K8sDaemonSetGet(tenantId, name)
	})
	if diags != nil {
		return diags
	}

	log.Printf("[TRACE] resourceKubernetesDaemonSetV1Create(%s): end", tenantId)
	return resourceKubernetesDaemonSetV1Read(ctx, d, meta)
}

func resourceKubernetesDaemonSetV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId, name, err := parseK8sDaemonSetIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Reading DaemonSet %s/%s", tenantId, name)

	c := meta.(*duplosdk.Client)
	ds, cerr := c.K8sDaemonSetGet(tenantId, name)
	if cerr != nil {
		if cerr.Status() == 404 {
			log.Printf("[TRACE] resourceKubernetesDaemonSetV1Read(%s, %s): object not found", tenantId, name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read DaemonSet. API error: %s", cerr)
	}
	if ds == nil {
		d.SetId("")
		return nil
	}

	d.Set("tenant_id", tenantId)
	d.Set("is_tenant_local", ds.IsTenantLocal)

	if metaErr := d.Set("metadata", flattenMetadata(ds.Metadata, d, meta)); metaErr != nil {
		return diag.FromErr(metaErr)
	}

	specOut, specErr := flattenDaemonSetSpec(ds.Spec, d, meta)
	if specErr != nil {
		return diag.FromErr(specErr)
	}
	if err := d.Set("spec", specOut); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func resourceKubernetesDaemonSetV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceKubernetesDaemonSetV1Update(%s): start", tenantId)

	name, err := getK8sDaemonSetName(d)
	if err != nil {
		return diag.FromErr(err)
	}

	metadata := expandMetadata(d.Get("metadata").([]interface{}))
	spec, err := expandDaemonSetSpec(d.Get("spec").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	rq := duplosdk.DuploK8sDaemonSet{
		TenantId:      tenantId,
		Metadata:      metadata,
		Spec:          spec,
		IsTenantLocal: d.Get("is_tenant_local").(bool),
	}

	c := meta.(*duplosdk.Client)
	if err := c.K8sDaemonSetUpdate(tenantId, name, &rq); err != nil {
		return diag.Errorf("Failed to update DaemonSet. API error: %s", err)
	}
	log.Printf("[INFO] Submitted updated DaemonSet %s/%s", tenantId, name)

	id := fmt.Sprintf("v3/subscriptions/%s/k8s/daemonSet/%s", tenantId, name)
	d.SetId(id)

	diags := resourceKubernetesDaemonSetV1Read(ctx, d, meta)
	log.Printf("[TRACE] resourceKubernetesDaemonSetV1Update(%s): end", tenantId)
	return diags
}

func resourceKubernetesDaemonSetV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tenantId := d.Get("tenant_id").(string)
	name, err := getK8sDaemonSetName(d)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceKubernetesDaemonSetV1Delete(%s, %s): start", tenantId, name)

	c := meta.(*duplosdk.Client)
	existing, err := c.K8sDaemonSetGet(tenantId, name)
	if err != nil {
		return diag.FromErr(err)
	}
	if existing != nil {
		if clientError := c.K8sDaemonSetDelete(tenantId, name); clientError != nil {
			if clientError.Status() == 404 {
				log.Printf("[TRACE] resourceKubernetesDaemonSetV1Delete(%s, %s): object not found", tenantId, name)
				d.SetId("")
				return nil
			}
			return diag.Errorf("Failed to delete DaemonSet. API error: %s", clientError)
		}
	}

	d.SetId("")
	log.Printf("[TRACE] resourceKubernetesDaemonSetV1Delete(%s, %s): end", tenantId, name)
	return nil
}

func getK8sDaemonSetName(d *schema.ResourceData) (string, error) {
	metadata, exists := d.GetOk("metadata")
	if !exists || len(metadata.([]interface{})) < 1 {
		return "", fmt.Errorf("metadata must be specified")
	}
	m := metadata.([]interface{})[0].(map[string]interface{})
	nameRaw, exists := m["name"]
	if !exists || nameRaw == "" {
		return "", fmt.Errorf("name must be specified inside the metadata block")
	}
	return nameRaw.(string), nil
}

func parseK8sDaemonSetIdParts(id string) (tenantId, name string, err error) {
	r := regexp.MustCompile(`^v3/subscriptions/([a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12})/k8s/daemonSet/([^/]+)$`)
	matches := r.FindStringSubmatch(id)
	if len(matches) == 3 {
		tenantId, name = matches[1], matches[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
