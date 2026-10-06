package duplocloud

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func gcpBigtableClusterSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cluster_id": {
			Description: "The ID of the Bigtable cluster.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"zone": {
			Description: "The zone in which the cluster runs (e.g. `us-east1-b`).",
			Type:        schema.TypeString,
			Required:    true,
		},
		"num_nodes": {
			Description: "The number of nodes for manual scaling. Conflicts with `autoscaling_config`; " +
				"when that is set, leave this unset and it reflects the current node count.",
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"autoscaling_config": {
			Description: "Autoscaling configuration for the cluster. When set, the cluster scales automatically " +
				"and `num_nodes` must be left unset.",
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"min_nodes": {
						Description:  "Minimum number of nodes for autoscaling.",
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntAtLeast(1),
					},
					"max_nodes": {
						Description:  "Maximum number of nodes for autoscaling. Must be at least `min_nodes`.",
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntAtLeast(1),
					},
					"cpu_target": {
						Description:  "The target CPU utilization percentage that drives autoscaling (10-80).",
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(10, 80),
					},
					"storage_target": {
						Description: "The target storage utilization in GiB per node that drives autoscaling. " +
							"Must be 2560-5120 for `SSD` or 8192-16384 for `HDD`. Defaults to the GCP recommended value when unset.",
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
				},
			},
		},
		"state": {
			Description: "The current state of the cluster (e.g. `READY`).",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}
}

func gcpBigtableInstanceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "GUID of the tenant the Bigtable instance will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The ID of the Bigtable instance. Used verbatim as the instance ID in GCP. " +
				"Must be 6-33 characters of lowercase letters, digits and hyphens, starting with a letter " +
				"and not ending with a hyphen.",
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][a-z0-9-]{4,31}[a-z0-9]$`),
				"must be 6-33 characters of lowercase letters, digits and hyphens, start with a letter and not end with a hyphen"),
		},
		"instance_type": {
			Description: "The type of the Bigtable instance. Must be one of `PRODUCTION` or `DEVELOPMENT`.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "PRODUCTION",
			ValidateFunc: validation.StringInSlice([]string{
				"PRODUCTION", "DEVELOPMENT",
			}, false),
		},
		"display_name": {
			Description: "The human-readable display name of the Bigtable instance (4-30 characters). " +
				"Defaults to `name`, cut to 30 characters.",
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringLenBetween(4, 30),
		},
		"storage_type": {
			Description: "Storage type for the instance's clusters. Must be one of `SSD` or `HDD`. " +
				"All clusters in a Bigtable instance share the same storage type, and GCP does not allow " +
				"changing it after creation; changing this forces a new instance.",
			Type:     schema.TypeString,
			Optional: true,
			Default:  "SSD",
			ForceNew: true,
			ValidateFunc: validation.StringInSlice([]string{
				"SSD", "HDD",
			}, false),
		},
		"labels": {
			Description: "Resource labels for user-provided metadata.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"state": {
			Description: "The current state of the Bigtable instance (e.g. `READY`).",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"cluster": {
			Description: "The clusters that belong to the Bigtable instance. At least one cluster is required. " +
				"Clusters are matched to the backend by `cluster_id`; list them in a stable order, " +
				"as reordering the blocks in configuration produces a diff.",
			Type:     schema.TypeList,
			Required: true,
			MinItems: 1,
			Elem:     &schema.Resource{Schema: gcpBigtableClusterSchema()},
		},
		"wait_until_ready": {
			Description: "Whether or not to wait until the Bigtable instance is ready, after creation.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
	}
}

func resourceGcpBigtableInstance() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_gcp_bigtable_instance` manages a GCP Bigtable instance and its clusters in Duplo.",

		ReadContext:   resourceGcpBigtableInstanceRead,
		CreateContext: resourceGcpBigtableInstanceCreate,
		UpdateContext: resourceGcpBigtableInstanceUpdate,
		DeleteContext: resourceGcpBigtableInstanceDelete,
		Importer: &schema.ResourceImporter{
			// wait_until_ready is a provider-side behavior flag with no backend
			// representation, so seed it to its default on import to avoid a
			// spurious post-import diff.
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				d.Set("wait_until_ready", true)
				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema:        gcpBigtableInstanceSchema(),
		CustomizeDiff: validateBigtableClusters,
	}
}

func resourceGcpBigtableInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpBigtableInstanceRead ******** start")

	tenantID, name, err := parseGcpBigtableInstanceIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	c := m.(*duplosdk.Client)
	instance, clientErr := c.GcpBigtableInstanceGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s Bigtable instance '%s': %s", tenantID, name, clientErr)
	}
	// The backend returns a null body (zero-value struct) when the instance is gone.
	if instance == nil || instance.Name == "" {
		d.SetId("")
		return nil
	}

	clusters, clientErr := c.GcpBigtableClusterList(tenantID, name)
	if clientErr != nil {
		return diag.Errorf("Unable to retrieve tenant %s Bigtable instance '%s' clusters: %s", tenantID, name, clientErr)
	}

	resourceGcpBigtableInstanceSetData(d, tenantID, name, instance, clusters)

	log.Printf("[TRACE] resourceGcpBigtableInstanceRead ******** end")
	return nil
}

func resourceGcpBigtableInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpBigtableInstanceCreate ******** start")

	c := m.(*duplosdk.Client)
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	rq := expandGcpBigtableCreateRequest(d)

	_, clientErr := c.GcpBigtableInstanceCreate(tenantID, rq)
	if clientErr != nil {
		return diag.Errorf("Error creating tenant %s Bigtable instance '%s': %s", tenantID, name, clientErr)
	}

	id := fmt.Sprintf("%s/%s", tenantID, name)

	// Wait for the instance to be present after the async create operation.
	if diags := waitForResourceToBePresentAfterCreate(ctx, d, "Bigtable instance", id, func() (interface{}, duplosdk.ClientError) {
		instance, err := c.GcpBigtableInstanceGet(tenantID, name)
		if err != nil || instance == nil || instance.Name == "" {
			return nil, err
		}
		return instance, nil
	}); diags != nil {
		return diags
	}

	d.SetId(id)

	if d.Get("wait_until_ready").(bool) {
		if err := gcpBigtableInstanceWaitUntilReady(ctx, c, tenantID, name, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.FromErr(err)
		}
		if err := gcpBigtableWaitUntilClustersReady(ctx, c, tenantID, name, configuredBigtableClusterIDs(d), d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.FromErr(err)
		}
	} else if err := gcpBigtableWaitUntilClustersPresent(ctx, c, tenantID, name, configuredBigtableClusterIDs(d), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}

	diags := resourceGcpBigtableInstanceRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpBigtableInstanceCreate ******** end")
	return diags
}

func resourceGcpBigtableInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpBigtableInstanceUpdate ******** start")

	tenantID, name, err := parseGcpBigtableInstanceIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)

	// Update instance-level fields.
	if d.HasChanges("display_name", "instance_type", "labels") {
		rq := &duplosdk.DuploBigtableInstanceUpdateRequest{
			DisplayName: d.Get("display_name").(string),
			Type:        bigtableTypeToInt(d.Get("instance_type").(string)),
		}
		if d.HasChange("labels") {
			// expandAsStringMap returns nil for an empty map, which would leave the
			// removed labels in place.
			labels := expandAsStringMap("labels", d)
			if labels == nil {
				labels = map[string]string{}
			}
			rq.Labels = &labels
		}
		if _, clientErr := c.GcpBigtableInstanceUpdate(tenantID, name, rq); clientErr != nil {
			return diag.Errorf("Error updating tenant %s Bigtable instance '%s': %s", tenantID, name, clientErr)
		}
	}

	// Reconcile clusters.
	if d.HasChange("cluster") {
		if diags := reconcileGcpBigtableClusters(ctx, d, c, tenantID, name); diags != nil {
			return diags
		}
	}

	if d.Get("wait_until_ready").(bool) {
		if err := gcpBigtableInstanceWaitUntilReady(ctx, c, tenantID, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
		if err := gcpBigtableWaitUntilClustersReady(ctx, c, tenantID, name, configuredBigtableClusterIDs(d), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	} else if err := gcpBigtableWaitUntilClustersPresent(ctx, c, tenantID, name, configuredBigtableClusterIDs(d), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return diag.FromErr(err)
	}

	diags := resourceGcpBigtableInstanceRead(ctx, d, m)
	log.Printf("[TRACE] resourceGcpBigtableInstanceUpdate ******** end")
	return diags
}

func resourceGcpBigtableInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[TRACE] resourceGcpBigtableInstanceDelete ******** start")

	tenantID, name, err := parseGcpBigtableInstanceIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	c := m.(*duplosdk.Client)

	if clientErr := c.GcpBigtableInstanceDelete(tenantID, name); clientErr != nil {
		return diag.Errorf("Error deleting tenant %s Bigtable instance '%s': %s", tenantID, name, clientErr)
	}

	if diags := waitForResourceToBeMissingAfterDelete(ctx, d, "Bigtable instance", d.Id(), func() (interface{}, duplosdk.ClientError) {
		instance, err := c.GcpBigtableInstanceGet(tenantID, name)
		if err != nil || instance == nil || instance.Name == "" {
			return nil, err
		}
		return instance, nil
	}); diags != nil {
		return diags
	}

	log.Printf("[TRACE] resourceGcpBigtableInstanceDelete ******** end")
	return nil
}

