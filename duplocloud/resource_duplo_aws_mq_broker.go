package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
			ValidateFunc: validation.StringInSlice([]string{"ACTIVEMQ", "RABBITMQ"}, false),
			Description:  "The type of broker engine. Valid values: ACTIVEMQ, RABBITMQ.",
		},
		"deployment_mode": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"ACTIVE_STANDBY_MULTI_AZ", "CLUSTER_MULTI_AZ", "SINGLE_INSTANCE"}, false),
			Description:  "The deployment mode of the broker. Valid values: ACTIVE_STANDBY_MULTI_AZ, CLUSTER_MULTI_AZ, SINGLE_INSTANCE.",
		},
		"broker_storage_type": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"EBS", "EFS"}, false),
			Description:  "The storage type of the broker. Valid values: EBS, EFS.",
		},
		"broker_name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the broker.",
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
		},
		"authentication_strategy": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"LDAP", "SIMPLE"}, false),
			Description:  "The authentication strategy. Valid values: LDAP, SIMPLE.",
		},
		"auto_minor_version_upgrade": {
			Type:        schema.TypeBool,
			Required:    true,
			Description: "Enables automatic upgrades to new minor versions.",
		},
		"users": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "List of users for the broker.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"user_name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The username.",
					},
					"password": {
						Type:        schema.TypeString,
						Required:    true,
						Sensitive:   true,
						Description: "The password.",
					},
					"groups": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Groups for the user.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
		"ldap_server_metadata": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "LDAP server metadata.",
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
			Description: "ARN of the primary broker for data replication. Required when data_replication_mode is CRDR.",
		},
		"encryption_options": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Encryption options for the broker.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"kms_key_id": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "KMS Key ID for encryption.",
					},
					"use_aws_owned_key": {
						Type:        schema.TypeBool,
						Optional:    true,
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
			Description: "List of security group IDs.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"subnet_ids": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "List of subnet IDs.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "A map of tags to assign to the resource.",
			Elem:        &schema.Schema{Type: schema.TypeString},
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
		Schema: duploAwsMqBrokerSchema(),
		//CustomizeDiff: validateSQSParameter,
	}
}

func resourceAwsMQRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	id := d.Id()
	tenantID, url, err := parseAwsSqsQueueIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)

	accountID, err := c.TenantGetAwsAccountID(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	fullname, err := c.ExtractSqsFullname(tenantID, url)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsSqsQueueRead(%s, %s): start", tenantID, url)

	queue, clientErr := c.DuploSQSQueueGetV3(tenantID, fullname)
	if queue == nil {
		d.SetId("") // object missing
		return nil
	}
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s sqs queue %s : %s", tenantID, url, clientErr)
	}

	prefix, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("arn", queue.Arn)
	d.Set("tenant_id", tenantID)
	d.Set("url", queue.Url)
	d.Set("fullname", fullname)
	d.Set("fifo_queue", queue.QueueType == 1)
	d.Set("content_based_deduplication", queue.ContentBasedDeduplication)
	d.Set("message_retention_seconds", queue.MessageRetentionPeriod)
	d.Set("visibility_timeout_seconds", queue.VisibilityTimeout)
	d.Set("delay_seconds", queue.DelaySeconds)
	if queue.QueueType == 1 {
		if queue.DeduplicationScope == 0 {
			d.Set("deduplication_scope", "queue")
		} else {
			d.Set("deduplication_scope", "messageGroup")
		}
		if queue.FifoThroughputLimit == 0 {
			d.Set("fifo_throughput_limit", "perQueue")
		} else {
			d.Set("fifo_throughput_limit", "perMessageGroupId")
		}
	}
	if queue.DeadLetterTargetQueueName != "" {
		dlq_config := make(map[string]interface{})
		dlq_config["target_sqs_dlq_name"] = queue.DeadLetterTargetQueueName
		dlq_config["max_message_receive_attempts"] = queue.MaxMessageTimesReceivedBeforeDeadLetterQueue
		d.Set("dead_letter_queue_configuration", []interface{}{dlq_config})
	}

	name, _ := duplosdk.UnwrapName(prefix, accountID, fullname, true)
	if queue.QueueType == 1 {
		name = strings.TrimSuffix(name, ".fifo")
	}
	d.Set("name", name)
	log.Printf("[TRACE] resourceAwsSqsQueueRead(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsMQCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	c := m.(*duplosdk.Client)
	rq := expandAwsMqBroker(d)
	log.Printf("[TRACE] resourceAwsSqsQueueCreate(%s, %s): start", tenantID, rq.BrokerName)

	cerr := c.DuploAWSMQBrokerCreate(tenantID, rq)
	if cerr != nil {
		return diag.Errorf("Error creating tenant %s SQS queue '%s': %s", tenantID, rq.BrokerName, cerr)
	}
	//	diags := waitForResourceToBePresentAfterCreate(ctx, d, "SQS Queue", fmt.Sprintf("%s/%s", tenantID, name), func() (interface{}, duplosdk.ClientError) {
	//		resp, err = c.DuploSQSQueueGetV3(tenantID, fullname)
	//		// wait for an Arn to be available
	//		if err == nil && resp != nil && resp.Arn == "" {
	//			return nil, nil
	//		}
	//		return c.DuploSQSQueueGetV3(tenantID, fullname)
	//	})
	//	if diags != nil {
	//		return diags
	//	}
	id := fmt.Sprintf("%s/%s", tenantID, rq.BrokerName)
	d.SetId(id)

	//	diags = resourceAwsSqsQueueRead(ctx, d, m)
	//	log.Printf("[TRACE] resourceAwsSqsQueueCreate(%s, %s): end", tenantID, name)
	return nil
}

func resourceAwsMQUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	if d.HasChanges("message_retention_seconds", "visibility_timeout_seconds", "content_based_deduplication", "deduplication_scope", "fifo_throughput_limit", "delay_seconds", "dead_letter_queue_configuration") {
		var err error

		tenantID := d.Get("tenant_id").(string)
		fullname := d.Get("fullname").(string)
		url := d.Get("url").(string)
		log.Printf("[TRACE] resourceAwsSqsQueueUpdate(%s, %s): start", tenantID, fullname)
		c := m.(*duplosdk.Client)

		rq := expandAwsSqsQueue(d)
		rq.Name = fullname
		rq.Url = url
		_, err = c.DuploSQSQueueUpdateV3(tenantID, rq)
		if err != nil {
			return diag.Errorf("Error updating tenant %s SQS queue '%s': %s", tenantID, fullname, err)
		}
		diags := waitForResourceToBePresentAfterCreate(ctx, d, "SQS Queue", fmt.Sprintf("%s/%s", tenantID, fullname), func() (interface{}, duplosdk.ClientError) {
			resp, err := c.DuploSQSQueueGetV3(tenantID, fullname)

			if err == nil && resp != nil && resp.Arn == "" {
				return nil, nil
			}
			return c.DuploSQSQueueGetV3(tenantID, fullname)
		})
		if diags != nil {
			return diags
		}

		diags = resourceAwsSqsQueueRead(ctx, d, m)
		log.Printf("[TRACE] resourceAwsSqsQueueUpdate(%s, %s): end", tenantID, fullname)
		return diags
	}
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
	req.Users = expandAwsMqBrokerUsers(d.Get("users"))

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
			req.Logs.Audit = a.(bool)
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
			UserName: userMap["user_name"].(string),
			Password: userMap["password"].(string),
		}
		if groups, ok := userMap["groups"]; ok && groups != nil {
			user.Groups = expandStringList(groups.([]interface{}))
		}
		users = append(users, user)
	}
	return users
}

func resourceAwsMQDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	tenantID, url, err := parseAwsSqsQueueIdParts(id)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[TRACE] resourceAwsSqsQueueDelete(%s, %s): start", tenantID, url)

	c := m.(*duplosdk.Client)

	fullname, err := c.ExtractSqsFullname(tenantID, url)
	if err != nil {
		return diag.FromErr(err)
	}

	clientErr := c.DuploSQSQueueDeleteV3(tenantID, fullname)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Unable to delete tenant %s sqs queue '%s': %s", tenantID, fullname, clientErr)
	}

	diag := waitForResourceToBeMissingAfterDelete(ctx, d, "SQS Queue", id, func() (interface{}, duplosdk.ClientError) {
		return c.DuploSQSQueueGetV3(tenantID, fullname)
	})
	if diag != nil {
		return diag
	}

	log.Printf("[TRACE] resourceAwsSqsQueueDelete(%s, %s): end", tenantID, fullname)
	return nil
}

