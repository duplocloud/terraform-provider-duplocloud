package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const DEFAULT_INFRA = "default"

func infrastructureVnetSecurityGroupsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The security group ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The security group name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"read_only": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"type": {
				Description: "The type of security group.  Will be one of: `\"host\"` or `\"lb\"`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"rules": {
				Description: "Security group rules",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"priority": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"action": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"direction": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"source_port_range": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"source_address_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"source_rule_type": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"destination_rule_type": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func infrastructureVnetSubnetSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The subnet ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "The subnet name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cidr_block": {
				Description: "The subnet CIDR block.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"type": {
				Description: "The type of subnet.  Will be one of: `\"public\"` or `\"private\"`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"zone": {
				Description: "The Duplo zone that the subnet resides in.  Will be one of:  `\"A\"`, `\"B\"`, `\"C\"`, or `\"D\"`",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"tags": {
				Description: "The subnet's tags.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        KeyValueSchema(),
			},
		},
	}
}

// SCHEMA for resource crud
func resourceInfrastructure() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_infrastructure` manages an infrastructure in Duplo." +
			"<p>**DuploCloud infrastructure** refers to the cloud resources and configurations managed within the DuploCloud platform." +
			"It includes the setup, organization, and management of cloud services like networks, compute instances, databases, and other cloud-native services within a specific environment or tenant." +
			"</p>",

		ReadContext:   resourceInfrastructureRead,
		CreateContext: resourceInfrastructureCreate,
		UpdateContext: resourceInfrastructureUpdate,
		DeleteContext: resourceInfrastructureDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(50 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"infra_name": {
				Description:  "The name of the infrastructure.  Infrastructure names are globally unique and less than 13 characters.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(2, 30),
			},
			"account_id": {
				Description: "The cloud account ID.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
			},
			"cloud": {
				Description: "The numerical index of cloud provider to use for the infrastructure.\n" +
					"Should be one of:\n\n" +
					"   - `0` : AWS\n" +
					"   - `2` : Azure\n" +
					"   - `3` : Google\n",
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  0,
			},
			"region": {
				Description: "The cloud provider region.  The Duplo portal must have already been configured to support this region.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"azcount": {
				Description: "The number of availability zones.  Must be one of: `2`, `3`, or `4`. This is applicable only for AWS.",
				Type:        schema.TypeInt,
				ForceNew:    true,
				Optional:    true,
			},
			"enable_k8_cluster": {
				Description: "Whether or not to provision a kubernetes cluster.",
				Type:        schema.TypeBool,
				ForceNew:    true,
				Required:    true,
			},
			"is_serverless_kubernetes": {
				Description: "Whether or not to make GKE with autopilot.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"enable_ecs_cluster": {
				Description: "Whether or not to provision an ECS cluster.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"enable_container_insights": {
				Description: "Whether or not to enable container insights for an ECS cluster.",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"custom_data": {
				Description:   "A list of configuration settings to apply on creation, expressed as key / value pairs.",
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          KeyValueSchema(),
				Deprecated:    "The custom_data argument is only applied on creation, and is deprecated in favor of the settings argument.",
				ConflictsWith: []string{"setting"},
			},
			"setting": {
				Description:   "A list of configuration settings to manage, expressed as key / value pairs.",
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          KeyValueSchema(),
				ConflictsWith: []string{"custom_data"},
			},
			"delete_unspecified_settings": {
				Description: "Whether or not this resource should delete any settings not specified by this resource. " +
					"**WARNING:**  It is not recommended to change the default value of `false`.",
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"all_settings": {
				Description: "A complete list of configuration settings for this infrastructure, even ones not being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        KeyValueSchema(),
			},
			"specified_settings": {
				Description: "A list of configuration setting key being managed by this resource.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"address_prefix": {
				Description: "The CIDR to use for the VPC or VNet.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"subnet_cidr": {
				Description: "The CIDR subnet size (in bits) for the automatically created subnets. This is applicable only for AWS.",
				Type:        schema.TypeInt,
				ForceNew:    true,
				Optional:    true,
			},
			"subnet_name": {
				Description: "The name of the subnet. This is applicable only for Azure.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
			},
			"subnet_fullname": {
				Description: "The full name of the subnet. This is applicable only for Azure.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"subnet_address_prefix": {
				Description: "The address prefixe to use for the subnet. This is applicable only for Azure",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
			},
			"status": {
				Description: "The status of the infrastructure.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vpc_id": {
				Description: "The VPC or VNet ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vpc_name": {
				Description: "The VPC or VNet name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"private_subnets": {
				Description: "The private subnets for the VPC or VNet.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        infrastructureVnetSubnetSchema(),
			},
			"public_subnets": {
				Description: "The public subnets for the VPC or VNet.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        infrastructureVnetSubnetSchema(),
			},
			"security_groups": {
				Description: "The security groups for the VPC or VNet.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        infrastructureVnetSecurityGroupsSchema(),
			},
			"wait_until_deleted": {
				Description:      "Whether or not to wait until Duplo has destroyed the infrastructure.",
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: diffSuppressFuncIgnore,
			},
			"cluster_ip_cidr": {
				Description: "cluster IP CIDR defines a private IP address range used for internal Kubernetes services.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// READ resource
func resourceInfrastructureRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) < 4 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	name := idParts[3]

	log.Printf("[TRACE] resourceInfrastructureRead(%s): start", name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	missing, err := infrastructureRead(c, d, name)
	if err != nil {
		return diag.Errorf("Unable to retrieve infrastructure '%s': %s", name, err)
	}
	if missing {
		d.SetId("") // object missing
		return nil
	}

	log.Printf("[TRACE] resourceInfrastructureRead(%s): end", name)
	return nil
}

// CREATE resource
func resourceInfrastructureCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	rq := duploInfrastructureConfigFromState(d)

	log.Printf("[TRACE] resourceInfrastructureCreate(%s): start", rq.Name)

	diags := validateInfraSchema(d)
	if diags != nil {
		return diags
	}
	// Post the object to Duplo.
	c := m.(*duplosdk.Client)
	err = c.InfrastructureCreate(rq)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait up to 60 seconds for Duplo to be able to return the infrastructure details.
	id := fmt.Sprintf("v2/admin/InfrastructureV2/%s", rq.Name)
	diags = waitForResourceToBePresentAfterCreate(ctx, d, "infrastructure", id, func() (interface{}, duplosdk.ClientError) {
		return c.InfrastructureGetConfig(rq.Name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	// Then, wait until the infrastructure is completely ready.
	err = duploInfrastructureWaitUntilReady(ctx, c, rq.Name, d.Timeout("create"))
	if err != nil {
		return diag.FromErr(err)
	}

	diags = resourceInfrastructureRead(ctx, d, m)
	log.Printf("[TRACE] resourceInfrastructureCreate(%s): end", rq.Name)
	return diags
}

// UPDATE resource
func resourceInfrastructureUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	infraName, diags := parseInfrastructureId(d)
	if diags != nil {
		return diags
	}

	log.Printf("[TRACE] resourceInfrastructureUpdate(%s): start", infraName)
	c := m.(*duplosdk.Client)

	// Apply any ECS changes.
	if d.HasChanges("enable_ecs_cluster", "enable_container_insights") {
		rq := duplosdk.DuploInfrastructureECSConfigUpdate{
			EnableECSCluster:        d.Get("enable_ecs_cluster").(bool),
			EnableContainerInsights: d.Get("enable_container_insights").(bool),
		}
		err := c.InfrastructureUpdateECSConfig(infraName, rq)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Apply any settings changes.
	config, err := c.InfrastructureGetSetting(infraName)
	if err != nil {
		return diag.Errorf("Error retrieving infrastructure settings for '%s': %s", infraName, err)
	}
	if d.HasChange("setting") {
		var existing *[]duplosdk.DuploKeyStringValue
		if v, ok := getAsStringArray(d, "specified_settings"); ok && v != nil {
			existing = selectKeyValues(config.Setting, *v)
		} else {
			existing = &[]duplosdk.DuploKeyStringValue{}
		}

		// Collect the desired state of settings specified by the user.
		settings := keyValueFromState("setting", d)
		specified := make([]string, len(*settings))
		for i, kv := range *settings {
			specified[i] = kv.Key
		}
		d.Set("specified_settings", specified)

		// Apply the changes via Duplo
		if d.Get("delete_unspecified_settings").(bool) {
			err = c.InfrastructureReplaceSetting(duplosdk.DuploInfrastructureSetting{InfraName: infraName, Setting: settings})
		} else {
			err = c.InfrastructureChangeSetting(infraName, existing, settings)
		}
		if err != nil {
			return diag.Errorf("Error updating infrastructure settings for '%s': %s", infraName, err)
		}
	}

	// Wait for 60 seconds, at first.
	time.Sleep(time.Minute)

	// Then, wait until the infrastructure is completely ready.
	waitErr := duploInfrastructureWaitUntilReady(ctx, c, infraName, d.Timeout("update"))
	if waitErr != nil {
		return diag.FromErr(waitErr)
	}

	diags = resourceInfrastructureRead(ctx, d, m)
	log.Printf("[TRACE] resourceInfrastructureUpdate(%s): end", infraName)
	return diags
}

// DELETE resource
func resourceInfrastructureDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	infraName, diags := parseInfrastructureId(d)
	if diags != nil {
		return diags
	}

	log.Printf("[TRACE] resourceInfrastructureDelete(%s): start", infraName)

	c := m.(*duplosdk.Client)
	err := c.InfrastructureDelete(infraName)
	if err != nil {
		if err.Status() == 404 {
			return nil
		}
		return diag.FromErr(err)
	}

	// Wait for 20 minutes to allow infrastructure deletion.
	// TODO: wait for it completely deleted (add an API that will actually show the status)
	if d.Get("wait_until_deleted").(bool) {
		log.Printf("[TRACE] resourceInfrastructureDelete(%s): waiting for 20 minutes because 'wait_until_deleted' is 'true'", infraName)
		time.Sleep(time.Duration(20) * time.Minute)
	}

	log.Printf("[TRACE] resourceInfrastructureDelete(%s): end", infraName)
	return nil
}

func parseInfrastructureId(d *schema.ResourceData) (string, diag.Diagnostics) {
	id := d.Id()
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) < 4 {
		return "", diag.Errorf("Invalid resource ID: %s", id)
	}
	name := idParts[3]
	return name, nil
}

func duploInfrastructureConfigFromState(d *schema.ResourceData) duplosdk.DuploInfrastructureConfig {
	var IsServerlessKubernetes bool

	// to check if boolean which is optional and doesn't have a default value there is no replacement for GetOkExist
	// https://discuss.hashicorp.com/t/terraform-sdk-usage-which-out-of-get-getok-getokexists-with-boolean/41815/8

	if v, ok := d.GetOkExists("is_serverless_kubernetes"); ok { //nolint:all
		IsServerlessKubernetes = v.(bool)
	} else {
		if d.Get("cloud").(int) == 3 {
			IsServerlessKubernetes = true
		}
	}

	config := duplosdk.DuploInfrastructureConfig{
		Name:                    d.Get("infra_name").(string),
		AccountId:               d.Get("account_id").(string),
		Cloud:                   d.Get("cloud").(int),
		Region:                  d.Get("region").(string),
		AzCount:                 d.Get("azcount").(int),
		EnableK8Cluster:         d.Get("enable_k8_cluster").(bool),
		EnableECSCluster:        d.Get("enable_ecs_cluster").(bool),
		IsServerlessKubernetes:  IsServerlessKubernetes,
		EnableContainerInsights: d.Get("enable_container_insights").(bool),
		Vnet: &duplosdk.DuploInfrastructureVnet{
			AddressPrefix: d.Get("address_prefix").(string),
			SubnetCidr:    d.Get("subnet_cidr").(int),
			Subnets:       &[]duplosdk.DuploInfrastructureVnetSubnet{},
		},
	}

	//Azure -> if needed only there, this subnet should be added only in Azure
	if config.Cloud == 2 {
		subnet := duplosdk.DuploInfrastructureVnetSubnet{}
		if v, ok := d.GetOk("subnet_name"); ok {
			subnet.Name = v.(string)
		}
		if v, ok := d.GetOk("subnet_address_prefix"); ok {
			subnet.AddressPrefix = v.(string)
		}
		config.Vnet.Subnets = &[]duplosdk.DuploInfrastructureVnetSubnet{
			subnet,
		}
	}

	if config.Cloud == 3 && !config.IsServerlessKubernetes {
		config.ClusterIpv4Cidr = d.Get("cluster_ip_cidr").(string)
	}

	if d.HasChange("setting") {
		config.CustomData = keyValueFromState("setting", d)
	} else if d.HasChange("custom_data") {
		config.CustomData = keyValueFromState("custom_data", d)
	}

	return config
}

func duploInfrastructureWaitUntilReady(ctx context.Context, c *duplosdk.Client, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.InfrastructureGetConfig(name)
			status := "pending"
			if rp != nil && err == nil {
				log.Printf("[DEBUG] Infrastructure provisioning status is %s", rp.ProvisioningStatus)
				if rp.ProvisioningStatus == "Complete" || strings.Contains(rp.ProvisioningStatus, "Ready") {
					status = "ready"
				}
				return rp, status, nil
			}
			return nil, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 45 sec anyway
		PollInterval: 45 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] duploInfrastructureWaitUntilReady(%s)", name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func infrastructureRead(c *duplosdk.Client, d *schema.ResourceData, name string) (bool, error) {
	var infra *duplosdk.DuploInfrastructureConfig
	config, err := c.InfrastructureGetConfig(name)
	if err != nil {
		return false, err
	}
	// Once backend API is fixed for default infra, Remove this.
	if DEFAULT_INFRA == name {
		infra = config
	} else {
		infra, err = c.InfrastructureGet(name)
		if err != nil {
			return false, err
		}
		if config == nil || infra == nil {
			return true, nil // object missing
		}
	}

	d.Set("infra_name", infra.Name)
	d.Set("account_id", infra.AccountId)
	d.Set("cloud", infra.Cloud)
	d.Set("region", infra.Region)
	d.Set("azcount", infra.AzCount)
	d.Set("enable_k8_cluster", infra.EnableK8Cluster)
	d.Set("is_serverless_kubernetes", infra.IsServerlessKubernetes)
	d.Set("enable_ecs_cluster", infra.EnableECSCluster)
	d.Set("enable_container_insights", infra.EnableContainerInsights)
	d.Set("address_prefix", infra.Vnet.AddressPrefix)
	d.Set("subnet_cidr", infra.Vnet.SubnetCidr)
	d.Set("status", infra.ProvisioningStatus)

	d.Set("all_settings", keyValueToState("all_settings", config.CustomData))

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := getAsStringArray(d, "specified_settings"); ok && v != nil {
		d.Set("setting", keyValueToState("setting", selectKeyValues(config.CustomData, *v)))
	} else {
		d.Set("specified_settings", make([]interface{}, 0))
	}

	// Set extended infrastructure information.
	if config.Vnet != nil {
		d.Set("vpc_id", config.Vnet.ID)
		d.Set("vpc_name", config.Vnet.Name)

		if config.Vnet.Subnets != nil {
			publicSubnets := make([]map[string]interface{}, 0, len(*config.Vnet.Subnets))
			privateSubnets := make([]map[string]interface{}, 0, len(*config.Vnet.Subnets))

			for _, vnetSubnet := range *config.Vnet.Subnets {

				// Skip it unless it's a duplo managed subnet.
				isDuploSubnet := true // older systems do not return tags
				if vnetSubnet.Tags != nil {
					isDuploSubnet = false
					for _, tag := range *vnetSubnet.Tags {
						if tag.Key == "aws:cloudformation:stack-name" && strings.HasPrefix(tag.Value, "duplo") {
							isDuploSubnet = true
							break
						}
					}
				}
				if !isDuploSubnet {
					continue
				}
				zone := vnetSubnet.Zone
				subnetType := vnetSubnet.SubnetType
				// The server may or may not have the new fields.
				// If cloud is aws
				if config.Cloud == 0 {
					if strings.HasPrefix(vnetSubnet.Name, config.Name) {
						// Split on infra name first
						parts := strings.SplitN(vnetSubnet.Name, config.Name, 2)
						nameParts := strings.SplitN(parts[1], "-", 3)
						if zone == "" {
							zone = nameParts[1]
						}
						if subnetType == "" {
							subnetType = nameParts[2]
						}
					} else {
						nameParts := strings.SplitN(vnetSubnet.Name, " ", 2)
						if zone == "" {
							zone = nameParts[0]
						}
						if subnetType == "" {
							subnetType = nameParts[1]
						}
					}
					subnet := map[string]interface{}{
						"id":         vnetSubnet.ID,
						"name":       vnetSubnet.Name,
						"cidr_block": vnetSubnet.AddressPrefix,
						"type":       subnetType,
						"zone":       zone,
						"tags":       keyValueToState("tags", vnetSubnet.Tags),
					}

					if subnetType == "private" {
						privateSubnets = append(privateSubnets, subnet)
					} else if subnetType == "public" {
						publicSubnets = append(publicSubnets, subnet)
					}
				}

				if config.Cloud == 2 && c.IsAzureCustomPrefixesEnabled() {
					d.Set("subnet_fullname", c.TrimPrefixSuffixFromResourceName(vnetSubnet.Name, "subnet", true))
				} else {
					d.Set("subnet_fullname", vnetSubnet.Name)
				}

				d.Set("subnet_address_prefix", vnetSubnet.AddressPrefix)
			}

			d.Set("private_subnets", privateSubnets)
			d.Set("public_subnets", publicSubnets)
		}

		if config.Vnet.SecurityGroups != nil {
			securityGroups := make([]map[string]interface{}, 0, len(*config.Vnet.SecurityGroups))

			for _, vnetSG := range *config.Vnet.SecurityGroups {
				sg := map[string]interface{}{
					"id":        vnetSG.SystemId,
					"name":      vnetSG.Name,
					"read_only": vnetSG.ReadOnly,
					"type":      vnetSG.SgType,
				}
				sgRules := make([]map[string]interface{}, 0, len(*vnetSG.Rules))
				for _, rule := range *vnetSG.Rules {
					r := map[string]interface{}{
						"priority":              rule.Priority,
						"action":                rule.RuleAction,
						"direction":             rule.Direction,
						"protocol":              rule.Protocol,
						"source_port_range":     rule.SourcePortRange,
						"source_address_prefix": rule.SrcAddressPrefix,
						"source_rule_type":      rule.SrcRuleType,
						"destination_rule_type": rule.DstRuleType,
					}
					sgRules = append(sgRules, r)
				}
				sg["rules"] = sgRules
				securityGroups = append(securityGroups, sg)
			}
			d.Set("security_groups", securityGroups)
		}
	}

	return false, nil
}

func validateInfraSchema(d *schema.ResourceData) diag.Diagnostics {
	log.Printf("[TRACE] validateInfraSchema: start")
	cloud := d.Get("cloud").(int)

	if cloud == 0 {
		if _, ok := d.GetOk("azcount"); !ok {
			return diag.Errorf("Attribute 'azcount' is required for aws cloud.")
		}
		if _, ok := d.GetOk("subnet_cidr"); !ok {
			return diag.Errorf("Attribute 'subnet_cidr' is required for aws cloud.")
		}
	} else if cloud == 2 {
		if _, ok := d.GetOk("subnet_address_prefix"); !ok {
			return diag.Errorf("Attribute 'subnet_address_prefix' is required for azure cloud.")
		}
		if _, ok := d.GetOk("account_id"); !ok {
			return diag.Errorf("Attribute 'account_id' is required for azure cloud.")
		}
	}
	log.Printf("[TRACE] validateInfraSchema: end")
	return nil
}