// reconcileGcpBigtableClusters adds, updates, and removes clusters to match the
// desired configuration, using the per-cluster Bigtable endpoints.
func reconcileGcpBigtableClusters(ctx context.Context, d *schema.ResourceData, c *duplosdk.Client, tenantID, name string) diag.Diagnostics {
	oldRaw, newRaw := d.GetChange("cluster")
	oldClusters := indexBigtableClustersByID(oldRaw.([]interface{}))
	newClusters := indexBigtableClustersByID(newRaw.([]interface{}))

	// Collect clusters that are no longer present, but defer the deletions until
	// after creates/updates. Bigtable rejects deleting the last remaining cluster,
	// so creating replacements first lets a single apply relocate a cluster
	// (remove + add a different cluster_id) without ever dropping below one cluster.
	toDelete := []string{}
	for id := range oldClusters {
		if _, ok := newClusters[id]; !ok {
			toDelete = append(toDelete, id)
		}
	}

	// storage_type is instance-level and ForceNew, so it never changes here.
	storageType := bigtableStorageToInt(d.Get("storage_type").(string))

	// Add new clusters and update existing ones in place.
	created := []string{}
	updated := map[string]*duplosdk.DuploBigtableCluster{}
	for id, cfg := range newClusters {
		_, exists := oldClusters[id]
		rq := expandGcpBigtableCluster(cfg, storageType)

		if !exists {
			if _, clientErr := c.GcpBigtableClusterCreate(tenantID, name, id, rq); clientErr != nil {
				return diag.Errorf("Error creating cluster '%s' of Bigtable instance '%s': %s", id, name, clientErr)
			}
			created = append(created, id)
			continue
		}

		// Update serve nodes / autoscaling in place.
		upd := &duplosdk.DuploBigtableCluster{
			ServeNodes:    rq.ServeNodes,
			ClusterConfig: rq.ClusterConfig,
		}
		if clientErr := c.GcpBigtableClusterUpdate(tenantID, name, id, upd); clientErr != nil {
			return diag.Errorf("Error updating cluster '%s' of Bigtable instance '%s': %s", id, name, clientErr)
		}
		updated[id] = upd
	}

	// The backend starts the cluster update and returns without waiting for it, so the
	// post-apply read could record the old node count or autoscaling settings and plan
	// the same change again. Wait until the listing reflects what was requested.
	if len(updated) > 0 {
		if err := gcpBigtableWaitUntilClustersUpdated(ctx, c, tenantID, name, updated, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("Error waiting for clusters of Bigtable instance '%s' to apply their updates: %s", name, err)
		}
	}

	// Cluster creation is a long-running operation that the backend does not wait on, so
	// wait for the replacements to be ready before removing anything. Otherwise the old
	// cluster could be deleted before its replacement exists.
	if len(toDelete) > 0 && len(created) > 0 {
		if err := gcpBigtableWaitUntilClustersReady(ctx, c, tenantID, name, created, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("Error waiting for new clusters of Bigtable instance '%s' before removing old ones: %s", name, err)
		}
	}

	// Now that replacements exist, remove the clusters that are no longer present.
	for _, id := range toDelete {
		if clientErr := c.GcpBigtableClusterDelete(tenantID, name, id); clientErr != nil {
			return diag.Errorf("Error deleting cluster '%s' of Bigtable instance '%s': %s", id, name, clientErr)
		}
	}

	return nil
}

