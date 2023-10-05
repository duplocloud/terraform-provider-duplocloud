package resources

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplocloud"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAzureSqlManagedDatabaseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the azure sql managed database will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the SQL Managed Instance. This needs to be globally unique within Azure.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"sku_name": {
			Description: "Specifies the SKU Name for the SQL Managed Instance. Valid values include `GP_Gen4`, `GP_Gen5`, `BC_Gen4`, `BC_Gen5`. Changing this forces a new resource to be created.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"GP_Gen4",
				"GP_Gen5",
				"BC_Gen4",
				"BC_Gen5",
			}, false),
		},
		"administrator_login": {
			Description:  "The administrator login name for the new server. Changing this forces a new resource to be created.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"administrator_login_password": {
			Description:  "The password associated with the `administrator_login` user. Needs to comply with Azure's Password Policy",
			Type:         schema.TypeString,
			Required:     true,
			Sensitive:    true,
			ValidateFunc: validation.StringIsNotEmpty,
		},
		"vcores": {
			Description: " Number of cores that should be assigned to your instance. Values can be `8`, `16`, or `24` if `sku_name` is `GP_Gen4`, or `8`, `16`, `24`, `32`, or `40` if `sku_name` is `GP_Gen5`",
			Type:        schema.TypeInt,
			Required:    true,
			ValidateFunc: validation.IntInSlice([]int{
				4,
				8,
				16,
				24,
				32,
				40,
				64,
				80,
			}),
		},
		"storage_size_in_gb": {
			Description:  "Maximum storage space for your instance. It should be a multiple of 32GB.",
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntBetween(32, 8192),
		},

		"subnet_id": {
			Description: "The subnet resource id that the SQL Managed Instance will be associated with.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"tags": {
			Type:     schema.TypeMap,
			Computed: true,
			Elem:     schema.TypeString,
		},
		"collation": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"public_data_endpoint_enabled": {
			Description: "Is the public data endpoint enabled? Default value is `false`.",
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
		},
		"minimum_tls_version": {
			Description: "The Minimum TLS Version for all SQL managed Database and SQL Data Warehouse databases associated with the server. Valid values are: `1.0`, `1.1` and `1.2`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ValidateFunc: validation.StringInSlice([]string{
				"1.0",
				"1.1",
				"1.2",
			}, false),
		},
		"fqdn": {
			Description: "The fully qualified domain name of the sql managed instance.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func resourceAzureSqlManagedDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_azure_sql_managed_database` manages an azure sql managed database in Duplo.",

		ReadContext:   resourceAzureSqlManagedDatabaseRead,
		CreateContext: resourceAzureSqlManagedDatabaseCreate,
		UpdateContext: resourceAzureSqlManagedDatabaseUpdate,
		DeleteContext: resourceAzureSqlManagedDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: duploAzureSqlManagedDatabaseSchema(),
	}
}

func resourceAzureSqlManagedDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureSqlManagedDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureSqlManagedDatabaseRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, clientErr := c.SqlManagedDatabaseGet(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s azure sql managed database %s : %s", tenantID, name, clientErr)
	}

	// TODO Set ccomputed attributes from duplo object to tf state.
	flattenAzureSqlManagedDatabase(d, duplo)

	log.Printf("[TRACE] resourceAzureSqlManagedDatabaseRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAzureSqlManagedDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] resourceAzureSqlManagedDatabaseCreate(%s, %s): start", tenantID, name)
	c := m.(*duplosdk.Client)

	rq := expandAzureSqlManagedDatabase(d)
	err = c.SqlManagedDatabaseCreate(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s azure sql managed database '%s': %s", tenantID, name, err)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)
	diags := duplocloud.waitForResourceToBePresentAfterCreate(ctx, d, "azure sql managed database", id, func() (interface{}, duplosdk.ClientError) {
		return c.SqlManagedDatabaseGet(tenantID, name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	if d.Get("wait_until_ready") == nil || d.Get("wait_until_ready").(bool) {
		err = sqlManagedDatabaseWaitUntilReady(ctx, c, tenantID, name, d.Timeout("create"))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	diags = resourceAzureSqlManagedDatabaseRead(ctx, d, m)
	log.Printf("[TRACE] resourceAzureSqlManagedDatabaseCreate(%s, %s): end", tenantID, name)
	return diags
}

func resourceAzureSqlManagedDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceAzureSqlManagedDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, name, err := parseAzureSqlManagedDatabaseIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAzureSqlManagedDatabaseDelete(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	clientErr := c.SqlManagedDatabaseDelete(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s azure sql managed database '%s': %s", tenantID, name, clientErr)
	}

	diag := duplocloud.waitForResourceToBeMissingAfterDelete(ctx, d, "azure sql managed database", id, func() (interface{}, duplosdk.ClientError) {
		return c.SqlManagedDatabaseGet(tenantID, name)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAzureSqlManagedDatabaseDelete(%s, %s): end", tenantID, name)
	return nil
}

func expandAzureSqlManagedDatabase(d *schema.ResourceData) *duplosdk.DuploAzureSqlManagedDatabaseRequest {
	sku, _ := expandManagedInstanceSkuName(d.Get("sku_name").(string))
	return &duplosdk.DuploAzureSqlManagedDatabaseRequest{
		NameEx:                               d.Get("name").(string),
		PropertiesSubnetID:                   d.Get("subnet_id").(string),
		PropertiesAdministratorLogin:         d.Get("administrator_login").(string),
		PropertiesAdministratorLoginPassword: d.Get("administrator_login_password").(string),
		PropertiesVCores:                     strconv.Itoa(d.Get("vcores").(int)),
		PropertiesStorageSizeInGB:            d.Get("storage_size_in_gb").(int),
		Sku:                                  sku,
	}
}

func parseAzureSqlManagedDatabaseIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenAzureSqlManagedDatabase(d *schema.ResourceData, duplo *duplosdk.DuploAzureSqlManagedDatabaseInstance) {
	d.Set("name", duplo.Name)
	d.Set("storage_size_in_gb", duplo.PropertiesStorageSizeInGB)
	d.Set("subnet_id", duplo.PropertiesSubnetID)
	d.Set("sku_name", duplo.Sku.Name)
	d.Set("vcores", duplo.PropertiesVCores)
	d.Set("administrator_login", duplo.PropertiesAdministratorLogin)
	d.Set("tags", duplo.Tags)
	d.Set("fqdn", duplo.PropertiesFullyQualifiedDomainName)
	d.Set("collation", duplo.PropertiesCollation)
	d.Set("minimum_tls_version", duplo.PropertiesMinimalTLSVersion)
	d.Set("public_data_endpoint_enabled", duplo.PropertiesPublicDataEndpointEnabled)
}

func sqlManagedDatabaseWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID string, name string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.SqlManagedDatabaseGet(tenantID, name)
			log.Printf("[TRACE] Sql managed instance provisioning state is (%s).", rp.PropertiesProvisioningState)
			status := "pending"
			if err == nil {
				if rp.PropertiesState == "Succeeded" {
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
	log.Printf("[DEBUG] sqlManagedDatabaseWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func expandManagedInstanceSkuName(skuName string) (*duplosdk.DuploAzureSqlManagedDatabaseSku, error) {
	parts := strings.Split(skuName, "_")
	if len(parts) != 2 {
		return nil, fmt.Errorf("sku_name (%s) has the wrong number of parts (%d) after splitting on _", skuName, len(parts))
	}

	var tier string
	switch parts[0] {
	case "GP":
		tier = "GeneralPurpose"
	case "BC":
		tier = "BusinessCritical"
	default:
		return nil, fmt.Errorf("sku_name %s has unknown sku tier %s", skuName, parts[0])
	}

	return &duplosdk.DuploAzureSqlManagedDatabaseSku{
		Name:   skuName,
		Tier:   tier,
		Family: parts[1],
	}, nil
}
