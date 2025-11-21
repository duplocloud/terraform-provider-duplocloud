package duplosdk

import "fmt"

type DuploTenantK8Quota struct {
	Metadata *DuploTenantK8QuotaMetadata `json:"metadata"`
	Spec     *DuploTenantK8QuotaSpec     `json:"spec"`
}

type DuploTenantK8QuotaMetadata struct {
	Name string `json:"name"`
}

type DuploTenantK8QuotaSpec struct {
	Hard          map[string]interface{} `json:"hard"`
	ScopeSelector map[string]interface{} `json:"scopeSelector"`
}

func (c *Client) DuploTenantK8QuotaCreate(tenantId string, rq *DuploTenantK8Quota) ClientError {
	return c.postAPI(fmt.Sprintf("DuploTenantK8QuotaCreate(%s,%s)", tenantId, rq.Metadata.Name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/resourceQuota", tenantId),
		rq,
		nil,
	)
}

func (c *Client) DuploTenantK8QuotaGet(tenantId, name string) (*DuploTenantK8Quota, ClientError) {
	rp := []DuploTenantK8Quota{}
	err := c.getAPI(fmt.Sprintf("DuploTenantK8QuotaGet(%s,%s)", tenantId, name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/resourceQuota", tenantId),
		&rp,
	)
	if err != nil {
		return nil, err
	}
	for _, val := range rp {
		if val.Metadata.Name == name {
			return &val, nil
		}
	}
	return nil, NewCustomError("not found", 404)
}

func (c *Client) DuploTenantK8QuotaUpdate(tenantId string, rq *DuploTenantK8Quota) ClientError {
	return c.putAPI(fmt.Sprintf("DuploTenantK8QuotaUpdate(%s,%s)", tenantId, rq.Metadata.Name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/resourceQuota/%s", tenantId, rq.Metadata.Name),
		rq,
		nil,
	)
}

func (c *Client) DuploTenantK8QuotaDelete(tenantId, name string) ClientError {
	return c.deleteAPI(fmt.Sprintf("DuploTenantK8QuotaDelete(%s,%s)", tenantId, name),
		fmt.Sprintf("v3/admin/k8s/subscriptions/%s/resourceQuota/%s", tenantId, name),
		nil,
	)
}
