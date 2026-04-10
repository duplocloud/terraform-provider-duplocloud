package duplocloud

import (
	"context"
	"fmt"
	"log"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceDuploRdsInstance() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_rds_instance` retrieves the current configuration of an RDS instance (or promoted read replica) for a given tenant. " +
			"Because this datasource always fetches live data from the API, it reflects the current writer endpoint after a failover event.",

		ReadContext: dataSourceDuploRdsInstanceRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The GUID of the tenant.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.IsUUID,
			},
			"name": {
				Description: "The short name of the RDS instance (without the duplo- prefix).",
				Type:        schema.TypeString,
				Required:    true,
			},
			"identifier": {
				Description: "The full name of the RDS instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"arn": {
				Description: "The ARN of the RDS instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"endpoint": {
				Description: "The current connection endpoint (host:port) of the RDS instance. Reflects the promoted writer after a failover.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"host": {
				Description: "The DNS hostname of the RDS instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"port": {
				Description: "The listening port of the RDS instance.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"master_username": {
				Description: "The master username of the RDS instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"engine": {
				Description: "The numerical index of database engine to use the for the RDS instance.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"engine_version": {
				Description: "The database engine version.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"snapshot_id": {
				Description: "The database snapshot the RDS instance was initialized from.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"db_subnet_group_name": {
				Description: "Name of DB subnet group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"db_name": {
				Description: "The name of the database.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"parameter_group_name": {
				Description: "The RDS parameter group name applied to the instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cluster_parameter_group_name": {
				Description: "Parameter group associated with this instance's DB Cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"store_details_in_secret_manager": {
				Description: "Whether or not RDS details are stored in the AWS secrets manager.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"size": {
				Description: "The instance type of the RDS instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"storage_type": {
				Description: "Storage type used for RDS instance storage.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"iops": {
				Description: "The IOPS value for the RDS instance.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"allocated_storage": {
				Description: "The allocated storage in gigabytes.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"encrypt_storage": {
				Description: "Whether or not the RDS instance storage is encrypted.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"enable_logging": {
				Description: "Whether or not RDS instance logging is enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"backup_retention_period": {
				Description: "The backup retention period in days.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"multi_az": {
				Description: "Whether the RDS instance is multi-AZ.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"instance_status": {
				Description: "The current status of the RDS instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"deletion_protection": {
				Description: "Whether deletion protection is enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"kms_key_id": {
				Description: "The globally unique identifier for the KMS key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cluster_identifier": {
				Description: "The RDS Cluster Identifier.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"v2_scaling_configuration": {
				Description: "Serverless v2 scaling configuration.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_capacity": {
							Description: "Minimum scaling capacity.",
							Type:        schema.TypeFloat,
							Computed:    true,
						},
						"max_capacity": {
							Description: "Maximum scaling capacity.",
							Type:        schema.TypeFloat,
							Computed:    true,
						},
					},
				},
			},
			"skip_final_snapshot": {
				Description: "Whether the final snapshot is skipped on delete.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"enable_iam_auth": {
				Description: "Whether RDS IAM authentication is enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"enhanced_monitoring": {
				Description: "The enhanced monitoring interval in seconds.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"performance_insights": {
				Description: "Amazon RDS Performance Insights configuration.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Description: "Whether Performance Insights is enabled.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"kms_key_id": {
							Description: "ARN of the KMS key used to encrypt Performance Insights data.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"retention_period": {
							Description: "Retention period in days.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
					},
				},
			},
			"availability_zone": {
				Description: "The Availability Zone of the RDS instance.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"auto_minor_version_upgrade": {
				Description: "Whether auto minor version upgrade is enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"storage_autoscaling": {
				Description: "Storage autoscaling configuration.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Description: "Whether storage autoscaling is enabled.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"max_allocated_storage": {
							Description: "The upper limit in GiB for storage autoscaling.",
							Type:        schema.TypeInt,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDuploRdsInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)
	log.Printf("[TRACE] dataSourceDuploRdsInstanceRead(%s, %s): start", tenantID, name)

	c := m.(*duplosdk.Client)
	duplo, err := c.RdsInstanceDataSourceGetByName(tenantID, name)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to retrieve RDS instance '%s' in tenant '%s': %s", name, tenantID, err))
	}
	if duplo == nil {
		return diag.Errorf("RDS instance '%s' not found in tenant '%s'", name, tenantID)
	}

	d.SetId(fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/%s", tenantID, name))
	jo := rdsInstanceToState(duplo, d)
	for key, val := range jo {
		d.Set(key, val)
	}

	log.Printf("[TRACE] dataSourceDuploRdsInstanceRead(%s, %s): end", tenantID, name)
	return nil
}
