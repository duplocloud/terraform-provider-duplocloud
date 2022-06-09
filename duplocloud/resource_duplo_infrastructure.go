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
		Description: "`duplocloud_infrastructure` manages a tenant in Duplo.",

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
				ValidateFunc: validation.StringLenBetween(2, 12),
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
			"enable_ecs_cluster": {
				Description: "Whether or not to provision an ECS cluster.",
				Type:        schema.TypeBool,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
			},
			"enable_container_insights": {
				Description: "Whether or not to enable container insights for an ECS cluster.",
				Type:        schema.TypeBool,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
			},
			"custom_data": {
				Description: "Custom configuration options for the infrastructure.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem:        KeyValueSchema(),
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
				Optional:    true,
				Default:     "default",
			},
			"subnet_address_prefix": {
				Description: "The address prefixe to use for the subnet. This is applicable only for Azure",
				Type:        schema.TypeString,
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
		},
	}
}

/// READ resource
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

/// CREATE resource
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

/// UPDATE resource
func resourceInfrastructureUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	rq := duploInfrastructureFromState(d)

	log.Printf("[TRACE] resourceInfrastructureUpdate(%s): start", rq.Name)

	// Put the object to Duplo.
	c := m.(*duplosdk.Client)
	_, err = c.InfrastructureUpdate(rq)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for 60 seconds, at first.
	time.Sleep(time.Minute)

	// Then, wait until the infrastructure is completely ready.
	err = duploInfrastructureWaitUntilReady(ctx, c, rq.Name, d.Timeout("update"))
	if err != nil {
		return diag.FromErr(err)
	}

	diags := resourceInfrastructureRead(ctx, d, m)
	log.Printf("[TRACE] resourceInfrastructureUpdate(%s): end", rq.Name)
	return diags
}

/// DELETE resource
func resourceInfrastructureDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) < 4 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	name := idParts[3]

	log.Printf("[TRACE] resourceInfrastructureDelete(%s): start", name)

	c := m.(*duplosdk.Client)
	err := c.InfrastructureDelete(name)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for 20 minutes to allow infrastructure deletion.
	// TODO: wait for it completely deleted (add an API that will actually show the status)
	if d.Get("wait_until_deleted").(bool) {
		log.Printf("[TRACE] resourceInfrastructureDelete(%s): waiting for 20 minutes because 'wait_until_deleted' is 'true'", name)
		time.Sleep(time.Duration(20) * time.Minute)
	}

	log.Printf("[TRACE] resourceInfrastructureDelete(%s): end", name)
	return nil
}

func duploInfrastructureFromState(d *schema.ResourceData) duplosdk.DuploInfrastructure {
	return duplosdk.DuploInfrastructure{
		Name:                    d.Get("infra_name").(string),
		AccountId:               d.Get("account_id").(string),
		Cloud:                   d.Get("cloud").(int),
		Region:                  d.Get("region").(string),
		AzCount:                 d.Get("azcount").(int),
		EnableK8Cluster:         d.Get("enable_k8_cluster").(bool),
		EnableECSCluster:        d.Get("enable_ecs_cluster").(bool),
		EnableContainerInsights: d.Get("enable_container_insights").(bool),
		AddressPrefix:           d.Get("address_prefix").(string),
		SubnetCidr:              d.Get("subnet_cidr").(int),
		CustomData:              keyValueFromState("custom_data", d),
	}
}

func duploInfrastructureConfigFromState(d *schema.ResourceData) duplosdk.DuploInfrastructureConfig {
	subnet := duplosdk.DuploInfrastructureVnetSubnet{}

	if v, ok := d.GetOk("subnet_name"); ok {
		subnet.Name = v.(string)
	}
	if v, ok := d.GetOk("subnet_address_prefix"); ok {
		subnet.AddressPrefix = v.(string)
	}
	return duplosdk.DuploInfrastructureConfig{
		Name:                    d.Get("infra_name").(string),
		AccountId:               d.Get("account_id").(string),
		Cloud:                   d.Get("cloud").(int),
		Region:                  d.Get("region").(string),
		AzCount:                 d.Get("azcount").(int),
		EnableK8Cluster:         d.Get("enable_k8_cluster").(bool),
		EnableECSCluster:        d.Get("enable_ecs_cluster").(bool),
		EnableContainerInsights: d.Get("enable_container_insights").(bool),
		Vnet: &duplosdk.DuploInfrastructureVnet{
			AddressPrefix: d.Get("address_prefix").(string),
			SubnetCidr:    d.Get("subnet_cidr").(int),
			Subnets: &[]duplosdk.DuploInfrastructureVnetSubnet{
				subnet,
			},
		},
		CustomData: keyValueFromState("custom_data", d),
	}
}

func duploInfrastructureWaitUntilReady(ctx context.Context, c *duplosdk.Client, name string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.InfrastructureGetConfig(name)
			status := "pending"
			if err == nil && rp.ProvisioningStatus == "Complete" {
				status = "ready"
			}
			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] duploInfrastructureWaitUntilReady(%s)", name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func infrastructureRead(c *duplosdk.Client, d *schema.ResourceData, name string) (bool, error) {

	config, err := c.InfrastructureGetConfig(name)
	if err != nil {
		return false, err
	}
	infra, err := c.InfrastructureGet(name)
	if err != nil {
		return false, err
	}
	if config == nil || infra == nil {
		return true, nil // object missing
	}

	d.Set("infra_name", infra.Name)
	d.Set("account_id", infra.AccountId)
	d.Set("cloud", infra.Cloud)
	d.Set("region", infra.Region)
	d.Set("azcount", infra.AzCount)
	d.Set("enable_k8_cluster", infra.EnableK8Cluster)
	d.Set("enable_ecs_cluster", infra.EnableECSCluster)
	d.Set("enable_container_insights", infra.EnableContainerInsights)
	d.Set("address_prefix", infra.Vnet.AddressPrefix)
	d.Set("subnet_cidr", infra.Vnet.SubnetCidr)
	d.Set("status", infra.ProvisioningStatus)
	d.Set("custom_data", keyValueToState("custom_data", config.CustomData))

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

				// The server may or may not have the new fields.
				nameParts := strings.SplitN(vnetSubnet.Name, " ", 2)
				zone := vnetSubnet.Zone
				subnetType := vnetSubnet.SubnetType
				if zone == "" {
					zone = nameParts[0]
				}
				if subnetType == "" {
					subnetType = nameParts[1]
				}

				if len(nameParts) == 2 {
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
					d.Set("subnet_name", vnetSubnet.Name)
					d.Set("subnet_address_prefix", vnetSubnet.AddressPrefix)
				}
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
