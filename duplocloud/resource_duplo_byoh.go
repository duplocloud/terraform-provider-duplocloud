package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploByohSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the BYHO will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the BYOH instance. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"direct_address": {
			Description: "IP address of the BYOH instance.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"username": {
			Description: "Username of the BYOH instance.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"password": {
			Description: "Password of the BYOH instance.",
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
		},
		"private_key": {
			Description: "Private key for BYOH instance.",
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
		},
		"agent_platform": {
			Description: "The numeric ID of the container agent pool that this instance is added to.",
			Type:        schema.TypeInt,
			Optional:    true,
			ForceNew:    true,
			Default:     0,
		},
		"allocation_tag": {
			Description: "Allocation tag for BYOH instance.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			ForceNew: true,
			Elem:     KeyValueSchema(),
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until BYOH instance to be connected to the fleet, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"connection_url": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"network_agent_url": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceByoh() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_byoh` manages BYOH in Duplo.",

		ReadContext:   resourceByohRead,
		CreateContext: resourceByohCreate,
		UpdateContext: resourceByohUpdate,
		DeleteContext: resourceByohDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploByohSchema(),
	}
}

func resourceByohRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseByohIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceByohRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.TenantGetByoh(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s BYOH %s : %s", tenantID, name, clientErr)
	}

	flattenByoh(tenantID, d, duplo)

	cred, err := c.TenantHostCredentialsGet(tenantID, duplosdk.DuploHostOOBData{
		IPAddress: duplo.DirectAddress,
		Cloud:     4,
	})
	if err != nil {
		// TODO - Fix backend API for missing data.
		log.Printf("[TRACE] Error : %s", err)
	}

	if cred != nil {
		if len(cred.Username) > 0 {
			d.Set("username", cred.Username)
		}
		if len(cred.Password) > 0 {
			d.Set("password", cred.Password)
		}
		if len(cred.Privatekey) > 0 {
			d.Set("private_key", cred.Privatekey)
		}
	}
	log.Printf("[TRACE] resourceByohRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceByohCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	username := d.Get("username").(string)
	allocationTag := d.Get("allocation_tag").(string)
	log.Printf("[TRACE] resourceByohCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandByoh(d)
	err = c.TenantMinionUpdate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s byoh '%s': %s", tenantID, name, err)
	}

	if len(allocationTag) > 0 {
		err = c.TenantUpdateCustomData(tenantID, duplosdk.CustomDataUpdate{
			ComponentId:   name,
			ComponentType: duplosdk.MINION,
			Key:           "AllocationTags",
			Value:         allocationTag,
		})

		if err != nil {
			return diag.Errorf("Error updating allocation tag for byoh instance '%s': %s", name, err)
		}
	}
	if len(username) > 0 {
		err = c.TenantHostCredentialsUpdate(tenantID, duplosdk.DuploHostOOBData{
			IPAddress: rq.DirectAddress,
			Cloud:     4,
			Credentials: &[]duplosdk.DuploHostCredential{
				{
					Username:   username,
					Password:   d.Get("password").(string),
					Privatekey: d.Get("private_key").(string),
				},
			},
		})
		if err != nil {
			return diag.Errorf("Error updating tenant %s byoh credentials.'%s': %s", tenantID, name, err)
		}
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "duplo byoh", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetByoh(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = byohWaitUntilReady(ctx, c, tenantID, name, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceByohRead(ctx, d, m)
	log.Printf("[TRACE] resourceByohCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceByohUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseByohIdParts(id)
	log.Printf("[TRACE] resourceByohUpdate(%s, %s): start", tenantID, name)
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)

	duplo, clientErr := c.TenantGetByoh(tenantID, name)
	if duplo == nil {
		return diag.Errorf("BYOH does not exist with name '%s': %s", name, err)
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return diag.Errorf("BYOH does not exist with name '%s': %s", name, err)
		}
		return diag.Errorf("Unable to retrieve tenant %s BYOH %s : %s", tenantID, name, clientErr)
	}

	if d.HasChange("allocation_tag") {
		err = c.TenantUpdateCustomData(tenantID, duplosdk.CustomDataUpdate{
			ComponentId:   name,
			ComponentType: duplosdk.MINION,
			Key:           "AllocationTags",
			Value:         d.Get("allocation_tag").(string),
		})

		if err != nil {
			return diag.Errorf("Error updating allocation tag for byoh instance '%s': %s", name, err)
		}
	}

	if d.HasChange("username") || d.HasChange("password") || d.HasChange("private_key") {
		err = c.TenantHostCredentialsUpdate(tenantID, duplosdk.DuploHostOOBData{
			IPAddress: duplo.DirectAddress,
			Cloud:     4,
			Credentials: &[]duplosdk.DuploHostCredential{
				{
					Username:   d.Get("username").(string),
					Password:   d.Get("password").(string),
					Privatekey: d.Get("private_key").(string),
				},
			},
		})
		if err != nil {
			return diag.Errorf("Error updating tenant %s byoh credentials.'%s': %s", tenantID, name, err)
		}
	}
	diags := resourceByohRead(ctx, d, m)

	log.Printf("[TRACE] resourceByohUpdate(%s, %s): end", tenantID, name)
	return diags
}

func resourceByohDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseByohIdParts(id)
	directAddress := d.Get("direct_address").(string)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceByohDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.TenantMinionDelete(tenantID, duplosdk.DuploMinionDeleteReq{
		Name:          name,
		DirectAddress: directAddress,
		State:         "delete",
	})
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s byoh '%s': %s", tenantID, name, clientErr)
	}

	diags := waitForResourceToBeMissingAfterDelete(ctx, d, "duplo byoh", id, func() (interface{}, duplosdk.ClientError) {
		return c.TenantGetByoh(tenantID, name)
	})

	log.Printf("[TRACE] resourceByohDelete(%s, %s): end", tenantID, name)
	return diags
}

func expandByoh(d *schema.ResourceData) duplosdk.DuploMinion {
	return duplosdk.DuploMinion{
		Name:            d.Get("name").(string),
		DirectAddress:   d.Get("direct_address").(string),
		Cloud:           4,
		AgentPlatform:   d.Get("agent_platform").(int),
		Tier:            "default",
		Tunnel:          0,
		ConnectionURL:   fmt.Sprintf("http://%s:4243", d.Get("direct_address").(string)),
		NetworkAgentURL: fmt.Sprintf("http://%s:60035", d.Get("direct_address").(string)),
		//Tags:          keyValueFromState("tags", d),
	}
}

func parseByohIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenByoh(tenantId string, d *schema.ResourceData, duplo *duplosdk.DuploMinion) {
	d.Set("tenant_id", tenantId)
	d.Set("name", duplo.Name)
	d.Set("direct_address", duplo.DirectAddress)
	d.Set("agent_platform", duplo.AgentPlatform)
	d.Set("connection_url", duplo.ConnectionURL)
	d.Set("network_agent_url", duplo.NetworkAgentURL)
	d.Set("tags", keyValueToState("tags", duplo.Tags))
	flattenByohAllocationTag(d, duplo.Tags)

}

func flattenByohAllocationTag(d *schema.ResourceData, duploObjects *[]duplosdk.DuploKeyStringValue) {
	log.Printf("[TRACE] flattenByohAllocationTag ******** start")
	for _, duploObject := range *duploObjects {
		if duploObject.Key == "AllocationTags" {
			d.Set("allocation_tag", duploObject.Value)
			break
		}
	}
	log.Printf("[TRACE] flattenByohAllocationTag ******** start")
}

func byohWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.TenantGetByoh(tenantID, name)
			log.Printf("[TRACE] BYOH connection status is (%s).", rp.ConnectionStatus)
			status := "pending"
			if err == nil {
				if rp.ConnectionStatus == "Connected" {
					status = "ready"
				} else {
					status = "pending"
				}
			}

			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] byohWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
