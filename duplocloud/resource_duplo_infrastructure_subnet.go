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
		DeleteContext: resourceInfrastructureSubnetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"private", "public"}, false),
			},
			"zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tagsSchemaForceNew(),
			"tags_all": tagsSchemaComputed(),

			// Which tags were specified by the config?
			"specified_tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
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
	duplo, err := c.InfrastructureGetSubnet(rq.InfrastructureName, "duploinfra-"+rq.Name, rq.AddressPrefix)
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
	d.Set("tags_all", duplosdk.KeyValueToMap(duplo.Tags))
	d.Set("type", duplo.SubnetType)

	// Build a list of current state, to replace the user-supplied settings.
	if v, ok := getAsStringArray(d, "specified_tags"); ok && v != nil {
		d.Set("tags", duplosdk.KeyValueToMap(selectKeyValues(duplo.Tags, *v)))
	}

	log.Printf("[TRACE] resourceInfrastructureSubnetRead(%s): end", id)
	return nil
}

func resourceInfrastructureSubnetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Start building the request.
	rq := duplosdk.DuploInfrastructureVnetSubnet{
		InfrastructureName: d.Get("infra_name").(string),
		Name:               d.Get("name").(string),
		AddressPrefix:      d.Get("cidr_block").(string),
		Zone:               d.Get("zone").(string),
		SubnetType:         d.Get("type").(string),
		Tags:               duplosdk.KeyValueFromMap("tags", d.Get("tags").(map[string]interface{})),
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

	// Collect the desired state of settings specified by the user.
	specified := make([]string, len(*rq.Tags))
	for i, kv := range *rq.Tags {
		specified[i] = kv.Key
	}
	d.Set("specified_tags", specified)

	diags := resourceInfrastructureSubnetRead(ctx, d, m)
	log.Printf("[TRACE] resourceInfrastructureSubnetCreate(%s): end", id)
	return diags
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
		return nil, fmt.Errorf("Invalid resource ID: %s", id)
	}

	return &duplosdk.DuploInfrastructureVnetSubnet{
		InfrastructureName: idParts[0],
		Name:               idParts[1],
		AddressPrefix:      idParts[2] + "/" + idParts[3],
	}, nil
}