func indexBigtableClustersByID(list []interface{}) map[string]map[string]interface{} {
	out := map[string]map[string]interface{}{}
	for _, raw := range list {
		m := raw.(map[string]interface{})
		out[m["cluster_id"].(string)] = m
	}
	return out
}

func resourceGcpBigtableInstanceSetData(d *schema.ResourceData, tenantID, name string, instance *duplosdk.DuploBigtableInstance, clusters *[]duplosdk.DuploBigtableCluster) {
	d.Set("tenant_id", tenantID)
	d.Set("name", name)
	d.Set("display_name", instance.DisplayName)
	d.Set("instance_type", bigtableTypeToString(instance.Type))
	d.Set("state", bigtableStateToString(instance.State))
	flattenGcpLabels(d, instance.Labels)
	// All clusters share the instance's storage type, so derive it from the first one.
	if clusters != nil && len(*clusters) > 0 {
		d.Set("storage_type", bigtableStorageToString((*clusters)[0].DefaultStorageType))
	}
	d.Set("cluster", flattenGcpBigtableClusters(d, clusters))
}

func flattenGcpBigtableClusters(d *schema.ResourceData, clusters *[]duplosdk.DuploBigtableCluster) []interface{} {
	if clusters == nil {
		return nil
	}

	// Flatten each cluster, keyed by its ID.
	byID := map[string]interface{}{}
	for _, cl := range *clusters {
		id := lastPathSegment(cl.Name)
		m := map[string]interface{}{
			"cluster_id": id,
			"zone":       lastPathSegment(cl.Location),
			"num_nodes":  cl.ServeNodes,
			"state":      bigtableStateToString(cl.State),
		}
		if cl.ClusterConfig != nil && cl.ClusterConfig.ClusterAutoscalingConfig != nil {
			ac := cl.ClusterConfig.ClusterAutoscalingConfig
			m["autoscaling_config"] = []interface{}{
				map[string]interface{}{
					"min_nodes":      ac.AutoscalingLimits.MinServeNodes,
					"max_nodes":      ac.AutoscalingLimits.MaxServeNodes,
					"cpu_target":     ac.AutoscalingTargets.CpuUtilizationPercent,
					"storage_target": ac.AutoscalingTargets.StorageUtilizationGibPerNode,
				},
			}
		}
		byID[id] = m
	}

	// Emit clusters in the order they already appear in config/state so that an
	// API listing in a different order does not produce spurious reorder diffs.
	// Clusters not yet tracked in config (e.g. on import) are appended in a
	// deterministic, cluster_id-sorted order.
	out := make([]interface{}, 0, len(byID))
	for _, raw := range d.Get("cluster").([]interface{}) {
		cfg := raw.(map[string]interface{})
		id := cfg["cluster_id"].(string)
		if m, ok := byID[id]; ok {
			out = append(out, m)
			delete(byID, id)
		}
	}
	remaining := make([]string, 0, len(byID))
	for id := range byID {
		remaining = append(remaining, id)
	}
	sort.Strings(remaining)
	for _, id := range remaining {
		out = append(out, byID[id])
	}
	return out
}

func expandGcpBigtableCreateRequest(d *schema.ResourceData) *duplosdk.DuploBigtableCreateInstanceRequest {
	rq := &duplosdk.DuploBigtableCreateInstanceRequest{
		InstanceId: d.Get("name").(string),
		Instance: duplosdk.DuploBigtableInstance{
			DisplayName: bigtableDisplayName(d.Get("display_name").(string), d.Get("name").(string)),
			Type:        bigtableTypeToInt(d.Get("instance_type").(string)),
			Labels:      expandAsStringMap("labels", d),
		},
		Clusters: map[string]duplosdk.DuploBigtableCluster{},
	}
	storageType := bigtableStorageToInt(d.Get("storage_type").(string))
	for _, raw := range d.Get("cluster").([]interface{}) {
		cfg := raw.(map[string]interface{})
		rq.Clusters[cfg["cluster_id"].(string)] = *expandGcpBigtableCluster(cfg, storageType)
	}
	return rq
}

