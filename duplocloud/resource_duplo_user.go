package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description:   "`duplocloud_user` manages a user in Duplo.",
		ReadContext:   resourceUserRead,
		CreateContext: resourceUserCreate,
		UpdateContext: resourceUserCreate,
		DeleteContext: resourceUserDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(120 * time.Minute),
			Delete: schema.DefaultTimeout(120 * time.Minute),
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
			"roles": {
				Description: "The list of roles to be assigned to thh created user. Valid values are - `User`, `Administrator`, `SignupUser`, `SecurityAdmin`.",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    4,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"is_readonly": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"reallocate_vpn_address": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"regenerate_vpn_password": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"vpn_static_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_vpn_config_created": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_confirmation_email_sent": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"current_session_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

/// READ resource
func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()

	if id == "" {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	log.Printf("[TRACE] resourceUserRead(%s): start", id)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.UserGet(id)
	if err != nil {
		return diag.Errorf("Unable to retrieve User '%s': %s", id, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	d.Set("username", duplo.Username)
	d.Set("roles", duplo.Roles)
	d.Set("is_readonly", duplo.IsReadOnly)
	d.Set("reallocate_vpn_address", duplo.ReallocateVpnAddress)
	d.Set("regenerate_vpn_password", duplo.RegenerateVpnPassword)
	d.Set("vpn_static_ip", duplo.VpnStaticIp)
	d.Set("is_vpn_config_created", duplo.IsVpnConfigCreated)
	d.Set("is_confirmation_email_sent", duplo.IsConfirmationEmailSent)
	d.Set("current_session_token", duplo.CurrentSessionToken)

	log.Printf("[TRACE] resourceUserRead(%s): end", id)
	return nil
}

/// CREATE resource
func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	userName := d.Get("username").(string)
	log.Printf("[TRACE] resourceUserCreate(%s): start", userName)
	rq := duplosdk.DuploUser{
		Username:              userName,
		IsReadOnly:            d.Get("is_readonly").(bool),
		ReallocateVpnAddress:  d.Get("reallocate_vpn_address").(bool),
		RegenerateVpnPassword: d.Get("regenerate_vpn_password").(bool),
	}
	if v, ok := getAsStringArray(d, "roles"); ok && v != nil {
		rq.Roles = v
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	resp, err := c.UserCreate(rq)
	if err != nil {
		return diag.Errorf("Unable to create User '%s': %s", resp.Username, err)
	}

	var rp *duplosdk.DuploUser
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "User", userName, func() (interface{}, duplosdk.ClientError) {
		rp, err = c.UserGet(userName)
		return rp, err
	})
	if diags != nil {
		return diags
	}

	d.SetId(userName)
	d.Set("vpn_static_ip", resp.VpnStaticIp)
	d.Set("is_vpn_config_created", resp.IsVpnConfigCreated)
	d.Set("is_confirmation_email_sent", resp.IsConfirmationEmailSent)
	d.Set("current_session_token", resp.CurrentSessionToken)
	diags = resourceUserRead(ctx, d, m)
	if diags != nil {
		return diags
	}

	log.Printf("[TRACE] resourceUserCreate(%s): end", userName)
	return nil
}

/// DELETE resource
func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()

	if id == "" {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	log.Printf("[TRACE] resourceUserDelete(%s): start", id)

	c := m.(*duplosdk.Client)
	err := c.UserDelete(id)
	if err != nil {
		return diag.Errorf("Error deleting User '%s': %s", id, err)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "User", id, func() (interface{}, duplosdk.ClientError) {
		if rp, err := c.UserExists(id); rp || err != nil {
			return rp, err
		}
		return nil, nil
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceUserDelete(%s): end", id)
	return nil
}
