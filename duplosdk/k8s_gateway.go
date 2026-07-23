package duplosdk

import "fmt"

type DuploK8sGateway struct {
	Name             string                     `json:"name"`
	GatewayClassName string                     `json:"gatewayClassName,omitempty"`
	IsPublic         bool                       `json:"isPublic,omitempty"`
	Listeners        *[]DuploK8sGatewayListener `json:"listeners,omitempty"`
	Addresses        *[]DuploK8sGatewayAddress  `json:"addresses,omitempty"`
	Annotations      map[string]string          `json:"annotations,omitempty"`
	Labels           map[string]string          `json:"labels,omitempty"`
}

type DuploK8sGatewayListener struct {
	Name          string                        `json:"name"`
	Port          int                           `json:"port"`
	Protocol      string                        `json:"protocol"`
	Hostname      string                        `json:"hostname,omitempty"`
	TLS           *DuploK8sGatewayListenerTLS   `json:"tls,omitempty"`
	AllowedRoutes *DuploK8sGatewayAllowedRoutes `json:"allowedRoutes,omitempty"`
}

type DuploK8sGatewayListenerTLS struct {
	Mode              string                    `json:"mode,omitempty"`
	CertificateRefs   *[]DuploK8sCertificateRef `json:"certificateRefs,omitempty"`
	CertMapAnnotation string                    `json:"certMapAnnotation,omitempty"`
}

type DuploK8sCertificateRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Group     string `json:"group,omitempty"`
}

type DuploK8sGatewayAllowedRoutes struct {
	From              string            `json:"from,omitempty"`
	NamespaceSelector map[string]string `json:"namespaceSelector,omitempty"`
}

type DuploK8sGatewayAddress struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value"`
}

type DuploK8sHTTPRoute struct {
	Name        string                   `json:"name"`
	ParentRefs  *[]DuploK8sParentRef     `json:"parentRefs,omitempty"`
	Hostnames   []string                 `json:"hostnames,omitempty"`
	Rules       *[]DuploK8sHTTPRouteRule `json:"rules,omitempty"`
	Annotations map[string]string        `json:"annotations,omitempty"`
	Labels      map[string]string        `json:"labels,omitempty"`
}

type DuploK8sParentRef struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace,omitempty"`
	SectionName string `json:"sectionName,omitempty"`
	Port        string `json:"port,omitempty"`
}

type DuploK8sHTTPRouteRule struct {
	Matches     *[]DuploK8sHTTPRouteMatch `json:"matches,omitempty"`
	BackendRefs *[]DuploK8sHTTPBackendRef `json:"backendRefs,omitempty"`
	Filters     []interface{}             `json:"filters,omitempty"`
}

type DuploK8sHTTPRouteMatch struct {
	Path    *DuploK8sPathMatch     `json:"path,omitempty"`
	Headers *[]DuploK8sHeaderMatch `json:"headers,omitempty"`
	Method  string                 `json:"method,omitempty"`
}

type DuploK8sPathMatch struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

type DuploK8sHeaderMatch struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
	Type  string `json:"type,omitempty"`
}

type DuploK8sHTTPBackendRef struct {
	Name      string `json:"name"`
	Port      int    `json:"port,omitempty"`
	Weight    *int   `json:"weight,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type DuploK8sReferenceGrant struct {
	Name string                        `json:"name"`
	From *[]DuploK8sReferenceGrantFrom `json:"from,omitempty"`
	To   *[]DuploK8sReferenceGrantTo   `json:"to,omitempty"`
}

type DuploK8sReferenceGrantFrom struct {
	Group     string `json:"group"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
}

type DuploK8sReferenceGrantTo struct {
	Group string `json:"group"`
	Kind  string `json:"kind"`
	Name  string `json:"name,omitempty"`
}

func (c *Client) DuploK8sGatewayCreate(tenantID string, rq *DuploK8sGateway) ClientError {
	rp := DuploK8sGateway{}
	return c.postAPI(
		fmt.Sprintf("DuploK8sGatewayCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/gateway", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) DuploK8sGatewayUpdate(tenantID, name string, rq *DuploK8sGateway) ClientError {
	rp := DuploK8sGateway{}
	return c.putAPI(
		fmt.Sprintf("DuploK8sGatewayUpdate(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/gateway/%s", tenantID, name),
		&rq,
		&rp,
	)
}

func (c *Client) DuploK8sGatewayGet(tenantID, name string) (*DuploK8sGateway, ClientError) {
	rp := DuploK8sGateway{}
	err := c.getAPI(
		fmt.Sprintf("DuploK8sGatewayGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/gateway/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploK8sGatewayDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploK8sGatewayDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/gateway/%s", tenantID, name),
		nil,
	)
}

func (c *Client) DuploK8sHTTPRouteCreate(tenantID string, rq *DuploK8sHTTPRoute) ClientError {
	rp := DuploK8sHTTPRoute{}
	return c.postAPI(
		fmt.Sprintf("DuploK8sHTTPRouteCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/httproute", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) DuploK8sHTTPRouteUpdate(tenantID, name string, rq *DuploK8sHTTPRoute) ClientError {
	rp := DuploK8sHTTPRoute{}
	return c.putAPI(
		fmt.Sprintf("DuploK8sHTTPRouteUpdate(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/httproute/%s", tenantID, name),
		&rq,
		&rp,
	)
}

func (c *Client) DuploK8sHTTPRouteGet(tenantID, name string) (*DuploK8sHTTPRoute, ClientError) {
	rp := DuploK8sHTTPRoute{}
	err := c.getAPI(
		fmt.Sprintf("DuploK8sHTTPRouteGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/httproute/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploK8sHTTPRouteDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploK8sHTTPRouteDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/httproute/%s", tenantID, name),
		nil,
	)
}

func (c *Client) DuploK8sReferenceGrantCreate(tenantID string, rq *DuploK8sReferenceGrant) ClientError {
	rp := DuploK8sReferenceGrant{}
	return c.postAPI(
		fmt.Sprintf("DuploK8sReferenceGrantCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/referencegrant", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) DuploK8sReferenceGrantUpdate(tenantID, name string, rq *DuploK8sReferenceGrant) ClientError {
	rp := DuploK8sReferenceGrant{}
	return c.putAPI(
		fmt.Sprintf("DuploK8sReferenceGrantUpdate(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/referencegrant/%s", tenantID, name),
		&rq,
		&rp,
	)
}

func (c *Client) DuploK8sReferenceGrantGet(tenantID, name string) (*DuploK8sReferenceGrant, ClientError) {
	rp := DuploK8sReferenceGrant{}
	err := c.getAPI(
		fmt.Sprintf("DuploK8sReferenceGrantGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/referencegrant/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploK8sReferenceGrantDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploK8sReferenceGrantDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/referencegrant/%s", tenantID, name),
		nil,
	)
}
