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

func duploLbConfigSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"replication_controller_name": {
			Description: "The name of the duplo service.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "The name of the duplo service.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"lb_type": {
			Description: "The numerical index of the type of load balancer configuration to create.\n" +
				"Should be one of:\n\n" +
				"   - `0` : ELB (Classic Load Balancer)\n" +
				"   - `1` : ALB (Application Load Balancer)\n" +
				"   - `2` : Health-check Only (No Load Balancer)\n" +
				"   - `3` : K8S Service w/ Cluster IP (No Load Balancer)\n" +
				"   - `4` : K8S Service w/ Node Port (No Load Balancer)\n" +
				"   - `5` : Azure Shared Application Gateway\n" +
				"   - `6` : NLB (Network Load Balancer)\n" +
				"   - `7` : Target Group Only\n",
			Type:     schema.TypeInt,
			Required: true,
			ForceNew: true,
		},
		"protocol": {
			Description: "The backend protocol associated with this load balancer configuration.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"port": {
			Description: "The backend port associated with this load balancer configuration.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"host_port": {
			Description: "The automatically assigned host port.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"external_port": {
			Description: "The frontend port associated with this load balancer configuration.",
			Type:        schema.TypeInt,
			Required:    true,
		},
		"custom_cidr": {
			Description: "Specify CIDR Values. This is applicable only for Network Load Balancer if `lb_type` is `6`.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"is_infra_deployment": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"dns_name": {
			Description: "The DNS name of the cloud load balancer (if applicable).",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"certificate_arn": {
			Description: "The ARN of an ACM certificate to associate with this load balancer.  Only applicable for HTTPS.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
		"cloud_name": {
			Description: "The name of the cloud load balancer (if applicable).",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"health_check_url": {
			Description: "The health check URL to associate with this load balancer configuration.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
		"external_traffic_policy": {
			Description: "Only for K8S Node Port (`lb_type = 4`) or load balancers in Kubernetes.  Set the kubernetes service `externalTrafficPolicy` attribute.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"extra_selector_label": {
			Description: "Only for K8S services or load balancers in Kubernetes.  Sets an additional selector label to narrow which pods can receive traffic.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Elem:        KeyValueSchema(),
		},
		"set_ingress_health_check": {
			Description: "Only for K8S services or load balancers in Kubernetes.  Set to `true` to set health check annotations for ingress.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"backend_protocol_version": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"frontend_ip": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"is_internal": {
			Description: "Whether or not to create an internal load balancer.",
			Type:        schema.TypeBool,
			Computed:    true,
			Optional:    true,
		},
		"is_native": {
			Type:     schema.TypeBool,
			Computed: true,
			Optional: true,
		},
		"host_name": {
			Description: "(Azure Only) Set only if Azure Shared Application Gateway is used (`lb_type = 5`).",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"target_group_arn": {
			Description: "The ARN of the Target Group to which to route traffic.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"index": {
			Description: "The load balancer Index.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
	}
}

// SCHEMA for resource crud
func resourceDuploServiceLbConfigs() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_duplo_service_lbconfigs` manages load balancer configuration(s) for a container-based service in Duplo.\n\n" +
			"NOTE: For Amazon ECS services, see the `duplocloud_ecs_service` resource.",

		ReadContext:   resourceDuploServiceLbConfigsRead,
		CreateContext: resourceDuploServiceLBConfigsCreate,
		UpdateContext: resourceDuploServiceLBConfigsUpdate,
		DeleteContext: resourceDuploServiceLbConfigsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploServiceLbConfigsSchema(),
	}
}

// DuploServiceLBConfigsSchema returns a Terraform resource schema for a service's load balancer
func duploServiceLbConfigsSchema() map[string]*schema.Schema {
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
				Schema: duploLbConfigSchema(),
			},
		},
	}
}

// READ resource
func resourceDuploServiceLbConfigsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploServiceLBConfigsRead: start")

	// Parse the identifying attributes
	tenantID, name := parseDuploServiceLbConfigsIdParts(d.Id())
	if tenantID == "" || name == "" {
		return diag.Errorf("Invalid resource ID: %s", d.Id())
	}

	// Get the object from Duplo, detecting a missing object
	var err error
	c := m.(*duplosdk.Client)
	list, err := c.ReplicationControllerLbConfigurationList(tenantID, name)
	if list == nil || len(*list) == 0 {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s service '%s' load balancer configs: %s", tenantID, name, err)
	}
	// Apply the TF state
	d.Set("tenant_id", tenantID)
	d.Set("replication_controller_name", name)
	// Handle each LB config
	lbconfigs := make([]interface{}, 0, len(*list))
	isCloudLb := false
	for _, lb := range *list {
		if lb.LbType != 2 && lb.LbType != 3 && lb.LbType != 4 {
			isCloudLb = true
		}
		lbconfigs = append(lbconfigs, flattenDuploServiceLbConfiguration(&lb))
	}
	if err = d.Set("lbconfigs", lbconfigs); err != nil {
		return diag.FromErr(err)
	}
	// Handle a cloud load balancer.
	if isCloudLb {
		lbdetails, lberr := c.TenantGetLbDetailsInService(tenantID, name)
		if lberr != nil && lberr.Status() != 404 {
			return diag.FromErr(err)
		}

		if lbdetails != nil {
			d.Set("arn", lbdetails.LoadBalancerArn)
			if lbdetails.State != nil && lbdetails.State.Code != nil {
				d.Set("status", lbdetails.State.Code.Value)
			}
		}
	}

	log.Printf("[TRACE] resourceDuploServiceLBConfigsRead: end")
	return nil
}

// CREATE resource
func resourceDuploServiceLBConfigsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceDuploServiceLBConfigsCreate: start")
	diags := resourceDuploServiceLBConfigsCreateOrUpdate(ctx, d, m, false)
	log.Printf("[TRACE] resourceDuploServiceLBConfigsCreate: end")
	return diags
}

// UPDATE resource
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
	var list []duplosdk.DuploLbConfiguration

	// Append all load balancer configs to the request.
	if v, ok := d.GetOk("lbconfigs"); ok {
		lbconfigs := v.([]interface{})

		if len(lbconfigs) > 0 {
			list = make([]duplosdk.DuploLbConfiguration, 0, len(lbconfigs))

			for _, vLbc := range lbconfigs {
				lbc := vLbc.(map[string]interface{})

				item := duplosdk.DuploLbConfiguration{
					TenantId:                  tenantID,
					ReplicationControllerName: name,
					LbType:                    lbc["lb_type"].(int),
					Protocol:                  lbc["protocol"].(string),
					Port:                      lbc["port"].(string),
					ExternalPort:              lbc["external_port"].(int),
					HealthCheckURL:            lbc["health_check_url"].(string),
					CertificateArn:            lbc["certificate_arn"].(string),
					IsNative:                  lbc["is_native"].(bool),
					IsInternal:                lbc["is_internal"].(bool),
					ExternalTrafficPolicy:     lbc["external_traffic_policy"].(string),
					SetIngressHealthCheck:     lbc["set_ingress_health_check"].(bool),
					ExtraSelectorLabels:       keyValueFromStateList("extra_selector_label", lbc),
				}
				if item.LbType == 5 {
					item.HostNames = &[]string{lbc["host_name"].(string)}
				}
				if v, ok := lbc["custom_cidr"]; ok && v != nil && len(v.([]interface{})) > 0 && item.LbType == 6 {
					item.CustomCidrs = expandStringList(v.([]interface{}))
				}
				list = append(list, item)
			}
		}
	}

	// Post the object to Duplo
	id := fmt.Sprintf("v2/subscriptions/%s/ServiceLBConfigsV2/%s", tenantID, name)
	c := m.(*duplosdk.Client)
	err = c.ReplicationControllerLbConfigurationBulkUpdate(tenantID, name, &list)
	if err != nil {
		return diag.Errorf("Error applying Duplo service '%s' load balancer configs: %s", id, err)
	}
	if !updating {
		d.SetId(id)
	}

	// Optionally wait for the load balancers to be ready.
	if d.Get("wait_until_ready").(bool) && len(list) > 0 {
		err = duploServiceLbConfigsWaitUntilReady(ctx, c, tenantID, name)
		if err != nil {
			return diag.Errorf("Error waiting for Duplo service '%s' load balancer configs to be ready: %s", id, err)
		}
	}

	// Read the latest status from Duplo
	diags := resourceDuploServiceLbConfigsRead(ctx, d, m)
	log.Printf("[TRACE] resourceDuploServiceLBConfigsCreateOrUpdate: end")
	return diags
}

// DELETE resource
func resourceDuploServiceLbConfigsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name := parseDuploServiceLbConfigsIdParts(id)
	if tenantID == "" || name == "" {
		return diag.Errorf("Invalid resource ID: %s", id)
	}

	log.Printf("[TRACE] resourceDuploServiceLbConfigsDelete(%s, %s): start", tenantID, name)

	// Delete the object from Duplo
	c := m.(*duplosdk.Client)
	err := c.ReplicationControllerLbConfigurationBulkUpdate(tenantID, name, &[]duplosdk.DuploLbConfiguration{})
	if err != nil {
		return diag.Errorf("Error deleting Duplo service '%s' load balancer configs: %s", id, err)
	}

	// Wait for it to be deleted
	diags := waitForResourceToBeMissingAfterDelete(ctx, d, "duplo service load balancer configs", id, func() (interface{}, duplosdk.ClientError) {
		list, errget := c.ReplicationControllerLbConfigurationList(tenantID, name)
		if errget == nil && (list == nil || len(*list) == 0) {
			list = nil
		}
		return list, errget
	})

	// Wait 40 more seconds to deal with consistency issues.
	if diags == nil {
		time.Sleep(40 * time.Second)
	}

	log.Printf("[TRACE] resourceDuploServiceLbConfigsDelete(%s, %s): end", tenantID, name)
	return diags
}

func parseDuploServiceLbConfigsIdParts(id string) (tenantID, name string) {
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) == 5 {
		tenantID, name = idParts[2], idParts[4]
	}
	return
}

// DuploServiceLBConfigsWaitForCreation waits for creation of an service's load balancer by the Duplo API
func duploServiceLbConfigsWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID, name string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"missing", "pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {

			// Get the list of load balancers.
			list, err := c.ReplicationControllerLbConfigurationList(tenantID, name)
			if err != nil || list == nil || len(*list) == 0 {
				return nil, "missing", err
			}

			// Find a cloud load balancer, and get it's status.
			isCloudLb := false
			for _, lb := range *list {
				if lb.LbType == 7 && len(lb.TgArn) == 0 {
					return name, "pending", nil
				}

				if lb.LbType != 0 && lb.LbType != 2 && lb.LbType != 3 && lb.LbType != 4 && lb.LbType != 7 {
					isCloudLb = true
				}
			}
			if isCloudLb {
				lbdetails, lberr := c.TenantGetLbDetailsInService(tenantID, name)
				if lberr != nil && lberr.Status() != 404 {
					return nil, "error", lberr
				}

				// Detect a not-ready cloud load balancer.
				if lbdetails == nil || lbdetails.State == nil || lbdetails.State.Code == nil || lbdetails.State.Code.Value != "active" {
					return name, "pending", nil
				}
			}

			// If we got this far, we either have no cloud LB, or it's ready.
			return name, "ready", nil
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 10 * time.Second,
		Timeout:      10 * time.Minute,
	}
	log.Printf("[DEBUG] duploServiceLBConfigsWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func flattenDuploServiceLbConfiguration(lb *duplosdk.DuploLbConfiguration) map[string]interface{} {
	log.Printf("[DEBUG] flattenDuploServiceLbConfiguration... Start")
	m := map[string]interface{}{
		"name":                        lb.ReplicationControllerName,
		"replication_controller_name": lb.ReplicationControllerName,
		"lb_type":                     lb.LbType,
		"protocol":                    lb.Protocol,
		"port":                        lb.Port,
		"host_port":                   lb.HostPort,
		"external_port":               lb.ExternalPort,
		"is_infra_deployment":         lb.IsInfraDeployment,
		"dns_name":                    lb.DnsName,
		"certificate_arn":             lb.CertificateArn,
		"cloud_name":                  lb.CloudName,
		"health_check_url":            lb.HealthCheckURL,
		"external_traffic_policy":     lb.ExternalTrafficPolicy,
		"backend_protocol_version":    lb.BeProtocolVersion,
		"index":                       lb.LbIndex,
		"frontend_ip":                 lb.FrontendIP,
		"is_native":                   lb.IsNative,
		"is_internal":                 lb.IsInternal,
		"set_ingress_health_check":    lb.SetIngressHealthCheck,
		"extra_selector_label":        keyValueToState("extra_selector_label", lb.ExtraSelectorLabels),
		"target_group_arn":            lb.TgArn,
		"custom_cidr":                 lb.CustomCidrs,
	}

	if lb.LbType == 5 && lb.HostNames != nil && len(*lb.HostNames) > 0 {
		log.Printf("[DEBUG] HostNames... %v", lb.HostNames)
		m["host_name"] = (*lb.HostNames)[0]
	}

	log.Printf("[DEBUG] flattenDuploServiceLbConfiguration... End")
	return m
}
