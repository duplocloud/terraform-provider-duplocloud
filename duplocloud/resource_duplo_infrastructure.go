package duplocloud

import (
	"context"
	"fmt"

	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SCHEMA for resource crud
func resourceInfrastructure() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceInfrastructureRead,
		CreateContext: resourceInfrastructureCreate,
		UpdateContext: resourceInfrastructureUpdate,
		DeleteContext: resourceInfrastructureDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"infra_name": {
				Type:     schema.TypeString,
				Optional: false,
				Required: true,
				ForceNew: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
				Required: false,
				ForceNew: true,
			},
			"cloud": {
				Type:     schema.TypeInt,
				Optional: true,
				Required: false,
				ForceNew: true,
				Default:  0,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: false,
				ForceNew: true,
				Required: true,
			},
			"azcount": {
				Type:     schema.TypeInt,
				Optional: false,
				ForceNew: true,
				Required: true,
			},
			"enable_k8_cluster": {
				Type:     schema.TypeBool,
				Optional: false,
				Required: true,
			},
			"address_prefix": {
				Type:     schema.TypeString,
				Optional: false,
				ForceNew: true,
				Required: true,
			},
			"subnet_cidr": {
				Type:     schema.TypeInt,
				Optional: false,
				ForceNew: true,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

/// READ resource
func resourceInfrastructureRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	name := d.Get("infra_name").(string)

	log.Printf("[TRACE] resourceInfrastructureRead(%s): start", name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.InfrastructureGet(name)
	if err != nil {
		return diag.Errorf("Unable to retrieve infrastructure '%s': %s", name, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	d.Set("infra_name", duplo.Name)
	d.Set("account_id", duplo.AccountId)
	d.Set("cloud", duplo.Cloud)
	d.Set("region", duplo.Region)
	d.Set("azcount", duplo.AzCount)
	d.Set("enable_k8_cluster", duplo.EnableK8Cluster)
	d.Set("address_prefix", duplo.AddressPrefix)
	d.Set("subnet_cidr", duplo.SubnetCidr)
	d.Set("status", duplo.ProvisioningStatus)

	log.Printf("[TRACE] resourceInfrastructureRead(%s): end", name)
	return nil
}

/// CREATE resource
func resourceInfrastructureCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rq := duploInfrastructureFromState(d)

	log.Printf("[TRACE] resourceInfrastructureCreate(%s): start", rq.Name)

	// Post the object to Duplo.
	c := m.(*duplosdk.Client)
	_, err := c.InfrastructureCreate(rq)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait up to 60 seconds for Duplo to be able to return the infrastructure details.
	id := fmt.Sprintf("v2/admin/InfrastructureV2/%s", rq.Name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "infrastructure", id, func() (interface{}, error) {
		return c.InfrastructureGet(rq.Name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	// Then, wait until the infrastructure is completely ready.
	err = duploInfrastructureWaitUntilReady(c, rq.Name, d.Timeout("create"))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceInfrastructureRead(ctx, d, m)
	log.Printf("[TRACE] resourceInfrastructureCreate(%s): end", rq.Name)
	return nil
}

/// UPDATE resource
func resourceInfrastructureUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rq := duploInfrastructureFromState(d)

	log.Printf("[TRACE] resourceInfrastructureUpdate(%s): start", rq.Name)

	// Put the object to Duplo.
	c := m.(*duplosdk.Client)
	_, err := c.InfrastructureUpdate(rq)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for 60 seconds, at first.
	time.Sleep(time.Minute)

	// Then, wait until the infrastructure is completely ready.
	err = duploInfrastructureWaitUntilReady(c, rq.Name, d.Timeout("update"))
	if err != nil {
		return diag.FromErr(err)
	}

	resourceInfrastructureRead(ctx, d, m)
	log.Printf("[TRACE] resourceInfrastructureUpdate(%s): end", rq.Name)
	return nil
}

/// DELETE resource
func resourceInfrastructureDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	name := d.Get("infra_name").(string)

	log.Printf("[TRACE] resourceInfrastructureDelete(%s): start", name)

	c := m.(*duplosdk.Client)
	err := c.InfrastructureDelete(name)
	if err != nil {
		return diag.FromErr(err)
	}

	// TODO: wait for it completely deleted (is there an API that will actually show the status?)

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

func duploInfrastructureWaitUntilReady(c *duplosdk.Client, name string, timeout time.Duration) error {
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
	_, err := stateConf.WaitForState()
	return err
}
