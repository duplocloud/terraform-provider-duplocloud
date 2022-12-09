package duplosdk

import (
	"fmt"
)

type DuploK8sPvc struct {
	Name        string            `json:"name"`
	Spec        *DuploK8sPvcSpec  `json:"spec,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

type DuploK8sPvcSpec struct {
	AccessModes      []string                  `json:"accessModes,omitempty"`
	StorageClassName string                    `json:"storageClassName,omitempty"`
	VolumeMode       string                    `json:"volumeMode,omitempty"`
	Resources        *DuploK8sPvcSpecResources `json:"resources,omitempty"`
	VolumeName       string                    `json:"volumeName,omitempty"`
}

type DuploK8sPvcSpecResources struct {
	Requests map[string]string `json:"requests,omitempty"`
	Limits   map[string]string `json:"limits,omitempty"`
}

func (c *Client) K8PvcGetList(tenantID string) (*[]DuploK8sPvc, ClientError) {
	rp := []DuploK8sPvc{}
	err := c.getAPI(
		fmt.Sprintf("K8PvcGetList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/k8s/pvc", tenantID),
		&rp)

	return &rp, err
}

func (c *Client) K8PvcGet(tenantID, PvcFullName string) (*DuploK8sPvc, ClientError) {

	// Retrieve the list of Pvcs
	list, err := c.K8PvcGetList(tenantID)
	if err != nil || list == nil {
		return nil, err
	}

	// Return the Pvc, if it exists.
	for i := range *list {
		if (*list)[i].Name == PvcFullName {
			return &(*list)[i], nil
		}
	}

	return nil, nil
}

func (c *Client) K8PvcCreate(tenantID string, rq *DuploK8sPvc) (*DuploK8sPvc, ClientError) {
	return c.K8PvcCreateOrUpdate(tenantID, rq, false)
}

func (c *Client) K8PvcUpdate(tenantID string, rq *DuploK8sPvc) (*DuploK8sPvc, ClientError) {
	return c.K8PvcCreateOrUpdate(tenantID, rq, true)
}

func (c *Client) K8PvcCreateOrUpdate(tenantID string, rq *DuploK8sPvc, updating bool) (*DuploK8sPvc, ClientError) {
	resp := DuploK8sPvc{}
	err := c.postAPI(
		fmt.Sprintf("K8PvcCreateOrUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/pvc", tenantID),
		&rq,
		&resp,
	)
	return &resp, err
}

func (c *Client) K8PvcDelete(tenantID, PvcFullName string) (*DuploK8sPvc, ClientError) {
	resp := DuploK8sPvc{}
	err := c.deleteAPI(
		fmt.Sprintf("K8PvcDelete(%s, %s)", tenantID, PvcFullName),
		fmt.Sprintf("v3/subscriptions/%s/k8s/pvc/%s", tenantID, PvcFullName),
		&resp)
	return &resp, err
}
