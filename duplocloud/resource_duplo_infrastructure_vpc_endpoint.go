package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceInfrastructureVpcEndpoint() *schema.Resource {
	return &schema.Resource{
		Description:   "`duplocloud_infrastructure_vpc_endpoint` manages a VPC endpoint for a DuploCloud infrastructure.",
		ReadContext:   resourceInfrastructureVpcEndpointRead,
		CreateContext: resourceInfrastructureVpcEndpointCreate,
		DeleteContext: resourceInfrastructureVpcEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"infra_name": {
				Description: "The name of the infrastructure.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"service_name": {
				Description: "The AWS service name for the VPC endpoint (e.g. `com.amazonaws.us-west-2.s3`).",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"vpc_endpoint_type": {
				Description:  "The type of VPC endpoint. Must be one of `Interface` or `Gateway`.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "Interface",
				ValidateFunc: validation.StringInSlice([]string{"Interface", "Gateway"}, false),
			},
			"vpc_endpoint_id": {
				Description: "The ID of the VPC endpoint.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the VPC endpoint (e.g. `available`, `pending`).",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vpc_id": {
				Description: "The ID of the VPC.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"private_dns_enabled": {
				Description: "Whether private DNS is enabled for the VPC endpoint.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"route_table_ids": {
				Description: "The route table IDs associated with the VPC endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"subnet_ids": {
				Description: "The subnet IDs associated with the VPC endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"network_interface_ids": {
				Description: "The network interface IDs associated with the VPC endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"dns_entries": {
				Description: "The DNS entries for the VPC endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_name": {
							Description: "The DNS name.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"hosted_zone_id": {
							Description: "The hosted zone ID.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"security_groups": {
				Description: "The security groups associated with the VPC endpoint.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_id": {
							Description: "The security group ID.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"group_name": {
							Description: "The security group name.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func resourceInfrastructureVpcEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	log.Printf("[TRACE] resourceInfrastructureVpcEndpointRead(%s): start", id)

	infraName, endpointId, err := parseInfrastructureVpcEndpointId(id)
	if err != nil {
		return diag.FromErr(err)
	}

	c := m.(*duplosdk.Client)
	duplo, cerr := c.InfrastructureGetVpcEndpoint(infraName, endpointId)
	if cerr != nil {
		if cerr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve VPC endpoint '%s': %s", id, cerr)
	}
	if duplo == nil {
		d.SetId("")
		return nil
	}

	flattenInfrastructureVpcEndpoint(d, infraName, duplo)

	log.Printf("[TRACE] resourceInfrastructureVpcEndpointRead(%s): end", id)
	return nil
}

func resourceInfrastructureVpcEndpointCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	infraName := d.Get("infra_name").(string)
	serviceName := d.Get("service_name").(string)
	endpointType := d.Get("vpc_endpoint_type").(string)

	log.Printf("[TRACE] resourceInfrastructureVpcEndpointCreate(%s, %s): start", infraName, serviceName)

	rq := duplosdk.DuploVpcEndpointCreateRequest{
		ServiceName:     serviceName,
		VpcEndpointType: duplosdk.DuploStringValue{Value: endpointType},
	}

	c := m.(*duplosdk.Client)
	rp, cerr := c.InfrastructureCreateVpcEndpoint(infraName, rq)
	if cerr != nil {
		return diag.Errorf("Error creating VPC endpoint for infrastructure '%s': %s", infraName, cerr)
	}

	// If the response includes the endpoint ID, use it directly.
	if rp != nil && rp.VpcEndpointId != "" {
		id := fmt.Sprintf("%s/%s", infraName, rp.VpcEndpointId)
		d.SetId(id)
	} else {
		// Otherwise, find the endpoint by service name.
		endpoint, err := waitForVpcEndpointCreated(ctx, c, infraName, serviceName, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return diag.Errorf("Error waiting for VPC endpoint '%s' to be created: %s", serviceName, err)
		}
		id := fmt.Sprintf("%s/%s", infraName, endpoint.VpcEndpointId)
		d.SetId(id)
	}

	diags := resourceInfrastructureVpcEndpointRead(ctx, d, m)
	log.Printf("[TRACE] resourceInfrastructureVpcEndpointCreate(%s, %s): end", infraName, serviceName)
	return diags
}

func resourceInfrastructureVpcEndpointDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	log.Printf("[TRACE] resourceInfrastructureVpcEndpointDelete(%s): start", id)

	infraName, endpointId, err := parseInfrastructureVpcEndpointId(id)
	if err != nil {
		return diag.FromErr(err)
	}

	c := m.(*duplosdk.Client)
	cerr := c.InfrastructureDeleteVpcEndpoint(infraName, endpointId)
	if cerr != nil {
		if cerr.Status() == 404 {
			return nil
		}
		return diag.Errorf("Error deleting VPC endpoint '%s': %s", id, cerr)
	}

	// Wait for the endpoint to be fully deleted.
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *retry.RetryError {
		ep, cerr := c.InfrastructureGetVpcEndpoint(infraName, endpointId)
		if cerr != nil {
			return retry.NonRetryableError(fmt.Errorf("error checking VPC endpoint status: %s", cerr))
		}
		if ep == nil {
			return nil
		}
		return retry.RetryableError(fmt.Errorf("VPC endpoint %s still exists in state %s", endpointId, ep.State.Value))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceInfrastructureVpcEndpointDelete(%s): end", id)
	return nil
}

func flattenInfrastructureVpcEndpoint(d *schema.ResourceData, infraName string, duplo *duplosdk.DuploVpcEndpoint) {
	d.Set("infra_name", infraName)
	d.Set("service_name", duplo.ServiceName)
	d.Set("vpc_endpoint_type", duplo.VpcEndpointType.Value)
	d.Set("vpc_endpoint_id", duplo.VpcEndpointId)
	d.Set("state", duplo.State.Value)
	d.Set("vpc_id", duplo.VpcId)
	d.Set("private_dns_enabled", duplo.PrivateDnsEnabled)
	d.Set("route_table_ids", duplo.RouteTableIds)
	d.Set("subnet_ids", duplo.SubnetIds)
	d.Set("network_interface_ids", duplo.NetworkInterfaceIds)

	if duplo.DnsEntries != nil {
		dnsEntries := make([]map[string]interface{}, 0, len(duplo.DnsEntries))
		for _, entry := range duplo.DnsEntries {
			dnsEntries = append(dnsEntries, map[string]interface{}{
				"dns_name":       entry.DnsName,
				"hosted_zone_id": entry.HostedZoneId,
			})
		}
		d.Set("dns_entries", dnsEntries)
	}

	if duplo.Groups != nil {
		groups := make([]map[string]interface{}, 0, len(duplo.Groups))
		for _, group := range duplo.Groups {
			groups = append(groups, map[string]interface{}{
				"group_id":   group.GroupId,
				"group_name": group.GroupName,
			})
		}
		d.Set("security_groups", groups)
	}
}

func parseInfrastructureVpcEndpointId(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid resource ID: %s", id)
	}
	return parts[0], parts[1], nil
}

func waitForVpcEndpointCreated(ctx context.Context, c *duplosdk.Client, infraName, serviceName string, timeout time.Duration) (*duplosdk.DuploVpcEndpoint, error) {
	var endpoint *duplosdk.DuploVpcEndpoint

	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		list, cerr := c.InfrastructureGetVpcEndpoints(infraName)
		if cerr != nil {
			return retry.NonRetryableError(fmt.Errorf("error listing VPC endpoints: %s", cerr))
		}
		if list != nil {
			for _, ep := range *list {
				if ep.ServiceName == serviceName {
					endpoint = &ep
					return nil
				}
			}
		}
		return retry.RetryableError(fmt.Errorf("VPC endpoint for service %s not yet created", serviceName))
	})

	return endpoint, err
}
