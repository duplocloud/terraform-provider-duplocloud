package duplocloud

import (
	"context"
	"fmt"
	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// SCHEMA for resource crud
func resourceInfrastructureOnprem() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_infrastructure_onprem` adds support to integrate on premise infra into duplocloud",

		ReadContext:   resourceInfrastructureOnpremRead,
		CreateContext: resourceInfrastructureOnpremCreate,
		DeleteContext: resourceInfrastructureOnpremDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(50 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"infra_name": {
				Description:  "The name of the infrastructure.  Infrastructure names are globally unique and less than 13 characters.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(2, 30),
			},
			"account_id": {
				Description: "The cloud account ID. Used with GCP cloud",
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
			},
			"region": {
				Description: "The cloud provider region.  The Duplo portal must have already been configured to support this region.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"azcount": {
				Description: "The number of availability zones.  Must be one of: `2`, `3`, or `4`. This is applicable only for AWS.",
				Type:        schema.TypeInt,
				ForceNew:    true,
				Optional:    true,
			},
			"enable_k8_cluster": {
				Description: "Whether or not to provision a kubernetes cluster.",
				Type:        schema.TypeBool,
				ForceNew:    true,
				Required:    true,
			},

			"vendor": {
				Description:  "Type of on premise vendor <br>0 - Rancher<br>1 - Generic<br>2 - EKS<br>",
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntInSlice([]int{0, 1, 2}),
			},
			"cluster_name": {
				Description: "Name of the on premise k8 cluster",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"cluster_endpoint": {
				Description:  "Endpoint URL of K8 cluster",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPS,
			},
			"api_token": {
				Description: "Token to access cluster API's",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"cluster_certificate_authority_data": {
				Description: "Required to validate API server certificates and kubelet client certificates",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"data_center": {
				Description: "Datacenter name of the onpremise cluster",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"eks_config": {
				Description: "EKS configuration for on premise infra if vendor is selected as 2",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_subnets": {
							Description: "The private subnets for the VPC.",
							Type:        schema.TypeList,
							Required:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							ForceNew: true,
						},
						"public_subnets": {
							Description: "The public subnets for the VPC.",
							Type:        schema.TypeList,
							Required:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							ForceNew: true,
						},
						"ingress_security_group_ids": {
							Description: "The security group IDs",
							Type:        schema.TypeList,
							Required:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							ForceNew: true,
						},
						"vpc_id": {
							Description: "The the ID of a Virtual Private Cloud ",
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
						},
					},
				},
			},
			"custom_data": {
				Description: "A list of configuration settings to apply on creation, expressed as key / value pairs.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        KeyValueSchema(),
				ForceNew:    true,
			},
			"status": {
				Description: "The status of the infrastructure.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
		CustomizeDiff: validateOnPremAttribute,
	}
}

// READ resource
func resourceInfrastructureOnpremRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) < 4 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	name := idParts[4]

	log.Printf("[TRACE] resourceInfrastructureOnpremRead(%s): start", name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	config, err := c.InfrastructureGetConfig(name)
	if err != nil {
		return diag.Errorf("%s", err.Error())
	}
	if config == nil {
		d.SetId("") // object missing
		return nil
	}
	flattenInfraOnprem(config, d)
	log.Printf("[TRACE] resourceInfrastructureOnpremRead(%s): end", name)
	return nil
}

// CREATE resource
func resourceInfrastructureOnpremCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var err error

	rq := duploInfrastructureOnpremConfigFromState(d)

	log.Printf("[TRACE] resourceInfrastructureOnpremCreate(%s): start", rq.Name)

	// Post the object to Duplo.
	c := m.(*duplosdk.Client)
	err = c.InfrastructureCreate(rq)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait up to 60 seconds for Duplo to be able to return the infrastructure details.
	id := fmt.Sprintf("v2/admin/Infrastructure/OnPremises/%s", rq.Name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "infrastructure", id, func() (interface{}, duplosdk.ClientError) {
		return c.InfrastructureGetConfig(rq.Name)
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	// Then, wait until the infrastructure is completely ready.
	err = duploInfrastructureWaitUntilReady(ctx, c, rq.Name, d.Timeout("create"))
	if err != nil {
		return diag.FromErr(err)
	}

	diags = resourceInfrastructureOnpremRead(ctx, d, m)
	log.Printf("[TRACE] resourceInfrastructureOnpremCreate(%s): end", rq.Name)
	return diags
}

// DELETE resource
func resourceInfrastructureOnpremDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	idParts := strings.SplitN(id, "/", 5)
	if len(idParts) < 4 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	infraName := idParts[4]
	log.Printf("[TRACE] resourceInfrastructureOnpremDelete(%s): start", infraName)

	c := m.(*duplosdk.Client)
	err := c.InfrastructureDelete(infraName)
	if err != nil {
		if err.Status() == 404 {
			return nil
		}
		return diag.FromErr(err)
	}
	//sleep added for safer side. get api used for polling returns 500 instead of 404
	time.Sleep(30 * time.Second)
	log.Printf("[TRACE] resourceInfrastructureOnpremDelete(%s): end", infraName)
	return nil
}

func duploInfrastructureOnpremConfigFromState(d *schema.ResourceData) duplosdk.DuploInfrastructureConfig {

	obj := duplosdk.DuploInfrastructureConfig{
		Name:            d.Get("infra_name").(string),
		AccountId:       d.Get("account_id").(string),
		Cloud:           10,
		Region:          d.Get("region").(string),
		AzCount:         d.Get("azcount").(int),
		EnableK8Cluster: d.Get("enable_k8_cluster").(bool),
		CustomData:      keyValueFromState("custom_data", d),
		Vnet:            &duplosdk.DuploInfrastructureVnet{},
	}
	config := duplosdk.DuploOnPremK8Config{
		Name:                           d.Get("cluster_name").(string),
		Vendor:                         d.Get("vendor").(int),
		ClusterEndpoint:                d.Get("cluster_endpoint").(string),
		ApiToken:                       d.Get("api_token").(string),
		CertificateAuthorityDataBase64: d.Get("cluster_certificate_authority_data").(string),
	}
	if config.Vendor == 2 {
		eksConfig := duplosdk.DuploOnPremEKSConfig{}
		for _, ps := range d.Get("eks_config.0.private_subnets").([]interface{}) {
			eksConfig.PrivateSubnets = append(eksConfig.PrivateSubnets, ps.(string))
		}
		for _, ps := range d.Get("eks_config.0.public_subnets").([]interface{}) {
			eksConfig.PublicSubnets = append(eksConfig.PublicSubnets, ps.(string))
		}
		for _, gid := range d.Get("eks_config.0.ingress_security_group_ids").([]interface{}) {
			eksConfig.IngressSecurityGroupIds = append(eksConfig.IngressSecurityGroupIds, gid.(string))
		}
		eksConfig.VpcId = d.Get("eks_config.0.vpc_id").(string)
		config.OnPremEKSConfig = &eksConfig
	}
	obj.OnPrem = &duplosdk.DuploOnPrem{
		DataCenter:     d.Get("data_center").(string),
		OnPremK8Config: &config,
	}
	return obj
}

func flattenInfraOnprem(infra *duplosdk.DuploInfrastructureConfig, d *schema.ResourceData) {
	// Once backend API is fixed for default infra, Remove this.

	d.Set("infra_name", infra.Name)
	d.Set("account_id", infra.AccountId)
	d.Set("cloud", infra.Cloud)
	d.Set("region", infra.Region)
	d.Set("azcount", infra.AzCount)
	d.Set("enable_k8_cluster", infra.EnableK8Cluster)
	if len(*infra.CustomData) > 0 {
		d.Set("custom_data", keyValueToState("custom_data", infra.CustomData))
	}
	if infra.OnPrem != nil {
		d.Set("data_center", infra.OnPrem.DataCenter)
		if infra.OnPrem.OnPremK8Config != nil {
			d.Set("name", infra.OnPrem.OnPremK8Config.Name)
			d.Set("api_token", infra.OnPrem.OnPremK8Config.ApiToken)
			d.Set("cluster_endpoint", infra.OnPrem.OnPremK8Config.ClusterEndpoint)
			d.Set("cluster_certificate_authority_data", infra.OnPrem.OnPremK8Config.CertificateAuthorityDataBase64)
			d.Set("vendor", infra.OnPrem.OnPremK8Config.Vendor)
			if infra.OnPrem.OnPremK8Config.OnPremEKSConfig != nil {
				d.Set("eks_config.0.private_subnets", infra.OnPrem.OnPremK8Config.OnPremEKSConfig.PrivateSubnets)
				d.Set("eks_config.0.public_subnets", infra.OnPrem.OnPremK8Config.OnPremEKSConfig.PublicSubnets)
				d.Set("eks_config.0.ingress_security_group_ids", infra.OnPrem.OnPremK8Config.OnPremEKSConfig.IngressSecurityGroupIds)
				d.Set("eks_config.0.vpc_id", infra.OnPrem.OnPremK8Config.OnPremEKSConfig.VpcId)

			}
		}
	}
	d.Set("status", infra.ProvisioningStatus)

}

func validateOnPremAttribute(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	vendor := diff.Get("vendor").(int)
	eks := diff.Get("eks_config").([]interface{})

	if vendor == 2 && len(eks) == 0 {
		return fmt.Errorf("for vendor eks, eks_config required")
	}

	return nil
}
