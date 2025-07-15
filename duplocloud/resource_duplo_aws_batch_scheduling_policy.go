package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAwsBatchSchedulingPolicySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the aws batch scheduling policy will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Specifies the name of the scheduling policy.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the scheduling policy.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"arn": {
			Description: "The Amazon Resource Name of the scheduling policy.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"fair_share_policy": {
			Description: "A fairshare policy block specifies the `compute_reservation`, `share_delay_seconds`, and `share_distribution` of the scheduling policy. The `fairshare_policy block` is documented below.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"compute_reservation": {
						Description:  "A value used to reserve some of the available maximum vCPU for fair share identifiers that have not yet been used.",
						Type:         schema.TypeInt,
						Optional:     true,
						Computed:     true,
						ValidateFunc: validation.IntBetween(0, 99),
					},
					"share_decay_seconds": {
						Description:  "The time period to use to calculate a fair share percentage for each fair share identifier in use, in seconds.",
						Type:         schema.TypeInt,
						Optional:     true,
						Computed:     true,
						ValidateFunc: validation.IntBetween(0, 604800),
					},
					"share_distribution": {
						Description: "One or more share distribution blocks which define the weights for the fair share identifiers for the fair share policy.",
						Type:        schema.TypeSet,
						// There can be no more than 500 fair share identifiers active in a job queue.
						MaxItems: 500,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"share_identifier": {
									Description:  "A fair share identifier or fair share identifier prefix.",
									Type:         schema.TypeString,
									Required:     true,
									ValidateFunc: validShareIdentifier,
								},
								"weight_factor": {
									Description:  "The weight factor for the fair share identifier.",
									Type:         schema.TypeFloat,
									Optional:     true,
									ValidateFunc: validation.FloatBetween(0.0001, 999.9999),
								},
							},
						},
					},
				},
			},
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

func resourceAwsBatchSchedulingPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_batch_scheduling_policy` manages an aws batch scheduling policy in Duplo.",

		ReadContext:   resourceAwsBatchSchedulingPolicyRead,
		CreateContext: resourceAwsBatchSchedulingPolicyCreate,
		UpdateContext: resourceAwsBatchSchedulingPolicyUpdate,
		DeleteContext: resourceAwsBatchSchedulingPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsBatchSchedulingPolicySchema(),
	}
}

func resourceAwsBatchSchedulingPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsBatchSchedulingPolicyIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsBatchSchedulingPolicyRead(%s, %s): start", tenantID, name)

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	policy, clientErr := c.AwsBatchSchedulingPolicyGet(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(clientErr)
	}
	if policy == nil {
		d.SetId("") // object missing
		return nil
	}
	flattenBatchSchedulingPolicy(d, c, policy, tenantID)

	log.Printf("[TRACE] resourceAwsBatchSchedulingPolicyRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsBatchSchedulingPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsBatchSchedulingPolicyCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	rq := expandAwsBatchSchedulingPolicy(d)
	err = c.AwsBatchSchedulingPolicyCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s aws batch scheduling policy '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws batch scheduling policy", id, func() (interface{}, duplosdk.ClientError) {
		return c.AwsBatchSchedulingPolicyGet(tenantID, fullName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsBatchSchedulingPolicyRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsBatchSchedulingPolicyCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsBatchSchedulingPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	fullname := d.Get("fullname").(string)
	arn := d.Get("arn").(string)
	log.Printf("[TRACE] resourceAwsBatchSchedulingPolicyUpdate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	req := &duplosdk.DuploAwsBatchSchedulingPolicy{
		Arn: arn,
	}
	updateFound := false
	if d.HasChange("fair_share_policy") {
		updateFound = true
		req.FairsharePolicy = expandFairsharePolicy(d.Get("fair_share_policy").([]interface{}))
	}
	if updateFound {
		err = c.AwsBatchSchedulingPolicyUpdate(tenantID, req)
		if err != nil {
			return diag.Errorf("Error updating tenant %s aws batch scheduling policy '%s': %s", tenantID, name, err)
		}

		diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws batch scheduling policy", arn, func() (interface{}, duplosdk.ClientError) {
			return c.AwsBatchSchedulingPolicyGet(tenantID, fullname)
		})
		if diags != nil {
			return diags
		}
		diags = resourceAwsBatchSchedulingPolicyRead(ctx, d, m)
		return diags
	}
	log.Printf("[TRACE] resourceAwsBatchSchedulingPolicyUpdate(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsBatchSchedulingPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsBatchSchedulingPolicyIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsBatchSchedulingPolicyDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	clientErr := c.AwsBatchSchedulingPolicyDelete(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s aws batch scheduling policy '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "aws batch scheduling policy", id, func() (interface{}, duplosdk.ClientError) {
		return c.AwsBatchSchedulingPolicyGet(tenantID, fullName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsBatchSchedulingPolicyDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAwsBatchSchedulingPolicy(d *schema.ResourceData) *duplosdk.DuploAwsBatchSchedulingPolicy {
	fairsharePolicy := expandFairsharePolicy(d.Get("fair_share_policy").([]interface{}))
	return &duplosdk.DuploAwsBatchSchedulingPolicy{
		Name:            d.Get("name").(string),
		FairsharePolicy: fairsharePolicy,
		Tags:            expandAsStringMap("tags", d),
	}
}

func parseAwsBatchSchedulingPolicyIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenBatchSchedulingPolicy(d *schema.ResourceData, c *duplosdk.Client, duplo *duplosdk.DuploAwsBatchSchedulingPolicy, tenantId string) diag.Diagnostics {
	prefix, err := c.GetDuploServicesPrefix(tenantId)
	if err != nil {
		return diag.FromErr(err)
	}
	name, _ := duplosdk.UnprefixName(prefix, duplo.Name)
	d.Set("tenant_id", tenantId)
	d.Set("name", name)
	d.Set("arn", duplo.Arn)
	d.Set("fullname", duplo.Name)
	d.Set("tags", duplo.Tags)
	d.Set("fair_share_policy", flattenFairsharePolicy(duplo.FairsharePolicy))
	return nil
}

func validShareIdentifier(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[a-zA-Z0-9]{0,254}[a-zA-Z0-9*]?$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q (%q) must be limited to 255 alphanumeric characters, where the last character can be an asterisk (*).", k, v))
	}
	return
}

func flattenFairsharePolicy(fairsharePolicy *duplosdk.DuploAwsFairsharePolicy) []interface{} {
	if fairsharePolicy == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"compute_reservation": fairsharePolicy.ComputeReservation,
		"share_decay_seconds": fairsharePolicy.ShareDecaySeconds,
	}

	shareDistributions := fairsharePolicy.ShareDistribution

	fairsharePolicyShareDistributions := []interface{}{}
	for _, shareDistribution := range *shareDistributions {
		sdValues := map[string]interface{}{
			"share_identifier": shareDistribution.ShareIdentifier,
			"weight_factor":    shareDistribution.WeightFactor,
		}
		fairsharePolicyShareDistributions = append(fairsharePolicyShareDistributions, sdValues)
	}
	values["share_distribution"] = fairsharePolicyShareDistributions

	return []interface{}{values}
}

func expandFairsharePolicy(fairsharePolicy []interface{}) *duplosdk.DuploAwsFairsharePolicy {
	if len(fairsharePolicy) == 0 || fairsharePolicy[0] == nil {
		return nil
	}

	tfMap, ok := fairsharePolicy[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &duplosdk.DuploAwsFairsharePolicy{
		ComputeReservation: tfMap["compute_reservation"].(int),
		ShareDecaySeconds:  tfMap["share_decay_seconds"].(int),
	}

	shareDistributions := tfMap["share_distribution"].(*schema.Set).List()

	fairsharePolicyShareDistributions := []duplosdk.DuploAwsFairsharePolicyShareDistribution{}

	for _, shareDistribution := range shareDistributions {
		data := shareDistribution.(map[string]interface{})

		schedulingPolicyConfig := duplosdk.DuploAwsFairsharePolicyShareDistribution{
			ShareIdentifier: data["share_identifier"].(string),
			WeightFactor:    data["weight_factor"].(float64),
		}
		fairsharePolicyShareDistributions = append(fairsharePolicyShareDistributions, schedulingPolicyConfig)
	}

	result.ShareDistribution = &fairsharePolicyShareDistributions

	return result
}
