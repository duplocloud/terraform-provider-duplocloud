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
	EngineType                      EngineType                 `json:"engineType"`               //ACTIVEMQ, RABBITMQ
	DeploymentMode                  DeploymentMode             `json:"deploymentMode,omitempty"` //ACTIVE_STANDBY_MULTI_AZ, CLUSTER_MULTI_AZ, SINGLE_INSTANCE
	BrokerStorageType               BrokerStorageType          `json:"storageType,omitempty"`    //EBS, EFS
	BrokerName                      string                     `json:"brokerName"`
	HostInstanceType                string                     `json:"hostInstanceType"`
	EngineVersion                   string                     `json:"engineVersion"`
	AuthenticationStrategy          AuthenticationStrategy     `json:"authenticationStrategy"` //LDAP, SIMPLE
	AutoMinorVersionUpgrade         bool                       `json:"autoMinorVersionUpgrade"`
	Users                           []DuploAWSMQUser           `json:"users"`
	LdapServerMetadata              *DuploMQLDAPMetadata       `json:"ldapServerMetadata,omitempty"`
	Configuration                   *DuplocloudMQConfiguration `json:"configuration,omitempty"`
	CreatorRequestId                string                     `json:"creatorRequestId,omitempty"`                //make this field compute. add field is_app_idempotent, is set to true create uuid and set it to CreatorRequestId
	DataReplicationMode             DataReplicationMode        `json:"dataReplicationMode"`                       //CRDR, NONE
	DataReplicationPrimaryBrokerArn string                     `json:"dataReplicationPrimaryBrokerArn,omitempty"` // required when CRDR
	EncryptionOptions               *DuploMQEncryptionOptions  `json:"encryptionOptions,omitempty"`
	Logs                            *DuploMQLogs               `json:"logs"`
	MaintenanceWindow               *DuploMQMaintenanceWindow  `json:"maintenanceWindowStartTime,omitempty"`
	PubliclyAccessible              bool                       `json:"publiclyAccessible"`
	SecurityGroups                  []string                   `json:"securityGroups"`
	SubnetIds                       []string                   `json:"subnetIds"`
	Tags                            map[string]string          `json:"tags"`
	BrokerId                        string                     `json:"name"`
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

func (c *Client) DuploAWSMQBrokerCreate(tenantID string, rq *DuploAWSMQ) (*DuploMQBrokerResponse, ClientError) {
	rp := DuploMQBrokerResponse{}
	err := c.postAPI(
		fmt.Sprintf("DuploAWSMQBrokerCreate(%s, %s)", tenantID, rq.BrokerName),
		fmt.Sprintf("v3/subscriptions/%s/aws/mq/broker", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploAWSMQBrokerDelete(tenantID string, brokerID string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploSnsTopicDelete(%s, %s)", tenantID, brokerID),
		fmt.Sprintf("v3/subscriptions/%s/aws/mq/broker/%s", tenantID, brokerID),
		nil,
	)
}

func (c *Client) DuploAWSMQBrokerGet(tenantID, brokerID string) (*DuploMQBrokerResponse, ClientError) {
	rp := DuploMQBrokerResponse{}
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

type DuploMQBrokerResponse struct {
	AutoMinorVersionUpgrade bool     `json:"AutoMinorVersionUpgrade"`
	ActionsRequired         []string `json:"ActionsRequired"`
	AuthenticationStrategy  struct {
		Value string `json:"Value"`
	} `json:"AuthenticationStrategy"`
	BrokerInstances []interface{} `json:"BrokerInstances"`
	Configurations  struct {
		History []interface{} `json:"History"`
	} `json:"Configurations"`
	EncryptionOptions *DuploMQEncryptionOptions `json:"EncryptionOptions"`
	EngineVersion     string                    `json:"EngineVersion"`
	Logs              struct {
		Audit           bool   `json:"Audit"`
		AuditLogGroup   string `json:"AuditLogGroup"`
		General         bool   `json:"General"`
		GeneralLogGroup string `json:"GeneralLogGroup"`
	} `json:"Logs"`
	MaintenanceWindowStartTime struct {
		DayOfWeek struct {
			Value string `json:"Value"`
		} `json:"DayOfWeek"`
		TimeOfDay string `json:"TimeOfDay"`
		TimeZone  string `json:"TimeZone"`
	} `json:"MaintenanceWindowStartTime"`
	PendingSecurityGroups []string `json:"PendingSecurityGroups"`
	PubliclyAccessible    bool     `json:"PubliclyAccessible"`
	SecurityGroups        []string `json:"SecurityGroups"`
	StorageType           struct {
		Value string `json:"Value"`
	} `json:"StorageType"`
	SubnetIds []string          `json:"SubnetIds"`
	Tags      map[string]string `json:"Tags"`
	Users     []struct {
		PendingChange struct {
			Value string `json:"Value"`
		} `json:"PendingChange"`
		Username string   `json:"Username"`
		Password string   `json:"Password"`
		Groups   []string `json:"Groups"`
	} `json:"Users"`
	BrokerArn   string `json:"BrokerArn"`
	BrokerId    string `json:"BrokerId"`
	BrokerName  string `json:"BrokerName"`
	BrokerState struct {
		Value string `json:"Value"`
	} `json:"BrokerState"`
	Created        string `json:"Created"`
	DeploymentMode struct {
		Value string `json:"Value"`
	} `json:"DeploymentMode"`
	EngineType struct {
		Value string `json:"Value"`
	} `json:"EngineType"`
	HostInstanceType string `json:"HostInstanceType"`
	ResourceType     int    `json:"ResourceType"`
	Name             string `json:"Name"`
}
