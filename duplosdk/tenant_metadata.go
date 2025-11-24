package duplosdk

import (
	"fmt"
)

type DuploTenantMetadata struct {
	TenantId string `json:"ComponentId,omitempty"`
	Type     string `json:"Type"` //aws_console, text, url
	Key      string `json:"Key"`
	Value    string `json:"Value"`
	State    string `json:"State,omitempty"`
}

func (c *Client) TenantMetadatasGetByKey(tenantID string, key string) (*DuploTenantMetadata, ClientError) {
	rp := []DuploTenantMetadata{}
	err := c.getAPI(
		fmt.Sprintf("TenantMetadataGetByKey(%s,%s)", tenantID, key),
		fmt.Sprintf("admin/GetTenantConfigData/%s", tenantID),
		&rp)
	if err != nil {
		return nil, NewCustomError(err.Error(), err.Status())
	}
	for _, v := range rp {
		if v.Key == key {
			return &v, err
		}
	}
	return nil, NewCustomError("not found", 404)
}

func (c *Client) TenantMetadataManage(tenantID string, rq *DuploTenantMetadata) ClientError {
	err := c.postAPI(
		fmt.Sprintf("TenantMetadataCreate(%s)", tenantID),
		"admin/UpdateTenantConfigData",
		&rq, nil)
	return err
}
