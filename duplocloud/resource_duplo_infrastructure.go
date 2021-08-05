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
			Create: schema.DefaultTimeout(30 * time.Minute),
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
				Computed:    true,
			},
			"cloud": {
				Description: "The numerical index of cloud provider to use for the infrastructure.\n" +
					"Should be one of:\n\n" +
					"   - `0` : AWS\n" +
					"   - `2` : Azure\n",
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
				Description: "The number of availability zones.  Must be one of: `2`, `3`, or `4`.",
				Type:        schema.TypeInt,
				ForceNew:    true,
				Required:    true,
			},
			"enable_k8_cluster": {
				Description: "Whether or not to provision a kubernetes cluster.",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"address_prefix": {
				Description: "The CIDR to use for the VPC or VNet.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"subnet_cidr": {
				Description: "The CIDR subnet size (in bits) for the automatically created subnets.",
				Type:        schema.TypeInt,
				ForceNew:    true,
				Required:    true,
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
			"wait_until_deleted": {
				Description:      "Whether or not to wait until Duplo has destroyed the Infrastructure",
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

	rq := duploInfrastructureFromState(d)

	log.Printf("[TRACE] resourceInfrastructureCreate(%s): start", rq.Name)

	// Post the object to Duplo.
	c := m.(*duplosdk.Client)
	_, err = c.InfrastructureCreate(rq)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait up to 60 seconds for Duplo to be able to return the infrastructure details.
	id := fmt.Sprintf("v2/admin/InfrastructureV2/%s", rq.Name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "infrastructure", id, func() (interface{}, duplosdk.ClientError) {
		return c.InfrastructureGet(rq.Name)
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
		Name:            d.Get("infra_name").(string),
		AccountId:       d.Get("account_id").(string),
		Cloud:           d.Get("cloud").(int),
		Region:          d.Get("region").(string),
		AzCount:         d.Get("azcount").(int),
		EnableK8Cluster: d.Get("enable_k8_cluster").(bool),
		AddressPrefix:   d.Get("address_prefix").(string),
		SubnetCidr:      d.Get("subnet_cidr").(int),
	}
}

func duploInfrastructureWaitUntilReady(ctx context.Context, c *duplosdk.Client, name string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.InfrastructureGet(name)
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

	infra, err := c.InfrastructureGet(name)
	if err != nil {
		return false, err
	}
	config, err := c.InfrastructureGetConfig(name)
	if err != nil {
		return false, err
	}
	if infra == nil || config == nil {
		return true, nil // object missing
	}

	d.Set("infra_name", infra.Name)
	d.Set("account_id", infra.AccountId)
	d.Set("cloud", infra.Cloud)
	d.Set("region", infra.Region)
	d.Set("azcount", infra.AzCount)
	d.Set("enable_k8_cluster", infra.EnableK8Cluster)
	d.Set("address_prefix", infra.AddressPrefix)
	d.Set("subnet_cidr", infra.SubnetCidr)
	d.Set("status", infra.ProvisioningStatus)

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
				}
			}

			d.Set("private_subnets", privateSubnets)
			d.Set("public_subnets", publicSubnets)
		}
	}

	return false, nil
}
