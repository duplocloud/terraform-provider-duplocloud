package duplosdk

import (
	"fmt"
	"time"
)

type DuploK8sSecretProviderClass struct {
	Name          string                                     `json:"name"`
	Provider      string                                     `json:"provider"`
	Annotations   map[string]string                          `json:"annotations,omitempty"`
	Labels        map[string]string                          `json:"labels,omitempty"`
	Parameters    *DuploK8sSecretProviderClassParameters     `json:"parameters,omitempty"`
	SecretObjects *[]DuploK8sSecretProviderClassSecretObject `json:"secretObjects,omitempty"`
}

type DuploK8sSecretProviderClassSecretObject struct {
	SecretName  string                                         `json:"secretName"`
	Type        string                                         `json:"type"`
	Annotations map[string]string                              `json:"annotations,omitempty"`
	Labels      map[string]string                              `json:"labels,omitempty"`
	Data        *[]DuploK8sSecretProviderClassSecretObjectData `json:"data,omitempty"`
}

type DuploK8sSecretProviderClassSecretObjectData struct {
	Key        string `json:"key,omitempty"`
	ObjectName string `json:"objectName,omitempty"`
}

type DuploK8sSecretProviderClassParameters struct {
	Objects string `json:"objects,omitempty"`
}

type DuploK8sSecretProviderClassDetails struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Annotations struct {
			DuplocloudNetDuploObject string `json:"duplocloud.net/duplo-object"`
		} `json:"annotations"`
		CreationTimestamp time.Time `json:"creationTimestamp"`
		Generation        int       `json:"generation"`
		Labels            struct {
			DuplocloudNetOwner      string `json:"duplocloud.net/owner"`
			DuplocloudNetTenantid   string `json:"duplocloud.net/tenantid"`
			DuplocloudNetTenantname string `json:"duplocloud.net/tenantname"`
		} `json:"labels"`
		ManagedFields []struct {
			APIVersion string    `json:"apiVersion"`
			Manager    string    `json:"manager"`
			Operation  string    `json:"operation"`
			Time       time.Time `json:"time"`
		} `json:"managedFields"`
		Name            string `json:"name"`
		Namespace       string `json:"namespace"`
		ResourceVersion string `json:"resourceVersion"`
		UID             string `json:"uid"`
	} `json:"metadata"`
	Spec struct {
		Provider   string `json:"provider"`
		Parameters struct {
			Objects string `json:"objects"`
		} `json:"parameters"`
		SecretObjects []struct {
			SecretName string `json:"secretName"`
			Type       string `json:"type"`
			Data       []struct {
				Key        string `json:"key"`
				ObjectName string `json:"objectName"`
			} `json:"data"`
		} `json:"secretObjects"`
	} `json:"spec"`
}

func (c *Client) DuploK8sSecretProviderClassCreate(tenantID string, rq *DuploK8sSecretProviderClass) ClientError {
	rp := DuploK8sSecretProviderClass{}
	return c.postAPI(
		fmt.Sprintf("DuploK8sSecretProviderClassCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/secretproviderclass", tenantID),
		&rq,
		&rp,
	)
}

func (c *Client) DuploK8sSecretProviderClassUpdate(tenantID, name string, rq *DuploK8sSecretProviderClass) ClientError {
	rp := DuploK8sSecretProviderClass{}
	return c.putAPI(
		fmt.Sprintf("DuploK8sSecretProviderClassUpdate(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/secretproviderclass/%s", tenantID, name),
		&rq,
		&rp,
	)
}

func (c *Client) DuploK8sSecretProviderClassGet(tenantID, name string) (*DuploK8sSecretProviderClass, ClientError) {
	rp := DuploK8sSecretProviderClass{}
	err := c.getAPI(
		fmt.Sprintf("DuploK8sSecretProviderClassGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/secretproviderclass/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploK8sSecretProviderClassK8sGet(tenantID, name string) (*DuploK8sSecretProviderClassDetails, ClientError) {
	rp := DuploK8sSecretProviderClassDetails{}
	err := c.getAPI(
		fmt.Sprintf("DuploK8sSecretProviderClassGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/k8-secretproviderclass/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploK8sSecretProviderClassDelete(tenantID string, name string) ClientError {
	rp := DuploK8sSecretProviderClass{}
	return c.deleteAPI(
		fmt.Sprintf("DuploK8sSecretProviderClassDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/k8s/secretproviderclass/%s", tenantID, name),
		&rp,
	)
}
