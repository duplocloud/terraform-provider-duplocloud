package duplosdk

import (
	"fmt"
)

type DuploTargetGroupTargetRegister struct {
	TargetGroupARN string          `json:"TargetGroupArn"`
	Targets        []DuploTargetId `json:"Targets,omitempty"`
}

type DuploTargetId struct {
	Id               string `json:"Id"`
	AvailabilityZone string `json:"-"`
	Port             int    `json:"-"`
}

func (c *Client) DuploAwsTargetGroupTargetCreate(tenantID, name string, rq *DuploTargetGroupTargetRegister) ClientError {
	return c.postAPI(
		fmt.Sprintf("DuploAwsTargetGroupTargetCreate(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/target/%s", tenantID, name),
		&rq,
		nil,
	)
}

func (c *Client) DuploAwsTargetGroupTargetGet(tenantID, name string) (*DuploTargetGroupTargetRegister, ClientError) {
	obj := DuploTargetGroupTargetRegister{}
	rp := []interface{}{}
	err := c.getAPI(
		fmt.Sprintf("DuploAwsTargetGroupTargetGet(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/target/%s", tenantID, name),
		&rp,
	)
	if err == nil {
		for _, v := range rp {
			m := v.(map[string]interface{})
			tId := DuploTargetId{}
			if v, ok := m["Id"]; ok {
				tId.Id = v.(string)
			}
			if v, ok := m["AvailabilityZone"]; ok {
				tId.AvailabilityZone = v.(string)
			}
			if v, ok := m["Port"]; ok {
				tId.Port = int(v.(float64))
			}
			obj.Targets = append(obj.Targets, tId)
		}
	}

	return &obj, err
}

func (c *Client) DuploAwsTargetGroupTargetDelete(tenantID, name string, rq *DuploTargetGroupTargetRegister) ClientError {
	return c.putAPI(
		fmt.Sprintf("DuploAwsTargetGroupTargetCreate(%s,%s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/target/%s", tenantID, name),
		&rq,
		nil,
	)
}
