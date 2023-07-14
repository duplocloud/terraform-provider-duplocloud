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
)

func duploAdminSystemSettingSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"key": {
			Description: "Key name for the system setting.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"value": {
			Description: "Value for the system setting.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"type": {
			Description: "Type of the system setting.",
			Type:        schema.TypeString,
			Required:    true,
		},
	}
}

func resourceAdminSystemSetting() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_admin_system_setting` manages an admin system setting in Duplo.",

		ReadContext:   resourceAdminSystemSettingRead,
		CreateContext: resourceAdminSystemSettingCreate,
		UpdateContext: resourceAdminSystemSettingUpdate,
		DeleteContext: resourceAdminSystemSettingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAdminSystemSettingSchema(),
	}
}

func resourceAdminSystemSettingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	keyType, key, err := parseAdminSystemSettingIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAdminSystemSettingRead(%s,%s): start", keyType, key)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.SystemSettingGet(key)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve admin system setting %s of type %s : %s", key, keyType, clientErr)
	}
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}

	flattenAdminSystemSetting(d, duplo)

	log.Printf("[TRACE] resourceAdminSystemSettingRead(%s, %s): end", keyType, key)
	return nil
}

func resourceAdminSystemSettingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	key := d.Get("key").(string)
	keyType := d.Get("type").(string)
	log.Printf("[TRACE] resourceAdminSystemSettingCreate(%s,%s): start", keyType, key)
	c := m.(*duplosdk.Client)

	rq := expandAdminSystemSetting(d)
	err = c.SystemSettingCreate(rq)
	if err != nil {
		return diag.Errorf("Error creating key type %s admin system setting '%s': %s", keyType, key, err)
	}

	id := fmt.Sprintf("%s/%s", keyType, key)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "admin system setting", id, func() (interface{}, duplosdk.ClientError) {
		return c.SystemSettingGet(key)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	diags = resourceAdminSystemSettingRead(ctx, d, m)
	log.Printf("[TRACE] resourceAdminSystemSettingCreate(%s, %s): end", keyType, key)
	return diags
}

func resourceAdminSystemSettingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAdminSystemSettingCreate(ctx, d, m)
}

func resourceAdminSystemSettingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	keyType, key, err := parseAdminSystemSettingIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAdminSystemSettingDelete(%s, %s): start", keyType, key)

	c := m.(*duplosdk.Client)
	clientErr := c.SystemSettingDelete(key)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete key type %s admin system setting '%s': %s", keyType, key, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "admin system setting", id, func() (interface{}, duplosdk.ClientError) {
		return c.SystemSettingGet(key)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAdminSystemSettingDelete(%s, %s): end", keyType, key)
	return nil
}

func expandAdminSystemSetting(d *schema.ResourceData) *duplosdk.DuploCustomDataEx {
	return &duplosdk.DuploCustomDataEx{
		Key:   d.Get("key").(string),
		Value: d.Get("value").(string),
		Type:  d.Get("type").(string),
	}
}

func parseAdminSystemSettingIdParts(id string) (keyType, key string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		keyType, key = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAdminSystemSetting(d *schema.ResourceData, duplo *duplosdk.DuploCustomDataEx) {
	d.Set("key", duplo.Key)
	d.Set("value", duplo.Value)
	d.Set("type", duplo.Type)
}
