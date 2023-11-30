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

func awsSsmParameterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the SSM parameter will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the SSM parameter.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"type": {
			Description: "The type of the SSM parameter. Valid values are `String`, `StringList`, and `SecureString`.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"String",
				"StringList",
				"SecureString"},
				false),
		},
		"value": {
			Description: "The value of the SSM parameter.",
			Type:        schema.TypeString,
			Optional:    false,
			Computed:    true,
		},
		"description": {
			Description: "The description of the SSM parameter.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"key_id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"allowed_pattern": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"last_modified_user": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"last_modified_date": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

// Resource for managing an AWS SSM parameter
func resourceAwsSsmParameter() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_ssm_parameter` manages an AWS SSM parameter in Duplo.",

		ReadContext:   resourceAwsSsmParameterRead,
		CreateContext: resourceAwsSsmParameterCreate,
		UpdateContext: resourceAwsSsmParameterUpdate,
		DeleteContext: resourceAwsSsmParameterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		Schema: awsSsmParameterSchema(),
	}
}

// READ resource
func resourceAwsSsmParameterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseAwsSsmParameterIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsSsmParameterRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	ssmParam, clientErr := c.SsmParameterGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("") // object missing
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s SSM parameter '%s': %s", tenantID, name, clientErr)
	}

	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("type", ssmParam.Type)
	d.Set("value", ssmParam.Value)
	d.Set("key_id", ssmParam.KeyId)
	d.Set("description", ssmParam.Description)
	d.Set("allowed_pattern", ssmParam.AllowedPattern)
	d.Set("last_modified_user", ssmParam.LastModifiedUser)
	d.Set("last_modified_date", ssmParam.LastModifiedDate)

	log.Printf("[TRACE] resourceAwsSsmParameterRead(%s, %s): end", tenantID, name)
	return nil
}

// CREATE resource
func resourceAwsSsmParameterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsSsmParameterCreate(%s, %s): start", tenantID, name)

	// Create the request object.
	rq := duplosdk.DuploSsmParameterRequest{
		Name:           name,
		Type:           d.Get("type").(string),
		Value:          d.Get("value").(string),
		Description:    d.Get("description").(string),
		KeyId:          d.Get("key_id").(string),
		AllowedPattern: d.Get("allowed_pattern").(string),
	}

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	_, err = c.SsmParameterCreate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s SSM parameter '%s': %s", tenantID, name, err)
	}

	// Wait for Duplo to be able to return the table's details.
	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "SSM parameter", id, func() (interface{}, duplosdk.ClientError) {
		return c.SsmParameterGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAwsSsmParameterRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsSsmParameterCreate(%s, %s): end", tenantID, name)
	return diags
}

// UPDATE resource
func resourceAwsSsmParameterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsSsmParameterUpdate(%s, %s): start", tenantID, name)

	// Create the request object.
	rq := duplosdk.DuploSsmParameterRequest{
		Name:        name,
		Type:        d.Get("type").(string),
		Value:       d.Get("value").(string),
		Description: d.Get("description").(string),
	}

	c := m.(*duplosdk.Client)

	// Post the object to Duplo
	_, err = c.SsmParameterUpdate(tenantID, &rq)
	if err != nil {
		return diag.Errorf("Error update tenant %s SSM parameter '%s': %s", tenantID, name, err)
	}

	diags := resourceAwsSsmParameterRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsSsmParameterUpdate(%s, %s): end", tenantID, name)
	return diags
}

// DELETE resource
func resourceAwsSsmParameterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Parse the identifying attributes
	id := d.Id()
	tenantID, name, err := parseAwsSsmParameterIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsSsmParameterDelete(%s, %s): start", tenantID, name)

	// Delete the function.
	c := m.(*duplosdk.Client)
	clientErr := c.SsmParameterDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s SSM parameter '%s': %s", tenantID, name, clientErr)
	}

	// Wait up to 60 seconds for Duplo to delete the cluster.
	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "SSM parameter", id, func() (interface{}, duplosdk.ClientError) {
		return c.SsmParameterGet(tenantID, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsSsmParameterDelete(%s, %s): end", tenantID, name)
	return nil
}

func parseAwsSsmParameterIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}
