package duplocloud

import (
	"context"
	"fmt"
	"strings"

	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// SCHEMA for resource crud
func resourceDuploServiceLBConfigs() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_duplo_service_lbconfigs` manages load balancer configuration(s) for a container-based service in Duplo.\n\n" +
			"NOTE: For Amazon ECS services, see the `duplocloud_ecs_service` resource.",

		ReadContext:   resourceDuploServiceLBConfigsRead,
		CreateContext: resourceDuploServiceLBConfigsCreate,
		UpdateContext: resourceDuploServiceLBConfigsUpdate,
		DeleteContext: resourceDuploServiceLBConfigsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploServiceLBConfigsSchema(),
	}
}

// DuploServiceLBConfigsSchema returns a Terraform resource schema for a service's load balancer
func duploServiceLBConfigsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that hosts the duplo service.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true, //switch tenant
			ValidateFunc: validation.IsUUID,
		},
		"replication_controller_name": {
			Description: "The name of the duplo service.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"arn": {
			Description: "The load balancer ARN.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"status": {
			Description: "The load balancer's current status.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"wait_until_ready": {
			Description:      "Whether or not to wait until Duplo considers all of the load balancers ready",
			Type:             schema.TypeBool,
			Optional:         true,
			Default:          true,
			DiffSuppressFunc: diffSuppressWhenNotCreating,
		},
		"lbconfigs": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"lb_type": {
						Description: "The numerical index of the type of load balancer configuration to create.\n" +
							"Should be one of:\n\n" +
							"   - `0` : ELB (Classic Load Balancer)\n" +
							"   - `1` : ALB (Application Load Balancer)\n" +
							"   - `2` : Health-check Only (No Load Balancer)\n" +
							"   - `3` : K8S Service w/ Cluster IP (No Load Balancer)\n" +
							"   - `4` : K8S Service w/ Node Port (No Load Balancer)\n" +
							"   - `5` : NLB (Network Load Balancer)\n",
						Type:     schema.TypeInt,
						Required: true,
						ForceNew: true,
					},
					"protocol": {
						Description: "The frontend protocol associated with this load balancer configuration.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"port": {
						Description: "The backend port associated with this load balancer configuration.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"external_port": {
						Description: "The frontend port associated with this load balancer configuration.",
						Type:        schema.TypeInt,
						Required:    true,
					},
					"health_check_url": {
						Description: "The health check URL to associate with this load balancer configuration.",
						Type:        schema.TypeString,
						Computed:    true,
						Optional:    true,
					},
					"certificate_arn": {
						Description: "The ARN of an ACM certificate to associate with this load balancer.  Only applicable for HTTPS.",
						Type:        schema.TypeString,
						Computed:    true,
						Optional:    true,
					},
					"replication_controller_name": {
						Type:       schema.TypeString,
						Computed:   true,
						Optional:   true,
						Deprecated: "Set the replication_controller_name field instead of lbconfigs.replication_controller_name",
					},
					"is_native": {
						Type:     schema.TypeBool,
						Computed: true,
						Optional: true,
					},
					"external_traffic_policy": {
						Description: "Only for K8S Node Port (`lb_type = 4`).  Set the kubernetes service `externalTrafficPolicy` attribute.",
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
					},
					"is_internal": {
						Description: "Whether or not to create an internal load balancer.",
						Type:        schema.TypeBool,
						Computed:    true,
						Optional:    true,
					},
				},
			},
		},
	}
}

/// READ resource
func resourceDuploServiceLBConfigsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploServiceLBConfigsRead: start")

	// Parse the identifying attributes
	tenantID, name := parseDuploServiceLBConfigsIdParts(d.Id())
	if tenantID == "" || name == "" {
		return diag.Errorf("Invalid resource ID: %s", d.Id())
	}

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.DuploServiceLBConfigsGet(tenantID, name)
	if duplo == nil || duplo.LBConfigs == nil || len(*duplo.LBConfigs) == 0 {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s service '%s' load balancer configs: %s", tenantID, name, err)
	}

	// Apply the TF state
	d.Set("tenant_id", duplo.TenantID)
	d.Set("replication_controller_name", duplo.ReplicationControllerName)
	d.Set("arn", duplo.Arn)
	d.Set("status", duplo.Status)
	d.Set("lbconfigs", flattenDuploServiceLBConfigurations(duplo.LBConfigs))

	log.Printf("[TRACE] resourceDuploServiceLBConfigsRead: end")
	return nil
}

/// CREATE resource
func resourceDuploServiceLBConfigsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploServiceLBConfigsCreate: start")
	diags := resourceDuploServiceLBConfigsCreateOrUpdate(ctx, d, m, false)
	log.Printf("[TRACE] resourceDuploServiceLBConfigsCreate: end")
	return diags
}

/// UPDATE resource
func resourceDuploServiceLBConfigsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploServiceLBConfigsUpdate: start")
	diags := resourceDuploServiceLBConfigsCreateOrUpdate(ctx, d, m, true)
	return diags
}

func resourceDuploServiceLBConfigsCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}, updating bool) diag.Diagnostics {
	var err error

	log.Printf("[TRACE] resourceDuploServiceLBConfigsCreateOrUpdate: start")

	// Start the build the reqeust.
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("replication_controller_name").(string)
	rq := duplosdk.DuploServiceLBConfigs{
		ReplicationControllerName: name,
		TenantID:                  tenantID,
		Arn:                       d.Get("arn").(string),
		Status:                    d.Get("status").(string),
	}

	// Append all load balancer configs to the request.
	if v, ok := d.GetOk("lbconfigs"); ok {
		lbconfigs := v.([]interface{})

		if len(lbconfigs) > 0 {
			var lbcs []duplosdk.DuploLBConfiguration

			for _, vLbc := range lbconfigs {
				lbc := vLbc.(map[string]interface{})
				lbcs = append(lbcs, duplosdk.DuploLBConfiguration{
					LBType:                    lbc["lb_type"].(int),
					Protocol:                  lbc["protocol"].(string),
					Port:                      lbc["port"].(string),
					ExternalPort:              lbc["external_port"].(int),
					HealthCheckURL:            lbc["health_check_url"].(string),
					CertificateArn:            lbc["certificate_arn"].(string),
					ReplicationControllerName: name,
					IsNative:                  lbc["is_native"].(bool),
					IsInternal:                lbc["is_internal"].(bool),
					ExternalTrafficPolicy:     lbc["external_traffic_policy"].(string),
				})
			}

			rq.LBConfigs = &lbcs
		}
	}

	// Post the object to Duplo
	id := fmt.Sprintf("v2/subscriptions/%s/ServiceLBConfigsV2/%s", tenantID, name)
	c := m.(*duplosdk.Client)
	_, err = c.DuploServiceLBConfigsCreateOrUpdate(tenantID, &rq, updating)
	if err != nil {
		return diag.Errorf("Error applying Duplo service '%s' load balancer configs: %s", id, err)
	}
	if !updating {
		d.SetId(id)
	}

	// Optionally wait for the load balancers to be ready.
	if d.Get("wait_until_ready").(bool) {
		err = duploServiceLBConfigsWaitUntilReady(ctx, c, tenantID, name)
		if err != nil {
			return diag.Errorf("Error waiting for Duplo service '%s' load balancer configs to be ready: %s", id, err)
		}
	}

	// Read the latest status from Duplo
	diags := resourceDuploServiceLBConfigsRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploServiceLBConfigsCreateOrUpdate: end")
	return diags
}

/// DELETE resource
func resourceDuploServiceLBConfigsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("[TRACE] resourceDuploServiceLBConfigsDelete: start")

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name := parseDuploServiceLBConfigsIdParts(id)
	if tenantID == "" || name == "" {
		return diag.Errorf("Invalid resource ID: %s", id)
	}

	// Delete the object from Duplo
	c := m.(*duplosdk.Client)

	// Check if service is exists.
	exists := c.DuploServiceLBConfigsExists(tenantID, name)

	if exists {
		err := c.DuploServiceLBConfigsDelete(tenantID, name)
		if err != nil {
			return diag.Errorf("Error deleting Duplo service '%s' load balancer configs: %s", id, err)
		}

		// Wait for it to be deleted
		diags := waitForResourceToBeMissingAfterDelete(ctx, d, "duplo service load balancer configs", id, func() (interface{}, duplosdk.ClientError) {
			rp, errget := c.DuploServiceLBConfigsGet(tenantID, name)
			if errget == nil && (rp == nil || rp.LBConfigs == nil || len(*rp.LBConfigs) == 0) {
				rp = nil
			}
			return rp, errget
		})

		// Wait 40 more seconds to deal with consistency issues.
		if diags == nil {
			time.Sleep(40 * time.Second)
		}
	}

	log.Printf("[TRACE] resourceDuploServiceLBConfigsDelete: end")
	return diags
}

func flattenDuploServiceLBConfigurations(list *[]duplosdk.DuploLBConfiguration) []interface{} {
	if list == nil {
		return []interface{}{}
	}

	lbconfigs := make([]interface{}, 0, len(*list))

	for _, lbc := range *list {
		lbconfigs = append(lbconfigs, map[string]interface{}{
			"lb_type":                     lbc.LBType,
			"protocol":                    lbc.Protocol,
			"port":                        lbc.Port,
			"external_port":               lbc.ExternalPort,
			"health_check_url":            lbc.HealthCheckURL,
			"certificate_arn":             lbc.CertificateArn,
			"replication_controller_name": lbc.ReplicationControllerName,
			"is_native":                   lbc.IsNative,
			"is_internal":                 lbc.IsInternal,
			"external_traffic_policy":     lbc.ExternalTrafficPolicy,
		})
	}

	return lbconfigs
}

func parseDuploServiceLBConfigsIdParts(id string) (tenantID, name string) {
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) == 5 {
		tenantID, name = idParts[2], idParts[4]
	}
	return
}

// DuploServiceLBConfigsWaitForCreation waits for creation of an service's load balancer by the Duplo API
func duploServiceLBConfigsWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID, name string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"missing", "pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.DuploServiceLBConfigsGet(tenantID, name)
			if err != nil || rp == nil {
				return nil, "missing", err
			}

			status := "pending"
			if rp.Status == "Ready" {
				status = "ready"
			}
			return rp, status, nil
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      20 * time.Minute,
	}
	log.Printf("[DEBUG] duploServiceLBConfigsWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
