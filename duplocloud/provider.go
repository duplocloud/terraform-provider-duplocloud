package duplocloud

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"terraform-provider-duplocloud/duplocloud/data_sources"
	"terraform-provider-duplocloud/duplocloud/resources"
	"terraform-provider-duplocloud/duplosdk"

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
		},
		ResourcesMap: map[string]*schema.Resource{
			"duplocloud_oci_containerengine_node_pool":   resources.resourceOciContainerEngineNodePool(),
			"duplocloud_admin_system_setting":            resources.resourceAdminSystemSetting(),
			"duplocloud_aws_cloudfront_distribution":     resources.resourceAwsCloudfrontDistribution(),
			"duplocloud_aws_dynamodb_table":              resources.resourceAwsDynamoDBTable(),
			"duplocloud_aws_dynamodb_table_v2":           resources.resourceAwsDynamoDBTableV2(),
			"duplocloud_aws_elasticsearch":               resources.resourceDuploAwsElasticSearch(),
			"duplocloud_aws_host":                        resources.resourceAwsHost(),
			"duplocloud_aws_load_balancer":               resources.resourceAwsLoadBalancer(),
			"duplocloud_aws_load_balancer_listener":      resources.resourceAwsLoadBalancerListener(),
			"duplocloud_aws_kafka_cluster":               resources.resourceAwsKafkaCluster(),
			"duplocloud_aws_lambda_function":             resources.resourceAwsLambdaFunction(),
			"duplocloud_aws_ssm_parameter":               resources.resourceAwsSsmParameter(),
			"duplocloud_duplo_service":                   resources.resourceDuploService(),
			"duplocloud_duplo_service_lbconfigs":         resources.resourceDuploServiceLbConfigs(),
			"duplocloud_duplo_service_params":            resources.resourceDuploServiceParams(),
			"duplocloud_ecache_instance":                 resources.resourceDuploEcacheInstance(),
			"duplocloud_ecs_task_definition":             resources.resourceDuploEcsTaskDefinition(),
			"duplocloud_ecs_service":                     resources.resourceDuploEcsService(),
			"duplocloud_gcp_cloud_function":              resources.resourceGcpCloudFunction(),
			"duplocloud_gcp_pubsub_topic":                resources.resourceGcpPubsubTopic(),
			"duplocloud_gcp_scheduler_job":               resources.resourceGcpSchedulerJob(),
			"duplocloud_gcp_storage_bucket":              resources.resourceGcpStorageBucket(),
			"duplocloud_k8_config_map":                   resources.resourceK8ConfigMap(),
			"duplocloud_k8_secret":                       resources.resourceK8Secret(),
			"duplocloud_k8_ingress":                      resources.resourceK8Ingress(),
			"duplocloud_k8_secret_provider_class":        resources.resourceK8SecretProviderClass(),
			"duplocloud_infrastructure":                  resources.resourceInfrastructure(),
			"duplocloud_infrastructure_setting":          resources.resourceInfrastructureSetting(),
			"duplocloud_infrastructure_subnet":           resources.resourceInfrastructureSubnet(),
			"duplocloud_plan_certificates":               resources.resourcePlanCertificates(),
			"duplocloud_plan_configs":                    resources.resourcePlanConfigs(),
			"duplocloud_plan_settings":                   resources.resourcePlanSettings(),
			"duplocloud_plan_images":                     resources.resourcePlanImages(),
			"duplocloud_rds_instance":                    resources.resourceDuploRdsInstance(),
			"duplocloud_rds_read_replica":                resources.resourceDuploRdsReadReplica(),
			"duplocloud_s3_bucket":                       resources.resourceS3Bucket(),
			"duplocloud_tenant":                          resources.resourceTenant(),
			"duplocloud_user":                            resources.resourceUser(),
			"duplocloud_tenant_config":                   resources.resourceTenantConfig(),
			"duplocloud_tenant_tag":                      resources.resourceTenantTag(),
			"duplocloud_tenant_secret":                   resources.resourceTenantSecret(),
			"duplocloud_tenant_network_security_rule":    resources.resourceTenantSecurityRule(),
			"duplocloud_emr_cluster":                     resources.resourceAwsEmrCluster(),
			"duplocloud_asg_profile":                     resources.resourceAwsASG(),
			"duplocloud_docker_credentials":              resources.resourceDockerCreds(),
			"duplocloud_aws_appautoscaling_target":       resources.resourceAwsAppautoscalingTarget(),
			"duplocloud_aws_appautoscaling_policy":       resources.resourceAwsAppautoscalingPolicy(),
			"duplocloud_aws_cloudwatch_event_rule":       resources.resourceAwsCloudWatchEventRule(),
			"duplocloud_aws_cloudwatch_event_target":     resources.resourceAwsCloudWatchEventTarget(),
			"duplocloud_aws_lambda_permission":           resources.resourceAwsLambdaPermission(),
			"duplocloud_aws_cloudwatch_metric_alarm":     resources.resourceAwsCloudWatchMetricAlarm(),
			"duplocloud_aws_ecr_repository":              resources.resourceAwsEcrRepository(),
			"duplocloud_aws_api_gateway_integration":     resources.resourceAwsApiGatewayIntegration(),
			"duplocloud_aws_target_group_attributes":     resources.resourceAwsTargetGroupAttributes(),
			"duplocloud_aws_lb_target_group":             resources.resourceTargetGroup(),
			"duplocloud_aws_sqs_queue":                   resources.resourceAwsSqsQueue(),
			"duplocloud_aws_sns_topic":                   resources.resourceAwsSnsTopic(),
			"duplocloud_aws_lb_listener_rule":            resources.resourceAwsLbListenerRule(),
			"duplocloud_azure_key_vault_secret":          resources.resourceAzureKeyVaultSecret(),
			"duplocloud_azure_storage_account":           resources.resourceAzureStorageAccount(),
			"duplocloud_azure_mysql_database":            resources.resourceAzureMysqlDatabase(),
			"duplocloud_azure_redis_cache":               resources.resourceAzureRedisCache(),
			"duplocloud_azure_virtual_machine":           resources.resourceAzureVirtualMachine(),
			"duplocloud_azure_sql_managed_database":      resources.resourceAzureSqlManagedDatabase(),
			"duplocloud_azure_postgresql_database":       resources.resourceAzurePostgresqlDatabase(),
			"duplocloud_azure_mssql_server":              resources.resourceAzureMssqlServer(),
			"duplocloud_azure_mssql_database":            resources.resourceAzureMssqlDatabase(),
			"duplocloud_azure_mssql_elasticpool":         resources.resourceAzureMssqlElasticPool(),
			"duplocloud_azure_virtual_machine_scale_set": resources.resourceAzureVirtualMachineScaleSet(),
			"duplocloud_azure_storage_share_file":        resources.resourceAzureStorageShareFile(),
			"duplocloud_azure_log_analytics_workspace":   resources.resourceAzureLogAnalyticsWorkspace(),
			"duplocloud_azure_recovery_services_vault":   resources.resourceAzureRecoveryServicesVault(),
			"duplocloud_azure_vm_feature":                resources.resourceAzureVmFeature(),
			"duplocloud_azure_vault_backup_policy":       resources.resourceAzureVaultBackupPolicy(),
			"duplocloud_azure_network_security_rule":     resources.resourceAzureNetworkSgRule(),
			"duplocloud_azure_k8_node_pool":              resources.resourceAzureK8NodePool(),
			"duplocloud_azure_sql_virtual_network_rule":  resources.resourceAzureSqlServerVnetRule(),
			"duplocloud_azure_sql_firewall_rule":         resources.resourceAzureSqlFirewallRule(),
			"duplocloud_other_agents":                    resources.resourceOtherAgents(),
			"duplocloud_byoh":                            resources.resourceByoh(),
			"duplocloud_aws_mwaa_environment":            resources.resourceMwaaAirflow(),
			"duplocloud_aws_efs_file_system":             resources.resourceAwsEFS(),
			"duplocloud_k8_persistent_volume_claim":      resources.resourceK8PVC(),
			"duplocloud_k8_storage_class":                resources.resourceK8StorageClass(),
			"duplocloud_aws_batch_scheduling_policy":     resources.resourceAwsBatchSchedulingPolicy(),
			"duplocloud_aws_batch_compute_environment":   resources.resourceAwsBatchComputeEnvironment(),
			"duplocloud_aws_batch_job_queue":             resources.resourceAwsBatchJobQueue(),
			"duplocloud_aws_batch_job_definition":        resources.resourceAwsBatchJobDefinition(),
			"duplocloud_aws_timestreamwrite_database":    resources.resourceAwsTimestreamDatabase(),
			"duplocloud_aws_timestreamwrite_table":       resources.resourceAwsTimestreamTable(),
			"duplocloud_aws_rds_tag":                     resources.resourceAwsRdsTag(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"duplocloud_admin_aws_credentials":     data_sources.dataSourceAdminAwsCredentials(),
			"duplocloud_aws_account":               data_sources.dataSourceAwsAccount(),
			"duplocloud_aws_lb_listeners":          data_sources.dataSourceTenantAwsLbListeners(),
			"duplocloud_aws_lb_target_groups":      data_sources.dataSourceTenantAwsLbTargetGroups(),
			"duplocloud_aws_ssm_parameter":         data_sources.dataSourceAwsSsmParameter(),
			"duplocloud_aws_ssm_parameters":        data_sources.dataSourceAwsSsmParameters(),
			"duplocloud_eks_credentials":           data_sources.dataSourceEksCredentials(),
			"duplocloud_duplo_service":             data_sources.dataSourceDuploService(),
			"duplocloud_duplo_services":            data_sources.dataSourceDuploServices(),
			"duplocloud_duplo_service_lbconfigs":   data_sources.dataSourceDuploServiceLbConfigs(),
			"duplocloud_duplo_service_params":      data_sources.dataSourceDuploServiceParams(),
			"duplocloud_ecs_service":               data_sources.dataSourceDuploEcsService(),
			"duplocloud_ecs_services":              data_sources.dataSourceDuploEcsServices(),
			"duplocloud_ecs_task_definition":       data_sources.dataSourceDuploEcsTaskDefinition(),
			"duplocloud_ecs_task_definitions":      data_sources.dataSourceDuploEcsTaskDefinitions(),
			"duplocloud_infrastructure":            data_sources.dataSourceInfrastructure(),
			"duplocloud_infrastructures":           data_sources.dataSourceInfrastructures(),
			"duplocloud_k8_config_map":             data_sources.dataSourceK8ConfigMap(),
			"duplocloud_k8_config_maps":            data_sources.dataSourceK8ConfigMaps(),
			"duplocloud_k8_secret":                 data_sources.dataSourceK8Secret(),
			"duplocloud_k8_secrets":                data_sources.dataSourceK8Secrets(),
			"duplocloud_native_hosts":              data_sources.dataSourceNativeHosts(),
			"duplocloud_native_host_image":         data_sources.dataSourceNativeHostImage(),
			"duplocloud_native_host_images":        data_sources.dataSourceNativeHostImages(),
			"duplocloud_plan":                      data_sources.dataSourcePlan(),
			"duplocloud_plan_image":                data_sources.dataSourcePlanImage(),
			"duplocloud_plan_images":               data_sources.dataSourcePlanImages(),
			"duplocloud_plans":                     data_sources.dataSourcePlans(),
			"duplocloud_tenant":                    data_sources.dataSourceTenant(),
			"duplocloud_tenants":                   data_sources.dataSourceTenants(),
			"duplocloud_tenant_aws_region":         data_sources.dataSourceTenantAwsRegion(),
			"duplocloud_tenant_aws_credentials":    data_sources.dataSourceTenantAwsCredentials(),
			"duplocloud_tenant_aws_kms_key":        data_sources.dataSourceTenantAwsKmsKey(),
			"duplocloud_tenant_aws_kms_keys":       data_sources.dataSourceTenantAwsKmsKeys(),
			"duplocloud_tenant_eks_credentials":    data_sources.dataSourceTenantEksCredentials(),
			"duplocloud_tenant_internal_subnets":   data_sources.dataSourceTenantInternalSubnets(),
			"duplocloud_tenant_external_subnets":   data_sources.dataSourceTenantExternalSubnets(),
			"duplocloud_tenant_config":             data_sources.dataSourceTenantConfig(),
			"duplocloud_tenant_secret":             data_sources.dataSourceTenantSecret(),
			"duplocloud_tenant_secrets":            data_sources.dataSourceTenantSecrets(),
			"duplocloud_emr_cluster":               data_sources.dataSourceEmrClusters(),
			"duplocloud_plan_certificate":          data_sources.dataSourcePlanCert(),
			"duplocloud_plan_certificates":         data_sources.dataSourcePlanCerts(),
			"duplocloud_asg_profiles":              data_sources.dataSourceAsgProfiles(),
			"duplocloud_plan_nat_gateways":         data_sources.dataSourcePlanNgws(),
			"duplocloud_ecr_repository":            data_sources.dataSourceEcrRepository(),
			"duplocloud_azure_storage_account_key": data_sources.dataSourceAzureStorageAccountKey(),
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
	}
	if host == "" {
		host = os.Getenv("duplo_host")
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
		c.HTTPClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return c, diags
}
