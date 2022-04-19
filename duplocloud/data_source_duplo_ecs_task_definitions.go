package duplocloud

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"log"
	"terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ecsTaskDefinitionListElementSchema returns a Terraform resource schema for an element of an ECS Task Definition list
func ecsTaskDefinitionListElementSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"arn": {
			Description: "The ARN of the task definition.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"family": {
			Description: "The family the task definition.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "The short name of the task definition.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"revision": {
			Description: "The current revision of the task definition.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"latest": {
			Description: "The current revision of the task definition.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
	}
}

// SCHEMA for resource data/search
func dataSourceDuploEcsTaskDefinitions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDuploEcsTaskDefinitionsRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsUUID,
			},
			"family": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"family"},
			},
			"latest": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"task_definitions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: ecsTaskDefinitionListElementSchema(),
				},
			},
		},
	}
}

func dataSourceDuploEcsTaskDefinitionsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	tenantID := d.Get("tenant_id").(string)
	log.Printf("[TRACE] dataSourceDuploServicesRead(%s): start", tenantID)

	c := m.(*duplosdk.Client)

	// Get the tenant name and the resource prefix
	tenant, err := c.GetTenantForUser(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}
	prefix := fmt.Sprintf("duploservices-%s-", tenant.AccountName)

	// Determine the filtering to use.
	onlyLatest := d.Get("latest").(bool)
	onlyFamily := d.Get("family").(string)
	if onlyFamily == "" {
		if v, ok := d.GetOk("name"); ok && v.(string) != "" {
			onlyFamily = strings.Join([]string{prefix, v.(string)}, "")
		}
	}

	// Get the ARNs from Duplo.
	list, err := c.EcsTaskDefinitionListArns(tenantID)
	if err != nil {
		return diag.Errorf("Unable to list tenant %s ECS task definitions: %s", tenantID, err)
	}
	d.SetId(tenantID)

	// Build the list
	itemCount := 0
	if list != nil {
		itemCount = len(*list)
	}
	taskDefs := make([]map[string]interface{}, 0, itemCount)
	if itemCount > 0 {

		// Track latest revisions of items.
		latestRevisions := make(map[string]int64)

		// First, build an unfiltered list.
		unfiltered := make([]map[string]interface{}, 0, itemCount)
		for _, arn := range *list {

			// Split out the information from the ARN.
			arnParts := strings.SplitN(arn, "/", 2)
			idParts := strings.SplitN(arnParts[1], ":", 2)
			family := idParts[0]
			revision, err := strconv.ParseInt(idParts[1], 10, 0)
			if err != nil {
				return diag.Errorf("%s: failed to parse revision in ARN: %s", arn, err)
			}
			name := strings.TrimPrefix(family, prefix)

			// Produce the initial attributes (latest will be updated afterwards)
			taskDef := map[string]interface{}{
				"arn":      arn,
				"family":   family,
				"name":     name,
				"revision": revision,
				"latest":   false,
			}

			// Track the latest revision we have seen for this family.
			if latest, ok := latestRevisions[family]; !ok || latest < revision {
				latestRevisions[family] = revision
			}

			// Add it to the unfiltered list.
			unfiltered = append(unfiltered, taskDef)
		}

		// Now, mark the items that are "latest" and handle filtering.
		if itemCount > 0 {
			for _, taskDef := range unfiltered {

				// Mark a "latest" item.
				if taskDef["revision"] == latestRevisions[taskDef["family"].(string)] {
					taskDef["latest"] = true
				}

				// Filter this item
				if (onlyFamily == "" || taskDef["family"].(string) == onlyFamily) && (!onlyLatest || taskDef["latest"].(bool)) {
					taskDefs = append(taskDefs, taskDef)
				}
			}
		}
	}

	if err := d.Set("task_definitions", taskDefs); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] dataSourceDuploServicesRead(%s): end", tenantID)
	return nil
}
