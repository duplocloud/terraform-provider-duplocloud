package duplosdk

import (
	"fmt"
	"strings"
	"time"
)

type (
	DataReplicationMode    string
	AuthenticationStrategy string
	EngineType             string
	DeploymentMode         string
	BrokerStorageType      string
	DayOfWeek              string
)
type DuploAWSMQ struct {
	EngineType                      EngineType                 `json:"EngineType"`                  //ACTIVEMQ, RABBITMQ
	DeploymentMode                  DeploymentMode             `json:"DeploymentMode,omitempty"`    //ACTIVE_STANDBY_MULTI_AZ, CLUSTER_MULTI_AZ, SINGLE_INSTANCE
	BrokerStorageType               BrokerStorageType          `json:"BrokerStorageType,omitempty"` //EBS, EFS
	BrokerName                      string                     `json:"BrokerName"`
	HostInstanceType                string                     `json:"HostInstanceType"`
	EngineVersion                   string                     `json:"EngineVersion"`
	AuthenticationStrategy          AuthenticationStrategy     `json:"AuthenticationStrategy"` //LDAP, SIMPLE
	AutoMinorVersionUpgrade         bool                       `json:"AutoMinorVersionUpgrade"`
	Users                           []DuploAWSMQUser           `json:"Users"`
	LdapServerMetadata              *DuploMQLDAPMetadata       `json:"LdapServerMetadata,omitempty"`
	Configuration                   *DuplocloudMQConfiguration `json:"Configuration,omitempty"`
	CreatorRequestId                string                     `json:"CreatorRequestId,omitempty"`                //make this field compute. add field is_app_idempotent, is set to true create uuid and set it to CreatorRequestId
	DataReplicationMode             DataReplicationMode        `json:"DataReplicationMode"`                       //CRDR, NONE
	DataReplicationPrimaryBrokerArn string                     `json:"DataReplicationPrimaryBrokerArn,omitempty"` // required when CRDR
	EncryptionOptions               *DuploMQEncryptionOptions  `json:"EncryptionOptions,omitempty"`
	Logs                            *DuploMQLogs               `json:"Logs"`
	MaintenanceWindow               *DuploMQMaintenanceWindow  `json:"MaintenanceWindowStartTime,omitempty"`
	PubliclyAccessible              bool                       `json:"PubliclyAccessible"`
	SecurityGroups                  []string                   `json:"SecurityGroups"`
	SubnetIds                       []string                   `json:"SubnetIds"`
	Tags                            map[string]string          `json:"Tags"`
}

type DuploAWSMQUser struct {
	UserName string   `json:"UserName"`
	Password string   `json:"Password"`
	Groups   []string `json:"Groups"`
}

type DuploMQLDAPMetadata struct {
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

type DuplocloudMQConfiguration struct {
	Id       string `json:"Id"`
	Revision int    `json:"Revision"`
}

type DuploMQEncryptionOptions struct {
	KmsKeyId       string `json:"KmsKeyId"`
	UseAwsOwnedKey bool   `json:"UseAwsOwnedKey"`
}

type DuploMQLogs struct {
	Audit   bool `json:"Audit"` //not aplicable for rabbit mq
	General bool `json:"General"`
}

type DuploMQMaintenanceWindow struct {
	TimeOfDay string    `json:"TimeOfDay"`
	TimeZone  string    `json:"TimeZone"`
	DayOfWeek DayOfWeek `json:"DayOfWeek"`
}

func (c *Client) DuploAWSMQBrokerCreate(tenantID string, rq *DuploAWSMQ) ClientError {
	var rp interface{}
	err := c.postAPI(
		fmt.Sprintf("DuploAWSMQBrokerCreate(%s, %s)", tenantID, rq.BrokerName),
		fmt.Sprintf("v3/subscriptions/%s/aws/mq/broker", tenantID),
		&rq,
		&rp,
	)
	return err
}

func (c *Client) DuploAWSMQBrokerDelete(tenantID string, brokerID string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploSnsTopicDelete(%s, %s)", tenantID, brokerID),
		fmt.Sprintf("v3/subscriptions/%s/aws/mq/broker/%s", tenantID, brokerID),
		nil,
	)
}

func (c *Client) DuploAWSMQBrokerGet(tenantID, brokerID string) (*DuploAWSMQ, ClientError) {
	rp := DuploAWSMQ{}
	err := c.getAPI(
		fmt.Sprintf("DuploAWSMQBrokerGet(%s,%s)", tenantID, brokerID),
		fmt.Sprintf("v3/subscriptions/%s/aws/mq/broker/%s", tenantID, brokerID),
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
