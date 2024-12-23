package duplocloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"terraform-provider-duplocloud/duplosdk"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func k8sIngressSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"tenant_id": {
			Description:  "The GUID of the tenant that the Ingress will be created in.",
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validation.IsUUID,
		},
		"name": {
			Description: "The name of the Ingress.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"ingress_class_name": {
			Description: "The ingress class name references an IngressClass resource that contains additional configuration including the name of the controller that should implement the class.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"lbconfig": {
			Description: "The load balancer configuration. This is required when `ingress_class_name` is set to `alb`.",
			Type:        schema.TypeList,
			Computed:    true,
			Optional:    true,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"is_internal": {
						Description: "Whether or not to create an internal load balancer.",
						Type:        schema.TypeBool,
						Required:    true,
					},
					"dns_prefix": {
						Description: "The DNS prefix to expose services using Route53 domain.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"certificate_arn": {
						Description: "The ARN of an ACM certificate to associate with this load balancer.  Only applicable for HTTPS.",
						Type:        schema.TypeString,
						Computed:    true,
						Optional:    true,
					},
					"http_port": {
						Description: "HTTP Listener Port.",
						Type:        schema.TypeInt,
						Optional:    true,
					},
					"https_port": {
						Description: "HTTPS Listener Port.",
						Type:        schema.TypeInt,
						Optional:    true,
					},
				},
			},
		},
		"annotations": {
			Description: "An unstructured key value map stored with the ingress that may be used to store arbitrary metadata.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"labels": {
			Description: "Map of string keys and values that can be used to organize and categorize (scope and select) the service. May match selectors of replication controllers and services.",
			Type:        schema.TypeMap,
			Optional:    true,
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"rule": {
			Description: "A list of host rules used to configure the Ingress.",
			Type:        schema.TypeList,
			MinItems:    1,
			Computed:    true,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"path": {
						Description: "Specify the path (for e.g. /api /v1/api/) to do a path base routing. If host is specified then both path and host should be match for the incoming request.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"path_type": {
						Description: "Type of the path to be used.",
						Type:        schema.TypeString,
						Required:    true,
						ValidateFunc: validation.StringInSlice([]string{
							"Prefix",
							"Exact",
							"ImplementationSpecific",
						}, false),
					},
					"host": {
						Description: "If a host is provided (for e.g. example, foo.bar.com), the rules apply to that host.",
						Type:        schema.TypeString,
						Computed:    true,
						Optional:    true,
					},
					"service_name": {
						Description: "Name of the kubernetes service which Ingress will use as backend to serve the request.",
						Type:        schema.TypeString,
						Required:    true,
					},
					"port_name": {
						Description: "Port name from the kubernetes service that ingress will use as backend port to serve the requests.",
						Type:        schema.TypeString,
						Optional:    true,
					},
					"port": {
						Description:  "Port from the kubernetes service that ingress will use as backend port to serve the requests.",
						ValidateFunc: validation.IntBetween(1, 65535),
						Type:         schema.TypeInt,
						Optional:     true,
					},
				},
			},
		},
		"tls": {
			Description: "Block represents the TLS configuration. Currently the Ingress only supports a single TLS port, 443. If multiple members of this list specify different hosts, they will be multiplexed on the same port according to the hostname specified through the SNI TLS extension, if the ingress controller fulfilling the ingress supports SNI",
			Type:        schema.TypeList,
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"hosts": {
						Description: "The list of hosts included in the TLS certificate. Each value in this list must match the name(s) specified in the TLS secret. If not specified, it defaults to the wildcard host setting for the load balancer controller managing this Ingress.",
						Type:        schema.TypeList,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Required: true,
					},
					"secret_name": {
						Description: "The name of the secret used to terminate TLS traffic on port 443. This field is optional, enabling TLS routing based solely on the SNI hostname. If the SNI host in a listener conflicts with the 'Host' header in an IngressRule, the SNI host is used for termination, while the 'Host' header value is used for routing.",
						Type:        schema.TypeString,
						Required:    true,
					},
				},
			},
		},
	}
}

// SCHEMA for resource crud
func resourceK8Ingress() *schema.Resource {
	return &schema.Resource{
		Description: "`duplocloud_k8_ingress` manages a kubernetes Ingress in a Duplo tenant.",

		ReadContext:   resourceK8IngressRead,
		CreateContext: resourceK8IngressCreate,
		UpdateContext: resourceK8IngressUpdate,
		DeleteContext: resourceK8IngressDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},
		Schema: k8sIngressSchema(),
	}
}

