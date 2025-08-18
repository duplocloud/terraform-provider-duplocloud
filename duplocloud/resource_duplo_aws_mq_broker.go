package duplocloud

import (
	"bytes"
	"context"
	"fmt"
	"hash/crc32"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func duploAwsMqBrokerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the SQS queue will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"engine_type": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"ACTIVEMQ", "RABBITMQ"}, false),
			Description:  "The type of broker engine. Valid values: ACTIVEMQ, RABBITMQ.",
		},
		"deployment_mode": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"ACTIVE_STANDBY_MULTI_AZ", "CLUSTER_MULTI_AZ", "SINGLE_INSTANCE"}, false),
			Description:  "The deployment mode of the broker. Valid values: ACTIVE_STANDBY_MULTI_AZ, CLUSTER_MULTI_AZ, SINGLE_INSTANCE.",
		},
		"broker_storage_type": {
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: validation.StringInSlice([]string{"EBS", "EFS"}, false),
			Description:  "The storage type of the broker. Valid values: EBS, EFS.",
		},
		"broker_name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The name of the broker.",
		},
		"broker_fullname": {
			Type:     schema.TypeString,
			Computed: true,
		},

		"host_instance_type": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The broker's instance type.",
		},
		"engine_version": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The version of the broker engine.",
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				// Suppress diff if only the patch part differs (e.g., "5.17.6.1" vs "5.17.6")
				oldParts := strings.SplitN(old, ".", 3)
				newParts := strings.SplitN(new, ".", 3)
				if len(oldParts) == 3 && len(newParts) == 2 {
					// Compare major.minor
					return oldParts[0] == newParts[0] && oldParts[1] == newParts[1]
				}
				return false
			},
		},
		"authentication_strategy": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"LDAP", "SIMPLE"}, false),
			Description:  "The authentication strategy. Valid values: LDAP, SIMPLE., RABBITMQ only supports SIMPLE and its not updatable after creation of RABBITMQ.",
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				// Suppress diff if the engine type is RABBITMQ and the authentication strategy is SIMPLE
				if strings.EqualFold(d.Get("engine_type").(string), "RABBITMQ") && !strings.EqualFold(d.Get("authentication_strategy").(string), "SIMPLE") {
					return true
				}
				return false
			},
		},
		"auto_minor_version_upgrade": {
			Type:        schema.TypeBool,
			Required:    true,
			Description: "Enables automatic upgrades to new minor versions.",
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"users": {
			Type:        schema.TypeSet,
			Optional:    true,
			Description: "List of users for the broker. User not updatable after creation for RABBITMQ.",
			Set:         resourceUserHash,
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				// AWS currently does not support updating the RabbitMQ users beyond resource creation.
				// User list is not returned back after creation.
				// Updates to users can only be in the RabbitMQ UI.
				if v := d.Get("engine_type").(string); strings.EqualFold(v, "RABBITMQ") && d.Get("arn").(string) != "" {
					return true
				}

				return false
			}, Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"user_name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The username.",
						//ForceNew:    true,
					},
					"password": {
						Type:        schema.TypeString,
						Required:    true,
						Sensitive:   true,
						Description: "The password.",
						//ForceNew:    true,
					},
					"groups": {
						Type:        schema.TypeSet,
						Optional:    true,
						Computed:    true,
						Description: "Groups to which the user belongs.",
						Elem:        &schema.Schema{Type: schema.TypeString},
						//ForceNew:    true,
					},
					"console_access": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
					"replication_user": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
				},
			},
			ConflictsWith: []string{"ldap_server_metadata"},
		},
		"ldap_server_metadata": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "LDAP server metadata. Not applicable for RabbitMQ.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"hosts": {
						Type:        schema.TypeList,
						Required:    true,
						Description: "List of LDAP hosts.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"role_base": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "LDAP role base.",
					},
					"role_name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "LDAP role name.",
					},
					"role_search_matching": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "LDAP role search matching.",
					},
					"role_search_subtree": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "LDAP role search subtree.",
					},
					"service_account_password": {
						Type:        schema.TypeString,
						Required:    true,
						Sensitive:   true,
						Description: "LDAP service account password.",
					},
					"service_account_username": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "LDAP service account username.",
					},
					"user_base": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "LDAP user base.",
					},
					"user_role_name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "LDAP user role name.",
					},
					"user_search_matching": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "LDAP user search matching.",
					},
					"user_search_subtree": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "LDAP user search subtree.",
					},
				},
			},
			ConflictsWith: []string{"users"},
		},
		"configuration": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "MQ configuration.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Configuration ID.",
					},
					"revision": {
						Type:        schema.TypeInt,
						Required:    true,
						Description: "Configuration revision.",
					},
				},
			},
		},
		"creator_request_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "A unique, case-sensitive identifier to ensure idempotency of the request.",
		},
		"is_app_idempotent": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "If true, a UUID will be generated and set to creator_request_id.",
		},
		"data_replication_mode": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"CRDR", "NONE"}, false),
			Description:  "Data replication mode. Valid values: CRDR, NONE.",
		},
		"data_replication_primary_broker_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "ARN of the primary broker for data replication. Required when data_replication_mode is CRDR.",
		},
		"encryption_options": {
			Type:        schema.TypeList,
			Optional:    true,
			ForceNew:    true,
			MaxItems:    1,
			Description: "Encryption options for the broker.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"kms_key_id": {
						Type:        schema.TypeString,
						Optional:    true,
						ForceNew:    true,
						Description: "KMS Key ID for encryption.",
					},
					"use_aws_owned_key": {
						Type:        schema.TypeBool,
						Optional:    true,
						ForceNew:    true,
						Description: "Whether to use AWS owned key.",
					},
				},
			},
		},
		"logs": {
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Description: "Logging options for the broker.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"audit": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Enable audit logging (not applicable for RabbitMQ).",
						DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
							// Suppress diff if the engine type is RABBITMQ and the authentication strategy is SIMPLE
							return strings.EqualFold(d.Get("engine_type").(string), "RABBITMQ")
						},
					},
					"general": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Enable general logging.",
					},
				},
			},
		},
		"maintenance_window": {
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			MaxItems:    1,
			Description: "Maintenance window start time.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"time_of_day": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Time of day for maintenance window. 24 hours format",
					},
					"time_zone": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Time zone for maintenance window.",
					},
					"day_of_week": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"}, false),
						Description:  "Day of week for maintenance window.",
					},
				},
			},
		},
		"publicly_accessible": {
			Type:        schema.TypeBool,
			Required:    true,
			Description: "Whether the broker is publicly accessible.",
		},
		"security_groups": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "List of security group IDs. SG cannot be updated after creation for RABBITMQ.",
			Elem:        &schema.Schema{Type: schema.TypeString},
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				return strings.EqualFold(d.Get("engine_type").(string), "RABBITMQ")
			},
		},
		"subnet_ids": {
			Type:        schema.TypeList,
			Optional:    true,
			Computed:    true,
			Description: "List of subnet IDs.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "A map of tags to assign to the resource.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"broker_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

func resourceAwsMQBroker() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_aws_sqs_queue` manages a SQS queue in Duplo.",

		ReadContext:   resourceAwsMQRead,
		CreateContext: resourceAwsMQCreate,
		UpdateContext: resourceAwsMQUpdate,
		DeleteContext: resourceAwsMQDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema:        duploAwsMqBrokerSchema(),
		CustomizeDiff: validateMQParameter,
	}
}

func resourceAwsMQRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID, brokerID, name := idParts[0], idParts[1], idParts[2]
	log.Printf("[TRACE] resourceAwsMQRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)

	duplo, clientErr := c.DuploAWSMQBrokerGet(tenantID, brokerID)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s Amazon MQ broker %s : %s", tenantID, brokerID, clientErr)
	}
	flattenAwsMqBroker(d, duplo)
	log.Printf("[TRACE] resourceAwsMQRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsMQCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)
	rq := expandAwsMqBroker(d)
	log.Printf("[TRACE] resourceAwsMQCreate(%s, %s): start", tenantID, rq.BrokerName)

	rp, cerr := c.DuploAWSMQBrokerCreate(tenantID, rq)
	if cerr != nil {
		return diag.Errorf("Error creating tenant %s Amazon MQ broker '%s': %s", tenantID, rq.BrokerName, cerr)
	}
	err := waitUntilMQBrokerReady(ctx, c, tenantID, rp.BrokerId, d.Timeout("create"))
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	id := fmt.Sprintf("%s/%s/%s", tenantID, rp.BrokerId, rq.BrokerName)
	d.SetId(id)

	diags := resourceAwsMQRead(ctx, d, m)
	log.Printf("[TRACE] resourceAwsMQCreate(%s, %s): end", tenantID, rq.BrokerName)
	return diags
}

func resourceAwsMQUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID, brokerID, name := idParts[0], idParts[1], idParts[2]
	log.Printf("[TRACE] resourceAwsMQUpdate(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	rq := expandAwsMqBrokerUpdate(d, brokerID)
	cerr := c.DuploAWSMQBrokerUpdate(tenantID, brokerID, *rq)
	if cerr != nil {
		return diag.Errorf("error updating tenant %s aws amazon broker %s : %s", tenantID, brokerID, cerr.Error())
	}
	cerr = c.DuploAWSMQBrokerReeboot(tenantID, brokerID, name)
	if cerr != nil {
		return diag.Errorf("error rebooting tenant %s aws amazon broker %s : %s", tenantID, brokerID, cerr.Error())
	}

	err := waitUntilMQBrokerReady(ctx, c, tenantID, brokerID, d.Timeout("update"))
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	log.Printf("[TRACE] resourceAwsMQUpdate(%s, %s): end", tenantID, name)

	return nil
}

func expandAwsMqBroker(d *schema.ResourceData) *duplosdk.DuploAWSMQ {
	req := &duplosdk.DuploAWSMQ{
		EngineType:              duplosdk.EngineType(d.Get("engine_type").(string)),
		BrokerName:              d.Get("broker_name").(string),
		HostInstanceType:        d.Get("host_instance_type").(string),
		EngineVersion:           d.Get("engine_version").(string),
		AuthenticationStrategy:  duplosdk.AuthenticationStrategy(d.Get("authentication_strategy").(string)),
		AutoMinorVersionUpgrade: d.Get("auto_minor_version_upgrade").(bool),
		DataReplicationMode:     duplosdk.DataReplicationMode(d.Get("data_replication_mode").(string)),
		PubliclyAccessible:      d.Get("publicly_accessible").(bool),
		SecurityGroups:          expandStringList(d.Get("security_groups").([]interface{})),
		SubnetIds:               expandStringList(d.Get("subnet_ids").([]interface{})),
		Tags:                    expandStringMap(d.Get("tags").(map[string]interface{})),
	}

	if v, ok := d.GetOk("deployment_mode"); ok {
		req.DeploymentMode = duplosdk.DeploymentMode(v.(string))
	}
	if v, ok := d.GetOk("broker_storage_type"); ok {
		req.BrokerStorageType = duplosdk.BrokerStorageType(v.(string))
	}
	if v, ok := d.GetOk("is_app_idempotent"); ok && v.(bool) {
		req.CreatorRequestId = uuid.NewString()
	}
	if v, ok := d.GetOk("data_replication_primary_broker_arn"); ok {
		req.DataReplicationPrimaryBrokerArn = v.(string)
	}

	// Users
	req.Users = expandAwsMqBrokerUsers(d.Get("users").(*schema.Set).List())

	// LDAP server metadata
	if v, ok := d.GetOk("ldap_server_metadata"); ok && len(v.([]interface{})) > 0 {
		ldap := v.([]interface{})[0].(map[string]interface{})
		req.LdapServerMetadata = &duplosdk.DuploMQLDAPMetadata{
			Hosts:                  expandStringList(ldap["hosts"].([]interface{})),
			RoleBase:               ldap["role_base"].(string),
			RoleName:               ldap["role_name"].(string),
			RoleSearchMatching:     ldap["role_search_matching"].(string),
			RoleSearchSubtree:      ldap["role_search_subtree"].(bool),
			ServiceAccountPassword: ldap["service_account_password"].(string),
			ServiceAccountUsername: ldap["service_account_username"].(string),
			UserBase:               ldap["user_base"].(string),
			UserRoleName:           ldap["user_role_name"].(string),
			UserSearchMatching:     ldap["user_search_matching"].(string),
			UserSearchSubtree:      ldap["user_search_subtree"].(bool),
		}
	}

	// Configuration
	if v, ok := d.GetOk("configuration"); ok && len(v.([]interface{})) > 0 {
		conf := v.([]interface{})[0].(map[string]interface{})
		req.Configuration = &duplosdk.DuplocloudMQConfiguration{
			Id:       conf["id"].(string),
			Revision: conf["revision"].(int),
		}
	}

	// Encryption options
	if v, ok := d.GetOk("encryption_options"); ok && len(v.([]interface{})) > 0 {
		enc := v.([]interface{})[0].(map[string]interface{})
		req.EncryptionOptions = &duplosdk.DuploMQEncryptionOptions{}
		if kid, ok := enc["kms_key_id"]; ok && kid != nil {
			req.EncryptionOptions.KmsKeyId = kid.(string)
		}
		if uaws, ok := enc["use_aws_owned_key"]; ok && uaws != nil {
			req.EncryptionOptions.UseAwsOwnedKey = uaws.(bool)
		}
	}

	// Logs
	if v, ok := d.GetOk("logs"); ok && len(v.([]interface{})) > 0 {
		logs := v.([]interface{})[0].(map[string]interface{})
		req.Logs = &duplosdk.DuploMQLogs{
			General: logs["general"].(bool),
		}
		if a, ok := logs["audit"]; ok && a != nil {
			nullBool := a.(bool)
			req.Logs.Audit = &nullBool
		}
		if req.EngineType == "RABBITMQ" {
			req.Logs.Audit = nil
		}
	}

	// Maintenance window
	if v, ok := d.GetOk("maintenance_window"); ok && len(v.([]interface{})) > 0 {
		mw := v.([]interface{})[0].(map[string]interface{})
		req.MaintenanceWindow = &duplosdk.DuploMQMaintenanceWindow{
			TimeOfDay: mw["time_of_day"].(string),
			TimeZone:  mw["time_zone"].(string),
			DayOfWeek: duplosdk.DayOfWeek(mw["day_of_week"].(string)),
		}
	}

	return req
}

func expandAwsMqBrokerUsers(v interface{}) []duplosdk.DuploAWSMQUser {
	users := []duplosdk.DuploAWSMQUser{}
	if v == nil {
		return users
	}
	rawUsers := v.([]interface{})
	for _, u := range rawUsers {
		if u == nil {
			continue
		}
		userMap := u.(map[string]interface{})
		user := duplosdk.DuploAWSMQUser{
			UserName:        userMap["user_name"].(string),
			Password:        userMap["password"].(string),
			ConsoleAccess:   userMap["console_access"].(bool),
			ReplicationUser: userMap["replication_user"].(bool),
		}
		gs := userMap["groups"].(*schema.Set).List()
		gstr := []string{}
		for _, g := range gs {
			gstr = append(gstr, g.(string))
		}
		user.Groups = gstr
		users = append(users, user)
	}
	return users
}

func resourceAwsMQDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.Split(id, "/")
	tenantID, brokerID := idParts[0], idParts[1]

	log.Printf("[TRACE] resourceAwsMQDelete(%s, %s): start", tenantID, brokerID)

	c := m.(*duplosdk.Client)

	clientErr := c.DuploAWSMQBrokerDelete(tenantID, brokerID)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s Amazon MQ Broker '%s': %s", tenantID, brokerID, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "Amazon MQ Broker", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploAWSMQBrokerGet(tenantID, brokerID)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsMQDelete(%s, %s): end", tenantID, brokerID)
	return nil
}

// flattenAwsMqBroker updates ResourceData from DuploMQBrokerResponse.
func flattenAwsMqBroker(d *schema.ResourceData, resp *duplosdk.DuploMQBrokerResponse) error {
	if err := d.Set("engine_type", resp.EngineType.Value); err != nil {
		return err
	}
	if err := d.Set("broker_fullname", resp.BrokerName); err != nil {
		return err
	}
	if err := d.Set("arn", resp.BrokerArn); err != nil {
		return err
	}

	if err := d.Set("host_instance_type", resp.HostInstanceType); err != nil {
		return err
	}
	if err := d.Set("engine_version", removePatchVersion(resp.EngineVersion)); err != nil {
		return err
	}
	if err := d.Set("authentication_strategy", resp.AuthenticationStrategy.Value); err != nil {
		return err
	}
	if err := d.Set("auto_minor_version_upgrade", resp.AutoMinorVersionUpgrade); err != nil {
		return err
	}
	if err := d.Set("publicly_accessible", resp.PubliclyAccessible); err != nil {
		return err
	}
	if err := d.Set("security_groups", resp.SecurityGroups); err != nil {
		return err
	}
	if err := d.Set("subnet_ids", resp.SubnetIds); err != nil {
		return err
	}

	if err := d.Set("tags", filterDuploDefinedTagsAsMap(resp.Tags)); err != nil {
		return err
	}
	if resp.DeploymentMode.Value != "" {
		if err := d.Set("deployment_mode", resp.DeploymentMode.Value); err != nil {
			return err
		}
	}
	if resp.StorageType.Value != "" {
		if err := d.Set("broker_storage_type", resp.StorageType.Value); err != nil {
			return err
		}
	}
	// Users
	users := make([]map[string]interface{}, 0, len(resp.Users))
	for _, u := range resp.Users {
		user := map[string]interface{}{
			"user_name": u.Username,
			// password is not returned in response, so leave empty
			"password": u.Password,
		}
		if len(u.Groups) == 0 {
			user["groups"] = nil
		} else {
			user["groups"] = u.Groups
		}
		users = append(users, user)
	}
	if err := d.Set("users", users); err != nil {
		return err
	}
	// Encryption options
	if resp.EncryptionOptions != nil {
		enc := map[string]interface{}{}
		if resp.EncryptionOptions.KmsKeyId != "" {
			enc["kms_key_id"] = resp.EncryptionOptions.KmsKeyId
		}
		enc["use_aws_owned_key"] = resp.EncryptionOptions.UseAwsOwnedKey
		if err := d.Set("encryption_options", []interface{}{enc}); err != nil {
			return err
		}
	}
	// Logs
	logs := map[string]interface{}{
		"general": resp.Logs.General,
		"audit":   resp.Logs.Audit,
	}
	if err := d.Set("logs", []interface{}{logs}); err != nil {
		return err
	}
	// Maintenance window
	if resp.MaintenanceWindowStartTime.DayOfWeek.Value != "" ||
		resp.MaintenanceWindowStartTime.TimeOfDay != "" ||
		resp.MaintenanceWindowStartTime.TimeZone != "" {
		mw := map[string]interface{}{
			"day_of_week": resp.MaintenanceWindowStartTime.DayOfWeek.Value,
			"time_of_day": resp.MaintenanceWindowStartTime.TimeOfDay,
			"time_zone":   resp.MaintenanceWindowStartTime.TimeZone,
		}
		if err := d.Set("maintenance_window", []interface{}{mw}); err != nil {
			return err
		}
	}
	// Set computed broker_id
	if resp.BrokerId != "" {
		if err := d.Set("broker_id", resp.BrokerId); err != nil {
			return err
		}
	}
	return nil
}

func waitUntilMQBrokerReady(ctx context.Context, c *duplosdk.Client, tenantID string, brokerId string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.DuploAWSMQBrokerGet(tenantID, brokerId)
			//			log.Printf("[TRACE] Dynamodb status is (%s).", rp.TableStatus.Value)
			status := "pending"
			if rp != nil && rp.BrokerState.Value == "RUNNING" {
				status = "ready"
			}
			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] waitUntilMQBrokerReady(%s, %s)", tenantID, brokerId)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func expandAwsMqBrokerUpdate(d *schema.ResourceData, brokerId string) *duplosdk.DuploAWSMQBrokerUpdateRequest {
	req := &duplosdk.DuploAWSMQBrokerUpdateRequest{
		BrokerId: brokerId,
	}
	if d.HasChange("host_instance_type") {
		req.HostInstanceType = d.Get("host_instance_type").(string)

	}
	if d.HasChange("engine_version") {
		req.EngineVersion = d.Get("engine_version").(string)
	}
	if d.HasChange("auto_minor_version_upgrade") {
		req.AutoMinorVersionUpgrade = d.Get("auto_minor_version_upgrade").(bool)
	}

	if d.HasChange("data_replication_mode") {
		req.DataReplicationMode = duplosdk.DataReplicationMode(d.Get("data_replication_mode").(string))

	}

	if d.HasChange("authentication_strategy") {
		req.AuthenticationStrategy = duplosdk.AuthenticationStrategy(d.Get("authentication_strategy").(string))
	}
	if d.HasChange("security_groups") {
		req.SecurityGroups = expandStringList(d.Get("security_groups").([]interface{}))
	}
	if d.HasChange("ldap_server_metadata") {
		if v, ok := d.GetOk("ldap_server_metadata"); ok && len(v.([]interface{})) > 0 {
			ldap := v.([]interface{})[0].(map[string]interface{})
			req.LdapServerMetadata = &duplosdk.DuploMQLDAPMetadata{
				Hosts:                  expandStringList(ldap["hosts"].([]interface{})),
				RoleBase:               ldap["role_base"].(string),
				RoleName:               ldap["role_name"].(string),
				RoleSearchMatching:     ldap["role_search_matching"].(string),
				RoleSearchSubtree:      ldap["role_search_subtree"].(bool),
				ServiceAccountPassword: ldap["service_account_password"].(string),
				ServiceAccountUsername: ldap["service_account_username"].(string),
				UserBase:               ldap["user_base"].(string),
				UserRoleName:           ldap["user_role_name"].(string),
				UserSearchMatching:     ldap["user_search_matching"].(string),
				UserSearchSubtree:      ldap["user_search_subtree"].(bool),
			}
		}
	}
	// Configuration
	if d.HasChange("configuration") {
		if v, ok := d.GetOk("configuration"); ok && len(v.([]interface{})) > 0 {
			conf := v.([]interface{})[0].(map[string]interface{})
			req.Configuration = &duplosdk.DuplocloudMQConfiguration{
				Id:       conf["id"].(string),
				Revision: conf["revision"].(int),
			}
		}
	}
	// Logs
	if d.HasChange("logs") {

		if v, ok := d.GetOk("logs"); ok && len(v.([]interface{})) > 0 {
			logs := v.([]interface{})[0].(map[string]interface{})
			req.Logs = &duplosdk.DuploMQLogs{
				General: logs["general"].(bool),
			}
			if a, ok := logs["audit"]; ok && a != nil {
				nullBool := a.(bool)
				req.Logs.Audit = &nullBool
			}
			if d.Get("engine_type").(string) == "RABBITMQ" {
				// AWS does not return audit logging for RabbitMQ, so we set it to false.
				req.Logs.Audit = nil
			}

		}
	}
	// Maintenance window
	if d.HasChange("maintenance_window") {
		if v, ok := d.GetOk("maintenance_window"); ok && len(v.([]interface{})) > 0 {
			mw := v.([]interface{})[0].(map[string]interface{})
			req.MaintenanceWindow = &duplosdk.DuploMQMaintenanceWindow{
				TimeOfDay: mw["time_of_day"].(string),
				TimeZone:  mw["time_zone"].(string),
				DayOfWeek: duplosdk.DayOfWeek(mw["day_of_week"].(string)),
			}
		}
	}
	return req
}

func validateMQParameter(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	bst := diff.Get("broker_storage_type").(string)
	dm := diff.Get("deployment_mode").(string)
	hit := diff.Get("host_instance_type").(string)
	et := diff.Get("engine_type").(string)
	if et == "ACTIVE_MQ" {
		if dm == "SINGLE_INSTANCE" && bst == "EBS" && !strings.Contains(hit, "m5") {
			return fmt.Errorf("ACTIVE_MQ storage type EBS is supported by m5 instance type family ")
		}
		if dm == "ACTIVE_STANDBY_MULTI_AZ" && bst == "EBS" {
			return fmt.Errorf("ACTIVE_MQ storage type EBS is not supported for ACTIVE_STANDBY_MULTI_AZ deployment mode")

		}
	}
	if et == "RABBIT_MQ" {
		if dm == "ACTIVE_STANDBY_MULTI_AZ" {
			return fmt.Errorf("RABBIT_MQ deployment mode ACTIVE_STANDBY_MULTI_AZ is not supported")
		}
		if bst == "EFS" {
			return fmt.Errorf("RABBIT_MQ storage type EFS is not supported")
		}
		if dm == "CLUSTER_MULTI_AZ" && strings.Contains(hit, "t3") {
			return fmt.Errorf("RABBIT_MQ cluster mode do not support t3 family storage type")
		}
		if v, ok := diff.GetOk("ldap_server_metadata"); ok && len(v.([]interface{})) > 0 {
			return fmt.Errorf("RABBIT_MQ do not support ldap_server_metadata block")
		}
	}
	return nil
}

func removePatchVersion(version string) string {
	// Remove the patch version from the engine version string
	parts := strings.SplitN(version, ".", 3)
	if len(parts) < 2 {
		return version // Return as is if it doesn't have enough parts
	}
	return fmt.Sprintf("%s.%s", parts[0], parts[1]) // Return only major and minor versions
}

func resourceUserHash(v any) int {
	var buf bytes.Buffer

	m := v.(map[string]any)
	if ca, ok := m["console_access"]; ok {
		fmt.Fprintf(&buf, "%t-", ca.(bool))
	} else {
		buf.WriteString("false-")
	}
	if g, ok := m["groups"]; ok {
		fmt.Fprintf(&buf, "%v-", g.(*schema.Set).List())
	}
	if p, ok := m["password"]; ok {
		fmt.Fprintf(&buf, "%s-", p.(string))
	}
	fmt.Fprintf(&buf, "%s-", m["user_name"].(string))

	return stringHashcode(buf.String())
}

func stringHashcode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

func flattenUsers(users []map[string]interface{}, cfgUsers []duplosdk.DuploAWSMQUser) *schema.Set {
	out := make([]any, 0)

	for _, u := range cfgUsers {
		userMap := map[string]interface{}{
			"user_name":        u.UserName,
			"console_access":   u.ConsoleAccess,
			"replication_user": u.ReplicationUser,
			"groups":           flattenStringSet(u.Groups),
		}
		out = append(out, userMap)
	}

	return schema.NewSet(resourceUserHash, out)
}
