package duplocloud

import (
	"context"
	"log"
	"strconv"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func planSchema(single bool) map[string]*schema.Schema {
	result := map[string]*schema.Schema{
		"cloud": {
			Description: "The numerical index of the cloud provider for this plan" +
				"Will be one of:\n\n" +
				"   - `0` : AWS\n" +
				"   - `2` : Azure\n" +
				"   - `3` : GCP\n",
			Type:     schema.TypeInt,
			Computed: true,
		},
		"region": {
			Description: "The cloud provider region.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"account_id": {
			Description: "The cloud account ID.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"vpc_id": {
			Description: "The VPC or VNet ID.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"private_subnet_ids": {
			Description: "The private subnets for the VPC or VNet.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"public_subnet_ids": {
			Description: "The public subnets for the VPC or VNet.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"availability_zones": {
			Description: "A list of the Availability Zones available to the plan.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"waf_infos": {
			Description: "Plan web application firewalls that can be attached to load balancers",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"id": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"kms_keys": {
			Description: "Plan KMS keys that can be used for cloud-based encryption",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"arn": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"certificates": {
			Description: "Plan certificates that can be attached to load balancers",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"arn": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"images": {
			Description: "Plan images that can be used to launch native hosts",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"image_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"os": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"username": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"tags": {
						Type:     schema.TypeList,
						Computed: true,
						Elem:     KeyValueSchema(),
					},
				},
			},
		},
		"metadata": {
			Description: "Plan metadata",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        KeyValueSchema(),
		},
		"config": {
			Description: "Plan configuration data",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        CustomDataExSchema(),
		},
		"capabilities": {
			Description: "Map of capability flags",
			Type:        schema.TypeMap,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeBool},
		},
		"cloud_config": {
			Description: "Cloud-specific plan configuration data",
			Type:        schema.TypeMap,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"kubernetes_config": {
			Description: "Kubernetes-specific plan configuration data",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"api_server": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"token": {
						Type:      schema.TypeString,
						Computed:  true,
						Sensitive: true,
					},
					"provider": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"region": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"version": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"certificate_authority_data": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
	}

	if single {
		result["plan_id"] = &schema.Schema{
			Description: "The plan ID",
			Type:        schema.TypeString,
			Required:    true,
		}
	} else {
		result["plan_id"] = &schema.Schema{
			Description: "The plan ID",
			Type:        schema.TypeString,
			Computed:    true,
		}
	}
	return result
}

func dataSourcePlans() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plans` retrieves a list of plans from Duplo.",

		ReadContext: dataSourcePlansRead,
		Schema: map[string]*schema.Schema{
			"data": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: planSchema(false),
				},
			},
		},
	}
}

func dataSourcePlan() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_plan` retrieves details of a plan in Duplo.",

		ReadContext: dataSourcePlanRead,
		Schema:      planSchema(true),
	}
}

func dataSourcePlansRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourcePlansRead(): start")

	// Retrieve the objects from duplo.
	c := m.(*duplosdk.Client)
	list, err := c.PlanGetList()
	if err != nil {
		return diag.FromErr(err)
	}

	// Populate the results from the list.
	data := make([]interface{}, 0, len(*list))
	for _, duplo := range *list {
		plan := map[string]interface{}{
			"plan_id": duplo.Name,
		}
		flattenPlanCloudConfig(plan, &duplo)

		data = append(data, plan)
	}

	if err := d.Set("data", data); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	log.Printf("[TRACE] dataSourcePlansRead(): end")
	return nil
}

// READ/SEARCH resources
func dataSourcePlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] dataSourcePlanRead(): start")

	c := m.(*duplosdk.Client)
	planId := d.Get("plan_id").(string)

	// Get the object from Duplo, detecting a missing object
	plan, err := c.PlanGet(planId)
	if err != nil {
		return diag.Errorf("Unable to retrieve plan '%s': %s", planId, err)
	}
	if plan == nil {
		return diag.Errorf("No plan named '%s' was found", planId)
	}
	d.SetId(planId)

	attrs := map[string]interface{}{}
	flattenPlanCloudConfig(attrs, plan)
	for k, v := range attrs {
		d.Set(k, v)
	}

	log.Printf("[TRACE] dataSourcePlanRead(): end")
	return nil
}

