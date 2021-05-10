package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func kafkaClusterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description: "The GUID of the tenant that the Kafka cluster will be created in.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"name": {
			Description: "The short name of the Kafka cluster.  Duplo will add a prefix to the name.  You can retrieve the full name from the `fullname` attribute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"fullname": {
			Description: "The full name of the Kakfa cluster.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"arn": {
			Description: "The ARN of the Kafka cluster.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"kafka_version": {
			Description: "The version of the Kafka cluster.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"configuration_arn": {
			Description:  "An ARN of a Kafka configuration to apply to the cluster.",
			Type:         schema.TypeString,
			ForceNew:     true,
			Optional:     true,
			Computed:     true,
			RequiredWith: []string{"configuration_revision"},
		},
		"configuration_revision": {
			Description:  "An revision of a Kafka configuration to apply to the cluster.",
			Type:         schema.TypeInt,
			ForceNew:     true,
			Optional:     true,
			Computed:     true,
			RequiredWith: []string{"configuration_arn"},
		},
		"instance_type": {
			Description: "The Kafka node instance type to use.\n" +
				"See the [AWS documentation](https://docs.aws.amazon.com/msk/latest/developerguide/msk-create-cluster.html) for more information.",
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"storage_size": {
			Description: "The size of the Kafka storage, in gigabytes.",
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
		},
		"az_distribution": {
			Description: "The availability zone distribution used by the cluster.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"subnets": {
			Description: "The list of subnets that the cluster will be launched in.",
			Type:        schema.TypeList,
			ForceNew:    true,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"security_groups": {
			Description: "The list of security groups used by the cluster.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"state": {
			Description: "The current state of the cluster.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"plaintext_zookeeper_connect_string": {
			Description: "The Zookeeper connect string for plaintext (unencrypted) connections.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"tls_zookeeper_connect_string": {
			Description: "The Zookeeper connect string for TLS (encrypted) connections.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"plaintext_bootstrap_broker_string": {
			Description: "The bootstrap broker connect string for plaintext (unencrypted) connections.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"tls_bootstrap_broker_string": {
			Description: "The bootstrap broker connect string for TLS (encrypted) connections.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"tags": {
			Type:     schema.TypeMap,
			Computed: true,
			Elem:     schema.TypeString,
		},
	}
}

// Resource for managing an AWS Kafka cluster
func resourceAwsKafkaCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceKafkaClusterRead,
		CreateContext: resourceKafkaClusterCreate,
		//UpdateContext: resourceKafkaClusterUpdate,
		DeleteContext: resourceKafkaClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: kafkaClusterSchema(),
	}
}

/// READ resource
func resourceKafkaClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceKafkaClusterRead ******** start")

	// Parse the identifying attributes
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	duplo, err := c.TenantGetKafkaCluster(tenantID, name)
	if duplo == nil {
		d.SetId("") // object missing
		return nil
	}
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s kafka cluster '%s': %s", tenantID, name, err)
	}
	info, err := c.TenantGetKafkaClusterInfo(tenantID, duplo.Arn)
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s kafka cluster info '%s': %s", tenantID, name, err)
	}
	bootstrap, err := c.TenantGetKafkaClusterBootstrapBrokers(tenantID, duplo.Arn)
	if err != nil {
		return diag.Errorf("Unable to retrieve tenant %s kafka cluster bootstrap brokers '%s': %s", tenantID, name, err)
	}

	// Set simple fields first.
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("fullname", duplo.Name)
	d.Set("arn", duplo.Arn)

	// Next, set fields that come from extended information.
	if info != nil {
		if info.BrokerNodeGroup != nil {
			d.Set("instance_type", info.BrokerNodeGroup.InstanceType)
			d.Set("storage_size", info.BrokerNodeGroup.StorageInfo.EbsStorageInfo.VolumeSize)
			d.Set("plaintext_zookeeper_connect_string", info.ZookeeperConnectString)
			d.Set("tls_zookeeper_connect_string", info.ZookeeperConnectStringTls)
			if info.BrokerNodeGroup.AZDistribution != nil {
				d.Set("az_distribution", info.BrokerNodeGroup.AZDistribution.Value)
			}
			if info.BrokerNodeGroup.Subnets != nil {
				d.Set("subnets", info.BrokerNodeGroup.Subnets)
			}
			if info.BrokerNodeGroup.SecurityGroups != nil {
				d.Set("security_groups", info.BrokerNodeGroup.SecurityGroups)
			}
		}
		if info.CurrentSoftware != nil {
			d.Set("kafka_version", info.CurrentSoftware.KafkaVersion)
			d.Set("configuration_arn", info.CurrentSoftware.ConfigurationArn)
			d.Set("configuration_revision", info.CurrentSoftware.ConfigurationRevision)
		}
	}
	if bootstrap != nil {
		d.Set("plaintext_bootstrap_broker_string", bootstrap.BootstrapBrokerString)
		d.Set("tls_bootstrap_broker_string", bootstrap.BootstrapBrokerStringTls)
	}
	d.Set("state", info.State.Value)
	d.Set("tags", info.Tags)

	log.Printf("[TRACE] resourceKafkaClusterRead ******** end")
	return nil
}

