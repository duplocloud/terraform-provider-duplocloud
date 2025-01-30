package duplocloud

import (
	"context"
	"encoding/json"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Resource for managing an infrastructure's settings.
func resourceHelmRelease() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_helm_release` manages helm release at duplocloud",

		ReadContext:   resourceHelmReleaseRead,
		CreateContext: resourceHelmReleaseCreate,
		UpdateContext: resourceHelmReleaseUpdate,
		DeleteContext: resourceHelmReleaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description:  "The GUID of the tenant that the storage bucket will be created in.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Description:  "The name of the helm chart",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]{0,62}[a-zA-Z0-9]$`), "Invalid name format, name can be 64 character long and start with an alphabet or digit and can contain hypen or periods"),
			},
			"release_name": {
				Description:  "Provide release name to identify specific deployment of helm chart.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]{0,62}[a-zA-Z0-9]$`), "Invalid name format, name can be 64 character long and start with an alphabet or digit and can contain hypen or periods"),
			},
			"interval": {
				Description:  "Interval related to helm release",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "5m0s",
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([0-5]?\d)m([0-5]?\d)s$`), "invalid minute second format, valid format 0m0s or 00m00s m[0-59] s[0-59]"),
			},
			"chart": {
				Description: "Helm chart",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				ForceNew:    true, // relaunch instance
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description:  "Provide unique name for the helm chart.",
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]{0,62}[a-zA-Z0-9]$`), "Invalid name format, name can be 64 character long and start with an alphabet or digit and can contain hypen or periods"),
						},
						"interval": {
							Description:  "The interval associated to helm chart",
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "5m0s",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([0-5]?\d)m([0-5]?\d)s$`), "invalid minute second format, valid format 0m0s or 00m00s m[0-59] s[0-59]"),
						},
						"version": {
							Description: "The helm chart version",
							Type:        schema.TypeString,
							Required:    true,
						},
						"reconcile_strategy": {
							Description:  "The reconcile strategy should be chosen from ChartVersion or Revision. No new chart artifact is produced on updates to the source unless the version is changed in HelmRepository. Use `Revision` to produce new chart artifact on change in source revision.",
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"ChartVersion", "Revision"}, false),
							Default:      "ChartVersion",
						},
						"source_type": {
							Description:  "The helm chart source, currently only HelmRepository as source is supported",
							Type:         schema.TypeString,
							Default:      "HelmRepository",
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"HelmRepository"}, false),
						},
						"source_name": {
							Description:  "The name of the source, referred from helm repository resource.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]{0,62}[a-zA-Z0-9]$`), "Invalid name format, name can be 64 character long and start with an alphabet or digit and can contain hypen or periods"),
						},
					},
				},
			},
			"values": {
				Description: "Customise an helm chart.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func resourceHelmReleaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID := idParts[0]
	name := idParts[2]
	log.Printf("[TRACE] resourceHelmReleaseRead(%s,%s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	duplo, err := c.DuploHelmReleaseGet(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to retrieve helm release details for '%s': %s", name, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	flattenErr := flattenHelmRelease(d, *duplo)
	if err != nil {
		diag.Errorf("%s", flattenErr.Error())
	}
	log.Printf("[TRACE] resourceHelmReleaseRead(%s,%s): end", tenantID, name)
	return nil
}

func resourceHelmReleaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	rq, err := expandHelmRelease(d)
	log.Printf("[TRACE] resourceHelmReleaseCreate(%s,%s): start", tenantID, rq.Metadata.Name)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	c := m.(*duplosdk.Client)

	err = c.DuploHelmReleaseCreate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("resourceHelmReleaseCreate cannot create helm release %s for tenant %s error: %s", rq.Metadata.Name, tenantID, err.Error())
	}
	err = helmReleaseWaitUntilReady(ctx, c, tenantID, rq.Metadata.Name, d.Timeout("create"))
	if err != nil {
		diag.Errorf("%s", err.Error())
	}

	d.SetId(tenantID + "/helm-release/" + rq.Metadata.Name)

	diags := resourceHelmReleaseRead(ctx, d, m)
	log.Printf("[TRACE] resourceHelmReleaseCreate(%s,%s): end", tenantID, rq.Metadata.Name)
	return diags
}

func resourceHelmReleaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID := idParts[0]
	name := idParts[2]
	rq, err := expandHelmRelease(d)
	log.Printf("[TRACE] resourceHelmReleaseUpdate(%s,%s): start", tenantID, name)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	c := m.(*duplosdk.Client)

	err = c.DuploHelmReleaseUpdate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("resourceHelmReleaseUpdate cannot update helm release %s for tenant %s error: %s", rq.Metadata.Name, tenantID, err.Error())
	}
	err = helmReleaseWaitUntilReady(ctx, c, tenantID, rq.Metadata.Name, d.Timeout("update"))
	if err != nil {
		diag.Errorf("%s", err.Error())
	}
	diags := resourceHelmReleaseRead(ctx, d, m)
	log.Printf("[TRACE] resourceHelmReleaseUpdate(%s,%s): end", tenantID, rq.Metadata.Name)
	return diags
}

func resourceHelmReleaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID := idParts[0]
	name := idParts[2]
	log.Printf("[TRACE] resourceHelmReleaseDelete(%s,%s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	err := c.DuploHelmReleaseDelete(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to delete helm release %s for '%s': %s", name, tenantID, err)
	}

	log.Printf("[TRACE] resourceHelmReleaseDelete(%s,%s): end", tenantID, name)
	return nil
}

func expandHelmRelease(d *schema.ResourceData) (duplosdk.DuploHelmRelease, error) {
	obj := duplosdk.DuploHelmRelease{}
	obj.Metadata = duplosdk.DuploHelmMetadata{
		Name: d.Get("name").(string),
	}
	obj.Spec = duplosdk.DuploHelmSpec{
		Interval:    d.Get("interval").(string),
		ReleaseName: d.Get("release_name").(string),
	}
	obj.Spec.Chart = &duplosdk.Chart{
		Spec: duplosdk.ChartSpec{
			Interval:          d.Get("chart.0.interval").(string),
			Chart:             d.Get("chart.0.name").(string),
			Version:           d.Get("chart.0.version").(string),
			ReconcileStrategy: d.Get("chart.0.reconcile_strategy").(string),
			SourceRef: duplosdk.SourceRef{
				Kind: d.Get("chart.0.source_type").(string),
				Name: d.Get("chart.0.source_name").(string),
			},
		},
	}
	err := json.Unmarshal([]byte(d.Get("values").(string)), &obj.Spec.Values)
	return obj, err
}

func flattenHelmRelease(d *schema.ResourceData, rb duplosdk.DuploHelmRelease) error {
	d.Set("name", rb.Metadata.Name)
	d.Set("interval", rb.Spec.Interval)
	d.Set("release_name", rb.Spec.ReleaseName)
	d.Set("chart.0.name", rb.Spec.Chart.Spec.Chart)
	d.Set("chart.0.version", rb.Spec.Chart.Spec.Version)
	d.Set("chart.0.interval", rb.Spec.Chart.Spec.Interval)
	d.Set("chart.0.source_type", rb.Spec.Chart.Spec.SourceRef.Kind)
	d.Set("chart.0.source_name", rb.Spec.Chart.Spec.SourceRef.Name)
	d.Set("chart.0.reconcile_strategy", rb.Spec.Chart.Spec.ReconcileStrategy)
	v, err := json.Marshal(rb.Spec.Values)
	if err == nil {
		d.Set("values", string(v))
	}
	return err
}

func helmReleaseWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.DuploHelmReleaseGet(tenantID, name)
			status := "pending"
			if err == nil {
				for _, d := range rp.Status.Condition {
					if d.Type == "Ready" {
						status = "ready"
						break
					}
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] helmReleaseWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