// bigtableDisplayName returns the configured display name, or one derived from the
// instance name. GCP requires a display name of 4-30 characters, and the backend's own
// fallback never applies because an unset protobuf string arrives as "" rather than null.
// An instance name is 6-33 characters, so it is cut to 30.
func bigtableDisplayName(displayName, name string) string {
	if displayName != "" {
		return displayName
	}
	if len(name) > 30 {
		return name[:30]
	}
	return name
}

func expandGcpBigtableCluster(cfg map[string]interface{}, storageType int) *duplosdk.DuploBigtableCluster {
	cl := &duplosdk.DuploBigtableCluster{
		Location:           cfg["zone"].(string),
		DefaultStorageType: storageType,
	}
	if ac, ok := cfg["autoscaling_config"].([]interface{}); ok && len(ac) > 0 && ac[0] != nil {
		a := ac[0].(map[string]interface{})
		// storage_target is Optional+Computed, so it may be absent from the map.
		storageTarget, _ := a["storage_target"].(int)
		cl.ClusterConfig = &duplosdk.DuploBigtableClusterConfig{
			ClusterAutoscalingConfig: &duplosdk.DuploBigtableClusterAutoscalingConfig{
				AutoscalingLimits: duplosdk.DuploBigtableAutoscalingLimits{
					MinServeNodes: a["min_nodes"].(int),
					MaxServeNodes: a["max_nodes"].(int),
				},
				AutoscalingTargets: duplosdk.DuploBigtableAutoscalingTargets{
					CpuUtilizationPercent:        a["cpu_target"].(int),
					StorageUtilizationGibPerNode: storageTarget,
				},
			},
		}
	} else {
		cl.ServeNodes = cfg["num_nodes"].(int)
	}
	return cl
}

// validateBigtableClusters validates the cluster blocks at plan time. The
// checks themselves live in pure helpers so that they can be unit tested.
func validateBigtableClusters(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	oldRaw, _ := diff.GetChange("cluster")
	newClusters := diff.Get("cluster").([]interface{})

	if err := validateBigtableClusterBlocks(newClusters, bigtableUnknownNumNodes(diff.GetRawConfig())); err != nil {
		return err
	}
	if err := validateBigtableScalingModes(diff.GetRawConfig()); err != nil {
		return err
	}
	if err := validateBigtableStorageTargets(newClusters, diff.Get("storage_type").(string)); err != nil {
		return err
	}
	return validateBigtableClusterTransitions(oldRaw.([]interface{}), newClusters)
}

// validateBigtableClusterBlocks validates the planned cluster blocks on their own:
//   - cluster_id values are unique (clusters are matched by id, so duplicates
//     would silently collapse and reconcile the wrong cluster);
//   - each cluster has a manual node count or an autoscaling configuration;
//   - autoscaling max_nodes is not below min_nodes and not above 10 times min_nodes,
//     which GCP enforces as a hard limit. Each bound is checked to be positive by its
//     own ValidateFunc, and an unknown bound reads as 0 here, so these are only
//     checked once both are known.
//
// unknownNumNodes holds the indexes of clusters whose num_nodes is not known yet. The
// flattened diff reads such a value as 0, so the node-count check waits for it.
func validateBigtableClusterBlocks(clusters []interface{}, unknownNumNodes map[int]bool) error {
	seen := map[string]bool{}
	for i, raw := range clusters {
		cfg := raw.(map[string]interface{})
		id := cfg["cluster_id"].(string)
		if seen[id] {
			return fmt.Errorf("duplicate cluster_id %q: each cluster must have a unique cluster_id", id)
		}
		seen[id] = true

		ac, _ := cfg["autoscaling_config"].([]interface{})
		hasAutoscaling := len(ac) > 0
		numNodes := cfg["num_nodes"].(int)
		if !hasAutoscaling && numNodes <= 0 && !unknownNumNodes[i] {
			return fmt.Errorf("cluster %q: either 'num_nodes' (> 0) or 'autoscaling_config' must be set", id)
		}
		if hasAutoscaling && ac[0] != nil {
			a := ac[0].(map[string]interface{})
			minNodes, _ := a["min_nodes"].(int)
			maxNodes, _ := a["max_nodes"].(int)
			if minNodes > 0 && maxNodes > 0 && maxNodes < minNodes {
				return fmt.Errorf("cluster %q: autoscaling 'max_nodes' (%d) must be greater than or equal to 'min_nodes' (%d)", id, maxNodes, minNodes)
			}
			if minNodes > 0 && maxNodes > 10*minNodes {
				return fmt.Errorf("cluster %q: autoscaling 'max_nodes' (%d) cannot be more than 10 times 'min_nodes' (%d)", id, maxNodes, minNodes)
			}
		}
	}
	return nil
}

