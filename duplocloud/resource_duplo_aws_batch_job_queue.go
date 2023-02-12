package duplocloud

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAwsBatchJobQueueSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the aws batch Job queue will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Specifies the name of the Job queue.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the Job queue.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"compute_environments": {
			Description: "Specifies the set of compute environments mapped to a job queue and their order. The position of the compute environments in the list will dictate the order.",
			Type:        schema.TypeList,
			Required:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"priority": {
			Description: "The priority of the job queue. Job queues with a higher priority are evaluated first when associated with the same compute environment.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"scheduling_policy_arn": {
			Description: "The ARN of the fair share scheduling policy. If this parameter is specified, the job queue uses a fair share scheduling policy. If this parameter isn't specified, the job queue uses a first in, first out (FIFO) scheduling policy. After a job queue is created, you can replace but can't remove the fair share scheduling policy.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"state": {
			Description:  "The state of the job queue. Must be one of: `ENABLED` or `DISABLED`",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"ENABLED", "DISABLED"}, true),
		},
		"arn": {
			Description: "The Amazon Resource Name of the Job queue.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"wait_for_deployment": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
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

func resourceAwsBatchJobQueue() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_batch_job_queue` manages an aws batch Job queue in Duplo.",

		ReadContext:   resourceAwsBatchJobQueueRead,
		CreateContext: resourceAwsBatchJobQueueCreate,
		UpdateContext: resourceAwsBatchJobQueueUpdate,
		DeleteContext: resourceAwsBatchJobQueueDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAwsBatchJobQueueSchema(),
	}
}

func resourceAwsBatchJobQueueRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsBatchJobQueueIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsBatchJobQueueRead(%s, %s): start", tenantID, name)

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	jq, clientErr := c.AwsBatchJobQueueGet(tenantID, fullName)
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
	flattenBatchJobQueue(d, c, jq, tenantID)

	log.Printf("[TRACE] resourceAwsBatchJobQueueRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsBatchJobQueueCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsBatchJobQueueCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	rq := expandAwsBatchJobQueue(d)
	err = c.AwsBatchJobQueueCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s aws batch Job queue '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws batch Job queue", id, func() (interface{}, duplosdk.ClientError) {
		return c.AwsBatchJobQueueGet(tenantID, fullName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_for_deployment") == nil || d.Get("wait_for_deployment").(bool) {
		err = batchJobQueueUntilValid(ctx, c, tenantID, fullName, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAwsBatchJobQueueRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsBatchJobQueueCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAwsBatchJobQueueUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	arn := d.Get("arn").(string)
	log.Printf("[TRACE] resourceAwsBatchJobQueueUpdate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	if d.HasChanges("compute_environments", "priority", "scheduling_policy_arn", "state") {
		name := d.Get("name").(string)
		updateInput := &duplosdk.DuploAwsBatchJobQueue{
			ComputeEnvironmentOrder: createComputeEnvironmentOrder(d.Get("compute_environments").([]interface{})),
			JobQueue:                fullName,
			Priority:                d.Get("priority").(int),
			State:                   &duplosdk.DuploStringValue{Value: d.Get("state").(string)},
			JobQueueArn:             arn,
		}
		if d.HasChange("scheduling_policy_arn") {
			if v, ok := d.GetOk("scheduling_policy_arn"); ok {
				updateInput.SchedulingPolicyArn = v.(string)
			} else {
				return diag.Errorf("Cannot remove the fair share scheduling policy")
			}
		} else {
			// if a queue is a FIFO queue, SchedulingPolicyArn should not be set. Error is "Only fairshare queue can have scheduling policy"
			// hence, check for scheduling_policy_arn and set it in the inputs only if it exists already
			if v, ok := d.GetOk("scheduling_policy_arn"); ok {
				updateInput.SchedulingPolicyArn = v.(string)
			}
		}
		log.Printf("[DEBUG] Updating Batch Job queue: %v", updateInput)
		err := c.AwsBatchJobQueueUpdate(tenantID, updateInput)
		if err != nil {
			return diag.Errorf("Error creating tenant %s aws batch Job queue '%s': %s", tenantID, name, err)
		}

		diags := waitForResourceToBePresentAfterCreate(ctx, d, "aws batch Job queue", fullName, func() (interface{}, duplosdk.ClientError) {
			return c.AwsBatchJobQueueGet(tenantID, fullName)
		})
		if diags != nil {
			return diags
		}
		if d.Get("wait_for_deployment") == nil || d.Get("wait_for_deployment").(bool) {
			err := batchJobQueueUntilValid(ctx, c, tenantID, fullName, d.Timeout("create"))
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return nil
}

func resourceAwsBatchJobQueueDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsBatchJobQueueIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsBatchJobQueueDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)

	jq, err := c.AwsBatchJobQueueGet(tenantID, fullName)
	if err != nil {
		return diag.FromErr(err)
	}
	if jq == nil {
		return nil
	}

	if jq.State != nil && jq.State.Value == "ENABLED" {
		log.Printf("[TRACE] Disable batch Job queue before delete. (%s, %s): start", tenantID, fullName)

		clientErr := c.AwsBatchJobQueueDisable(tenantID, fullName)
		if clientErr != nil {
			if clientErr.Status() == 404 {
				return nil
			}
			return diag.Errorf("Unable to disable tenant %s aws  batch Job queue '%s': %s", tenantID, fullName, clientErr)
		}
		err = batchJobQueueUntilDisabled(ctx, c, tenantID, fullName, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
		// time.Sleep(time.Duration(20) * time.Second)
	}

	clientErr := c.AwsBatchJobQueueDelete(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s aws batch Job queue '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "aws batch Job queue", id, func() (interface{}, duplosdk.ClientError) {
		return c.AwsBatchJobQueueGet(tenantID, fullName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsBatchJobQueueDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAwsBatchJobQueue(d *schema.ResourceData) *duplosdk.DuploAwsBatchJobQueue {

	input := &duplosdk.DuploAwsBatchJobQueue{
		JobQueueName:            d.Get("name").(string),
		Priority:                d.Get("priority").(int),
		State:                   &duplosdk.DuploStringValue{Value: d.Get("state").(string)},
		ComputeEnvironmentOrder: createComputeEnvironmentOrder(d.Get("compute_environments").([]interface{})),
	}

	if v, ok := d.GetOk("scheduling_policy_arn"); ok {
		input.SchedulingPolicyArn = v.(string)
	}

	if v, ok := d.GetOk("tags"); ok && v != nil {
		input.Tags = expandAsStringMap("tags", d)
	}

	return input
}

func parseAwsBatchJobQueueIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenBatchJobQueue(d *schema.ResourceData, c *duplosdk.Client, duplo *duplosdk.DuploAwsBatchJobQueue, tenantId string) diag.Diagnostics {
	log.Printf("[TRACE]flattenBatchJobQueue... Start ")
	prefix, err := c.GetDuploServicesPrefix(tenantId)
	if err != nil {
		return diag.FromErr(err)
	}
	name, _ := duplosdk.UnprefixName(prefix, duplo.JobQueueName)
	d.Set("tenant_id", tenantId)
	d.Set("name", name)
	d.Set("priority", duplo.Priority)
	d.Set("arn", duplo.JobQueueArn)
	d.Set("fullname", duplo.JobQueueName)
	d.Set("scheduling_policy_arn", duplo.SchedulingPolicyArn)
	d.Set("tags", duplo.Tags)
	if duplo.State != nil {
		d.Set("state", duplo.State.Value)
	}
	if duplo.Status != nil {
		d.Set("status", duplo.Status.Value)
	}
	d.Set("status_reason", duplo.StatusReason)

	computeEnvironments := make([]string, 0, len(*duplo.ComputeEnvironmentOrder))

	sort.Slice(*duplo.ComputeEnvironmentOrder, func(i, j int) bool {
		return (*duplo.ComputeEnvironmentOrder)[i].Order < (*duplo.ComputeEnvironmentOrder)[j].Order
	})

	for _, computeEnvironmentOrder := range *duplo.ComputeEnvironmentOrder {
		computeEnvironments = append(computeEnvironments, computeEnvironmentOrder.ComputeEnvironment)
	}

	if err := d.Set("compute_environments", computeEnvironments); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE]flattenBatchJobQueue... End ")
	return nil
}

func batchJobQueueUntilDisabled(ctx context.Context, c *duplosdk.Client, tenantID string, fullname string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AwsBatchJobQueueGet(tenantID, fullname)
			log.Printf("[TRACE] Batch Job queue state is (%s).", rp.State.Value)
			status := "pending"
			if err == nil {
				if rp.State.Value == "DISABLED" && rp.Status.Value == "VALID" {
					status = "ready"
				} else {
					status = "pending"
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] batchJobQueueUntilDisabled(%s, %s)", tenantID, fullname)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func batchJobQueueUntilValid(ctx context.Context, c *duplosdk.Client, tenantID string, fullname string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.AwsBatchJobQueueGet(tenantID, fullname)
			log.Printf("[TRACE] Batch Job queue status is (%s).", rp.Status.Value)
			status := "pending"
			if err == nil {
				if rp.Status.Value == "VALID" {
					status = "ready"
				} else {
					status = "pending"
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] batchJobQueueUntilValid(%s, %s)", tenantID, fullname)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func createComputeEnvironmentOrder(order []interface{}) *[]duplosdk.DuploAwsBatchComputeEnvironmentOrder {
	envs := make([]duplosdk.DuploAwsBatchComputeEnvironmentOrder, 0, len(order))
	for i, env := range order {
		envs = append(envs, duplosdk.DuploAwsBatchComputeEnvironmentOrder{
			Order:              i,
			ComputeEnvironment: env.(string),
		})
	}
	return &envs
}
