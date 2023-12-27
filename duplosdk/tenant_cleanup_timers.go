package duplosdk

import "fmt"

type DuploTenantCleanUpTimers struct {
	ExpiryTime string `json:"ExpiryTime,omitempty"`
	PauseTime  string `json:"PauseTime,omitempty"`
}

type DuploTenantCleanUpTimersUpdateRequest struct {
	TenantId         string `json:"TenantId,omitempty"`
	ExpiryTime       string `json:"ExpiryTime,omitempty"`
	PauseTime        string `json:"PauseTime,omitempty"`
	RemoveExpiryTime bool   `json:"RemoveExpiryTime,omitempty"`
	RemovePauseTime  bool   `json:"RemovePauseTime,omitempty"`
}

// GetTenantCleanUpTimers gets the expiry of a tenant. It uses the Tenant API.
func (c *Client) GetTenantCleanUpTimers(tenantId string) (*DuploTenantCleanUpTimers, ClientError) {
	tenant, err := c.TenantGet(tenantId)
	if err != nil {
		return nil, err
	}

	tenantExpiry := DuploTenantCleanUpTimers{
		ExpiryTime: tenant.Expiry,
		PauseTime:  tenant.PauseTime,
	}

	return &tenantExpiry, nil
}

// UpdateTenantCleanUpTimers updates the clean-up timers of a tenant.
func (c *Client) UpdateTenantCleanUpTimers(expiry *DuploTenantCleanUpTimersUpdateRequest) ClientError {
	apiName := fmt.Sprintf("UpdateTenantCleanUpTimers(%s)", expiry.TenantId)
	return c.postAPI(apiName, "adminproxy/UpdateTenantCleanupTimers", expiry, nil)
}