func resourceK8IngressRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sIngressIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8IngressRead(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.DuploK8sIngressGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s ingress %s : %s", tenantID, name, clientErr)
	}
	if rp == nil || rp.Name == "" {
		d.SetId("")
		return nil
	}

	flattenK8sIngress(tenantID, d, rp)
	log.Printf("[TRACE] resourceK8IngressRead(%s, %s): end", tenantID, name)
	return nil
}

// CREATE resource
func resourceK8IngressCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID := d.Get("tenant_id").(string)
	name := d.Get("name").(string)

	log.Printf("[TRACE] resourceK8IngressCreate(%s, %s): start", tenantID, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sIngress(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	cerr := c.DuploK8sIngressCreate(tenantID, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}
	d.SetId(fmt.Sprintf("v3/subscriptions/%s/k8s/ingress/%s", tenantID, name))

	diags := resourceK8IngressRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8IngressCreate(%s, %s): end", tenantID, name)
	return diags
}

// UPDATE resource
func resourceK8IngressUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sIngressIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8IngressUpdate(%s, %s): start", tenantID, name)

	// Convert the Terraform resource data into a Duplo object
	rq, err := expandK8sIngress(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Post the object to Duplo
	c := m.(*duplosdk.Client)
	cerr := c.DuploK8sIngressUpdate(tenantID, name, rq)
	if cerr != nil {
		return diag.FromErr(cerr)
	}

	diags := resourceK8IngressRead(ctx, d, m)
	log.Printf("[TRACE] resourceK8IngressUpdate(%s, %s): end", tenantID, name)
	return diags
}

// DELETE resource
func resourceK8IngressDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tenantID, name, err := parseK8sIngressIdParts(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[TRACE] resourceK8IngressDelete(%s, %s): start", tenantID, name)

	// Get the object from Duplo, detecting a missing object
	c := m.(*duplosdk.Client)
	rp, clientErr := c.DuploK8sIngressGet(tenantID, name)
	if clientErr != nil {
		if clientErr.Status() == 404 {
			d.SetId("")
			return nil
		}
		return diag.Errorf("Unable to retrieve tenant %s k8s ingress %s : %s", tenantID, name, clientErr)
	}
	if rp != nil && rp.Name != "" {
		clientErr := c.DuploK8sIngressDelete(tenantID, name)
		if clientErr != nil {
			if clientErr.Status() == 404 {
				d.SetId("")
				return nil
			}
			return diag.Errorf("Unable to delete tenant %s k8s ingress %s : %s", tenantID, name, clientErr)
		}
	}

	log.Printf("[TRACE] resourceK8IngressDelete(%s, %s): end", tenantID, name)
	return nil
}

func parseK8sIngressIdParts(id string) (tenantID, name string, err error) {
	idParts := strings.SplitN(id, "/", 6)
	if len(idParts) == 6 {
		tenantID, name = idParts[2], idParts[5]
	} else {
		err = fmt.Errorf("invalid resource ID: %s", id)
	}
	return
}

func flattenK8sIngress(tenantId string, d *schema.ResourceData, duplo *duplosdk.DuploK8sIngress) {
	// First, set the simple fields.
	d.Set("tenant_id", tenantId)
	d.Set("name", duplo.Name)
	d.Set("ingress_class_name", duplo.IngressClassName)

	// Set LBConfig
	if duplo.LbConfig != nil {
		d.Set("lbconfig", []interface{}{flattenK8sIngressLBConfig(duplo.LbConfig)})
	}

	// Set Rules
	if duplo.Rules != nil && len(*duplo.Rules) > 0 {
		d.Set("rule", flattenK8sIngressRules(duplo.Rules))
	}

	// Finally, set the map
	d.Set("annotations", duplo.Annotations)
	d.Set("labels", duplo.Labels)
	if duplo.OtherSpec != nil {
		d.Set("tls", flattenTls(duplo.OtherSpec))
	}
}

func flattenK8sIngressLBConfig(duplo *duplosdk.DuploK8sLbConfig) map[string]interface{} {
	m := map[string]interface{}{
		"is_internal":     !duplo.IsPublic,
		"dns_prefix":      duplo.DnsPrefix,
		"certificate_arn": duplo.CertArn,
	}
	if duplo.Listeners != nil && len(duplo.Listeners.Http) > 0 {
		m["http_port"] = duplo.Listeners.Http[0]
	}
	if duplo.Listeners != nil && len(duplo.Listeners.Https) > 0 {
		m["https_port"] = duplo.Listeners.Https[0]
	}
	return m
}

func flattenK8sIngressRules(duplo *[]duplosdk.DuploK8sIngressRule) []interface{} {

	lst := []interface{}{}
	for _, v := range *duplo {
		lst = append(lst, flattenK8sIngressRule(v))
	}
	return lst
}

func flattenK8sIngressRule(duplo duplosdk.DuploK8sIngressRule) map[string]interface{} {
	m := map[string]interface{}{
		"path":      duplo.Path,
		"path_type": duplo.PathType,
		"port":      duplo.Port,
	}
	if len(duplo.Host) > 0 {
		m["host"] = duplo.Host
	}
	if len(duplo.ServiceName) > 0 {
		m["service_name"] = duplo.ServiceName
	}
	if len(duplo.PortName) > 0 {
		m["port_name"] = duplo.PortName
	}
	return m
}

func expandK8sIngress(d *schema.ResourceData) (*duplosdk.DuploK8sIngress, error) {
	rules, err := expandK8sIngressRules(d.Get("rule").([]interface{}))
	if err != nil {
		return nil, err
	}

	duplo := duplosdk.DuploK8sIngress{
		Name:             d.Get("name").(string),
		IngressClassName: d.Get("ingress_class_name").(string),
		Rules:            rules,
	}

	if v, ok := d.GetOk("lbconfig"); ok && !isInterfaceNil(v) && len(v.([]interface{})) == 1 {
		duplo.LbConfig = expandK8sIngressLBConfig(d.Get("lbconfig").([]interface{})[0].(map[string]interface{}))
	}
	// The annotations must be converted to a map of strings.
	if v, ok := d.GetOk("annotations"); ok && !isInterfaceNil(v) {
		duplo.Annotations = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			duplo.Annotations[key] = value.(string)
		}
	}

	// The labels must be converted to a map of strings.
	if v, ok := d.GetOk("labels"); ok && !isInterfaceNil(v) {
		duplo.Labels = map[string]string{}
		for key, value := range v.(map[string]interface{}) {
			duplo.Labels[key] = value.(string)
		}
	}
	if v, ok := d.GetOk("tls"); ok && !isInterfaceNil(v) {
		duplo.OtherSpec = expandTls(v.([]interface{}))
	}
	return &duplo, nil
}

