package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTenant() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_tenant` manages a tenant in Duplo." +
			"<p>A **DuploCloud tenant** is an isolated environment within the DuploCloud platform where you can manage and provision cloud resources." +
			" It essentially represents a distinct organizational unit or environment for deploying and managing infrastructure and applications.</p>",
		ReadContext:   resourceTenantRead,
		CreateContext: resourceTenantCreate,
		UpdateContext: resourceTenantRead, // NO-OP
		DeleteContext: resourceTenantDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"account_name": {
				Description: "The name of the tenant. Tenant names are globally unique, and cannot be a prefix of any other tenant name.",
				Type:        schema.TypeString,
				ForceNew:    true, // Change tenant name
				Required:    true,
			},
			"plan_id": {
				Description:  "The name of the plan under which the tenant will be created.",
				Type:         schema.TypeString,
				ForceNew:     true, // Change plan (infrastructure)
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 30),
			},
			"tenant_id": {
				Description: "A GUID identifying the tenant. This is automatically generated by Duplo.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"infra_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"existing_k8s_namespace": {
				Description: "Existing kubernetes namespace to use by the tenant. *NOTE: This is an advanced feature, please contact your DuploCloud administrator for help if you want to use this field.*",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`), "kubernetes namespace must contain only lower case alphanumeric and hypen, and cannot start or end with hypen"),
				),
			},
			"policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_volume_mapping": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"block_external_ep": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     KeyValueSchema(),
			},
			"wait_until_created": {
				Description: "Whether or not to wait until Duplo has created the tenant.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"allow_deletion": {
				Description: "Whether or not to even try and delete the tenant. *NOTE: This only works if you have disabled deletion protection for the tenant.*",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"wait_until_deleted": {
				Description: "Whether or not to wait until Duplo has destroyed the tenant.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

// READ resource
func resourceTenantRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()

	// Parse the identifying attributes
	tenantID := parseDuploTenantIdParts(id)
	if tenantID == "" {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	log.Printf("[TRACE] resourceTenantRead(%s): start", tenantID)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.TenantGetV2(tenantID)
	if err != nil {
		if err.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant '%s': %s", tenantID, err)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	// Set simple fields first.
	d.Set("account_name", duplo.AccountName)
	d.Set("tenant_id", duplo.TenantID)
	d.Set("plan_id", duplo.PlanID)
	d.Set("existing_k8s_namespace", duplo.ExistingK8sNamespace)
	d.Set("infra_owner", duplo.InfraOwner)

	// Next, set nested fields.
	if duplo.TenantPolicy != nil {
		d.Set("policy", []map[string]interface{}{{
			"allow_volume_mapping": true,
			"block_external_ep":    true,
		}})
	} else {
		d.Set("policy", []map[string]interface{}{})
	}
	d.Set("tags", keyValueToState("tags", duplo.Tags))

	log.Printf("[TRACE] resourceTenantRead(%s): end", tenantID)
	return nil
}

// CREATE resource
func resourceTenantCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rq := duplosdk.DuploTenant{
		AccountName:          d.Get("account_name").(string),
		PlanID:               d.Get("plan_id").(string),
		ExistingK8sNamespace: d.Get("existing_k8s_namespace").(string),
	}

	log.Printf("[TRACE] resourceTenantCreate(%s): start", rq.AccountName)

	// Post the object to Duplo
	c := m.(*duplosdk.Client)

	diags := validateTenantSchema(d, c)
	if diags != nil {
		return diags
	}

	infra, err := c.InfrastructureGetConfig(rq.PlanID)
	if err != nil {
		return diag.Errorf("Unable to retrieve duplo infrastructure '%s': %s", rq.PlanID, err)
	}
	if infra.Cloud == GCP_CLOUD && strings.Contains(rq.AccountName, "google") {
		return diag.Errorf("Restricted use of keyword google in account_name for gcp cloud")
	}
	if infra.Cloud == 2 {
		_, err = c.TenantCreateAzure(rq)
	} else {
		err = c.TenantCreate(rq)
	}

	if err != nil {
		return diag.Errorf("Unable to create tenant '%s': %s", rq.AccountName, err)
	}

	// Wait up to 60 seconds for Duplo to be able to return the tenant.
	var rp *duplosdk.DuploTenant
	diags = waitForResourceToBePresentAfterCreate(ctx, d, "tenant", rq.AccountName, func() (interface{}, duplosdk.ClientError) {
		rp, err = c.GetTenantByNameForUser(rq.AccountName)
		return rp, err
	})
	if diags != nil {
		return diags
	}

	d.SetId(fmt.Sprintf("v2/admin/TenantV2/%s", rp.TenantID))
	d.Set("tenant_id", rp.TenantID)

	// Wait for 2 minutes to allow tenant creation.
	if d.Get("wait_until_created").(bool) {
		log.Printf("[TRACE] resourceTenantCreate(%s): waiting for 2 minutes because 'wait_until_created' is 'true'", rq.AccountName)
		time.Sleep(time.Duration(120) * time.Second)
	}

	diags = waitForResourceToBePresentAfterCreate(ctx, d, "tenant", rq.AccountName, func() (interface{}, duplosdk.ClientError) {
		rp, err = c.GetTenantByNameForUser(rq.AccountName)
		return rp, err
	})
	if diags != nil {
		return diags
	}

	diags = resourceTenantRead(ctx, d, m)
	if diags != nil {
		return diags
	}
	log.Printf("[TRACE] resourceTenantCreate(%s): end", rq.AccountName)
	return nil
}

// DELETE resource
func resourceTenantDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()

	// Parse the identifying attributes
	tenantID := parseDuploTenantIdParts(id)
	if tenantID == "" {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	log.Printf("[TRACE] resourceTenantDelete(%s): start", tenantID)

	if d.Get("allow_deletion").(bool) {

		// Delete the object with Duplo
		c := m.(*duplosdk.Client)
		duplo, err := c.TenantGetV2(tenantID)
		if err != nil {
			if err.Status() == 404 {
				return nil
			}
			return diag.Errorf("Unable to retrieve tenant '%s': %s", tenantID, err)
		}
		if duplo == nil {
			return nil
		}
		err = c.TenantDelete(tenantID)
		if err != nil {
			if err.Status() == 404 {
				return nil
			}
			return diag.Errorf("Error deleting tenant '%s': %s", tenantID, err)
		}

		// Wait for 1 minute to allow tenant deletion.
		if d.Get("wait_until_deleted").(bool) {
			log.Printf("[TRACE] resourceTenantDelete(%s): waiting for 1 minute because 'wait_until_deleted' is 'true'", tenantID)
			time.Sleep(time.Duration(1) * time.Minute)
		}
	} else {
		log.Printf("[WARN] resourceTenantDelete(%s): will NOT delete the tenant - because 'allow_deletion' is 'false'", tenantID)
	}

	log.Printf("[TRACE] resourceTenantDelete(%s): end", tenantID)
	return nil
}

func parseDuploTenantIdParts(id string) (tenantID string) {
	idParts := strings.SplitN(id, "/", 4)
	if len(idParts) == 4 {
		tenantID = idParts[3]
	}
	return
}

func validateTenantSchema(d *schema.ResourceData, c *duplosdk.Client) diag.Diagnostics {
	log.Printf("[TRACE] validateTenantSchema: start")
	accountName := d.Get("account_name").(string)
	maxLength := 12
	features, _ := c.AdminGetSystemFeatures()
	if features != nil && features.TenantNameMaxLength > maxLength {
		maxLength = features.TenantNameMaxLength
	}
	if len(accountName) < 2 || len(accountName) > maxLength {
		return diag.Errorf("Length of attribute 'account_name' must be between 2 and %d inclusive, got: %d", maxLength, len(accountName))
	}
	return nil
}
