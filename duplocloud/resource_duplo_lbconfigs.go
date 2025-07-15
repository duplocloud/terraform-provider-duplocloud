package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
			Description: "The backend protocol associated with this load balancer configuration.\n" +
				"Supported protocol based on lb_type:\n\n" +
				"	- `0 (ELB)`: HTTP, HTTPS, TCP, UDP\n" +
				"	- `1 (ALB)` : HTTP, HTTPS\n" +
				"	- `3 (K8S Service w/ Cluster IP)`: TCP, UDP\n" +
				"	- `4 (K8S Service w/ Node Port)` : TCP, UDP\n" +
				"	- `5 (Azure Shared Application Gateway)`: HTTP, HTTPS\n" +
				"	- `6 (NLB)` : TCP, UDP, TLS\n" +
				"	- `7 (Target Group Only)` : HTTP, HTTPS, TCP, UDP, TLS\n",
			Type:             schema.TypeString,
			Required:         true,
			DiffSuppressFunc: diffSuppressStringCase,
			ValidateFunc:     validation.StringInSlice([]string{"HTTP", "HTTPS", "TCP", "UDP", "TLS"}, true),
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
			Description: "The frontend port associated with this load balancer configuration. Required if `lb_type` is not `7`.",
			Type:        schema.TypeInt,
			Optional:    true,
			Computed:    true,
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
			Description:      "Is used for communication between the load balancer and the target instances. This field is used to set protocol version for ALB load balancer. Only applicable when protocol is HTTP or HTTPS. The protocol version. Specify GRPC to send requests to targets using gRPC. Specify HTTP2 to send requests to targets using HTTP/2. The default is HTTP1, which sends requests to targets using HTTP/1.1",
			Type:             schema.TypeString,
			Optional:         true,
			DiffSuppressFunc: diffSuppressStringCase,
			ValidateFunc:     validation.StringInSlice([]string{"HTTP1", "HTTP2", "GRPC"}, true),
			Computed:         true,
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
			Description: "Set to true if the service for which the load balancer is being created is hosted on a docker native host, which is managed directly by DuploCloud, or false if the service is hosted on a cloud-provided platform like EKS, AKS, GKE, ECS, etc. The `duplocloud_native_hosts` data source lists the native hosts in a DuploCloud Tenant",
			Type:        schema.TypeBool,
			Computed:    true,
			Optional:    true,
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
		"allow_global_access": {
			Description: "Applicable for internal lb.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"skip_http_to_https": {
			Description: "Skip http to https.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"health_check": {
			Description: "Health Check configuration block.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"healthy_threshold": {
						Description:  "Number of consecutive health checks successes required before considering an unhealthy target healthy.",
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      3,
						ValidateFunc: validation.IntBetween(2, 10),
					},
					"unhealthy_threshold": {
						Description:  "Number of consecutive health check failures required before considering the target unhealthy.",
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      3,
						ValidateFunc: validation.IntBetween(2, 10),
					},
					"timeout": {
						Description:  "Amount of time, in seconds, during which no response means a failed health check.",
						Type:         schema.TypeInt,
						Optional:     true,
						Computed:     true,
						ValidateFunc: validation.IntBetween(2, 120),
					},
					"interval": {
						Description: "Approximate amount of time, in seconds, between health checks of an individual target. Minimum value 5 seconds, Maximum value 300 seconds.",
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     30,
					},
					"http_success_codes": {
						Description: "Response codes to use when checking for a healthy responses from a target. You can specify multiple values (for example, \"200,202\" for HTTP(s)) or a range of values (for example, \"200-299\"). Required for HTTP/HTTPS ALB. Only applies to Application Load Balancers (i.e., HTTP/HTTPS) not Network Load Balancers (i.e., TCP).",
						Type:        schema.TypeString,
						Computed:    true,
						Optional:    true,
					},
					"grpc_success_codes": {
						Description: "Response codes to use when checking for a healthy responses from a target. You can specify multiple values (for example, \"0,12\" for GRPC) or a range of values (for example, \"0-99\"). Required for GRPC ALB. Only applies to Application Load Balancers (i.e., GRPC) not Network Load Balancers (i.e., TCP).",
						Type:        schema.TypeString,
						Computed:    true,
						Optional:    true,
					},
				},
			},
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
		Schema:        duploServiceLbConfigsSchema(),
		CustomizeDiff: validateLBConfigParameters,
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
					Protocol:                  strings.ToUpper(lbc["protocol"].(string)),
					Port:                      lbc["port"].(string),
					HealthCheckURL:            lbc["health_check_url"].(string),
					CertificateArn:            lbc["certificate_arn"].(string),
					IsNative:                  lbc["is_native"].(bool),
					IsInternal:                lbc["is_internal"].(bool),
					ExternalTrafficPolicy:     lbc["external_traffic_policy"].(string),
					SetIngressHealthCheck:     lbc["set_ingress_health_check"].(bool),
					ExtraSelectorLabels:       keyValueFromStateList("extra_selector_label", lbc),
					SkipHttpToHttps:           lbc["skip_http_to_https"].(bool),
				}
				if v, ok := lbc["backend_protocol_version"]; ok && item.LbType == 1 {
					item.BeProtocolVersion = strings.ToUpper(v.(string))
				}
				if v, ok := lbc["health_check"]; ok && len(v.([]interface{})) > 0 {
					healthcheck := v.([]interface{})[0].(map[string]interface{})

					item.HealthCheckConfig = &duplosdk.DuploLbHealthCheckConfig{}
					item.HealthCheckConfig.HealthyThresholdCount = healthcheck["healthy_threshold"].(int)
					item.HealthCheckConfig.UnhealthyThresholdCount = healthcheck["unhealthy_threshold"].(int)
					item.HealthCheckConfig.HealthCheckTimeoutSeconds = healthcheck["timeout"].(int)
					item.HealthCheckConfig.LbHealthCheckIntervalSecondsype = healthcheck["interval"].(int)
					item.HealthCheckConfig.HttpSuccessCode = healthcheck["http_success_codes"].(string)
					item.HealthCheckConfig.GrpcSuccessCode = healthcheck["grpc_success_codes"].(string)
				}

				// if lbType is 7, then external_port is not required
				externalPortValue, exists := lbc["external_port"]
				if item.LbType != 7 && (!exists || externalPortValue.(int) == 0) {
					return diag.Errorf("'external_port' is required when 'lb_type' is set to %v", item.LbType)
				} else if exists {
					externalPort := externalPortValue.(int)
					item.ExternalPort = &externalPort
				} else {
					item.ExternalPort = nil
				}

				if item.LbType == 5 {
					item.HostNames = &[]string{lbc["host_name"].(string)}
				}
				if v, ok := lbc["custom_cidr"]; ok && v != nil && len(v.([]interface{})) > 0 && item.LbType == 6 {
					item.CustomCidrs = expandStringList(v.([]interface{}))
				}
				if item.IsInternal {
					item.AllowGlobalAccess = lbc["allow_global_access"].(bool)
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
		time.Sleep(140 * time.Second)
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
	stateConf := &retry.StateChangeConf{
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
				lbDetails, lberr := c.TenantGetLbDetailsInServiceNew(tenantID, name)
				if lberr != nil && lberr.Status() != 404 {
					return nil, "error", lberr
				}
				if lbDetails != nil {
					if val, ok := lbDetails["State"]; ok {
						if val != "active" {
							return name, "pending", nil

						}
					}
					return name, "ready", nil

				}
				log.Printf("[DEBUG] lbDetails %+v", lbDetails)

				// Detect a not-ready cloud load balancer.
				//	if lbdetails == nil || lbdetails.State == nil || lbdetails.State.Code == nil || lbdetails.State.Code.Value != "active" {
				//		return name, "pending", nil
				//	}
			} else {
				return name, "ready", nil
			}
			// If we got this far, we either have no cloud LB, or it's ready.
			return name, "pending", nil
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
		"protocol":                    strings.ToUpper(lb.Protocol),
		"port":                        lb.Port,
		"host_port":                   lb.HostPort,
		"external_port":               lb.ExternalPort,
		"is_infra_deployment":         lb.IsInfraDeployment,
		"dns_name":                    lb.DnsName,
		"certificate_arn":             lb.CertificateArn,
		"cloud_name":                  lb.CloudName,
		"health_check_url":            lb.HealthCheckURL,
		"external_traffic_policy":     lb.ExternalTrafficPolicy,
		"index":                       lb.LbIndex,
		"frontend_ip":                 lb.FrontendIP,
		"is_native":                   lb.IsNative,
		"is_internal":                 lb.IsInternal,
		"set_ingress_health_check":    lb.SetIngressHealthCheck,
		"extra_selector_label":        keyValueToState("extra_selector_label", lb.ExtraSelectorLabels),
		"target_group_arn":            lb.TgArn,
		"custom_cidr":                 lb.CustomCidrs,
		"allow_global_access":         lb.AllowGlobalAccess,
		"skip_http_to_https":          lb.SkipHttpToHttps,
		"backend_protocol_version":    strings.ToUpper(lb.BeProtocolVersion),
	}

	if lb.HealthCheckConfig != nil {
		healthcheckConfig := map[string]interface{}{
			"healthy_threshold":   lb.HealthCheckConfig.HealthyThresholdCount,
			"unhealthy_threshold": lb.HealthCheckConfig.UnhealthyThresholdCount,
			"timeout":             lb.HealthCheckConfig.HealthCheckTimeoutSeconds,
			"interval":            lb.HealthCheckConfig.LbHealthCheckIntervalSecondsype,
			"http_success_codes":  lb.HealthCheckConfig.HttpSuccessCode,
			"grpc_success_codes":  lb.HealthCheckConfig.GrpcSuccessCode,
		}
		m["health_check"] = []map[string]interface{}{healthcheckConfig}
	}

	if lb.LbType == 5 && lb.HostNames != nil && len(*lb.HostNames) > 0 {
		log.Printf("[DEBUG] HostNames... %v", lb.HostNames)
		m["host_name"] = (*lb.HostNames)[0]
	}
	if lb.LbType != 1 {
		m["backend_protocol_version"] = "HTTP1"

	}

	log.Printf("[DEBUG] flattenDuploServiceLbConfiguration... End")
	return m
}

