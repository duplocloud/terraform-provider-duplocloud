package duplocloud

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown

	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		if s.Deprecated != "" {
			desc += " " + s.Deprecated
		}
		return strings.TrimSpace(desc)
	}
}

// Provider return a Terraform provider schema
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"duplo_host": {
				Description: "This is the base URL to the Duplo REST API.  It must be provided, but it can also be sourced from the `duplo_host` environment variable.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"duplo_token": {
				Description: "This is a bearer token used to authenticate to the Duplo REST API.  It must be provided, but it can also be sourced from the `duplo_token` environment variable.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			"ssl_no_verify": {
				Description: "Disable SSL certificate verification.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"http_timeout": {
				Description: "Timeout for HTTP requests in seconds.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     30,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"duplocloud_oci_containerengine_node_pool":       resourceOciContainerEngineNodePool(),
			"duplocloud_admin_system_setting":                resourceAdminSystemSetting(),
			"duplocloud_aws_cloudfront_distribution":         resourceAwsCloudfrontDistribution(),
			"duplocloud_aws_dynamodb_table":                  resourceAwsDynamoDBTable(),
			"duplocloud_aws_dynamodb_table_v2":               resourceAwsDynamoDBTableV2(),
			"duplocloud_aws_elasticsearch":                   resourceDuploAwsElasticSearch(),
			"duplocloud_aws_host":                            resourceAwsHost(),
			"duplocloud_aws_load_balancer":                   resourceAwsLoadBalancer(),
			"duplocloud_aws_load_balancer_listener":          resourceAwsLoadBalancerListener(),
			"duplocloud_aws_kafka_cluster":                   resourceAwsKafkaCluster(),
			"duplocloud_aws_lambda_function":                 resourceAwsLambdaFunction(),
			"duplocloud_aws_ssm_parameter":                   resourceAwsSsmParameter(),
			"duplocloud_duplo_service":                       resourceDuploService(),
			"duplocloud_duplo_service_lbconfigs":             resourceDuploServiceLbConfigs(),
			"duplocloud_duplo_service_params":                resourceDuploServiceParams(),
			"duplocloud_ecache_instance":                     resourceDuploEcacheInstance(),
			"duplocloud_ecs_task_definition":                 resourceDuploEcsTaskDefinition(),
			"duplocloud_ecs_service":                         resourceDuploEcsService(),
			"duplocloud_gcp_cloud_function":                  resourceGcpCloudFunction(),
			"duplocloud_gcp_pubsub_topic":                    resourceGcpPubsubTopic(),
			"duplocloud_gcp_scheduler_job":                   resourceGcpSchedulerJob(),
			"duplocloud_gcp_storage_bucket":                  resourceGcpStorageBucket(),
			"duplocloud_k8_config_map":                       resourceK8ConfigMap(),
			"duplocloud_k8_secret":                           resourceK8Secret(),
			"duplocloud_k8_ingress":                          resourceK8Ingress(),
			"duplocloud_k8_secret_provider_class":            resourceK8SecretProviderClass(),
			"duplocloud_infrastructure":                      resourceInfrastructure(),
			"duplocloud_infrastructure_setting":              resourceInfrastructureSetting(),
			"duplocloud_infrastructure_subnet":               resourceInfrastructureSubnet(),
			"duplocloud_plan_certificates":                   resourcePlanCertificates(),
			"duplocloud_plan_configs":                        resourcePlanConfigs(),
			"duplocloud_plan_settings":                       resourcePlanSettings(),
			"duplocloud_plan_images":                         resourcePlanImages(),
			"duplocloud_rds_instance":                        resourceDuploRdsInstance(),
			"duplocloud_rds_read_replica":                    resourceDuploRdsReadReplica(),
			"duplocloud_s3_bucket":                           resourceS3Bucket(),
			"duplocloud_tenant":                              resourceTenant(),
			"duplocloud_tenant_access_grant":                 resourceTenantAccessGrant(),
			"duplocloud_tenant_cleanup_timers":               resourceTenantCleanUpTimers(),
			"duplocloud_user":                                resourceUser(),
			"duplocloud_tenant_config":                       resourceTenantConfig(),
			"duplocloud_tenant_tag":                          resourceTenantTag(),
			"duplocloud_tenant_secret":                       resourceTenantSecret(),
			"duplocloud_tenant_network_security_rule":        resourceTenantSecurityRule(),
			"duplocloud_emr_cluster":                         resourceAwsEmrCluster(),
			"duplocloud_asg_profile":                         resourceAwsASG(),
			"duplocloud_docker_credentials":                  resourceDockerCreds(),
			"duplocloud_aws_appautoscaling_target":           resourceAwsAppautoscalingTarget(),
			"duplocloud_aws_appautoscaling_policy":           resourceAwsAppautoscalingPolicy(),
			"duplocloud_aws_cloudwatch_event_rule":           resourceAwsCloudWatchEventRule(),
			"duplocloud_aws_cloudwatch_event_target":         resourceAwsCloudWatchEventTarget(),
			"duplocloud_aws_lambda_permission":               resourceAwsLambdaPermission(),
			"duplocloud_aws_cloudwatch_metric_alarm":         resourceAwsCloudWatchMetricAlarm(),
			"duplocloud_aws_ecr_repository":                  resourceAwsEcrRepository(),
			"duplocloud_aws_api_gateway_integration":         resourceAwsApiGatewayIntegration(),
			"duplocloud_aws_target_group_attributes":         resourceAwsTargetGroupAttributes(),
			"duplocloud_aws_lb_target_group":                 resourceTargetGroup(),
			"duplocloud_aws_sqs_queue":                       resourceAwsSqsQueue(),
			"duplocloud_aws_sns_topic":                       resourceAwsSnsTopic(),
			"duplocloud_aws_lb_listener_rule":                resourceAwsLbListenerRule(),
			"duplocloud_azure_key_vault_secret":              resourceAzureKeyVaultSecret(),
			"duplocloud_azure_tenant_key_vault":              resourceAzureTenantKeyVault(),
			"duplocloud_azure_tenant_key_vault_secret":       resourceAzureTenantKeyVaultSecret(),
			"duplocloud_azure_storage_account":               resourceAzureStorageAccount(),
			"duplocloud_azure_mysql_database":                resourceAzureMysqlDatabase(),
			"duplocloud_azure_redis_cache":                   resourceAzureRedisCache(),
			"duplocloud_azure_virtual_machine":               resourceAzureVirtualMachine(),
			"duplocloud_azure_sql_managed_database":          resourceAzureSqlManagedDatabase(),
			"duplocloud_azure_postgresql_database":           resourceAzurePostgresqlDatabase(),
			"duplocloud_azure_mssql_server":                  resourceAzureMssqlServer(),
			"duplocloud_azure_mssql_database":                resourceAzureMssqlDatabase(),
			"duplocloud_azure_mssql_elasticpool":             resourceAzureMssqlElasticPool(),
			"duplocloud_azure_virtual_machine_scale_set":     resourceAzureVirtualMachineScaleSet(),
			"duplocloud_azure_storage_share_file":            resourceAzureStorageShareFile(),
			"duplocloud_azure_log_analytics_workspace":       resourceAzureLogAnalyticsWorkspace(),
			"duplocloud_azure_recovery_services_vault":       resourceAzureRecoveryServicesVault(),
			"duplocloud_azure_vm_feature":                    resourceAzureVmFeature(),
			"duplocloud_azure_vault_backup_policy":           resourceAzureVaultBackupPolicy(),
			"duplocloud_azure_network_security_rule":         resourceAzureNetworkSgRule(),
			"duplocloud_azure_k8_node_pool":                  resourceAzureK8NodePool(),
			"duplocloud_azure_sql_virtual_network_rule":      resourceAzureSqlServerVnetRule(),
			"duplocloud_azure_sql_firewall_rule":             resourceAzureSqlFirewallRule(),
			"duplocloud_azure_k8s_cluster":                   resourceAzureK8sCluster(),
			"duplocloud_azure_private_endpoint":              resourceAzurePrivateEndpoint(),
			"duplocloud_other_agents":                        resourceOtherAgents(),
			"duplocloud_byoh":                                resourceByoh(),
			"duplocloud_aws_mwaa_environment":                resourceMwaaAirflow(),
			"duplocloud_aws_efs_file_system":                 resourceAwsEFS(),
			"duplocloud_aws_efs_lifecycle_policy":            resourceAwsEFSLifecyclePolicy(),
			"duplocloud_k8s_job":                             resourceKubernetesJobV1(),
			"duplocloud_k8s_cron_job":                        resourceKubernetesCronJobV1Beta1(),
			"duplocloud_k8_persistent_volume_claim":          resourceK8PVC(),
			"duplocloud_k8_storage_class":                    resourceK8StorageClass(),
			"duplocloud_aws_batch_scheduling_policy":         resourceAwsBatchSchedulingPolicy(),
			"duplocloud_aws_batch_compute_environment":       resourceAwsBatchComputeEnvironment(),
			"duplocloud_aws_batch_job_queue":                 resourceAwsBatchJobQueue(),
			"duplocloud_aws_batch_job_definition":            resourceAwsBatchJobDefinition(),
			"duplocloud_aws_timestreamwrite_database":        resourceAwsTimestreamDatabase(),
			"duplocloud_aws_timestreamwrite_table":           resourceAwsTimestreamTable(),
			"duplocloud_aws_rds_tag":                         resourceAwsRdsTag(),
			"duplocloud_gcp_sql_database_instance":           resourceGcpSqlDBInstance(),
			"duplocloud_gcp_node_pool":                       resourceGcpK8NodePool(),
			"duplocloud_gcp_firestore":                       resourceFirestore(),
			"duplocloud_plan_waf_v2":                         resourcePlanWafV2(),
			"duplocloud_plan_waf":                            resourcePlanWaf(),
			"duplocloud_plan_kms":                            resourcePlanKMS(),
			"duplocloud_plan_kms_v2":                         resourcePlanKMSV2(),
			"duplocloud_aws_apigateway_event":                resourceAwsApiGatewayEvent(),
			"duplocloud_s3_bucket_replication":               resourceS3BucketReplication(),
			"duplocloud_gcp_storage_bucket_v2":               resourceGCPStorageBucketV2(),
			"duplocloud_aws_lambda_function_event_config":    resourceAwsLambdaEventInvokeConfigFunction(),
			"duplocloud_azure_postgresql_flexible_database":  resourceAzurePostgresqlFlexibleDatabase(),
			"duplocloud_gcp_infra_maintenance_window":        resourceGCPInfraMaintenanceWindow(),
			"duplocloud_azure_storageclass_blob":             resourceAzureStorageBlob(),
			"duplocloud_azure_storageclass_queue":            resourceAzureStorageQueue(),
			"duplocloud_azure_storageclass_table":            resourceAzureStorageTable(),
			"duplocloud_azure_vm_maintenance_configuration":  resourceAzureVmMaintenanceConfig(),
			"duplocloud_user_tenant_access":                  resourceUserTenantAccess(),
			"duplocloud_gcp_redis_instance":                  resourceRedisInstance(),
			"duplocloud_gcp_host":                            resourceGcpHost(),
			"duplocloud_gcp_infra_security_rule":             resourceGCPInfraSecurityRule(),
			"duplocloud_gcp_tenant_security_rule":            resourceGCPTenantSecurityRule(),
			"duplocloud_infrastructure_onprem":               resourceInfrastructureOnprem(),
			"duplocloud_aws_launch_template":                 resourceAwsLaunchTemplate(),
			"duplocloud_k8_helm_repository":                  resourceHelmRepository(),
			"duplocloud_k8_helm_release":                     resourceHelmRelease(),
			"duplocloud_azure_datafactory":                   resourceAzureDataFactory(),
			"duplocloud_azure_availability_set":              resourceAzureAvailabilitySet(),
			"duplocloud_aws_launch_template_default_version": resourceAwsLaunchTemplateDefaultVersion(),
			"duplocloud_gcp_pubsub_subscription":             resourceGCPPubSubSubscription(),
			"duplocloud_aws_tag":                             resourceAwsCustomTag(),
			"duplocloud_aws_target_group_target_register":    resourceAwsTargetGroupTargetRegister(),
			"duplocloud_azure_cosmos_db_account":             resourceAzureCosmosDBAccount(),
			"duplocloud_azure_cosmos_db_database":            resourceAzureCosmosDB(),
			"duplocloud_azure_cosmos_db_container":           resourceAzureCosmosDBContainer(),
			"duplocloud_asg_instance_refresh":                resourceAsgInstanceRefresh(),
			"duplocloud_aws_cloudfront_function":             resourceAwsCloudfrontFunction(),
			"duplocloud_azure_mssqldb_retention_backup":      resourceMsSQLDBRetentionBackup(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"duplocloud_admin_aws_credentials":      dataSourceAdminAwsCredentials(),
			"duplocloud_aws_account":                dataSourceAwsAccount(),
			"duplocloud_aws_lb_listeners":           dataSourceTenantAwsLbListeners(),
			"duplocloud_aws_lb_target_groups":       dataSourceTenantAwsLbTargetGroups(),
			"duplocloud_aws_ssm_parameter":          dataSourceAwsSsmParameter(),
			"duplocloud_aws_ssm_parameters":         dataSourceAwsSsmParameters(),
			"duplocloud_eks_credentials":            dataSourceEksCredentials(),
			"duplocloud_gke_credentials":            dataSourceGKECredentials(),
			"duplocloud_duplo_service":              dataSourceDuploService(),
			"duplocloud_duplo_services":             dataSourceDuploServices(),
			"duplocloud_duplo_service_lbconfigs":    dataSourceDuploServiceLbConfigs(),
			"duplocloud_duplo_service_params":       dataSourceDuploServiceParams(),
			"duplocloud_ecs_service":                dataSourceDuploEcsService(),
			"duplocloud_ecs_services":               dataSourceDuploEcsServices(),
			"duplocloud_ecs_task_definition":        dataSourceDuploEcsTaskDefinition(),
			"duplocloud_ecs_task_definitions":       dataSourceDuploEcsTaskDefinitions(),
			"duplocloud_infrastructure":             dataSourceInfrastructure(),
			"duplocloud_infrastructures":            dataSourceInfrastructures(),
			"duplocloud_k8_config_map":              dataSourceK8ConfigMap(),
			"duplocloud_k8_config_maps":             dataSourceK8ConfigMaps(),
			"duplocloud_k8s_job":                    dataSourceK8sJob(),
			"duplocloud_k8s_cron_job":               dataSourceK8sCronJob(),
			"duplocloud_k8_secret":                  dataSourceK8Secret(),
			"duplocloud_k8_secrets":                 dataSourceK8Secrets(),
			"duplocloud_native_hosts":               dataSourceNativeHosts(),
			"duplocloud_native_host_image":          dataSourceNativeHostImage(),
			"duplocloud_native_host_images":         dataSourceNativeHostImages(),
			"duplocloud_plan":                       dataSourcePlan(),
			"duplocloud_plan_image":                 dataSourcePlanImage(),
			"duplocloud_plan_images":                dataSourcePlanImages(),
			"duplocloud_plan_settings":              dataSourcePlanSettings(),
			"duplocloud_plans":                      dataSourcePlans(),
			"duplocloud_tenant":                     dataSourceTenant(),
			"duplocloud_tenants":                    dataSourceTenants(),
			"duplocloud_tenant_aws_region":          dataSourceTenantAwsRegion(),
			"duplocloud_tenant_aws_credentials":     dataSourceTenantAwsCredentials(),
			"duplocloud_tenant_aws_kms_key":         dataSourceTenantAwsKmsKey(),
			"duplocloud_tenant_aws_kms_keys":        dataSourceTenantAwsKmsKeys(),
			"duplocloud_tenant_cleanup_timers":      dataSourceTenantCleanUpTimers(),
			"duplocloud_tenant_config":              dataSourceTenantConfig(),
			"duplocloud_tenant_eks_credentials":     dataSourceTenantEksCredentials(),
			"duplocloud_tenant_internal_subnets":    dataSourceTenantInternalSubnets(),
			"duplocloud_tenant_external_subnets":    dataSourceTenantExternalSubnets(),
			"duplocloud_tenant_secret":              dataSourceTenantSecret(),
			"duplocloud_tenant_secrets":             dataSourceTenantSecrets(),
			"duplocloud_emr_cluster":                dataSourceEmrClusters(),
			"duplocloud_plan_certificate":           dataSourcePlanCert(),
			"duplocloud_plan_certificates":          dataSourcePlanCerts(),
			"duplocloud_asg_profiles":               dataSourceAsgProfiles(),
			"duplocloud_plan_nat_gateways":          dataSourcePlanNgws(),
			"duplocloud_ecr_repository":             dataSourceEcrRepository(),
			"duplocloud_azure_storage_account_key":  dataSourceAzureStorageAccountKey(),
			"duplocloud_gcp_node_pool":              dataSourceGCPNodePool(),
			"duplocloud_gcp_node_pools":             dataSourceGCPNodePools(),
			"duplocloud_gcp_sql_database_instance":  dataSourceGCPCloudSQL(),
			"duplocloud_gcp_sql_database_instances": dataSourceGCPCloudSQLs(),
			"duplocloud_gcp_firestore":              dataSourceFirestore(),
			"duplocloud_gcp_firestores":             dataSourceFirestores(),
			"duplocloud_gcp_redis_instance":         dataSourceRedisInstance(),
			"duplocloud_plan_wafs":                  dataSourcePlanWafs(),
			"duplocloud_plan_waf":                   dataSourcePlanWaf(),
			"duplocloud_plan_wafs_v2":               dataSourcePlanWafsV2(),
			"duplocloud_plan_waf_v2":                dataSourcePlanWafV2(),
			"duplocloud_plan_kms":                   dataSourcePlanKMS(),
			"duplocloud_plan_kms_key":               dataSourcePlanKMSList(),
			"duplocloud_plan_kms_v2":                dataSourcePlanKMSV2(),
			"duplocloud_plan_kms_key_v2":            dataSourcePlanKMSListV2(),
			"duplocloud_azure_availability_set":     dataSourceAzureAvailabilitySet(),
			"duplocloud_aws_launch_template":        dataSourceAwsLaunchTemplate(),
			"duplocloud_system_features":            dataSourceDuploSystemFeatures(),
			"duplocloud_azure_cosmos_db_account":    dataSourceAzureCosmosDBAccount(),
			"duplocloud_azure_cosmos_db_database":   dataSourceAzureCosmosDBDatabase(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	token := d.Get("duplo_token").(string)
	host := d.Get("duplo_host").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	if token == "" {
		token = os.Getenv("duplo_token")
		if token == "" {
			token = os.Getenv("DUPLO_TOKEN")

		}
	}
	if host == "" {
		host = os.Getenv("duplo_host")
		if host == "" {
			host = os.Getenv("DUPLO_HOST")
		}
	}

	c, err := duplosdk.NewClient(host, token)

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Duplocloud Unable to create Duplocloud client.",
			Detail:   "Duplocloud Unable to create anonymous Duplocloud client - provide env for duplo_token, duplo_host",
		})
		return nil, diags
	}

	if sslNoVerify, ok := d.GetOk("ssl_no_verify"); ok && sslNoVerify.(bool) {
		log.Printf("[TRACE] ssl_no_verify provided in the provider configuration.")
		c.HTTPClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	if httpTimeout, ok := d.GetOk("http_timeout"); ok {
		timout := time.Duration(httpTimeout.(int)) * time.Second
		log.Printf("[TRACE] http_timeout provided in the provider configuration. Timeout : %s", timout)
		c.HTTPClient.Timeout = timout * time.Second
	}

	return c, diags
}
