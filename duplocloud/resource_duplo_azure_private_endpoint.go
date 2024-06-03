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

func duploAzurePrivateEndpointSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure private endpoint will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "Specifies the Name of the Private Endpoint.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"subnet_id": {
			Description: "The ID of the Subnet from which Private IP Addresses will be allocated for this Private Endpoint.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"private_link_service_connection": {
			Description: "Specifies private link service connections.",
			Type:        schema.TypeList,
			Required:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: "Specifies the Name of the Private Service Connection.",
						Type:        schema.TypeString,
						Required:    true,
						ForceNew:    true,
					},
					"private_connection_resource_id": {
						Description: "The ID of the Private Link Enabled Remote Resource which this Private Endpoint should be connected to.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"group_ids": {
						Type:     schema.TypeList,
						Required: true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
	}
}

func resourceAzurePrivateEndpoint() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_private_endpoint` manages an azure private endpoint in Duplo.",

		ReadContext:   resourceAzurePrivateEndpointRead,
		CreateContext: resourceAzurePrivateEndpointCreate,
		UpdateContext: resourceAzurePrivateEndpointUpdate,
		DeleteContext: resourceAzurePrivateEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzurePrivateEndpointSchema(),
	}
}

func resourceAzurePrivateEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, peName, err := parseAzurePrivateEndpointIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzurePrivateEndpointRead(%s, %s): start", tenantID, peName)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.PrivateEndpointGet(tenantID, peName)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure private endpoint %s : %s", tenantID, peName, clientErr)
	}

	flattenAzurePrivateEndpoint(d, duplo)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("tenant_id", tenantID)
	log.Printf("[TRACE] resourceAzurePrivateEndpointRead(%s, %s): end", tenantID, peName)
	return nil
}

func resourceAzurePrivateEndpointCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	peName := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzurePrivateEndpointCreate(%s, %s): start", tenantID, peName)
	c := m.(*duplosdk.Client)

	rq := expandAzurePrivateEndpoint(d)
	err = c.PrivateEndpointCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure private endpoint '%s': %s", tenantID, peName, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, peName)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "azure private endpoint", id, func() (interface{}, duplosdk.ClientError) {
		return c.PrivateEndpointGet(tenantID, peName)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAzurePrivateEndpointRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzurePrivateEndpointCreate(%s, %s): end", tenantID, peName)
	return diags
}

func resourceAzurePrivateEndpointUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzurePrivateEndpointDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, peName, err := parseAzurePrivateEndpointIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzurePrivateEndpointDelete(%s, %s): start", tenantID, peName)

	c := m.(*duplosdk.Client)
	clientErr := c.PrivateEndpointDelete(tenantID, peName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure private endpoint '%s': %s", tenantID, peName, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "azure private endpoint", id, func() (interface{}, duplosdk.ClientError) {
		return c.PrivateEndpointGet(tenantID, peName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzurePrivateEndpointDelete(%s, %s): end", tenantID, peName)
	return nil
}

func expandAzurePrivateEndpoint(d *schema.ResourceData) *duplosdk.DuploAzurePrivateEndpoint {
	return &duplosdk.DuploAzurePrivateEndpoint{
		Name: d.Get("name").(string),
		PropertiesSubnetID: &duplosdk.DuploAzurePrivateEndpointSubnetID{
			Id: d.Get("subnet_id").(string),
		},

		PrivateLinkServiceConnections: expandPrivateEndpointServiceConnections(d.Get("private_link_service_connection").([]interface{})),
	}
}

func expandPrivateEndpointServiceConnections(tfList []interface{}) *[]duplosdk.DuploAzurePrivateLinkServiceConnections {
	var apiObjects []duplosdk.DuploAzurePrivateLinkServiceConnections

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := duplosdk.DuploAzurePrivateLinkServiceConnections{}

		if v, ok := tfMap["name"].(string); ok && v != "" {
			apiObject.Name = v
		}

		if v, ok := tfMap["private_connection_resource_id"].(string); ok && v != "" {
			apiObject.PrivateLinkServiceId = v
		}

		if v, ok := tfMap["group_ids"]; ok && v != nil {
			apiObject.GroupIds, _ = getStringArray(tfMap, "group_ids")
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return &apiObjects
}

func parseAzurePrivateEndpointIdParts(id string) (tenantID, peName string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, peName = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzurePrivateEndpoint(d *schema.ResourceData, duplo *duplosdk.DuploAzurePrivateEndpoint) {
	d.Set("name", duplo.Name)
	d.Set("subnet_id", duplo.PropertiesSubnetID.Id)
	d.Set("private_link_service_connection", flattenAzurePrivateEndpointServiceConnections(duplo))
}

func flattenAzurePrivateEndpointServiceConnections(duplo *duplosdk.DuploAzurePrivateEndpoint) []map[string]interface{} {
	s := []map[string]interface{}{}
	for _, v := range *duplo.PrivateLinkServiceConnections {
		s = append(s, flattenAzurePrivateEndpointServiceConnection(v))
	}
	return s
}

func flattenAzurePrivateEndpointServiceConnection(duplo duplosdk.DuploAzurePrivateLinkServiceConnections) map[string]interface{} {
	m := make(map[string]interface{})
	m["name"] = duplo.Name
	m["private_connection_resource_id"] = duplo.PrivateLinkServiceId
	m["origin_path"] = duplo.GroupIds
	return m
}
