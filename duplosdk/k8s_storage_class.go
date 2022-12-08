package duplosdk

import (
	"fmt"
)

type DuploK8sStorageClassAllowedTopologies struct {
	Key    string   `json:"key,omitempty"`
	Values []string `json:"values,omitempty"`
}

type DuploK8sAllowedTopologiesMatchLabelExpressions struct {
	MatchLabelExpressions *[]DuploK8sStorageClassAllowedTopologies `json:"matchLabelExpressions,omitempty"`
}

type DuploK8sStorageClass struct {
	Name                 string                                            `json:"name"`
	Provisioner          string                                            `json:"provisioner"`
	ReclaimPolicy        string                                            `json:"reclaimPolicy,omitempty"`
	VolumeBindingMode    string                                            `json:"volumeBindingMode,omitempty"`
	Annotations          map[string]string                                 `json:"annotations,omitempty"`
	Labels               map[string]string                                 `json:"labels,omitempty"`
	Parameters           map[string]string                                 `json:"parameters,omitempty"`
	AllowVolumeExpansion bool                                              `json:"allowVolumeExpansion"`
	AllowedTopologies    *[]DuploK8sAllowedTopologiesMatchLabelExpressions `json:"allowedTopologies,omitempty"`
}

func (c *Client) K8StorageClassGetList(tenantID string) (*[]DuploK8sStorageClass, ClientError) {
	rp := []DuploK8sStorageClass{}
	err := c.getAPI(
		fmt.Sprintf("K8StorageClassGetList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/k8s/storageclass", tenantID),
		&rp)

	return &rp, err
}

func (c *Client) K8StorageClassGet(tenantID, StorageClassFullName string) (*DuploK8sStorageClass, ClientError) {

	// Retrieve the list of StorageClasss
	list, err := c.K8StorageClassGetList(tenantID)
	if err != nil || list == nil {
		return nil, err
	}

	// Return the StorageClass, if it exists.
	for i := range *list {
		if (*list)[i].Name == StorageClassFullName {
			return &(*list)[i], nil
		}
	}

	return nil, nil
}

// K8StorageClassCreate creates a k8s StorageClass via the Duplo API.
func (c *Client) K8StorageClassCreate(tenantID string, rq *DuploK8sStorageClass) ClientError {
	return c.K8StorageClassCreateOrUpdate(tenantID, rq, false)
}

// K8StorageClassUpdate updates a k8s StorageClass via the Duplo API.
func (c *Client) K8StorageClassUpdate(tenantID string, rq *DuploK8sStorageClass) ClientError {
	return c.K8StorageClassCreateOrUpdate(tenantID, rq, true)
}

// K8StorageClassCreateOrUpdate creates or updates a k8s StorageClass via the Duplo API.
func (c *Client) K8StorageClassCreateOrUpdate(tenantID string, rq *DuploK8sStorageClass, updating bool) ClientError {
	return c.postAPI(
		fmt.Sprintf("K8StorageClassCreateOrUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("subscriptions/%s/k8s/storageclass", tenantID),
		&rq,
		nil,
	)
}

// K8StorageClassDelete deletes a k8s StorageClass via the Duplo API.
func (c *Client) K8StorageClassDelete(tenantID, StorageClassFullName string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("K8StorageClassDelete(%s, %s)", tenantID, StorageClassFullName),
		fmt.Sprintf("v2/subscriptions/%s/k8s/storageclass/%s", tenantID, StorageClassFullName),
		nil)
}