func expandK8sIngressLBConfig(m map[string]interface{}) *duplosdk.DuploK8sLbConfig {
	dcb := &duplosdk.DuploK8sLbConfig{
		IsPublic:  !m["is_internal"].(bool),
		DnsPrefix: m["dns_prefix"].(string),
		CertArn:   m["certificate_arn"].(string),
	}
	l := duplosdk.DuploK8sIngressListeners{}
	if v, ok := m["http_port"]; ok && v.(int) > 0 {
		l.Http = []int{
			v.(int),
		}
	}
	if v, ok := m["https_port"]; ok && v.(int) > 0 {
		l.Https = []int{
			v.(int),
		}
	}
	dcb.Listeners = &l
	return dcb
}

func expandK8sIngressRules(lst []interface{}) (*[]duplosdk.DuploK8sIngressRule, error) {
	rules := make([]duplosdk.DuploK8sIngressRule, 0, len(lst))
	for _, v := range lst {
		rule, err := expandK8sIngressRule(v.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return &rules, nil
}

func expandK8sIngressRule(m map[string]interface{}) (duplosdk.DuploK8sIngressRule, error) {
	r := duplosdk.DuploK8sIngressRule{
		Path:     m["path"].(string),
		PathType: m["path_type"].(string),
	}
	if v, ok := m["port"]; ok {
		r.Port = v.(int)
	}
	if v, ok := m["port_name"]; ok {
		r.PortName = v.(string)
		if r.Port > 0 && r.PortName != "" {
			return r, fmt.Errorf("port and port_name are mutually exclusive")
		}
	}
	if v, ok := m["host"]; ok {
		r.Host = v.(string)
	}
	if v, ok := m["service_name"]; ok {
		r.ServiceName = v.(string)
	}
	return r, nil
}

func expandTls(i []interface{}) *duplosdk.DuploK8IngressOtherSpec {
	obj := duplosdk.DuploK8IngressOtherSpec{}
	hosts := []string{}
	tls := []duplosdk.DuploK8IngressSpecTls{}
	for _, s := range i {
		m := s.(map[string]interface{})
		hs := m["hosts"].([]interface{})
		for _, h := range hs {
			hosts = append(hosts, h.(string))
		}
		s := m["secret_name"].(string)

		obj := duplosdk.DuploK8IngressSpecTls{
			Host:       hosts,
			SecretName: s,
		}
		tls = append(tls, obj)

	}
	obj.Tls = tls
	return &obj
}

func flattenTls(obj *duplosdk.DuploK8IngressOtherSpec) []interface{} {
	i := make([]interface{}, 0, len(obj.Tls))

	for _, tls := range obj.Tls {
		m := map[string]interface{}{
			"hosts":       tls.Host,
			"secret_name": tls.SecretName,
		}
		i = append(i, m)
	}
	return i
}
