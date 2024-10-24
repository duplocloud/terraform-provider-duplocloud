package duplosdk

import (
	"fmt"
	"net/url"
)

type DuploUser struct {
	Username                string    `json:"Username,omitempty"`
	Roles                   *[]string `json:"Roles,omitempty"`
	ReallocateVpnAddress    bool      `json:"ReallocateVpnAddress"`
	IsReadOnly              bool      `json:"IsReadOnly"`
	RegenerateVpnPassword   bool      `json:"RegenerateVpnPassword,omitempty"`
	CurrentSessionToken     string    `json:"CurrentSessionToken,omitempty"`
	VpnStaticIp             string    `json:"VpnStaticIp,omitempty"`
	IsVpnConfigCreated      bool      `json:"IsVpnConfigCreated,omitempty"`
	IsConfirmationEmailSent bool      `json:"IsConfirmationEmailSent,omitempty"`
	State                   string    `json:"State,omitempty"`
}

func (c *Client) UserGet(userName string) (*DuploUser, ClientError) {
	list, err := c.UserList()
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, user := range *list {
			if user.Username == userName {
				return &user, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) UserExists(userName string) (bool, ClientError) {
	list, err := c.UserList()
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, user := range *list {
			if user.Username == userName {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) UserList() (*[]DuploUser, ClientError) {
	rp := []DuploUser{}
	err := c.getAPI("UserList", "admin/GetAllUserRoles", &rp)
	return &rp, err
}

func (c *Client) UserCreate(rq DuploUser) (*DuploUser, ClientError) {
	rp := DuploUser{}
	err := c.postAPI(fmt.Sprintf("UserCreate(%s)", rq.Username), "admin/UpdateUserRole", &rq, &rp)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

func (c *Client) UserDelete(userName string) ClientError {
	rq := DuploUser{
		Username: userName,
		State:    "deleted",
	}
	return c.postAPI(fmt.Sprintf("UserDelete(%s)", userName), "admin/UpdateUserRole", rq, nil)
}

func (c *Client) UserInfo() (*DuploUser, ClientError) {
	rp := DuploUser{}
	err := c.getAPI("UserList", "admin/GetUserRoleInfo", &rp)

	return &rp, err
}

type DuploUserTenantAccess struct {
	Username string                  `json:"Username"`
	Policy   DuploTenantAccessPolicy `json:"Policy,omitempty"`
	TenantId string                  `json:"TenantId"`
	State    string                  `json:"State,omitempty"`
}

type DuploTenantAccessPolicy struct {
	IsReadOnly bool `json:"IsReadOnly"`
}

type DuploUserTenantAccessResponse struct {
	AccountName string                  `json:"AccountName"`
	Policy      DuploTenantAccessPolicy `json:"Policy,omitempty"`
	TenantId    string                  `json:"TenantId"`
}

func (c *Client) GrantUserTenantAccess(rq *DuploUserTenantAccess) ClientError {

	return c.postAPI(fmt.Sprintf("GrantUserTenantAccess(%s)", rq.Username), "admin/UpdateUserAccess", rq, nil)
}

func (c *Client) GetUserTenantAccessInfo(userName, tenantId string) (*DuploUserTenantAccessResponse, ClientError) {
	rp := []DuploUserTenantAccessResponse{}
	err := c.getAPI("GetUserTenantAccessInfo", fmt.Sprintf("v3/admin/user/%s/tenantAccess", url.PathEscape(userName)), &rp)
	for _, d := range rp {
		if d.TenantId == tenantId {
			return &d, nil
		}
	}
	return nil, err
}
