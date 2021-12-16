package duplosdk

import (
	"fmt"
)

type DuploAzureRedisCacheSku struct {
	Capacity int    `json:"capacity"`
	Family   string `json:"family"`
	Name     string `json:"name"`
}

type DuploAzureRedisCacheProperties struct {
	ShardCount       int                     `json:"shardCount,omitempty"`
	Sku              DuploAzureRedisCacheSku `json:"sku"`
	EnableNonSslPort bool                    `json:"enableNonSslPort"`
	SubnetID         string                  `json:"subnetId,omitempty"`
}

type DuploAzureRedisCacheRequest struct {
	Properties DuploAzureRedisCacheProperties `json:"properties,omitempty"`
}

type DuploAzureRedisCache struct {
	PropertiesRedisConfiguration struct {
		Maxclients                     string `json:"maxclients"`
		MaxmemoryReserved              string `json:"maxmemory-reserved"`
		MaxfragmentationmemoryReserved string `json:"maxfragmentationmemory-reserved"`
		MaxmemoryDelta                 string `json:"maxmemory-delta"`
	} `json:"properties.redisConfiguration"`
	PropertiesEnableNonSslPort bool `json:"properties.enableNonSslPort"`
	PropertiesShardCount       int  `json:"properties.shardCount"`
	PropertiesSku              struct {
		Name     string `json:"name"`
		Family   string `json:"family"`
		Capacity int    `json:"capacity"`
	} `json:"properties.sku"`
	PropertiesSubnetID          string        `json:"properties.subnetId"`
	PropertiesStaticIP          string        `json:"properties.staticIP"`
	PropertiesRedisVersion      string        `json:"properties.redisVersion"`
	PropertiesProvisioningState string        `json:"properties.provisioningState"`
	PropertiesHostName          string        `json:"properties.hostName"`
	PropertiesPort              int           `json:"properties.port"`
	PropertiesSslPort           int           `json:"properties.sslPort"`
	PropertiesLinkedServers     []interface{} `json:"properties.linkedServers"`
	PropertiesInstances         []struct {
		SslPort    int  `json:"sslPort"`
		NonSslPort int  `json:"nonSslPort"`
		ShardID    int  `json:"shardId"`
		IsMaster   bool `json:"isMaster"`
	} `json:"properties.instances"`
	Tags     map[string]interface{} `json:"tags"`
	Location string                 `json:"location"`
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
}

func (c *Client) RedisCacheCreate(tenantID string, name string, rq *DuploAzureRedisCacheRequest) ClientError {
	return c.postAPI(
		fmt.Sprintf("RedisCacheCreate(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/CreateRedisInstance/%s", tenantID, name),
		&rq,
		nil,
	)
}

func (c *Client) RedisCacheGet(tenantID string, name string) (*DuploAzureRedisCache, ClientError) {
	list, err := c.RedisCacheList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, redis := range *list {
			if redis.Name == name {
				return &redis, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) RedisCacheList(tenantID string) (*[]DuploAzureRedisCache, ClientError) {
	rp := []DuploAzureRedisCache{}
	err := c.getAPI(
		fmt.Sprintf("RedisCacheList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/ListRedisInstances", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) RedisCacheExists(tenantID, name string) (bool, ClientError) {
	list, err := c.RedisCacheList(tenantID)
	if err != nil {
		return false, err
	}

	if list != nil {
		for _, redis := range *list {
			if redis.Name == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func (c *Client) RedisCacheDelete(tenantID string, name string) ClientError {
	return c.postAPI(
		fmt.Sprintf("RedisCacheDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/DeleteRedisInstance/%s", tenantID, name),
		nil,
		nil,
	)
}