// bigtableUnknownNumNodes returns the indexes of the configured clusters whose num_nodes
// is set to a value not known until apply, such as one taken from another resource.
// num_nodes is Computed, so only the raw configuration tells that apart from unset.
func bigtableUnknownNumNodes(config cty.Value) map[int]bool {
	out := map[int]bool{}
	if config.IsNull() || !config.IsKnown() || !config.Type().IsObjectType() || !config.Type().HasAttribute("cluster") {
		return out
	}
	clusters := config.GetAttr("cluster")
	if clusters.IsNull() || !clusters.IsKnown() || !clusters.CanIterateElements() {
		return out
	}
	i := 0
	for it := clusters.ElementIterator(); it.Next(); i++ {
		_, cl := it.Element()
		if !cl.IsNull() && cl.IsKnown() && !cl.GetAttr("num_nodes").IsKnown() {
			out[i] = true
		}
	}
	return out
}

// validateBigtableScalingModes rejects a cluster that configures both num_nodes and
// autoscaling_config. num_nodes is ignored while autoscaling is on and Read records the
// live node count in it, so setting both would plan a change on every run. num_nodes is
// Computed, so only the raw configuration says whether the user set it.
func validateBigtableScalingModes(config cty.Value) error {
	if config.IsNull() || !config.IsKnown() || !config.Type().IsObjectType() || !config.Type().HasAttribute("cluster") {
		return nil
	}
	clusters := config.GetAttr("cluster")
	if clusters.IsNull() || !clusters.IsKnown() || !clusters.CanIterateElements() {
		return nil
	}
	for it := clusters.ElementIterator(); it.Next(); {
		_, cl := it.Element()
		if cl.IsNull() || !cl.IsKnown() {
			continue
		}
		numNodes, autoscaling := cl.GetAttr("num_nodes"), cl.GetAttr("autoscaling_config")
		if numNodes.IsNull() || autoscaling.IsNull() || !autoscaling.IsKnown() || autoscaling.LengthInt() == 0 {
			continue
		}
		id := "?"
		if v := cl.GetAttr("cluster_id"); v.IsKnown() && !v.IsNull() {
			id = v.AsString()
		}
		return fmt.Errorf("cluster %q: 'num_nodes' and 'autoscaling_config' are mutually exclusive; remove 'num_nodes' when autoscaling is configured", id)
	}
	return nil
}

// validateBigtableStorageTargets checks each autoscaling storage_target against the
// range GCP accepts for the instance's storage type: 2560-5120 GiB per node for SSD and
// 8192-16384 for HDD. A target of 0 (or unset) asks for the GCP default.
func validateBigtableStorageTargets(clusters []interface{}, storageType string) error {
	lo, hi := 2560, 5120
	if storageType == "HDD" {
		lo, hi = 8192, 16384
	}
	for _, raw := range clusters {
		cfg := raw.(map[string]interface{})
		ac, _ := cfg["autoscaling_config"].([]interface{})
		if len(ac) == 0 || ac[0] == nil {
			continue
		}
		target, _ := ac[0].(map[string]interface{})["storage_target"].(int)
		if target != 0 && (target < lo || target > hi) {
			return fmt.Errorf("cluster %q: autoscaling 'storage_target' (%d) must be between %d and %d GiB per node for %s storage", cfg["cluster_id"], target, lo, hi, storageType)
		}
	}
	return nil
}

