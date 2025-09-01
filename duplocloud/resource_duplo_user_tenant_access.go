package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUserTenantAccess() *schema.Resource {
	return &schema.Resource{
		Description:   "`duplocloud_user_tenant_access` manages a user tenant access in Duplo.",
		ReadContext:   resourceUserTenantAccessRead,
		CreateContext: resourceUserTenantAccessCreate,
		UpdateContext: resourceUserTenantAccessUpdate,
		DeleteContext: resourceUserTenantAccessDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"username": {
				Description: "The unique user name or the email.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"is_readonly": {
				Description: "Specifiy readonly policy related to tenant",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"tenant_id": {
				Description: "Tenant Id to which user need to get access",
				Type:        schema.TypeString,
				Required:    true,
			},
			"tenant_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// READ resource
func resourceUserTenantAccessRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()

	if id == "" {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	log.Printf("[TRACE] resourceUserRead(%s): start", id)
	idParts := strings.Split(id, "/")
	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.GetUserTenantAccessInfo(idParts[0], idParts[1])
	if err != nil {
		return diag.Errorf("Duplocloud resource id '%s' \n[Error]: %s", id, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	d.Set("username", idParts[0])
	d.Set("is_readonly", duplo.Policy.IsReadOnly)
	d.Set("tenant_id", idParts[1])
	d.Set("tenant_name", duplo.AccountName)

	log.Printf("[TRACE] resourceUserRead(%s): end", id)
	return nil
}

// CREATE resource
func resourceUserTenantAccessCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	userName := d.Get("username").(string)
	log.Printf("[TRACE] resourceUserCreate(%s): start", userName)
	rq := duplosdk.DuploUserTenantAccess{
		Username: userName,
		Policy:   duplosdk.DuploTenantAccessPolicy{IsReadOnly: d.Get("is_readonly").(bool)},
		TenantId: d.Get("tenant_id").(string),
		State:    "create",
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	id := fmt.Sprintf("%s/%s", rq.Username, rq.TenantId)
	d.SetId(id)

	err := c.GrantUserTenantAccess(&rq)
	if err != nil {
		if err.Status() == 404 {
			log.Printf("Tenant config for tenant %s not found", id)
			d.SetId("")
			return diag.Errorf("Duplocloud resource id '%s' \n[Error]: %s", id, err)
		}
		return diag.Errorf("Duplocloud resource id '%s' \n[Error]: %s", id, err)
	}

	var rp *duplosdk.DuploUser
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "User", userName, func() (interface{}, duplosdk.ClientError) {
		rp, err = c.UserGet(userName)
		return rp, err
	})
	if diags != nil {
		return diags
	}

	diags = resourceUserTenantAccessRead(ctx, d, m)
	log.Printf("[TRACE] resourceUserTenantAccessCreate(%s): end", userName)
	return diags
}

func resourceUserTenantAccessUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	diags := resourceUserTenantAccessDelete(ctx, d, m)
	if diags != nil {
		return diags
	}
	diags = resourceUserTenantAccessCreate(ctx, d, m)
	return diags
}

// DELETE resource
func resourceUserTenantAccessDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	if id == "" {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	idParts := strings.Split(id, "/")
	log.Printf("[TRACE] resourceUserTenantAccessDelete(%s): start", idParts[0])
	rq := duplosdk.DuploUserTenantAccess{
		Username: idParts[0],
		Policy:   duplosdk.DuploTenantAccessPolicy{IsReadOnly: d.Get("is_readonly").(bool)},
		TenantId: idParts[1],
		State:    "deleted",
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	err := c.GrantUserTenantAccess(&rq)
	if err != nil {
		if err.Status() == 404 {
			return nil
		}
		return diag.Errorf("Duplocloud resource id '%s' \n[Error]: %s", id, err)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "User", id, func() (interface{}, duplosdk.ClientError) {
		if rp, err := c.GetUserTenantAccessInfo(rq.Username, rq.TenantId); rp != nil || err != nil {
			return rp, err
		}
		return nil, nil
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceUserTenantAccessDelete(%s): end", id)
	return nil
}