func validateLBConfigParameters(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {

	lbconfs := diff.Get("lbconfigs").([]interface{})
	for _, lb := range lbconfs {
		m := lb.(map[string]interface{})
		pr := m["protocol"].(string)
		p := strings.ToLower(pr)
		b := m["backend_protocol_version"].(string)
		bp := strings.ToLower(b)

		lb, ok := m["lb_type"].(int)
		if ok && lb != 1 && bp != "" && bp != "http1" {
			return fmt.Errorf("backend_protocol_version field is available only for ALB for others load balancer type use protocol")

		}
		//if ok && lb == 1 && bp == "" {
		//	return fmt.Errorf("backend_protocol_version is a required field for ALB load balancer type")
		//}
		if p == "http" && bp == "grpc" {
			return fmt.Errorf("cannot set backend_protocol_version = %s with protocol= %s", bp, pr)
		}

		if (lb == 1 || lb == 5) && (p != "http" && p != "https") {
			return fmt.Errorf("protocol = %s not supported for lb_type=%d", pr, lb)
		}
		if lb == 6 && (p != "tcp" && p != "udp" && p != "tls") {
			return fmt.Errorf("protocol = %s not supported for lb_type=%d", pr, lb)
		}
		if (lb == 3 || lb == 4) && (p != "tcp" && p != "udp") {
			return fmt.Errorf("protocol = %s not supported for lb_type=%d", pr, lb)
		}
		if lb == 0 && p == "tls" {
			return fmt.Errorf("protocol = %s not supported for lb_type=%d", pr, lb)
		}
		healthCheck := m["health_check"].([]interface{})
		if len(healthCheck) > 0 {
			hm := healthCheck[0].(map[string]interface{})
			if hm["timeout"].(int) >= hm["interval"].(int) {
				return fmt.Errorf("health check timeout must be less than health check interval")
			}
		}
	}
	return nil
}
