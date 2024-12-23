package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Resource for managing extra infrastructure subnets
func resourceInfrastructureSubnet() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceInfrastructureSubnetRead,
		CreateContext: resourceInfrastructureSubnetCreate,
		UpdateContext: resourceInfrastructureSubnetUpdate,
		DeleteContext: resourceInfrastructureSubnetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"infra_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Description:  "Specify subnet type. `private` and `public` is used for AWS subnet. Will be one of `none`, `appgwsubnet`, `appgw-internal-subnet`, `azurebastionsubnet`, `managedinstance`, `databrick-workspace`, `mysql-flexiserver`, `postgres-flexiserver` is used for azure.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"private", "public", "none", "appgwsubnet", "appgw-internal-subnet", "azurebastionsubnet", "managedinstance", "databrick-workspace", "mysql-flexiserver", "postgres-flexiserver"}, false),
			},
			"zone": {
				Description: "The Duplo zone that the subnet resides in.  Will be one of:  `\"A\"`, `\"B\"`, `\"C\"`, or `\"D\"`. This is applicable only for AWS subnets.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"isolated_network": {
				Description: "Determines whether the isolated network is enabled. This is applicable only for Azure subnets.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"service_endpoints": {
				Description: "The list of Service endpoints to associate with the azure subnet. Possible values include: `Microsoft.AzureActiveDirectory`, `Microsoft.AzureCosmosDB`, `Microsoft.ContainerRegistry`, `Microsoft.EventHub`, `Microsoft.KeyVault`, `Microsoft.ServiceBus`,`Microsoft.Sql`, `Microsoft.Storage` and `Microsoft.Web`. This is applicable only for Azure subnets.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceInfrastructureSubnetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	log.Printf("[TRACE] resourceInfrastructureSubnetRead(%s): start", id)
	rq, err := duploInfrastructureSubnetFromId(id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	subnetName := rq.Name

	if c.IsAzureCustomPrefixesEnabled() {
		subnetName = c.AddPrefixSuffixFromResourceName(subnetName, "subnet", true)
	} else {
		subnetName = "duploinfra-" + rq.Name
	}
	duplo, err := c.InfrastructureGetSubnet(rq.InfrastructureName, subnetName, rq.AddressPrefix)
	if err != nil {
		return diag.Errorf("Unable to retrieve infrastructure subnet '%s': %s", id, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set the simple fields first.
	d.Set("infra_name", rq.InfrastructureName)
	d.Set("subnet_name", duplo.Name)
	d.Set("subnet_id", duplo.ID)
	d.Set("cidr_block", duplo.AddressPrefix)
	d.Set("zone", duplo.Zone)
	d.Set("tags_all", keyValueToMap(duplo.Tags))
	d.Set("type", duplo.SubnetType)
	d.Set("isolated_network", duplo.IsolatedNetwork)
	d.Set("service_endpoints", duplo.ServiceEndpoints)

	x := d.Get("tags")
	log.Printf("[TRACE] infra subnet tags %v", x)

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := d.GetOk("tags"); ok && v != nil && len(v.(map[string]interface{})) > 0 {
		d.Set("tags", keyValueToMap(selectKeyValuesFromMap(duplo.Tags, v.(map[string]interface{}))))
	}

	log.Printf("[TRACE] resourceInfrastructureSubnetRead(%s): end", id)
	return nil
}

func resourceInfrastructureSubnetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	diags := validateSubnetSchema(d, m)
	if diags != nil {
		return diags
	}
	// Start building the request.
	rq := duplosdk.DuploInfrastructureVnetSubnet{
		InfrastructureName: d.Get("infra_name").(string),
		Name:               d.Get("name").(string),
		AddressPrefix:      d.Get("cidr_block").(string),
		Zone:               d.Get("zone").(string),
		SubnetType:         d.Get("type").(string),
		IsolatedNetwork:    d.Get("isolated_network").(bool),
		Tags:               keyValueFromMap(d.Get("tags").(map[string]interface{})),
	}
	if v, ok := d.GetOk("service_endpoints"); ok {
		endpoints := v.(*schema.Set)
		endpointList := make([]string, 0, endpoints.Len())
		for _, e := range endpoints.List() {
			endpointList = append(endpointList, e.(string))
		}
		rq.ServiceEndpoints = endpointList
	}
	// Build the ID - it is okay that the CIDR includes a slash
	id := fmt.Sprintf("%s/%s/%s", rq.InfrastructureName, rq.Name, rq.AddressPrefix)
	log.Printf("[TRACE] resourceInfrastructureSubnetCreate(%s): start", id)

	// Create the subnet in Duplo.
	c := m.(*duplosdk.Client)
	err := c.InfrastructureCreateOrUpdateSubnet(rq)
	if err != nil {
		return diag.Errorf("Error creating infrastructure subnet '%s': %s", id, err)
	}
	d.SetId(id)

	diags = resourceInfrastructureSubnetRead(ctx, d, m)
	log.Printf("[TRACE] resourceInfrastructureSubnetCreate(%s): end", id)
	return diags
}

func resourceInfrastructureSubnetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// NO-OP
	return resourceInfrastructureSubnetRead(ctx, d, m)
}

func resourceInfrastructureSubnetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	log.Printf("[TRACE] resourceInfrastructureSubnetRead(%s): start", id)
	rq, err := duploInfrastructureSubnetFromId(id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Delete the rule with Duplo
	c := m.(*duplosdk.Client)
	err = c.InfrastructureDeleteSubnet(rq.InfrastructureName, rq.Name, rq.AddressPrefix)
	if err != nil {
		return diag.Errorf("Error deleting infrastructure subnet '%s': %s", id, err)
	}

	log.Printf("[TRACE] resourceInfrastructureSubnetDelete(%s): end", id)
	return nil
}

func duploInfrastructureSubnetFromId(id string) (*duplosdk.DuploInfrastructureVnetSubnet, error) {
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) < 4 {
		return nil, fmt.Errorf("invalid resource ID: %s", id)
	}

	return &duplosdk.DuploInfrastructureVnetSubnet{
		InfrastructureName: idParts[0],
		Name:               idParts[1],
		AddressPrefix:      idParts[2] + "/" + idParts[3],
	}, nil
}

func validateSubnetSchema(d *schema.ResourceData, m interface{}) diag.Diagnostics {
	infraName := d.Get("infra_name").(string)
	log.Printf("[TRACE] validateSubnetSchema: start")

	c := m.(*duplosdk.Client)
	infraConfig, err := c.InfrastructureGetConfig(infraName)
	if err != nil {
		return diag.FromErr(err)
	}

	if infraConfig.Cloud == 0 {
		if _, ok := d.GetOk("zone"); !ok {
			return diag.Errorf("Attribute 'zone' is required for aws cloud.")
		}
	}
	log.Printf("[TRACE] validateSubnetSchema: end")
	return nil
}