func flattenAwsMqBroker(d *schema.ResourceData, req *duplosdk.DuploAWSMQ) error {
	if err := d.Set("engine_type", string(req.EngineType)); err != nil {
		return err
	}
	if err := d.Set("broker_name", req.BrokerName); err != nil {
		return err
	}
	if err := d.Set("host_instance_type", req.HostInstanceType); err != nil {
		return err
	}
	if err := d.Set("engine_version", req.EngineVersion); err != nil {
		return err
	}
	if err := d.Set("authentication_strategy", string(req.AuthenticationStrategy)); err != nil {
		return err
	}
	if err := d.Set("auto_minor_version_upgrade", req.AutoMinorVersionUpgrade); err != nil {
		return err
	}
	if err := d.Set("data_replication_mode", string(req.DataReplicationMode)); err != nil {
		return err
	}
	if err := d.Set("publicly_accessible", req.PubliclyAccessible); err != nil {
		return err
	}
	if err := d.Set("security_groups", req.SecurityGroups); err != nil {
		return err
	}
	if err := d.Set("subnet_ids", req.SubnetIds); err != nil {
		return err
	}
	if err := d.Set("tags", req.Tags); err != nil {
		return err
	}
	if req.DeploymentMode != "" {
		if err := d.Set("deployment_mode", string(req.DeploymentMode)); err != nil {
			return err
		}
	}
	if req.BrokerStorageType != "" {
		if err := d.Set("broker_storage_type", string(req.BrokerStorageType)); err != nil {
			return err
		}
	}
	if req.CreatorRequestId != "" {
		if err := d.Set("creator_request_id", req.CreatorRequestId); err != nil {
			return err
		}
	}
	if req.DataReplicationPrimaryBrokerArn != "" {
		if err := d.Set("data_replication_primary_broker_arn", req.DataReplicationPrimaryBrokerArn); err != nil {
			return err
		}
	}
	// Users
	users := make([]map[string]interface{}, 0, len(req.Users))
	for _, u := range req.Users {
		user := map[string]interface{}{
			"user_name": u.UserName,
			"password":  u.Password,
		}
		if len(u.Groups) > 0 {
			user["groups"] = u.Groups
		}
		users = append(users, user)
	}
	if err := d.Set("users", users); err != nil {
		return err
	}
	// LDAP server metadata
	if req.LdapServerMetadata != nil {
		ldap := map[string]interface{}{
			"hosts":                    req.LdapServerMetadata.Hosts,
			"role_base":                req.LdapServerMetadata.RoleBase,
			"role_name":                req.LdapServerMetadata.RoleName,
			"role_search_matching":     req.LdapServerMetadata.RoleSearchMatching,
			"role_search_subtree":      req.LdapServerMetadata.RoleSearchSubtree,
			"service_account_password": req.LdapServerMetadata.ServiceAccountPassword,
			"service_account_username": req.LdapServerMetadata.ServiceAccountUsername,
			"user_base":                req.LdapServerMetadata.UserBase,
			"user_role_name":           req.LdapServerMetadata.UserRoleName,
			"user_search_matching":     req.LdapServerMetadata.UserSearchMatching,
			"user_search_subtree":      req.LdapServerMetadata.UserSearchSubtree,
		}
		if err := d.Set("ldap_server_metadata", []interface{}{ldap}); err != nil {
			return err
		}
	}
	// Configuration
	if req.Configuration != nil {
		conf := map[string]interface{}{
			"id":       req.Configuration.Id,
			"revision": req.Configuration.Revision,
		}
		if err := d.Set("configuration", []interface{}{conf}); err != nil {
			return err
		}
	}
	// Encryption options
	if req.EncryptionOptions != nil {
		enc := map[string]interface{}{}
		if req.EncryptionOptions.KmsKeyId != "" {
			enc["kms_key_id"] = req.EncryptionOptions.KmsKeyId
		}
		enc["use_aws_owned_key"] = req.EncryptionOptions.UseAwsOwnedKey
		if err := d.Set("encryption_options", []interface{}{enc}); err != nil {
			return err
		}
	}
	// Logs
	if req.Logs != nil {
		logs := map[string]interface{}{
			"general": req.Logs.General,
		}
		logs["audit"] = req.Logs.Audit
		if err := d.Set("logs", []interface{}{logs}); err != nil {
			return err
		}
	}
	// Maintenance window
	if req.MaintenanceWindow != nil {
		mw := map[string]interface{}{
			"time_of_day": req.MaintenanceWindow.TimeOfDay,
			"time_zone":   req.MaintenanceWindow.TimeZone,
			"day_of_week": string(req.MaintenanceWindow.DayOfWeek),
		}
		if err := d.Set("maintenance_window", []interface{}{mw}); err != nil {
			return err
		}
	}
	return nil
}
