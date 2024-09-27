package duplosdk

import "fmt"

type DuploTenantAccessGrant struct {
	GrantorTenantId string `json:"GrantorTenantId,omitempty"`
	GrantedArea     string `json:"GrantedArea,omitempty"`
}

type DuploTenantAccessGrantStatus struct {
	DuploTenantAccessGrant
	Status string `json:"Status,omitempty"`
}

func (c *Client) CreateTenantAccessGrant(granteeTenantId string, duplo *DuploTenantAccessGrant) ClientError {
	// Create the bucket via Duplo.
	return c.postAPI(
		fmt.Sprintf("CreateTenantAccessGrant(%s, %s, %s)", granteeTenantId, duplo.GrantorTenantId, duplo.GrantedArea),
		fmt.Sprintf("3/admin/tenant/%s/accessGrant", granteeTenantId),
		&duplo,
		nil)
}

func (c *Client) GetTenantAccessGrant(granteeTenantId string, grantorTenantId string, grantedArea string) (*DuploTenantAccessGrant, ClientError) {
	rp := DuploTenantAccessGrant{}
	err := c.getAPI(fmt.Sprintf("GetTenantAccessGrant(%s, %s, %s)", granteeTenantId, grantorTenantId, grantedArea),
		fmt.Sprintf("3/admin/tenant/%s/accessGrant", granteeTenantId),
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

func (c *Client) DeleteTenantAccessGrant(granteeTenantId string, grantorTenantId string, grantedArea string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DeleteTenantAccessGrant(%s, %s, %s)", granteeTenantId, grantorTenantId, grantedArea),
		fmt.Sprintf("v3/admin/tenant/%s/accessGrant/%s/%s", granteeTenantId, grantorTenantId, grantedArea),
		nil)
}

func (c *Client) GetTenantAccessGrantStatus(granteeTenantId string, grantorTenantId string, grantedArea string) (*DuploTenantAccessGrantStatus, ClientError) {
	rp := DuploTenantAccessGrantStatus{}
	err := c.getAPI(fmt.Sprintf("GetTenantAccessGrantStatus(%s, %s, %s)", granteeTenantId, grantorTenantId, grantedArea),
		fmt.Sprintf("v3/admin/tenant/%s/accessGrant/%s/%s/status", granteeTenantId, grantorTenantId, grantedArea),
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, err
}
