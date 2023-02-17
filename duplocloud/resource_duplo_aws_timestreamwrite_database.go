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

func awsTimestreamDatabaseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the Timestream Database will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The short name of the Timestream Database.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(3, 64),
				// validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
			),
		},
		"fullname": {
			Description: "The full name of the Timestream Database.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"kms_key_id": {
			Description: "The ARN (not Alias ARN) of the KMS key to be used to encrypt the data stored in the database.",
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
		},
		"tags": {
			Description: "Tags in key-value format.",
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Elem:        KeyValueSchema(),
		},
		"arn": {
			Description: "The ARN that uniquely identifies this database.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"table_count": {
			Description: "The total number of tables found within the Timestream database.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
	}
}

func resourceAwsTimestreamDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_timestreamwrite_database` manages an aws Timestream database resource in Duplo.",

		ReadContext:   resourceAwsTimestreamDatabaseRead,
		CreateContext: resourceAwsTimestreamDatabaseCreate,
		UpdateContext: resourceAwsTimestreamDatabaseUpdate,
		DeleteContext: resourceAwsTimestreamDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: awsTimestreamDatabaseSchema(),
	}
}

func resourceAwsTimestreamDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsTimestreamDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsTimestreamDatabaseRead(%s, %s): start", tenantID, name)

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	db, clientErr := c.DuploTimestreamDBGet(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(clientErr)
	}
	if db == nil {
		d.SetId("") // object missing
		return nil
	}
	flattenTimestreamDatabase(d, c, db, tenantID)

	log.Printf("[TRACE] resourceAwsTimestreamDatabaseRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsTimestreamDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAwsTimestreamDatabaseCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	rq := expandAwsTimestreamDatabase(d)
	_, clientErr := c.DuploTimestreamDBCreate(tenantID, rq)
	if clientErr != nil {
		return diag.Errorf("Error creating tenant %s aws timestream database '%s': %s", tenantID, fullName, clientErr)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	d.SetId(id)

	diags := resourceAwsTimestreamDatabaseRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsTimestreamDatabaseCreate(%s, %s): end", tenantID, fullName)
	return diags
}

func resourceAwsTimestreamDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAwsTimestreamDatabaseCreate(ctx, d, m)
}

func resourceAwsTimestreamDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAwsTimestreamDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsTimestreamDatabaseDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	fullName, _ := c.GetDuploServicesName(tenantID, name)
	clientErr := c.DuploTimestreamDBDelete(tenantID, fullName)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s aws timestream database '%s': %s", tenantID, name, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "aws timestream database", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploTimestreamDBGet(tenantID, fullName)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsTimestreamDatabaseDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAwsTimestreamDatabase(d *schema.ResourceData) *duplosdk.DuploTimestreamDBCreateRequest {
	rq := &duplosdk.DuploTimestreamDBCreateRequest{
		DatabaseName: d.Get("name").(string),
		Tags:         keyValueFromState("tags", d),
	}
	if v, ok := d.GetOk("kms_key_id"); ok && v != "" {
		rq.KmsKeyId = v.(string)
	}
	return rq
}

func parseAwsTimestreamDatabaseIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenTimestreamDatabase(d *schema.ResourceData, c *duplosdk.Client, duplo *duplosdk.DuploTimestreamDBDetails, tenantId string) diag.Diagnostics {
	prefix, err := c.GetDuploServicesPrefix(tenantId)
	if err != nil {
		return diag.FromErr(err)
	}
	name, _ := duplosdk.UnprefixName(prefix, duplo.DatabaseName)
	d.Set("tenant_id", tenantId)
	d.Set("name", name)
	d.Set("arn", duplo.Arn)
	d.Set("fullname", duplo.DatabaseName)
	d.Set("table_count", duplo.TableCount)
	d.Set("kms_key_id", duplo.KmsKeyId)
	d.Set("tags", keyValueToState("tags", duplo.Tags))
	return nil
}
