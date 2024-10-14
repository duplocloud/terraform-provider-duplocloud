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

func tenantAccessGrantSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"grantor_tenant_id": {
			Description:  "The GUID of the tenant that will grant the access.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"grantee_tenant_id": {
			Description:  "The GUID of the tenant that will receive the granted access.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"grant_area": {
			Description: "The area the grant allows access to. Currently supported: ['s3', 'dynamodb', 'kms', 'apigw', 'rep']",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"s3",
				"dynamodb",
				"kms",
				"apigw",
				"rep",
			}, false),
		},
	}
}

func resourceTenantAccessGrant() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_tenant_access_grant` manages a tenant access grant in Duplo.",

		ReadContext:   resourceTenantAccessGrantRead,
		CreateContext: resourceTenantAccessGrantCreate,
		DeleteContext: resourceTenantAccessGrantDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: tenantAccessGrantSchema(),
	}
}

func resourceTenantAccessGrantRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	granteeTenantId, grantorTenantId, grantedArea, err := parseTenantAccessGrantIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceTenantAccessGrantRead(%s, %s, %s): start", grantorTenantId, granteeTenantId, grantedArea)

	// Get returns no new information for access grants
	// 404s are still useful to determine terraform plan
	c := m.(*duplosdk.Client)
	rp, clientErr := c.GetTenantAccessGrant(granteeTenantId, grantorTenantId, grantedArea)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s access grant { grantor(%s), grantedArea(%s)} - error %s", granteeTenantId, grantorTenantId, grantedArea, clientErr)
	}
	if rp == nil || rp.GrantedArea == "" || rp.GrantorTenantId == "" {
		d.SetId("")
		return nil
	}

	// d.Set("grantor_tenant_id", rp.GrantorTenantId)
	// d.Set("grantee_tenant_id", d.Get("grantee_tenant_id"))
	// d.Set("granted_area", rp.GrantedArea)

	// if no 404, nothing to update in TF state
	return nil
}

// CREATE resource
func resourceTenantAccessGrantCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	grantorTenantId := d.Get("grantor_tenant_id").(string)
	granteeTenantId := d.Get("grantee_tenant_id").(string)
	grantedArea := d.Get("grant_area").(string)

	log.Printf("[TRACE] resourceTenantAccessGrantCreate(%s, %s, %s): start", grantorTenantId, granteeTenantId, grantedArea)

	requestBody := &duplosdk.DuploTenantAccessGrant{
		GrantorTenantId: grantorTenantId,
		GrantedArea:     grantedArea,
	}

	// Post access grant to Duplo
	c := m.(*duplosdk.Client)
	clientError := c.CreateTenantAccessGrant(granteeTenantId, requestBody)
	if clientError != nil {
		return diag.FromErr(clientError)
	}
	id := fmt.Sprintf("%s/%s/%s", granteeTenantId, grantorTenantId, grantedArea)
	diag := waitForResourceToBePresentAfterCreate(ctx, d, "tenant access grant", id, func() (interface{}, duplosdk.ClientError) {
		resp, err := c.GetTenantAccessGrantStatus(granteeTenantId, grantorTenantId, grantedArea)
		if err != nil {
			return nil, err
		}

		// wait until grant policy has been applied in the cloud provider
		if resp.Status != "Deployed" {
			return nil, duplosdk.NewCustomError("Tenant Access Grant not ready", 404)
		}
		return resp, nil
	})
	if diag != nil {
		return diag
	}

	d.SetId(id)

	diags := resourceTenantAccessGrantRead(ctx, d, m)
	log.Printf("[TRACE] resourceTenantAccessGrantCreate(%s, %s, %s): end", granteeTenantId, grantorTenantId, grantedArea)
	return diags
}

// UPDATE is NO-OP, all tenant access grant properties are force new

// DELETE resource
func resourceTenantAccessGrantDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	granteeTenantId, grantorTenantId, grantedArea, err := parseTenantAccessGrantIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceTenantAccessGrantDelete(%s, %s, %s): start", granteeTenantId, grantorTenantId, grantedArea)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	clientErr := c.DeleteTenantAccessGrant(granteeTenantId, grantorTenantId, grantedArea)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s access grant { grantorTenantId(%s), grantedArea(%s) } - error %s", granteeTenantId, grantorTenantId, grantedArea, clientErr)
	}

	diags := waitForResourceToBeMissingAfterDelete(ctx, d, "RDS DB instance", d.Id(), func() (interface{}, duplosdk.ClientError) {
		status, err := c.GetTenantAccessGrantStatus(granteeTenantId, grantorTenantId, grantedArea)

		if err != nil {
			return nil, err
		}
		if status.Status == "NonExistent" {
			return nil, duplosdk.NewCustomError("grant does not exist", 404)
		}
		return status, nil
	})

	log.Printf("[TRACE] resourceTenantAccessGrantDelete(%s, %s, %s): end", granteeTenantId, grantorTenantId, grantedArea)
	return diags
}

func parseTenantAccessGrantIdParts(id string) (granteeTenantId string, grantorTenantId string, grantedArea string, err error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) == 3 {
		granteeTenantId, grantorTenantId, grantedArea = idParts[0], idParts[1], idParts[2]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
