package duplocloud

import (
	"context"
	"log"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dockerCredsSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the docker credentials will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"user_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"password": {
			Type:      schema.TypeString,
			Required:  true,
			Sensitive: true,
		},
		"email": {
			Type:     schema.TypeString,
			Required: true,
		},
		"registry": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
	}
}

func resourceDockerCreds() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_docker_credentials` manages the docker credentials for the tenant in Duplo.\n\n" +
			"This resource allows you take control of docker registry credentials for a specific tenant.",

		ReadContext:   resourceDockerCredsRead,
		CreateContext: resourceDockerCredsCreateOrUpdate,
		UpdateContext: resourceDockerCredsCreateOrUpdate,
		DeleteContext: resourceDockerCredsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: dockerCredsSchema(),
	}
}

func resourceDockerCredsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Id()
	log.Printf("[TRACE] resourceDockerCredsRead(%s): start", tenantID)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, err := c.TenantGetDockerCredentials(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	if rp == nil || rp["username"] == "" {
		d.SetId("")
		return nil
	}

	flattenDockerCreds(d, tenantID, rp)
	log.Printf("[TRACE] resourceDockerCredsRead(%s): end", tenantID)
	return nil
}

func resourceDockerCredsCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	//Read from Configuration and create duplo object
	tenantID := d.Get("tenant_id").(string)
	username := d.Get("user_name").(string)
	password := d.Get("password").(string)
	email := d.Get("email").(string)
	registry := d.Get("registry").(string)

	log.Printf("[TRACE] resourceDockerCredsCreateOrUpdate(%s): start", tenantID)

	// post the object to duplo

	c := m.(*duplosdk.Client)
	err := c.TenantUpdateDockerCredentials(tenantID, map[string]interface{}{
		"username": username,
		"password": password,
		"email":    email,
		"registry": registry,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tenantID)

	diags := resourceDockerCredsRead(ctx, d, m)
	log.Printf("[TRACE] resourceDockerCredsCreateOrUpdate(%s): end", tenantID)
	return diags
}

func resourceDockerCredsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	tenantID := d.Get("tenant_id").(string)
	log.Printf("[TRACE] resourceDockerCredsDelete(%s): start", tenantID)

	// post the object to duplo

	c := m.(*duplosdk.Client)
	err := c.TenantUpdateDockerCredentials(tenantID, map[string]interface{}{
		"username": "",
		"password": "",
		"email":    "",
		"registry": "",
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tenantID)

	diags := resourceDockerCredsRead(ctx, d, m)
	log.Printf("[TRACE] resourceDockerCredsDelete(%s): end", tenantID)
	return diags

}

func flattenDockerCreds(d *schema.ResourceData, tenantId string, data map[string]interface{}) {
	d.Set("tenant_id", tenantId)
	d.Set("user_name", data["username"])
	d.Set("email", data["email"])
	d.Set("password", data["password"])
	d.Set("registry", data["registry"])
}
