package duplosdk

import (
	"fmt"
	"strings"
	"time"
)

type DuploAWSMQ struct {
	EngineType              string           `json:"EngineType"`
	DeploymentMode          string           `json:"DeploymentMode,omitempty"`
	BrokerStorageType       string           `json:"BrokerStorageType,omitempty"`
	BrokerName              string           `json:"BrokerName"`
	HostInstanceType        string           `json:"HostInstanceType"`
	EngineVersion           string           `json:"EngineVersion"`
	AuthenticationStrategy  string           `json:"AuthenticationStrategy"` //LDAP, SIMPLE
	AutoMinorVersionUpgrade bool             `json:"AutoMinorVersionUpgrade"`
	Users                   []DuploAWSMQUser `json:"Users"`
	LdapServerMetadata      LDAPMetadata     `json:"LdapServerMetadata"`
	Configuration           MQConfiguration  `json:"Configuration"`
}

type DuploAWSMQUser struct {
	UserName string   `json:"UserName"`
	Password string   `json:"Password"`
	Groups   []string `json:"Groups"`
}

type LDAPMetadata struct {
	Hosts                  []string `json:"Hosts"`
	RoleBase               string   `json:"RoleBase"`
	RoleName               string   `json:"RoleName"`
	RoleSearchMatching     string   `json:"RoleSearchMatching"`
	RoleSearchSubtree      bool     `json:"RoleSearchSubtree"`
	ServiceAccountPassword string   `json:"ServiceAccountPassword"`
	ServiceAccountUsername string   `json:"ServiceAccountUsername"`
	UserBase               string   `json:"UserBase"`
	UserRoleName           string   `json:"UserRoleName"`
	UserSearchMatching     string   `json:"UserSearchMatching"`
	UserSearchSubtree      bool     `json:"UserSearchSubtree"`
}

type MQConfiguration struct {
	Id       string `json:"Id"`
	Revision int    `json:"Revision"`
}

func (c *Client) DuploSnsTopicCreate(tenantID string, rq *DuploSnsTopic) (*DuploSnsTopicResource, ClientError) {
	rp := &DuploSnsTopicResource{}
	err := c.postAPI(
		fmt.Sprintf("DuploSnsTopicCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/snsTopic", tenantID),
		&rq,
		&rp,
	)
	return rp, err
}

func (c *Client) DuploSnsTopicDelete(tenantID string, arn string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploSnsTopicDelete(%s, %s)", tenantID, arn),
		fmt.Sprintf("v3/subscriptions/%s/aws/snsTopic/%s", tenantID, arn),
		nil,
	)
}

func (c *Client) TenantListSnsTopic(tenantID string) (*[]DuploSnsTopicResource, ClientError) {
	rp := []DuploSnsTopicResource{}
	err := c.getAPI(
		fmt.Sprintf("TenantListSnsTopic(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/snsTopic", tenantID),
		&rp,
	)
	return &rp, err
}

func (c *Client) TenantGetSnsTopicAttributes(tenantID string, topicArn string) (*DuploSnsTopicAttributes, ClientError) {
	rp := DuploSnsTopicAttributes{}
	_, err := RetryWithExponentialBackoff(func() (interface{}, ClientError) {
		err := c.getAPI(
			fmt.Sprintf("TenantListSnsTopicAttributes(%s)", tenantID),
			fmt.Sprintf("v3/subscriptions/%s/aws/snsTopic/%s/attributes", tenantID, topicArn),
			&rp,
		)
		return &rp, err
	},
		RetryConfig{
			MinDelay:  1 * time.Second,
			MaxDelay:  5 * time.Second,
			MaxJitter: 2000,
			Timeout:   60 * time.Second,
			IsRetryable: func(error ClientError) bool {
				return error.Status() == 400 || strings.Contains(error.Error(), "context deadline exceeded")
			},
		})

	return &rp, err
}

func (c *Client) TenantGetSnsTopic(tenantID string, arn string) (*DuploSnsTopicResource, ClientError) {
	list, err := c.TenantListSnsTopic(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, topic := range *list {
			if topic.Name == arn {
				return &topic, nil
			}
		}
	}
	return nil, nil
}
