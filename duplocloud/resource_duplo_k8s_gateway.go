package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func k8sGatewaySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the Gateway will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the Gateway.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"gateway_class_name": {
			Description: "The name of the GatewayClass this Gateway uses. If not set, Duplo auto-selects a GatewayClass based on `is_public`.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
		},
		"is_public": {
			Description: "Whether Duplo should select a public (external) GatewayClass. When false, an internal GatewayClass is selected. Ignored when `gateway_class_name` is explicitly set.",
			Type:        schema.TypeBool,
			Optional:    true,
			ForceNew:    true,
			Default:     false,
		},
		"listener": {
			Description: "The listeners for the Gateway.",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: "The name of the listener.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"port": {
						Description:  "The network port of the listener.",
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntBetween(1, 65535),
					},
					"protocol": {
						Description: "The protocol of the listener. Must be one of: `HTTP`, `HTTPS`, `TLS`.",
						Type:        schema.TypeString,
						Required:    true,
						ValidateFunc: validation.StringInSlice([]string{
							"HTTP",
							"HTTPS",
							"TLS",
						}, false),
					},
					"hostname": {
						Description: "Optional hostname filter for the listener. Supports wildcards (e.g. `*.example.com`).",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"tls": {
						Description: "The TLS configuration for the listener. Applicable when `protocol` is `HTTPS` or `TLS`.",
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"mode": {
									Description: "The TLS mode. Must be one of: `Terminate`, `Passthrough`.",
									Type:        schema.TypeString,
									Optional:    true,
									Default:     "Terminate",
									ValidateFunc: validation.StringInSlice([]string{
										"Terminate",
										"Passthrough",
									}, false),
								},
								"certificate_ref": {
									Description: "Kubernetes Secret references holding TLS certificates. Applicable when `mode` is `Terminate`.",
									Type:        schema.TypeList,
									Optional:    true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"name": {
												Description: "The name of the referenced certificate resource.",
												Type:        schema.TypeString,
												Required:    true,
											},
											"namespace": {
												Description: "Optional namespace of the referenced resource. Cross-namespace references require a ReferenceGrant in the target namespace.",
												Type:        schema.TypeString,
												Optional:    true,
											},
											"kind": {
												Description: "The kind of the referenced resource.",
												Type:        schema.TypeString,
												Optional:    true,
												Default:     "Secret",
											},
											"group": {
												Description: "The API group of the referenced resource. Empty string for core Kubernetes resources (Secrets).",
												Type:        schema.TypeString,
												Optional:    true,
											},
										},
									},
								},
								"certificate_map": {
									Description: "Google Certificate Manager certificate map name. Sets the `networking.gke.io/certmap` annotation.",
									Type:        schema.TypeString,
									Optional:    true,
								},
							},
						},
					},
					"allowed_routes": {
						Description: "Which routes may attach to this listener.",
						Type:        schema.TypeList,
						Optional:    true,
						Computed:    true,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"from": {
									Description: "The namespaces from which routes may attach. Must be one of: `All`, `Same`, `Selector`.",
									Type:        schema.TypeString,
									Optional:    true,
									Computed:    true,
									ValidateFunc: validation.StringInSlice([]string{
										"All",
										"Same",
										"Selector",
									}, false),
								},
								"namespace_selector": {
									Description: "Label selector for namespaces. Required when `from` is `Selector`.",
									Type:        schema.TypeMap,
									Optional:    true,
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				},
			},
		},
		"address": {
			Description: "Addresses requested for the Gateway, e.g. a static IP.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Description: "The type of the address. Must be one of: `IPAddress`, `Hostname`.",
						Type:        schema.TypeString,
						Optional:    true,
						Default:     "IPAddress",
						ValidateFunc: validation.StringInSlice([]string{
							"IPAddress",
							"Hostname",
						}, false),
					},
					"value": {
						Description: "The address value.",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},
		"annotations": {
			Description: "An unstructured key value map stored with the Gateway that may be used to store arbitrary metadata. Annotations added by the backend or the gateway controller (e.g. `networking.gke.io/*`) are ignored and left in place.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"labels": {
			Description: "Map of string keys and values that can be used to organize and categorize (scope and select) the Gateway.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

// SCHEMA for resource crud
func resourceK8sGateway() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_k8s_gateway` manages a Kubernetes Gateway (gateway.networking.k8s.io/v1) in a Duplo tenant.",

		ReadContext:   resourceK8sGatewayRead,
		CreateContext: resourceK8sGatewayCreate,
		UpdateContext: resourceK8sGatewayUpdate,
		DeleteContext: resourceK8sGatewayDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: k8sGatewaySchema(),
	}
}

func resourceK8sGatewayRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sGatewayIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sGatewayRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.DuploK8sGatewayGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[TRACE] resourceK8sGatewayRead(%s, %s): object not found", tenantID, name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s gateway %s : %s", tenantID, name, clientErr)
	}
	if rp == nil || rp.Name == "" {
		d.SetId("")
		return nil
	}

	flattenK8sGateway(tenantID, d, rp)
	log.Printf("[TRACE] resourceK8sGatewayRead(%s, %s): end", tenantID, name)
	return nil
}

// CREATE resource
func resourceK8sGatewayCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] resourceK8sGatewayCreate(%s, %s): start", tenantID, name)

	rq := expandK8sGateway(d)

	c := m.(*duplosdk.Client)
	cerr := c.DuploK8sGatewayCreate(tenantID, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("v3/subscriptions/%s/k8s/gateway/%s", tenantID, name))

	diags := resourceK8sGatewayRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8sGatewayCreate(%s, %s): end", tenantID, name)
	return diags
}

// UPDATE resource
func resourceK8sGatewayUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sGatewayIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sGatewayUpdate(%s, %s): start", tenantID, name)

	rq := expandK8sGateway(d)

	c := m.(*duplosdk.Client)
	cerr := c.DuploK8sGatewayUpdate(tenantID, name, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}

	diags := resourceK8sGatewayRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8sGatewayUpdate(%s, %s): end", tenantID, name)
	return diags
}

// DELETE resource
func resourceK8sGatewayDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sGatewayIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sGatewayDelete(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.DuploK8sGatewayGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[TRACE] resourceK8sGatewayDelete(%s, %s): object not found", tenantID, name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s gateway %s : %s", tenantID, name, clientErr)
	}
	if rp != nil && rp.Name != "" {
		clientErr := c.DuploK8sGatewayDelete(tenantID, name)
		if clientErr != nil {
			if clientErr.Status() == 404 {
				d.SetId("")
				return nil
			}
			return diag.Errorf("Unable to delete tenant %s k8s gateway %s : %s", tenantID, name, clientErr)
		}
	}

	log.Printf("[TRACE] resourceK8sGatewayDelete(%s, %s): end", tenantID, name)
	return nil
}

func parseK8sGatewayIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 6)
	if len(idParts) == 6 {
		tenantID, name = idParts[2], idParts[5]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func expandK8sGateway(d *schema.ResourceData) *duplosdk.DuploK8sGateway {
	duplo := duplosdk.DuploK8sGateway{
		Name:             d.Get("name").(string),
		GatewayClassName: d.Get("gateway_class_name").(string),
		IsPublic:         d.Get("is_public").(bool),
		Listeners:        expandK8sGatewayListeners(d.Get("listener").([]interface{})),
	}

	if v, ok := d.GetOk("address"); ok && len(v.([]interface{})) > 0 {
		duplo.Addresses = expandK8sGatewayAddresses(v.([]interface{}))
	}

	if v, ok := d.GetOk("annotations"); ok && !isInterfaceNil(v) {
		duplo.Annotations = expandAsStringMap("annotations", d)
	}

	if v, ok := d.GetOk("labels"); ok && !isInterfaceNil(v) {
		duplo.Labels = expandAsStringMap("labels", d)
	}
	return &duplo
}

func expandK8sGatewayListeners(lst []interface{}) *[]duplosdk.DuploK8sGatewayListener {
	listeners := make([]duplosdk.DuploK8sGatewayListener, 0, len(lst))
	for _, v := range lst {
		listeners = append(listeners, expandK8sGatewayListener(v.(map[string]interface{})))
	}
	return &listeners
}

func expandK8sGatewayListener(m map[string]interface{}) duplosdk.DuploK8sGatewayListener {
	listener := duplosdk.DuploK8sGatewayListener{
		Name:     m["name"].(string),
		Port:     m["port"].(int),
		Protocol: m["protocol"].(string),
	}
	if v, ok := m["hostname"]; ok {
		listener.Hostname = v.(string)
	}
	if v, ok := m["tls"]; ok && len(v.([]interface{})) == 1 {
		listener.TLS = expandK8sGatewayListenerTLS(v.([]interface{})[0].(map[string]interface{}))
	}
	if v, ok := m["allowed_routes"]; ok && len(v.([]interface{})) == 1 {
		listener.AllowedRoutes = expandK8sGatewayAllowedRoutes(v.([]interface{})[0].(map[string]interface{}))
	}
	return listener
}

func expandK8sGatewayListenerTLS(m map[string]interface{}) *duplosdk.DuploK8sGatewayListenerTLS {
	tls := duplosdk.DuploK8sGatewayListenerTLS{
		Mode:              m["mode"].(string),
		CertMapAnnotation: m["certificate_map"].(string),
	}
	if v, ok := m["certificate_ref"]; ok && len(v.([]interface{})) > 0 {
		refs := make([]duplosdk.DuploK8sCertificateRef, 0, len(v.([]interface{})))
		for _, r := range v.([]interface{}) {
			rm := r.(map[string]interface{})
			refs = append(refs, duplosdk.DuploK8sCertificateRef{
				Name:      rm["name"].(string),
				Namespace: rm["namespace"].(string),
				Kind:      rm["kind"].(string),
				Group:     rm["group"].(string),
			})
		}
		tls.CertificateRefs = &refs
	}
	return &tls
}

func expandK8sGatewayAllowedRoutes(m map[string]interface{}) *duplosdk.DuploK8sGatewayAllowedRoutes {
	allowedRoutes := duplosdk.DuploK8sGatewayAllowedRoutes{
		From: m["from"].(string),
	}
	if v, ok := m["namespace_selector"]; ok && len(v.(map[string]interface{})) > 0 {
		allowedRoutes.NamespaceSelector = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			allowedRoutes.NamespaceSelector[key] = value.(string)
		}
	}
	return &allowedRoutes
}

func expandK8sGatewayAddresses(lst []interface{}) *[]duplosdk.DuploK8sGatewayAddress {
	addresses := make([]duplosdk.DuploK8sGatewayAddress, 0, len(lst))
	for _, v := range lst {
		m := v.(map[string]interface{})
		addresses = append(addresses, duplosdk.DuploK8sGatewayAddress{
			Type:  m["type"].(string),
			Value: m["value"].(string),
		})
	}
	return &addresses
}

func flattenK8sGateway(tenantId string, d *schema.ResourceData, duplo *duplosdk.DuploK8sGateway) {
	d.Set("tenant_id", tenantId)
	d.Set("name", duplo.Name)
	d.Set("gateway_class_name", duplo.GatewayClassName)

	if duplo.Listeners != nil {
		d.Set("listener", flattenK8sGatewayListeners(duplo.Listeners))
	}

	// The backend may report assigned (status) addresses; only track addresses
	// the user requested to avoid perpetual drift.
	if _, ok := d.GetOk("address"); ok && duplo.Addresses != nil {
		d.Set("address", flattenK8sGatewayAddresses(duplo.Addresses))
	}

	d.Set("annotations", filterServerManagedKeys(d, "annotations", duplo.Annotations))
	d.Set("labels", filterServerManagedKeys(d, "labels", duplo.Labels))
}

// Annotation/label keys written by the Duplo backend or the GKE gateway
// controller. These are dropped during flatten (unless explicitly configured)
// so they don't cause drift — including right after an import, when no prior
// state exists to filter against.
var k8sGatewayServerManagedKeys = []string{
	"networking.gke.io/",
	"duplocloud.net/is-any-host-allowed",
	"duplocloud.net/owner",
	"duplocloud.net/tenantid",
	"duplocloud.net/tenantname",
}

func isK8sGatewayServerManagedKey(key string) bool {
	for _, managed := range k8sGatewayServerManagedKeys {
		if key == managed || (strings.HasSuffix(managed, "/") && strings.HasPrefix(key, managed)) {
			return true
		}
	}
	return false
}

// filterServerManagedKeys keeps only the map keys the user manages: keys
// present in the config (prior state), plus — when there is no config, e.g.
// right after an import — any server key that isn't a known server-managed one.
func filterServerManagedKeys(d *schema.ResourceData, field string, server map[string]string) map[string]string {
	cfg := d.Get(field).(map[string]interface{})
	managed := map[string]string{}
	for k, v := range server {
		if _, ok := cfg[k]; ok {
			managed[k] = v
		} else if len(cfg) == 0 && !isK8sGatewayServerManagedKey(k) {
			managed[k] = v
		}
	}
	return managed
}

func flattenK8sGatewayListeners(duplo *[]duplosdk.DuploK8sGatewayListener) []interface{} {
	lst := []interface{}{}
	for _, v := range *duplo {
		lst = append(lst, flattenK8sGatewayListener(v))
	}
	return lst
}

func flattenK8sGatewayListener(duplo duplosdk.DuploK8sGatewayListener) map[string]interface{} {
	m := map[string]interface{}{
		"name":     duplo.Name,
		"port":     duplo.Port,
		"protocol": duplo.Protocol,
	}
	if len(duplo.Hostname) > 0 {
		m["hostname"] = duplo.Hostname
	}
	if duplo.TLS != nil {
		m["tls"] = []interface{}{flattenK8sGatewayListenerTLS(duplo.TLS)}
	}
	if duplo.AllowedRoutes != nil {
		m["allowed_routes"] = []interface{}{flattenK8sGatewayAllowedRoutes(duplo.AllowedRoutes)}
	}
	return m
}

func flattenK8sGatewayListenerTLS(duplo *duplosdk.DuploK8sGatewayListenerTLS) map[string]interface{} {
	// Default omitted fields to their schema defaults so a backend that leaves
	// them out doesn't cause perpetual drift.
	mode := duplo.Mode
	if mode == "" {
		mode = "Terminate"
	}
	m := map[string]interface{}{
		"mode":            mode,
		"certificate_map": duplo.CertMapAnnotation,
	}
	if duplo.CertificateRefs != nil && len(*duplo.CertificateRefs) > 0 {
		refs := make([]interface{}, 0, len(*duplo.CertificateRefs))
		for _, r := range *duplo.CertificateRefs {
			kind := r.Kind
			if kind == "" {
				kind = "Secret"
			}
			refs = append(refs, map[string]interface{}{
				"name":      r.Name,
				"namespace": r.Namespace,
				"kind":      kind,
				"group":     r.Group,
			})
		}
		m["certificate_ref"] = refs
	}
	return m
}

func flattenK8sGatewayAllowedRoutes(duplo *duplosdk.DuploK8sGatewayAllowedRoutes) map[string]interface{} {
	m := map[string]interface{}{
		"from": duplo.From,
	}
	if len(duplo.NamespaceSelector) > 0 {
		m["namespace_selector"] = duplo.NamespaceSelector
	}
	return m
}

func flattenK8sGatewayAddresses(duplo *[]duplosdk.DuploK8sGatewayAddress) []interface{} {
	lst := []interface{}{}
	for _, v := range *duplo {
		addrType := v.Type
		if addrType == "" {
			addrType = "IPAddress"
		}
		lst = append(lst, map[string]interface{}{
			"type":  addrType,
			"value": v.Value,
		})
	}
	return lst
}
