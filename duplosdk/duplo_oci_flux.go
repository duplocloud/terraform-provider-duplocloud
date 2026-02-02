package duplosdk

import (
	"fmt"
)

// DuploServiceParams represents a service's parameters in the Duplo SDK
type DuploOCIRepository struct {
	Metadata DuploOCIMetadata `json:"metadata"`
	Spec     *DuploOCISpec    `json:"spec"`
}
type DuploOCIMetadata struct {
	Name string `json:"name"`
}
type DuploOCISpec struct {
	URL       string                `json:"url,omitempty"`
	Interval  string                `json:"interval"`
	Ref       *DuploOCISpecRef      `json:"ref,omitempty"`
	Selector  *DuploOCILayerSeletor `json:"layerSelector,omitempty"`
	Secretref *DuploOCISecretRef    `json:"secretRef,omitempty"`
}
type DuploOCISpecRef struct {
	Tag string `json:"tag,omitempty"`
}

type DuploOCILayerSeletor struct {
	MediaType string `json:"mediaType,omitempty"`
	Operation string `json:"operation,omitempty"`
}

type DuploOCISecretRef struct {
	Name string `json:"name,omitempty"`
}

func (c *Client) DuploOCIRepositoryCreate(tenantID string, rq *DuploOCIRepository) ClientError {
	resp := map[string]interface{}{}
	err := c.postAPI(
		fmt.Sprintf("DuploOCIRepositoryCreate(%s, %s)", tenantID, rq.Metadata.Name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/fluxV2OciRepository", tenantID),
		&rq,
		&resp,
	)
	return err
}

func (c *Client) DuploOCIRepositoryUpdate(tenantID string, rq *DuploOCIRepository) ClientError {
	resp := map[string]interface{}{}
	err := c.putAPI(
		fmt.Sprintf("DuploOCIRepositoryUpdate(%s, %s)", tenantID, rq.Metadata.Name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/fluxV2OciRepository/%s", tenantID, rq.Metadata.Name),
		&rq,
		&resp,
	)
	return err
}

func (c *Client) DuploOCIRepositoryGet(tenantID string, name string) (*DuploOCIRepository, ClientError) {
	resp := DuploOCIRepository{}
	err := c.getAPI(
		fmt.Sprintf("DuploOCIRepositoryGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/fluxV2OciRepository/%s", tenantID, name),
		&resp,
	)
	return &resp, err
}

func (c *Client) DuploOCIRepositoryDelete(tenantID, name string) ClientError {
	err := c.deleteAPI(
		fmt.Sprintf("DuploOCIRepositoryDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/fluxV2OciRepository/%s", tenantID, name),
		nil,
	)
	return err
}