// validateBigtableClusterTransitions rejects the changes to an existing cluster
// that Bigtable will not accept in place:
//   - the zone (location) of a cluster cannot be changed, and an in-place update
//     would silently drop it;
//   - an autoscaling_config block cannot be removed. Clearing autoscaling
//     requires naming cluster_config.cluster_autoscaling_config in the update
//     mask, and the backend derives its mask from the fields it receives, so
//     omitting clusterConfig clears nothing, serve_nodes stays ignored while
//     autoscaling is on, and the plan would show the same removal forever.
//
// Existing clusters are matched by cluster_id; clusters being added or removed
// outright are not affected.
func validateBigtableClusterTransitions(oldClusters, newClusters []interface{}) error {
	previous := indexBigtableClustersByID(oldClusters)
	for _, raw := range newClusters {
		cfg := raw.(map[string]interface{})
		id := cfg["cluster_id"].(string)
		old, ok := previous[id]
		if !ok {
			continue
		}

		if oldZone := old["zone"].(string); oldZone != cfg["zone"].(string) {
			return fmt.Errorf("cluster %q: zone is immutable (cannot change from %q to %q); remove the cluster and add a new one to relocate it", id, oldZone, cfg["zone"])
		}

		oldAutoscaling, _ := old["autoscaling_config"].([]interface{})
		newAutoscaling, _ := cfg["autoscaling_config"].([]interface{})
		if len(oldAutoscaling) > 0 && len(newAutoscaling) == 0 {
			return fmt.Errorf("cluster %q: removing 'autoscaling_config' is not supported (autoscaling cannot be turned off in place); keep the 'autoscaling_config' block, or remove the cluster and add a new one to run it on a fixed 'num_nodes' count", id)
		}
	}
	return nil
}

func gcpBigtableInstanceWaitUntilReady(ctx context.Context, c *duplosdk.Client, tenantID, name string, timeout time.Duration) error {
	retryFlag := 3
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			rp, err := c.GcpBigtableInstanceGet(tenantID, name)
			status := "pending"
			if err == nil && rp != nil {
				if rp.State == duplosdk.BigtableStateReady {
					status = "ready"
				}
			} else if err != nil && retryFlag > 0 {
				retryFlag--
				err = nil
			}
			return rp, status, err
		},
		PollInterval: 20 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] gcpBigtableInstanceWaitUntilReady(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// gcpBigtableWaitUntilClustersReady waits until every cluster in clusterIDs is
// present in the instance's cluster listing and reports a READY state. Cluster
// create/update operations are asynchronous, so without this the post-apply read
// can capture a not-yet-ready cluster (leaving its computed `state` unset).
func gcpBigtableWaitUntilClustersReady(ctx context.Context, c *duplosdk.Client, tenantID, name string, clusterIDs []string, timeout time.Duration) error {
	return gcpBigtableWaitForClusters(ctx, c, tenantID, name, clusterIDs, true, timeout)
}

// gcpBigtableWaitUntilClustersPresent waits until every cluster in clusterIDs is
// present in the instance's cluster listing, in any state. It is used when
// wait_until_ready is false: the post-apply read would otherwise miss a cluster
// whose create is still in flight and leave the required cluster block partial.
func gcpBigtableWaitUntilClustersPresent(ctx context.Context, c *duplosdk.Client, tenantID, name string, clusterIDs []string, timeout time.Duration) error {
	return gcpBigtableWaitForClusters(ctx, c, tenantID, name, clusterIDs, false, timeout)
}

func gcpBigtableWaitForClusters(ctx context.Context, c *duplosdk.Client, tenantID, name string, clusterIDs []string, requireReady bool, timeout time.Duration) error {
	if len(clusterIDs) == 0 {
		return nil
	}
	retryFlag := 3
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"ready"},
		Refresh: func() (interface{}, string, error) {
			clusters, err := c.GcpBigtableClusterList(tenantID, name)
			if err != nil {
				if retryFlag > 0 {
					retryFlag--
					return clusters, "pending", nil
				}
				return clusters, "pending", err
			}
			ready := map[string]bool{}
			for _, cl := range *clusters {
				if !requireReady || cl.State == duplosdk.BigtableStateReady {
					ready[lastPathSegment(cl.Name)] = true
				}
			}
			for _, id := range clusterIDs {
				if !ready[id] {
					return clusters, "pending", nil
				}
			}
			return clusters, "ready", nil
		},
		PollInterval: 20 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] gcpBigtableWaitForClusters(%s, %s, %v, requireReady=%t)", tenantID, name, clusterIDs, requireReady)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// gcpBigtableWaitUntilClustersUpdated waits until every cluster in want reports the
