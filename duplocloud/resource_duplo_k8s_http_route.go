package duplocloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/duplocloud/terraform-provider-duplocloud/duplosdk"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func k8sHttpRouteSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the HTTPRoute will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the HTTPRoute.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"parent_ref": {
			Description: "The Gateways this HTTPRoute attaches to.",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Description: "The name of the Gateway.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"namespace": {
						Description: "The namespace of the Gateway. Cross-namespace routing requires administrator privileges.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"section_name": {
						Description: "The name of the listener on the Gateway to attach to.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"port": {
						Description:  "The port of the listener on the Gateway to attach to.",
						Type:         schema.TypeInt,
						Optional:     true,
						ValidateFunc: validation.IntBetween(1, 65535),
					},
				},
			},
		},
		"hostnames": {
			Description: "The hostnames used to match against the HTTP Host header.",
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"rule": {
			Description: "The routing rules for the HTTPRoute.",
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"match": {
						Description: "Conditions for matching incoming requests against this rule.",
						Type:        schema.TypeList,
						Optional:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"path": {
									Description: "Path matching condition.",
									Type:        schema.TypeList,
									Optional:    true,
									MaxItems:    1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"type": {
												Description: "The type of the path match. Must be one of: `Exact`, `PathPrefix`, `RegularExpression`.",
												Type:        schema.TypeString,
												Optional:    true,
												Default:     "PathPrefix",
												ValidateFunc: validation.StringInSlice([]string{
													"Exact",
													"PathPrefix",
													"RegularExpression",
												}, false),
											},
											"value": {
												Description: "The value of the path to match against.",
												Type:        schema.TypeString,
												Required:    true,
											},
										},
									},
								},
								"header": {
									Description: "HTTP request header matching conditions.",
									Type:        schema.TypeList,
									Optional:    true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"name": {
												Description: "The name of the HTTP header to match against.",
												Type:        schema.TypeString,
												Required:    true,
											},
											"value": {
												Description: "The value of the HTTP header to match against.",
												Type:        schema.TypeString,
												Required:    true,
											},
											"type": {
												Description: "The type of the header match.",
												Type:        schema.TypeString,
												Optional:    true,
											},
										},
									},
								},
								"method": {
									Description: "The HTTP method to match against (e.g. `GET`, `POST`).",
									Type:        schema.TypeString,
									Optional:    true,
								},
							},
						},
					},
					"backend_ref": {
						Description: "The backend services traffic should be sent to.",
						Type:        schema.TypeList,
						Required:    true,
						MinItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Description: "The name of the backend Kubernetes service.",
									Type:        schema.TypeString,
									Required:    true,
								},
								"port": {
									Description:  "The port of the backend Kubernetes service.",
									Type:         schema.TypeInt,
									Required:     true,
									ValidateFunc: validation.IntBetween(1, 65535),
								},
								"weight": {
									Description: "The proportion of traffic forwarded to this backend, relative to other backends in this rule.",
									Type:        schema.TypeInt,
									Optional:    true,
									Default:     1,
								},
								"namespace": {
									Description: "The namespace of the backend service. When different from the HTTPRoute's namespace, Duplo auto-creates a ReferenceGrant in the target namespace.",
									Type:        schema.TypeString,
									Optional:    true,
								},
							},
						},
					},
					"filters": {
						Description: "A JSON encoded list of HTTPRouteFilter objects (redirect, rewrite, header modification, etc). " +
							"You can use the `jsonencode()` function to build this from JSON.",
						Type:             schema.TypeString,
						Optional:         true,
						ValidateFunc:     ValidateJSONArrayString,
						DiffSuppressFunc: suppressEquivalentJSONDiffs,
					},
				},
			},
		},
		"annotations": {
			Description: "An unstructured key value map stored with the HTTPRoute that may be used to store arbitrary metadata. Annotations added by the backend or the gateway controller are ignored and left in place.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"labels": {
			Description: "Map of string keys and values that can be used to organize and categorize (scope and select) the HTTPRoute.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

// SCHEMA for resource crud
func resourceK8sHttpRoute() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_k8s_http_route` manages a Kubernetes HTTPRoute (gateway.networking.k8s.io/v1) in a Duplo tenant.",

		ReadContext:   resourceK8sHttpRouteRead,
		CreateContext: resourceK8sHttpRouteCreate,
		UpdateContext: resourceK8sHttpRouteUpdate,
		DeleteContext: resourceK8sHttpRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: k8sHttpRouteSchema(),
	}
}

func resourceK8sHttpRouteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sHttpRouteIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sHttpRouteRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.DuploK8sHTTPRouteGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[TRACE] resourceK8sHttpRouteRead(%s, %s): object not found", tenantID, name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s http route %s : %s", tenantID, name, clientErr)
	}
	if rp == nil || rp.Name == "" {
		d.SetId("")
		return nil
	}

	flattenK8sHttpRoute(tenantID, d, rp)
	log.Printf("[TRACE] resourceK8sHttpRouteRead(%s, %s): end", tenantID, name)
	return nil
}

// CREATE resource
func resourceK8sHttpRouteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] resourceK8sHttpRouteCreate(%s, %s): start", tenantID, name)

	rq, err := expandK8sHttpRoute(d)
	if err != nil {
		return diag.FromErr(err)
	}

	c := m.(*duplosdk.Client)
	cerr := c.DuploK8sHTTPRouteCreate(tenantID, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("v3/subscriptions/%s/k8s/httproute/%s", tenantID, name))

	diags := resourceK8sHttpRouteRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8sHttpRouteCreate(%s, %s): end", tenantID, name)
	return diags
}

// UPDATE resource
func resourceK8sHttpRouteUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sHttpRouteIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sHttpRouteUpdate(%s, %s): start", tenantID, name)

	rq, rqErr := expandK8sHttpRoute(d)
	if rqErr != nil {
		return diag.FromErr(rqErr)
	}

	c := m.(*duplosdk.Client)
	cerr := c.DuploK8sHTTPRouteUpdate(tenantID, name, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}

	diags := resourceK8sHttpRouteRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8sHttpRouteUpdate(%s, %s): end", tenantID, name)
	return diags
}

// DELETE resource
func resourceK8sHttpRouteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sHttpRouteIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8sHttpRouteDelete(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.DuploK8sHTTPRouteGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			log.Printf("[TRACE] resourceK8sHttpRouteDelete(%s, %s): object not found", tenantID, name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s http route %s : %s", tenantID, name, clientErr)
	}
	if rp != nil && rp.Name != "" {
		clientErr := c.DuploK8sHTTPRouteDelete(tenantID, name)
		if clientErr != nil {
			if clientErr.Status() == 404 {
				d.SetId("")
				return nil
			}
			return diag.Errorf("Unable to delete tenant %s k8s http route %s : %s", tenantID, name, clientErr)
		}
	}

	log.Printf("[TRACE] resourceK8sHttpRouteDelete(%s, %s): end", tenantID, name)
	return nil
}

func parseK8sHttpRouteIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 6)
	if len(idParts) == 6 {
		tenantID, name = idParts[2], idParts[5]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func expandK8sHttpRoute(d *schema.ResourceData) (*duplosdk.DuploK8sHTTPRoute, error) {
	rules, err := expandK8sHttpRouteRules(d.Get("rule").([]interface{}))
	if err != nil {
		return nil, err
	}

	duplo := duplosdk.DuploK8sHTTPRoute{
		Name:       d.Get("name").(string),
		ParentRefs: expandK8sHttpRouteParentRefs(d.Get("parent_ref").([]interface{})),
		Rules:      rules,
	}

	if v, ok := d.GetOk("hostnames"); ok && len(v.([]interface{})) > 0 {
		hostnames := make([]string, 0, len(v.([]interface{})))
		for _, h := range v.([]interface{}) {
			hostnames = append(hostnames, h.(string))
		}
		duplo.Hostnames = hostnames
	}

	if v, ok := d.GetOk("annotations"); ok && !isInterfaceNil(v) {
		duplo.Annotations = expandAsStringMap("annotations", d)
	}

	if v, ok := d.GetOk("labels"); ok && !isInterfaceNil(v) {
		duplo.Labels = expandAsStringMap("labels", d)
	}
	return &duplo, nil
}

func expandK8sHttpRouteParentRefs(lst []interface{}) *[]duplosdk.DuploK8sParentRef {
	refs := make([]duplosdk.DuploK8sParentRef, 0, len(lst))
	for _, v := range lst {
		m := v.(map[string]interface{})
		ref := duplosdk.DuploK8sParentRef{
			Name:        m["name"].(string),
			Namespace:   m["namespace"].(string),
			SectionName: m["section_name"].(string),
		}
		if port, ok := m["port"]; ok && port.(int) > 0 {
			ref.Port = strconv.Itoa(port.(int))
		}
		refs = append(refs, ref)
	}
	return &refs
}

func expandK8sHttpRouteRules(lst []interface{}) (*[]duplosdk.DuploK8sHTTPRouteRule, error) {
	rules := make([]duplosdk.DuploK8sHTTPRouteRule, 0, len(lst))
	for _, v := range lst {
		rule, err := expandK8sHttpRouteRule(v.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return &rules, nil
}

func expandK8sHttpRouteRule(m map[string]interface{}) (duplosdk.DuploK8sHTTPRouteRule, error) {
	rule := duplosdk.DuploK8sHTTPRouteRule{}

	if v, ok := m["match"]; ok && len(v.([]interface{})) > 0 {
		matches := make([]duplosdk.DuploK8sHTTPRouteMatch, 0, len(v.([]interface{})))
		for _, mv := range v.([]interface{}) {
			matches = append(matches, expandK8sHttpRouteMatch(mv.(map[string]interface{})))
		}
		rule.Matches = &matches
	}

	if v, ok := m["backend_ref"]; ok && len(v.([]interface{})) > 0 {
		refs := make([]duplosdk.DuploK8sHTTPBackendRef, 0, len(v.([]interface{})))
		for _, rv := range v.([]interface{}) {
			rm := rv.(map[string]interface{})
			weight := rm["weight"].(int)
			refs = append(refs, duplosdk.DuploK8sHTTPBackendRef{
				Name:      rm["name"].(string),
				Port:      rm["port"].(int),
				Weight:    &weight,
				Namespace: rm["namespace"].(string),
			})
		}
		rule.BackendRefs = &refs
	}

	if v, ok := m["filters"]; ok && v.(string) != "" {
		filters := []interface{}{}
		if err := json.Unmarshal([]byte(v.(string)), &filters); err != nil {
			return rule, fmt.Errorf("filters is not a valid JSON array: %s", err)
		}
		rule.Filters = filters
	}
	return rule, nil
}

func expandK8sHttpRouteMatch(m map[string]interface{}) duplosdk.DuploK8sHTTPRouteMatch {
	match := duplosdk.DuploK8sHTTPRouteMatch{
		Method: m["method"].(string),
	}
	if v, ok := m["path"]; ok && len(v.([]interface{})) == 1 {
		pm := v.([]interface{})[0].(map[string]interface{})
		match.Path = &duplosdk.DuploK8sPathMatch{
			Type:  pm["type"].(string),
			Value: pm["value"].(string),
		}
	}
	if v, ok := m["header"]; ok && len(v.([]interface{})) > 0 {
		headers := make([]duplosdk.DuploK8sHeaderMatch, 0, len(v.([]interface{})))
		for _, hv := range v.([]interface{}) {
			hm := hv.(map[string]interface{})
			headers = append(headers, duplosdk.DuploK8sHeaderMatch{
				Name:  hm["name"].(string),
				Value: hm["value"].(string),
				Type:  hm["type"].(string),
			})
		}
		match.Headers = &headers
	}
	return match
}

func flattenK8sHttpRoute(tenantId string, d *schema.ResourceData, duplo *duplosdk.DuploK8sHTTPRoute) {
	d.Set("tenant_id", tenantId)
	d.Set("name", duplo.Name)

	if duplo.ParentRefs != nil {
		d.Set("parent_ref", flattenK8sHttpRouteParentRefs(duplo.ParentRefs))
	}
	if len(duplo.Hostnames) > 0 {
		d.Set("hostnames", duplo.Hostnames)
	}
	if duplo.Rules != nil {
		d.Set("rule", flattenK8sHttpRouteRules(duplo.Rules))
	}

	d.Set("annotations", filterServerManagedKeys(d, "annotations", duplo.Annotations))
	d.Set("labels", filterServerManagedKeys(d, "labels", duplo.Labels))
}

func flattenK8sHttpRouteParentRefs(duplo *[]duplosdk.DuploK8sParentRef) []interface{} {
	lst := []interface{}{}
	for _, v := range *duplo {
		m := map[string]interface{}{
			"name":         v.Name,
			"namespace":    v.Namespace,
			"section_name": v.SectionName,
		}
		if len(v.Port) > 0 {
			if port, err := strconv.Atoi(v.Port); err == nil {
				m["port"] = port
			}
		}
		lst = append(lst, m)
	}
	return lst
}

func flattenK8sHttpRouteRules(duplo *[]duplosdk.DuploK8sHTTPRouteRule) []interface{} {
	lst := []interface{}{}
	for _, v := range *duplo {
		lst = append(lst, flattenK8sHttpRouteRule(v))
	}
	return lst
}

func flattenK8sHttpRouteRule(duplo duplosdk.DuploK8sHTTPRouteRule) map[string]interface{} {
	m := map[string]interface{}{}

	if duplo.Matches != nil && len(*duplo.Matches) > 0 {
		matches := make([]interface{}, 0, len(*duplo.Matches))
		for _, mv := range *duplo.Matches {
			matches = append(matches, flattenK8sHttpRouteMatch(mv))
		}
		m["match"] = matches
	}

	if duplo.BackendRefs != nil && len(*duplo.BackendRefs) > 0 {
		refs := make([]interface{}, 0, len(*duplo.BackendRefs))
		for _, rv := range *duplo.BackendRefs {
			rm := map[string]interface{}{
				"name":      rv.Name,
				"port":      rv.Port,
				"namespace": rv.Namespace,
			}
			if rv.Weight != nil {
				rm["weight"] = *rv.Weight
			}
			refs = append(refs, rm)
		}
		m["backend_ref"] = refs
	}

	if len(duplo.Filters) > 0 {
		if b, err := json.Marshal(duplo.Filters); err == nil {
			m["filters"] = string(b)
		}
	}
	return m
}

func flattenK8sHttpRouteMatch(duplo duplosdk.DuploK8sHTTPRouteMatch) map[string]interface{} {
	m := map[string]interface{}{
		"method": duplo.Method,
	}
	if duplo.Path != nil {
		m["path"] = []interface{}{map[string]interface{}{
			"type":  duplo.Path.Type,
			"value": duplo.Path.Value,
		}}
	}
	if duplo.Headers != nil && len(*duplo.Headers) > 0 {
		headers := make([]interface{}, 0, len(*duplo.Headers))
		for _, hv := range *duplo.Headers {
			headers = append(headers, map[string]interface{}{
				"name":  hv.Name,
				"value": hv.Value,
				"type":  hv.Type,
			})
		}
		m["header"] = headers
	}
	return m
}
