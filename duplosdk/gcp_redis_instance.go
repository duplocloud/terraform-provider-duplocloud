package duplosdk

import (
	"fmt"
	"log"
)

type DuploRedisInstanceBody struct {
	Name                     string            `json:"Name"`
	DisplayName              string            `json:"DisplayName"`
	ReadReplicasEnabled      bool              `json:"ReadReplicasEnabled"`
	RedisVersion             string            `json:"RedisVersion"`
	RedisConfigs             map[string]string `json:"RedisConfigs"`
	ReplicaCount             int               `json:"ReplicaCount"`
	MemorySizeGb             int               `json:"MemorySizeGb"`
	AuthEnabled              bool              `json:"AuthEnabled"`
	TransitEncryptionEnabled bool              `json:"TransitEncryptionEnabled"`
	Tier                     int               `json:"Tier"`
	Labels                   map[string]string `json:"Labels"`
	Port                     int               `json:"Port,omitempty"`
	SelfLink                 string            `json:"SelfLink,omitempty"`
	Status                   string            `json:"Status,omitempty"`
	ResourceType             int               `json:"ResourceType,omitempty"`
}

func (c *Client) RedisInstanceCreate(tenantID string, rq *DuploRedisInstanceBody) (*DuploRedisInstanceBody, ClientError) {
	log.Printf("[TRACE] \nRedisInstance request \n\n ******%+v\n*******", rq)
	resp := DuploRedisInstanceBody{}
	err := c.postAPI(
		fmt.Sprintf("RedisInstanceCreate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/redis", tenantID),
		&rq,
		&resp,
	)
	return &resp, err
}

func (c *Client) RedisInstanceGet(tenantID string, fullname string) (*DuploRedisInstanceBody, ClientError) {
	rp := DuploRedisInstanceBody{}
	err := c.getAPI(
		fmt.Sprintf("RedisInstanceGet(%s, %s)", tenantID, fullname),
		fmt.Sprintf("v3/subscriptions/%s/google/redis/%s", tenantID, fullname),
		&rp,
	)
	return &rp, err
}

func (c *Client) RedisInstanceList(tenantID string) (*[]DuploRedisInstanceBody, ClientError) {
	rp := []DuploRedisInstanceBody{}
	err := c.getAPI(
		fmt.Sprintf("RedisInstanceList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/redis", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) RedisInstanceDelete(tenantID, fullname string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("RedisInstanceDelete(%s, %s)", tenantID, fullname),
		fmt.Sprintf("v3/subscriptions/%s/google/redis/%s", tenantID, fullname),
		nil)
}

func (c *Client) RedisInstanceUpdate(tenantID, fullname string, rq *DuploRedisInstanceBody) (*DuploRedisInstanceBody, ClientError) {
	rp := DuploRedisInstanceBody{}
	err := c.putAPI(
		fmt.Sprintf("RedisInstanceUpdate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/redis/%s", tenantID, fullname),
		&rq,
		&rp,
	)
	return &rp, err
}