// node count or autoscaling settings that were requested for it.
func gcpBigtableWaitUntilClustersUpdated(ctx context.Context, c *duplosdk.Client, tenantID, name string, want map[string]*duplosdk.DuploBigtableCluster, timeout time.Duration) error {
	retryFlag := 3
	stateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"updated"},
		Refresh: func() (interface{}, string, error) {
			clusters, err := c.GcpBigtableClusterList(tenantID, name)
			if err != nil {
				if retryFlag > 0 {
					retryFlag--
					return clusters, "pending", nil
				}
				return clusters, "pending", err
			}
			got := map[string]*duplosdk.DuploBigtableCluster{}
			for i := range *clusters {
				got[lastPathSegment((*clusters)[i].Name)] = &(*clusters)[i]
			}
			for id, w := range want {
				if g, ok := got[id]; !ok || !bigtableClusterUpdateApplied(g, w) {
					return clusters, "pending", nil
				}
			}
			return clusters, "updated", nil
		},
		PollInterval: 20 * time.Second,
		Timeout:      timeout,
	}
	log.Printf("[DEBUG] gcpBigtableWaitUntilClustersUpdated(%s, %s)", tenantID, name)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// bigtableClusterUpdateApplied reports whether a listed cluster reflects a requested
// update. An autoscaled cluster is compared on its autoscaling settings, since its node
// count moves on its own; a storage target of 0 asks for the GCP default, so any value
// satisfies it. A manually scaled cluster is compared on its node count.
func bigtableClusterUpdateApplied(got, want *duplosdk.DuploBigtableCluster) bool {
	if want.ClusterConfig != nil && want.ClusterConfig.ClusterAutoscalingConfig != nil {
		if got.ClusterConfig == nil || got.ClusterConfig.ClusterAutoscalingConfig == nil {
			return false
		}
		g, w := got.ClusterConfig.ClusterAutoscalingConfig, want.ClusterConfig.ClusterAutoscalingConfig
		if g.AutoscalingLimits != w.AutoscalingLimits ||
			g.AutoscalingTargets.CpuUtilizationPercent != w.AutoscalingTargets.CpuUtilizationPercent {
			return false
		}
		return w.AutoscalingTargets.StorageUtilizationGibPerNode == 0 ||
			g.AutoscalingTargets.StorageUtilizationGibPerNode == w.AutoscalingTargets.StorageUtilizationGibPerNode
	}
	return got.ServeNodes == want.ServeNodes
}

// configuredBigtableClusterIDs returns the cluster IDs currently declared in config.
func configuredBigtableClusterIDs(d *schema.ResourceData) []string {
	raw := d.Get("cluster").([]interface{})
	ids := make([]string, 0, len(raw))
	for _, r := range raw {
		ids = append(ids, r.(map[string]interface{})["cluster_id"].(string))
	}
	return ids
}

func parseGcpBigtableInstanceIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) == 2 {
		tenantID, name = idParts[0], idParts[1]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func lastPathSegment(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "/")
	return parts[len(parts)-1]
}

func bigtableTypeToInt(s string) int {
	switch s {
	case "DEVELOPMENT":
		return duplosdk.BigtableTypeDevelopment
	default:
		return duplosdk.BigtableTypeProduction
	}
}

func bigtableTypeToString(i int) string {
	switch i {
	case duplosdk.BigtableTypeDevelopment:
		return "DEVELOPMENT"
	case duplosdk.BigtableTypeProduction:
		return "PRODUCTION"
	default:
		// The schema only allows PRODUCTION/DEVELOPMENT and defaults to
		// PRODUCTION, so treat any unknown/unspecified backend value the same
		// way to avoid an empty instance_type and a perpetual diff.
		return "PRODUCTION"
	}
}

func bigtableStorageToInt(s string) int {
	switch s {
	case "HDD":
		return duplosdk.BigtableStorageHDD
	default:
		return duplosdk.BigtableStorageSSD
	}
}

func bigtableStorageToString(i int) string {
	switch i {
	case duplosdk.BigtableStorageHDD:
		return "HDD"
	default:
		return "SSD"
	}
}

func bigtableStateToString(i int) string {
	switch i {
	case 1:
		return "READY"
	case 2:
		return "CREATING"
	case 3:
		return "RESIZING"
	case 4:
		return "DISABLED"
	default:
		return "STATE_NOT_KNOWN"
	}
}