func flattenPlanCloudConfig(plan map[string]interface{}, duplo *duplosdk.DuploPlan) {

	if duplo.CloudPlatforms != nil && len(*duplo.CloudPlatforms) > 0 {
		cp := (*duplo.CloudPlatforms)[0]
		plan["cloud"] = cp.Platform

		if cp.Platform == 2 {
			plan["cloud_config"] = cp.AzureConfig
			plan["account_id"] = cp.AzureConfig["SubscriptionId"]
			plan["vpc_id"] = cp.AzureConfig["VnetId"]
			plan["region"] = cp.AzureConfig["Region"]
		} else if cp.Platform == 3 {
			plan["cloud_config"] = cp.GoogleConfig
			plan["account_id"] = cp.GoogleConfig["GcpProjectId"]
			plan["vpc_id"] = cp.GoogleConfig["VnetId"]
			plan["region"] = cp.GoogleConfig["GcpRegion"]
		}

	} else if duplo.AwsConfig != nil && len(duplo.AwsConfig) > 0 {
		region := duplo.AwsConfig["AwsRegion"]
		plan["cloud"] = 0
		plan["cloud_config"] = duplo.AwsConfig
		plan["account_id"] = duplo.AwsConfig["AwsAccountId"]
		plan["vpc_id"] = duplo.AwsConfig["VpcId"]
		plan["region"] = region
		if duplo.AwsConfig["AwsElbSubnet"] != nil {
			plan["public_subnet_ids"] = strings.Split(duplo.AwsConfig["AwsElbSubnet"].(string), ";")
		}
		if duplo.AwsConfig["AwsInternalElbSubnet"] != nil {
			ids := strings.Split(duplo.AwsConfig["AwsInternalElbSubnet"].(string), ";")
			plan["private_subnet_ids"] = ids
			az := make([]string, 0, len(ids))
			for i := 97; i < 97+len(ids); i++ {
				az = append(az, region.(string)+string(rune(i)))
			}
			plan["availability_zones"] = az
		}
	}

	plan["metadata"] = keyValueToState("metadata", duplo.MetaData)
	plan["config"] = customDataExToState("config", duplo.PlanConfigData)
	plan["capabilities"] = duplo.Capabilities

	if duplo.Images != nil && len(*duplo.Images) > 0 {
		plan["images"] = flattenPlanImages(duplo.Images)
	}

	if duplo.Certificates != nil && len(*duplo.Certificates) > 0 {
		plan["certificates"] = flattenPlanCertificates(duplo.Certificates)
	}

	if duplo.WafInfos != nil && len(*duplo.WafInfos) > 0 {
		plan["waf_infos"] = flattenPlanWafs(duplo.WafInfos)
	}

	if duplo.KmsKeyInfos != nil && len(*duplo.KmsKeyInfos) > 0 {
		plan["kms_keys"] = flattenPlanKmsKeys(duplo.KmsKeyInfos)
	}

	if duplo.K8ClusterConfigs != nil && len(*duplo.K8ClusterConfigs) > 0 {
		plan["kubernetes_config"] = flattenKubernetesConfig(duplo.K8ClusterConfigs)
	}
}

func flattenPlanImages(list *[]duplosdk.DuploPlanImage) []interface{} {
	result := make([]interface{}, 0, len(*list))

	for _, image := range *list {
		result = append(result, map[string]interface{}{
			"name":     image.Name,
			"image_id": image.ImageId,
			"os":       image.OS,
			"username": image.Username,
			"tags":     keyValueToState("images[].tags", image.Tags),
		})
	}

	return result
}

func flattenPlanCertificates(list *[]duplosdk.DuploPlanCertificate) []interface{} {
	result := make([]interface{}, 0, len(*list))

	for _, cert := range *list {
		result = append(result, map[string]interface{}{
			"name": cert.CertificateName,
			"id":   cert.CertificateArn,
			"arn":  cert.CertificateArn,
		})
	}

	return result
}

func flattenPlanWafs(list *[]duplosdk.DuploPlanWAF) []interface{} {
	result := make([]interface{}, 0, len(*list))

	for _, waf := range *list {
		result = append(result, map[string]interface{}{
			"name":          waf.WebAclName,
			"id":            waf.WebAclId,
			"dashboard_url": waf.DashboardUrl,
		})
	}

	return result
}

func flattenPlanKmsKeys(list *[]duplosdk.DuploPlanKmsKeyInfo) []interface{} {
	result := make([]interface{}, 0, len(*list))

	for _, kms := range *list {
		result = append(result, map[string]interface{}{
			"name": kms.KeyName,
			"id":   kms.KeyId,
			"arn":  kms.KeyArn,
		})
	}

	return result
}

func flattenKubernetesConfig(list *[]duplosdk.DuploPlanK8ClusterConfig) []interface{} {
	kc := (*list)[0]

	return []interface{}{map[string]interface{}{
		"name":                       kc.Name,
		"api_server":                 kc.ApiServer,
		"token":                      kc.Token,
		"provider":                   kc.K8Provider,
		"region":                     kc.AwsRegion,
		"version":                    kc.K8sVersion,
		"certificate_authority_data": kc.CertificateAuthorityDataBase64,
	}}
}
