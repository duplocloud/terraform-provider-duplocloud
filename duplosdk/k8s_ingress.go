package duplosdk

import "fmt"

type DuploK8sIngress struct {
	Name             string                 `json:"name"`
	IngressClassName string                 `json:"ingressClassName"`
	Annotations      map[string]string      `json:"annotations,omitempty"`
	Labels           map[string]string      `json:"labels,omitempty"`
	LbConfig         *DuploK8sLbConfig      `json:"lbConfig,omitempty"`
	Rules            *[]DuploK8sIngressRule `json:"rules,omitempty"`
}

type DuploK8sLbConfig struct {
	IsPublic          bool                      `json:"isPublic,omitempty"`
	DnsPrefix         string                    `json:"dnsPrefix,omitempty"`
	WafArn            string                    `json:"wafArn,omitempty"`
	EnableAccessLogs  bool                      `json:"enableAccessLogs,omitempty"`
	DropInvalidHeader bool                      `json:"dropInvalidHeader,omitempty"`
	CertArn           string                    `json:"certArn,omitempty"`
	Listeners         *DuploK8sIngressListeners `json:"listeners,omitempty"`
}

type DuploK8sIngressListeners struct {
	Http  []int `json:"http,omitempty"`
	Https []int `json:"https,omitempty"`
	Tcp   []int `json:"tcp,omitempty"`
}

type DuploK8sIngressRule struct {
	Path        string `json:"path,omitempty"`
	PathType    string `json:"pathType,omitempty"`
	ServiceName string `json:"serviceName,omitempty"`
	Host        string `json:"host,omitempty"`
	Port        int    `json:"port,omitempty"`
}

func (c *Client) DuploK8sIngressCreate(tenantID string, rq *DuploK8sIngress) ClientError {
	rp := DuploK8sIngress{}
	return c.postAPI(
		fmt.Sprintf("DuploK8sIngressCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/ingress", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) DuploK8sIngressUpdate(tenantID, name string, rq *DuploK8sIngress) ClientError {
	rp := DuploK8sIngress{}
	return c.putAPI(
		fmt.Sprintf("DuploK8sIngressUpdate(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/ingress/%s", tenantID, name),
		&rq,
		&rp,
	)
}
func (c *Client) DuploK8sIngressGet(tenantID, name string) (*DuploK8sIngress, ClientError) {
	rp := DuploK8sIngress{}
	err := c.getAPI(
		fmt.Sprintf("DuploK8sIngressGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/ingress/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploK8sIngressDelete(tenantID string, name string) ClientError {
	rp := DuploK8sIngress{}
	return c.deleteAPI(
		fmt.Sprintf("DuploK8sIngressDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/ingress/%s", tenantID, name),
		&rp,
	)
}
