package duplosdk

import (
	"fmt"
)

// DuploServiceParams represents a service's parameters in the Duplo SDK
type DuploHelmRepository struct {
	Metadata DuploHelmMetadata `json:"Metadata"`
	Spec     DuploHelmSpec     `json:"spec"`
}
type DuploHelmMetadata struct {
	Name string `json:"name"`
}

type DuploHelmSpec struct {
	URL         string `json:"url,omitempty"`
	Interval    string `json:"interval"`
	ReleaseName string `json:"releaseName,omitempty"`
	Chart       *Chart `json:"chart,omitempty"`
	//Values      map[string]interface{} `json:"values,omitempty"`
	Values          interface{} `json:"values,omitempty"`
	Insecure        bool        `json:"Insecure"`
	PassCredentials bool        `json:"PassCredentials"`
	Provider        string      `json:"Provider"`
	Suspend         bool        `json:"Suspend"`
	Type            string      `json:"Type"`
}

type SourceRef struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type ChartSpec struct {
	Interval          string    `json:"interval"`
	Chart             string    `json:"chart"`
	Version           string    `json:"version"`
	ReconcileStrategy string    `json:"reconcileStrategy"`
	SourceRef         SourceRef `json:"sourceRef"`
}

type Chart struct {
	Spec ChartSpec `json:"spec"`
}

func (c *Client) DuploHelmRepositoryCreate(tenantID string, rq *DuploHelmRepository) ClientError {
	resp := map[string]interface{}{}
	err := c.postAPI(
		fmt.Sprintf("DuploHelmRepositoryCreate(%s, %s)", tenantID, rq.Metadata.Name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/fluxV2HelmRepository", tenantID),
		&rq,
		&resp,
	)
	return err
}

func (c *Client) DuploHelmRepositoryUpdate(tenantID string, rq *DuploHelmRepository) ClientError {
	resp := map[string]interface{}{}
	err := c.putAPI(
		fmt.Sprintf("DuploHelmRepositoryUpdate(%s, %s)", tenantID, rq.Metadata.Name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/fluxV2HelmRepository/%s", tenantID, rq.Metadata.Name),
		&rq,
		&resp,
	)
	return err
}

func (c *Client) DuploHelmRepositoryGet(tenantID string, name string) (*DuploHelmRepository, ClientError) {
	resp := DuploHelmRepository{}
	err := c.getAPI(
		fmt.Sprintf("DuploHelmRepositoryGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/fluxV2HelmRepository/%s", tenantID, name),
		&resp,
	)
	return &resp, err
}

func (c *Client) DuploHelmRepositoryDelete(tenantID, name string) ClientError {
	err := c.deleteAPI(
		fmt.Sprintf("DuploHelmRepositoryDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/fluxV2HelmRepository/%s", tenantID, name),
		nil,
	)
	return err
}

type DuploHelmRelease struct {
	Metadata DuploHelmMetadata      `json:"Metadata"`
	Spec     DuploHelmSpec          `json:"spec"`
	Status   DuploHelmReleaseStatus `json:"status"`
}
type DuploHelmReleaseStatus struct {
	Condition []DuploHelmReleaseStatusCondn `json:"conditions"`
}

type DuploHelmReleaseStatusCondn struct {
	Type string `json:"type"`
}

func (c *Client) DuploHelmReleaseCreate(tenantID string, rq *DuploHelmRelease) ClientError {
	resp := map[string]interface{}{}
	err := c.postAPI(
		fmt.Sprintf("DuploHelmReleaseCreate(%s, %s)", tenantID, rq.Metadata.Name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/fluxV2HelmRelease", tenantID),
		&rq,
		&resp,
	)
	return err
}

func (c *Client) DuploHelmReleaseUpdate(tenantID string, rq *DuploHelmRelease) ClientError {
	resp := map[string]interface{}{}
	err := c.putAPI(
		fmt.Sprintf("DuploHelmReleaseUpdate(%s, %s)", tenantID, rq.Metadata.Name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/fluxV2HelmRelease/%s", tenantID, rq.Metadata.Name),
		&rq,
		&resp,
	)
	return err
}

func (c *Client) DuploHelmReleaseGet(tenantID string, name string) (*DuploHelmRelease, ClientError) {
	resp := DuploHelmRelease{}
	err := c.getAPI(
		fmt.Sprintf("DuploHelmReleaseGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/fluxV2HelmRelease/%s", tenantID, name),
		&resp,
	)
	return &resp, err
}

func (c *Client) DuploHelmReleaseDelete(tenantID, name string) ClientError {
	err := c.deleteAPI(
		fmt.Sprintf("DuploHelmReleaseDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/fluxV2HelmRelease/%s", tenantID, name),
		nil,
	)
	return err
}
