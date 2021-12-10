package duplosdk

import (
	"fmt"
)

// DuploSsmParameter is a Duplo SDK object that represents an SSM parameter.
type DuploSsmParameter struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name             string `json:"Name"`
	Type             string `json:"Type"`
	Value            string `json:"Value"`
	Description      string `json:"Description"`
	AllowedPattern   string `json:"AllowedPattern,omitempty"`
	KeyId            string `json:"KeyId,omitempty"`
	LastModifiedDate string `json:"LastModifiedDate,omitempty"`
	LastModifiedUser string `json:"LastModifiedUser,omitempty"`
}

// DuploSsmParameterRequest is a Duplo SDK object that represents a request to create or update an SSM parameter.
type DuploSsmParameterRequest struct {
	Name           string `json:"Name"`
	Type           string `json:"Type"`
	Value          string `json:"Value"`
	Description    string `json:"Description"`
	AllowedPattern string `json:"AllowedPattern,omitempty"`
	KeyId          string `json:"KeyId,omitempty"`
}

// SsmParameterCreate creates an SSM parameter via the Duplo API.
func (c *Client) SsmParameterCreate(tenantID string, rq *DuploSsmParameterRequest) (*DuploSsmParameter, ClientError) {
	rp := DuploSsmParameter{}
	err := c.postAPI(
		fmt.Sprintf("SsmParameterCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/ssmParameter", tenantID),
		&rq,
		&rp,
	)
	rp.TenantID = tenantID
	return &rp, err
}

// SsmParameterUpdate updates an SSM parameter via the Duplo API.
func (c *Client) SsmParameterUpdate(tenantID string, rq *DuploSsmParameterRequest) (*DuploSsmParameter, ClientError) {
	rp := DuploSsmParameter{}
	err := c.putAPI(
		fmt.Sprintf("SsmParameterCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/ssmParameter/%s", tenantID, rq.Name),
		&rq,
		&rp,
	)
	rp.TenantID = tenantID
	return &rp, err
}

// SsmParameterDelete deletes an SSM parameter via the Duplo API.
func (c *Client) SsmParameterDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("SsmParameterDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/ssmParameter/%s", tenantID, name),
		nil)
}

// SsmParameterGet retrieves an SSM parameter via the Duplo API
func (c *Client) SsmParameterGet(tenantID string, name string) (*DuploSsmParameter, ClientError) {
	rp := DuploSsmParameter{}
	err := c.getAPI(
		fmt.Sprintf("SsmParameterGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/ssmParameter/%s", tenantID, name),
		&rp)
	if rp.Name == "" {
		return nil, err
	}

	rp.TenantID = tenantID
	return &rp, err
}

// SsmParameterGet retrieves a list of SSM parameters via the Duplo API
func (c *Client) SsmParameterList(tenantID string) (*[]DuploSsmParameter, ClientError) {
	list := []DuploSsmParameter{}
	err := c.getAPI(
		fmt.Sprintf("SsmParameterList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/ssmParameter", tenantID),
		&list)
	if err != nil {
		return nil, err
	}

	// Add the tenant ID to each element and return the list.
	for i := range list {
		list[i].TenantID = tenantID
	}
	return &list, nil
}
