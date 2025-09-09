package duplosdk

import (
	"fmt"
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
	Logs                            *DuploMQLogs               `json:"logs,omitempty"`
	MaintenanceWindow               *DuploMQMaintenanceWindow  `json:"maintenanceWindowStartTime,omitempty"`
	PubliclyAccessible              bool                       `json:"publiclyAccessible"`
	SecurityGroups                  []string                   `json:"securityGroups"`
	SubnetIds                       []string                   `json:"subnetIds"`
	Tags                            map[string]string          `json:"tags"`
	BrokerId                        string                     `json:"name"`
}

type DuploAWSMQUser struct {
	UserName        string   `json:"username"`
	Password        string   `json:"password"`
	Groups          []string `json:"groups"`
	ConsoleAccess   bool     `json:"consoleAccess"`   // for rabbitmq
	ReplicationUser bool     `json:"replicationUser"` // for rabbitmq
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
	Audit   *bool `json:"Audit,omitempty"` //not aplicable for rabbit mq
	General bool  `json:"General"`
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

func (c *Client) DuploAWSMQBrokerReeboot(tenantID, brokerId, name string) ClientError {
	err := c.postAPI(
		fmt.Sprintf("DuploAWSMQBrokerReeboot(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/mq/broker/%s/reboot", tenantID, brokerId),
		nil,
		nil,
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

func (c *Client) DuploAWSMQBrokerGet(tenantID, brokerID string) (*DuploMQBrokerResponse, ClientError) {
	rp := DuploMQBrokerResponse{}
	err := c.getAPI(
		fmt.Sprintf("DuploAWSMQBrokerGet(%s,%s)", tenantID, brokerID),
		fmt.Sprintf("v3/subscriptions/%s/aws/mq/broker/%s", tenantID, brokerID),
		&rp,
	)
	return &rp, err
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
		Audit           *bool  `json:"Audit,omitempty"`
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
	SubnetIds   []string          `json:"SubnetIds"`
	Tags        map[string]string `json:"Tags"`
	Users       []DuploAWSMQUser  `json:"Users"`
	BrokerArn   string            `json:"BrokerArn"`
	BrokerId    string            `json:"BrokerId"`
	BrokerName  string            `json:"BrokerName"`
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

type DuploAWSMQBrokerUpdateRequest struct {
	AuthenticationStrategy  AuthenticationStrategy     `json:"authenticationStrategy,omitempty"`
	AutoMinorVersionUpgrade bool                       `json:"autoMinorVersionUpgrade,omitempty"`
	BrokerId                string                     `json:"brokerId"`
	Configuration           *DuplocloudMQConfiguration `json:"configuration,omitempty"`
	DataReplicationMode     DataReplicationMode        `json:"dataReplicationMode,omitempty"`
	EngineVersion           string                     `json:"engineVersion,omitempty"`
	HostInstanceType        string                     `json:"hostInstanceType,omitempty"`
	LdapServerMetadata      *DuploMQLDAPMetadata       `json:"ldapServerMetadata,omitempty"`
	Logs                    *DuploMQLogs               `json:"logs,omitempty"`
	MaintenanceWindow       *DuploMQMaintenanceWindow  `json:"maintenanceWindowStartTime,omitempty"`
	SecurityGroups          []string                   `json:"securityGroups,omitempty"`
	Tags                    map[string]string          `json:"Tags"`
}

func (c *Client) DuploAWSMQBrokerUpdate(tenantID, brokerID string, rq DuploAWSMQBrokerUpdateRequest) ClientError {
	var rp interface{}
	err := c.putAPI(
		fmt.Sprintf("DuploAWSMQBrokerUpdate(%s,%s)", tenantID, brokerID),
		fmt.Sprintf("v3/subscriptions/%s/aws/mq/broker/%s", tenantID, brokerID), &rq,
		&rp,
	)
	return err
}

type DuploAwsMQConfig struct {
	AuthenticationStrategy string                 `json:"AuthenticationStrategy"`
	EngineType             string                 `json:"EngineType"`
	EngineVersion          string                 `json:"EngineVersion"`
	Name                   string                 `json:"Name"`
	Tags                   map[string]interface{} `json:"Tags"`
}

type DuploAwsMQConfigResponse struct {
	AuthenticationStrategy DuploStringValue       `json:"AuthenticationStrategy"`
	EngineType             DuploStringValue       `json:"EngineType"`
	EngineVersion          string                 `json:"EngineVersion"`
	Name                   string                 `json:"Name"`
	Tags                   map[string]interface{} `json:"Tags"`
	ConfigId               string                 `json:"Id"`
	Arn                    string                 `json:"Arn"`
	LatestRevision         struct {
		Description string `json:"Description"`
		Revision    int    `json:"Revision"`
	} `json:"LatestRevision"`
}

type DuploAwsMQConfigUpdate struct {
	ConfigurationId string `json:"ConfigurationId"`
	Data            string `json:"Data"`
	Description     string `json:"EngineVersion"`
}

func (c *Client) DuploAWSMQConfigUpdate(tenantID, brokerID string, rq DuploAwsMQConfigUpdate) ClientError {
	var rp interface{}
	err := c.putAPI(
		fmt.Sprintf("DuploAWSMQConfgiUpdate(%s,%s)", tenantID, rq.ConfigurationId),
		fmt.Sprintf("v3/subscriptions/%s/aws/mq/broker/config/%s", tenantID, rq.ConfigurationId),
		&rq,
		&rp)
	return err
}

func (c *Client) DuploAWSMQConfigCreate(tenantID string, rq DuploAwsMQConfig) (map[string]interface{}, ClientError) {
	rp := make(map[string]interface{})
	err := c.postAPI(
		fmt.Sprintf("DuploAWSMQConfigCreate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/mq/broker/config", tenantID),
		&rq,
		&rp,
	)
	return rp, err
}

func (c *Client) DuploAWSMQConfigGet(tenantID, cID string) (*DuploAwsMQConfigResponse, ClientError) {
	rp := DuploAwsMQConfigResponse{}
	err := c.getAPI(
		fmt.Sprintf("DuploAWSMQConfigGet(%s,%s)", tenantID, cID),
		fmt.Sprintf("v3/subscriptions/%s/aws/mq/broker/config/%s", tenantID, cID),
		&rp,
	)

	return &rp, err
}
