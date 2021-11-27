package duplosdk

import "fmt"

func (c *Client) SystemSettingCreate(rq *DuploCustomDataEx) ClientError {
	rp := DuploCustomDataEx{}
	return c.postAPI(
		fmt.Sprintf("SystemSettingCreate(%s)", rq.Key),
		fmt.Sprintf("v3/admin/systemSettings/config"),
		&rq,
		&rp,
	)
}

func (c *Client) SystemSettingGet(key string) (*DuploCustomDataEx, ClientError) {

	list, err := c.SystemSettingList()

	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, setting := range *list {
			if setting.Key == key {
				return &setting, nil
			}
		}
	}
	return nil, nil

}

func (c *Client) SystemSettingList() (*[]DuploCustomDataEx, ClientError) {
	rp := []DuploCustomDataEx{}
	err := c.getAPI(
		fmt.Sprintf("SystemSettingList"),
		fmt.Sprintf("v3/admin/systemSettings/config"),
		&rp,
	)
	return &rp, err
}

func (c *Client) SystemSettingDelete(key string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("SystemSettingDelete(%s)", key),
		fmt.Sprintf("v3/admin/systemSettings/config/%s", key),
		nil,
	)
}