/// CREATE resource
func resourceKafkaClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceKafkaClusterCreate ******** start")

	// Create the request object.
	rq := duplosdk.DuploKafkaClusterRequest{
		Name:            d.Get("name").(string),
		KafkaVersion:    d.Get("kafka_version").(string),
		BrokerNodeGroup: &duplosdk.DuploKafkaBrokerNodeGroupInfo{InstanceType: d.Get("instance_type").(string)},
	}
	rq.BrokerNodeGroup.StorageInfo.EbsStorageInfo.VolumeSize = d.Get("storage_size").(int)

	// Apply any subnet settings
	if subnets, ok := getAsStringArray(d, "subnets"); ok && len(*subnets) > 0 {
		rq.BrokerNodeGroup.Subnets = subnets
	}

	// Apply any custom configuration.
	if v, ok := d.GetOk("configuration_arn"); ok {
		rq.ConfigurationInfo = &duplosdk.DuploKafkaConfigurationInfo{
			Arn:      v.(string),
			Revision: int64(d.Get("configuration_revision").(int)),
		}
	}

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)

	// Post the object to Duplo
	err := c.TenantCreateKafkaCluster(tenantID, rq)
	if err != nil {
		return diag.Errorf("Error creating tenant %s kafka cluster '%s': %s", tenantID, rq.Name, err)
	}

	// Wait for Duplo to be able to return the cluster's details.
	var rp *duplosdk.DuploKafkaCluster
	var errget error
	id := fmt.Sprintf("%s/%s", tenantID, rq.Name)
	diags := waitForResourceToBePresentAfterCreate(ctx, d, "kafka cluster", id, func() (interface{}, error) {
		rp, errget = c.TenantGetKafkaCluster(tenantID, rq.Name)
		if rp != nil && rp.Arn != "" {
			return rp, errget
		}
		return nil, errget
	})
	if diags != nil {
		return diags
	}
	d.SetId(id)

	// Next, wait for the cluster to become active.
	err = duploKafkaClusterWaitUntilReady(ctx, c, tenantID, rp.Arn, d.Timeout("create"))
	if err != nil {
		return diag.FromErr(err)
	}

	diags = resourceKafkaClusterRead(ctx, d, m)
	log.Printf("[TRACE] resourceKafkaClusterCreate ******** end")
	return diags
}

/// DELETE resource
func resourceKafkaClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceKafkaClusterDelete ******** start")

	// Prepare for the request.
	c := m.(*duplosdk.Client)
	id := d.Id()
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) < 2 {
		return diag.Errorf("Invalid resource ID: %s", id)
	}
	tenantID, name := idParts[0], idParts[1]

	// See if the object still exists in Duplo.
	duplo, err := c.TenantGetKafkaCluster(tenantID, name)
	if err != nil {
		return diag.Errorf("Unable to get kafka cluster '%s': %s", id, err)
	}
	if duplo != nil {
		arn := duplo.Arn

		// Delete the cluster.
		err := c.TenantDeleteKafkaCluster(tenantID, arn)
		if err != nil {
			return diag.Errorf("Error deleting kafka cluster '%s': %s", id, err)
		}

		// Wait up to 60 seconds for Duplo to delete the cluster.
		diag := waitForResourceToBeMissingAfterDelete(ctx, d, "kafka cluster", id, func() (interface{}, error) {
			return c.TenantGetKafkaCluster(tenantID, name)
		})
		if diag != nil {
			return diag
		}
	}

	// Wait 10 more seconds to deal with consistency issues.
	time.Sleep(10 * time.Second)

	log.Printf("[TRACE] resourceKafkaClusterDelete ******** end")
	return nil
}

func duploKafkaClusterWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID, arn string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.TenantGetKafkaClusterInfo(tenantID, arn)
			status := "pending"
			if err == nil && rp != nil && rp.State != nil && rp.State.Value == "ACTIVE" {
				status = "ready"
			}
			return rp, status, err
		},
		// MinTimeout will be 10 sec freq, if times-out forces 30 sec anyway
		PollInterval: 30 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] duploKafkaClusterWaitUntilReady(%s, %s)", tenantID, arn)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}
