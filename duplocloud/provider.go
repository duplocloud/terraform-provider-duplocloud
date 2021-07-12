package duplocloud

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
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
			"duplocloud_aws_dynamodb_table":           resourceAwsDynamoDBTable(),
			"duplocloud_aws_elasticsearch":            resourceDuploAwsElasticSearch(),
			"duplocloud_aws_host":                     resourceAwsHost(),
			"duplocloud_aws_load_balancer":            resourceAwsLoadBalancer(),
			"duplocloud_aws_kafka_cluster":            resourceAwsKafkaCluster(),
			"duplocloud_aws_lambda_function":          resourceAwsLambdaFunction(),
			"duplocloud_duplo_service":                resourceDuploService(),
			"duplocloud_duplo_service_lbconfigs":      resourceDuploServiceLBConfigs(),
			"duplocloud_duplo_service_params":         resourceDuploServiceParams(),
			"duplocloud_ecache_instance":              resourceDuploEcacheInstance(),
			"duplocloud_ecs_task_definition":          resourceDuploEcsTaskDefinition(),
			"duplocloud_ecs_service":                  resourceDuploEcsService(),
			"duplocloud_k8_config_map":                resourceK8ConfigMap(),
			"duplocloud_k8_secret":                    resourceK8Secret(),
			"duplocloud_infrastructure":               resourceInfrastructure(),
			"duplocloud_infrastructure_subnet":        resourceInfrastructureSubnet(),
			"duplocloud_plan_certificates":            resourcePlanCertificates(),
			"duplocloud_rds_instance":                 resourceDuploRdsInstance(),
			"duplocloud_s3_bucket":                    resourceS3Bucket(),
			"duplocloud_tenant":                       resourceTenant(),
			"duplocloud_tenant_config":                resourceTenantConfig(),
			"duplocloud_tenant_secret":                resourceTenantSecret(),
			"duplocloud_tenant_network_security_rule": resourceTenantSecurityRule(),
			"duplocloud_emr_cluster":                  resourceAwsEmrCluster(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"duplocloud_admin_aws_credentials":   dataSourceAdminAwsCredentials(),
			"duplocloud_aws_account":             dataSourceAwsAccount(),
			"duplocloud_aws_lb_listeners":        dataSourceTenantAwsLbListeners(),
			"duplocloud_aws_lb_target_groups":    dataSourceTenantAwsLbTargetGroups(),
			"duplocloud_eks_credentials":         dataSourceEksCredentials(),
			"duplocloud_duplo_service":           dataSourceDuploService(),
			"duplocloud_duplo_services":          dataSourceDuploServices(),
			"duplocloud_duplo_service_lbconfigs": dataSourceDuploServiceLBConfigs(),
			"duplocloud_duplo_service_params":    dataSourceDuploServiceParams(),
			"duplocloud_infrastructure":          dataSourceInfrastructure(),
			"duplocloud_infrastructures":         dataSourceInfrastructures(),
			"duplocloud_k8_config_map":           dataSourceK8ConfigMap(),
			"duplocloud_k8_config_maps":          dataSourceK8ConfigMaps(),
			"duplocloud_k8_secret":               dataSourceK8Secret(),
			"duplocloud_native_hosts":            dataSourceNativeHosts(),
			"duplocloud_plan":                    dataSourcePlan(),
			"duplocloud_plans":                   dataSourcePlans(),
			"duplocloud_tenant":                  dataSourceTenant(),
			"duplocloud_tenants":                 dataSourceTenants(),
			"duplocloud_tenant_aws_region":       dataSourceTenantAwsRegion(),
			"duplocloud_tenant_aws_credentials":  dataSourceTenantAwsCredentials(),
			"duplocloud_tenant_aws_kms_key":      dataSourceTenantAwsKmsKey(),
			"duplocloud_tenant_aws_kms_keys":     dataSourceTenantAwsKmsKeys(),
			"duplocloud_tenant_eks_credentials":  dataSourceTenantEksCredentials(),
			"duplocloud_tenant_internal_subnets": dataSourceTenantInternalSubnets(),
			"duplocloud_tenant_external_subnets": dataSourceTenantExternalSubnets(),
			"duplocloud_tenant_config":           dataSourceTenantConfig(),
			"duplocloud_tenant_secret":           dataSourceTenantSecret(),
			"duplocloud_tenant_secrets":          dataSourceTenantSecrets(),
			"duplocloud_emr_cluster":             dataSourceEmrClusters(),
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
