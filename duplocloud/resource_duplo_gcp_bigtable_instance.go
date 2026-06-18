package duplocloud

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

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
			Description: "The number of nodes for manual scaling. Leave unset (or use `autoscaling_config`) " +
				"to enable autoscaling. When `autoscaling_config` is set, this reflects the current node count.",
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"autoscaling_config": {
			Description: "Autoscaling configuration for the cluster. When set, the cluster scales automatically " +
				"and `num_nodes` is ignored.",
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"min_nodes": {
						Description: "Minimum number of nodes for autoscaling.",
						Type:        schema.TypeInt,
						Required:    true,
					},
					"max_nodes": {
						Description: "Maximum number of nodes for autoscaling.",
						Type:        schema.TypeInt,
						Required:    true,
					},
					"cpu_target": {
						Description:  "The target CPU utilization percentage that drives autoscaling (10-80).",
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(10, 80),
					},
					"storage_target": {
						Description: "The target storage utilization in GiB per node that drives autoscaling. " +
							"Defaults to the GCP recommended value when unset.",
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
			Description: "The ID of the Bigtable instance. Used verbatim as the instance ID in GCP.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
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
			Description: "The human-readable display name of the Bigtable instance.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
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
		rq := &duplosdk.DuploBigtableInstance{
			DisplayName: d.Get("display_name").(string),
			Type:        bigtableTypeToInt(d.Get("instance_type").(string)),
			Labels:      expandAsStringMap("labels", d),
		}
		if _, clientErr := c.GcpBigtableInstanceUpdate(tenantID, name, rq); clientErr != nil {
			return diag.Errorf("Error updating tenant %s Bigtable instance '%s': %s", tenantID, name, clientErr)
		}
	}

	// Reconcile clusters.
	if d.HasChange("cluster") {
		if diags := reconcileGcpBigtableClusters(d, c, tenantID, name); diags != nil {
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
func reconcileGcpBigtableClusters(d *schema.ResourceData, c *duplosdk.Client, tenantID, name string) diag.Diagnostics {
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
	for id, cfg := range newClusters {
		_, exists := oldClusters[id]
		rq := expandGcpBigtableCluster(cfg, storageType)

		if !exists {
			if _, clientErr := c.GcpBigtableClusterCreate(tenantID, name, id, rq); clientErr != nil {
				return diag.Errorf("Error creating cluster '%s' of Bigtable instance '%s': %s", id, name, clientErr)
			}
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
			DisplayName: d.Get("display_name").(string),
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

// validateBigtableClusters validates the cluster blocks at plan time:
//   - each cluster has a manual node count or an autoscaling configuration;
//   - cluster_id values are unique (clusters are matched by id, so duplicates
//     would silently collapse and reconcile the wrong cluster);
//   - the zone of an existing cluster is not changed (cluster location is
//     immutable in Bigtable, and an in-place update would silently drop it).
func validateBigtableClusters(ctx context.Context, diff *schema.ResourceDiff, m interface{}) error {
	seen := map[string]bool{}
	for _, raw := range diff.Get("cluster").([]interface{}) {
		cfg := raw.(map[string]interface{})
		id := cfg["cluster_id"].(string)
		if seen[id] {
			return fmt.Errorf("duplicate cluster_id %q: each cluster must have a unique cluster_id", id)
		}
		seen[id] = true

		ac, _ := cfg["autoscaling_config"].([]interface{})
		hasAutoscaling := len(ac) > 0
		numNodes := cfg["num_nodes"].(int)
		if !hasAutoscaling && numNodes <= 0 {
			return fmt.Errorf("cluster %q: either 'num_nodes' (> 0) or 'autoscaling_config' must be set", id)
		}
	}

	// A cluster's zone (location) cannot be changed in place. Compare existing
	// clusters by cluster_id and reject a zone change with a clear message.
	oldRaw, newRaw := diff.GetChange("cluster")
	oldZones := map[string]string{}
	for _, raw := range oldRaw.([]interface{}) {
		cfg := raw.(map[string]interface{})
		oldZones[cfg["cluster_id"].(string)] = cfg["zone"].(string)
	}
	for _, raw := range newRaw.([]interface{}) {
		cfg := raw.(map[string]interface{})
		id := cfg["cluster_id"].(string)
		if oldZone, ok := oldZones[id]; ok && oldZone != cfg["zone"].(string) {
			return fmt.Errorf("cluster %q: zone is immutable (cannot change from %q to %q); remove the cluster and add a new one to relocate it", id, oldZone, cfg["zone"])
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
				if cl.State == duplosdk.BigtableStateReady {
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
	log.Printf("[DEBUG] gcpBigtableWaitUntilClustersReady(%s, %s, %v)", tenantID, name, clusterIDs)
	_, err := stateConf.WaitForStateContext(ctx)
	return err
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
