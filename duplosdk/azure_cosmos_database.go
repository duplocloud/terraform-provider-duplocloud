package duplosdk

import (
	"fmt"
)

type DuploAzureCosmosDBRequest struct {
	Name       string                                    `json:"name"`
	Kind       string                                    `json:"kind"`
	Identity   *DuploAzureCosmosDBManagedServiceIdentity `json:"identity"`
	Properties *DuploAzureCosmosDBProperties             `json:"properties"`
	Location   string                                    `json:"location"`
}

type DuploAzureCosmosDBProperties struct {
	ConsistencyPolicy                  *DuploAzureCosmosDBConsistencyPolicy              `json:"consistencyPolicy"`
	Locations                          []string                                          `json:"locations"`
	IpRules                            []string                                          `json:"ipRules"`
	IsVirtualNetworkFilterEnabled      bool                                              `json:"isVirtualNetworkFilterEnabled"`
	EnableAutomaticFailover            bool                                              `json:"enableAutomaticFailover"`
	Capabilities                       *[]DuploAzureCosmosDBCapability                   `json:"capabilities"`
	VirtualNetworkRules                *[]DuploAzureCosmosDBVirtualNetworkRule           `json:"virtualNetworkRules"`
	EnableMultipleWriteLocations       bool                                              `json:"enableMultipleWriteLocations"`
	EnableCassandraConnector           bool                                              `json:"enableCassandraConnector"`
	ConnectorOffer                     string                                            `json:"connectorOffer"`
	DisableKeyBasedMetadataWriteAccess bool                                              `json:"disableKeyBasedMetadataWriteAccess"`
	KeyVaultKeyUri                     string                                            `json:"keyVaultKeyUri"`
	DefaultIdentity                    string                                            `json:"defaultIdentity"`
	PublicNetworkAccess                string                                            `json:"publicNetworkAccess"`
	EnableFreeTier                     bool                                              `json:"enableFreeTier"`
	ApiProperties                      *DuploAzureCosmosDBApiProperties                  `json:"apiProperties"`
	EnableAnalyticalStorage            bool                                              `json:"enableAnalyticalStorage"`
	AnalyticalStorageConfiguration     *DuploAzureCosmosDBAnalyticalStorageConfiguration `json:"analyticalStorageConfiguration"`
	CreateMode                         string                                            `json:"createMode"`
	BackupPolicy                       *DuploAzureCosmosDBBackupPolicy                   `json:"backupPolicy"`
	Cors                               *DuploAzureCosmosDBCorsPolicy                     `json:"cors"`
	NetworkAclBypass                   string                                            `json:"networkAclBypass"` //None AzureServices
	NetworkAclBypassResourceIds        []string                                          `json:"networkAclBypassResourceIds"`
	DisableLocalAuth                   bool                                              `json:"disableLocalAuth"`
	RestoreParameters                  *DuploAzureCosmosDBRestoreParameters              `json:"restoreParameters"`
	Capacity                           *DuploAzureCosmosDBCapacity                       `json:"capacity"`
	DatabaseAccountOfferType           string                                            `json:"databaseAccountOff"`
}

type DuploAzureCosmosDBResponse struct {
	Name                               string                                            `json:"name"`
	Kind                               string                                            `json:"kind"`
	Identity                           *DuploAzureCosmosDBManagedServiceIdentity         `json:"identity"`
	ConsistencyPolicy                  *DuploAzureCosmosDBConsistencyPolicy              `json:"properties.consistencyPolicy"`
	Locations                          []string                                          `json:"properties.locations"`
	IpRules                            []string                                          `json:"properties.ipRules"`
	IsVirtualNetworkFilterEnabled      bool                                              `json:"properties.isVirtualNetworkFilterEnabled"`
	EnableAutomaticFailover            bool                                              `json:"properties.enableAutomaticFailover"`
	Capabilities                       *[]DuploAzureCosmosDBCapability                   `json:"properties.capabilities"`
	VirtualNetworkRules                *[]DuploAzureCosmosDBVirtualNetworkRule           `json:"properties.virtualNetworkRules"`
	EnableMultipleWriteLocations       bool                                              `json:"properties.enableMultipleWriteLocations"`
	EnableCassandraConnector           bool                                              `json:"properties.enableCassandraConnector"`
	ConnectorOffer                     string                                            `json:"properties.connectorOffer"`
	DisableKeyBasedMetadataWriteAccess bool                                              `json:"properties.disableKeyBasedMetadataWriteAccess"`
	KeyVaultKeyUri                     string                                            `json:"properties.keyVaultKeyUri"`
	DefaultIdentity                    string                                            `json:"properties.defaultIdentity"`
	PublicNetworkAccess                string                                            `json:"properties.publicNetworkAccess"`
	EnableFreeTier                     bool                                              `json:"properties.enableFreeTier"`
	ApiProperties                      *DuploAzureCosmosDBApiProperties                  `json:"properties.apiProperties"`
	EnableAnalyticalStorage            bool                                              `json:"properties.enableAnalyticalStorage"`
	AnalyticalStorageConfiguration     *DuploAzureCosmosDBAnalyticalStorageConfiguration `json:"properties.analyticalStorageConfiguration"`
	CreateMode                         string                                            `json:"properties.createMode"`
	BackupPolicy                       *DuploAzureCosmosDBBackupPolicy                   `json:"properties.backupPolicy"`
	Cors                               *DuploAzureCosmosDBCorsPolicy                     `json:"properties.cors"`
	NetworkAclBypass                   string                                            `json:"properties.networkAclBypass"` //None AzureServices
	NetworkAclBypassResourceIds        []string                                          `json:"properties.networkAclBypassResourceIds"`
	DisableLocalAuth                   bool                                              `json:"properties.disableLocalAuth"`
	RestoreParameters                  *DuploAzureCosmosDBRestoreParameters              `json:"properties.restoreParameters"`
	Capacity                           *DuploAzureCosmosDBCapacity                       `json:"properties.capacity"`
	DatabaseAccountOfferType           string                                            `json:"properties.databaseAccountOfferType"`
	Location                           string                                            `json:"location"`
}

type DuploAzureCosmosDBCapacity struct {
	TotalThroughputLimit int `json:"totalThroughputLimit"`
}

type DuploAzureCosmosDBRestoreParameters struct {
	RestoreMode        string                            `json:"restoreMode"`
	TablesToRestore    []string                          `json:"tablesToRestore"`
	DatabasesToRestore DuploAzureDatabaseRestoreResource `json:"databasesToRestore"`
}

type DuploAzureDatabaseRestoreResource struct {
	DatabaseName    string   `json:"databaseName"`
	CollectionNames []string `json:"collectionNames"`
}

type DuploAzureCosmosDBCorsPolicy struct {
	AllowedOrigins  string  `json:"allowedOrigins"`
	AllowedMethods  string  `json:"allowedMethods"`
	AllowedHeaders  string  `json:"allowedHeaders"`
	ExposedHeaders  string  `json:"exposedHeaders"`
	MaxAgeInSeconds float64 `json:"maxAgeInSeconds"`
}

type DuploAzureCosmosDBBackupPolicy struct {
	BackupPolicyMigrationState DuploAzureCosmosDBBackupPolicyMigrationState `json:"migrationState"`
}

type DuploAzureCosmosDBBackupPolicyMigrationState struct {
	Status     string `json:"Status"`
	TargetType string `json:"targetType"`
	StartTime  string `json:"startTime"`
}

type DuploAzureCosmosDBAnalyticalStorageConfiguration struct {
	SchemaType string `json:"schemaType"`
}

type DuploAzureCosmosDBApiProperties struct {
	ServerVersion string `json:"serverVersion"`
}

type DuploAzureCosmosDBVirtualNetworkRule struct {
	Id                               string `json:"id"`
	IgnoreMissingVNetServiceEndpoint bool   `json:"ignoreMissingVNetServiceEndpoint"`
}

type DuploAzureCosmosDBCapability struct {
	Name string `json:"name"`
}

type DuploAzureCosmosDBConsistencyPolicy struct {
	MaxStalenessPrefix      int         `json:"maxStalenessPrefix"`
	MaxIntervalInSeconds    int         `json:"maxIntervalInSeconds"`
	DefaultConsistencyLevel interface{} `json:"defaultConsistencyLevel"` // ENUM: Eventual,Session,BoundedStaleness,Strong,
}

type DuploAzureCosmosDBManagedServiceIdentity struct {
	PrincipalId            string                                                                    `json:"principalId"`
	TenantId               string                                                                    `json:"tenantId"`
	ResourceIdentityType   string                                                                    `json:"type"` //Enum: SystemAssigned,UserAssigned,SystemAssignedUserAssigned,None
	UserAssignedIdentities map[string]DuploAzureCosmosDBManagedServiceIdentityUserAssignedIdentities `json:"userAssignedIdentities"`
}

type DuploAzureCosmosDBManagedServiceIdentityUserAssignedIdentities struct {
	PrincipalId string `json:"principalId"`
	ClientId    string `json:"clientId"`
}

func (c *Client) CreateCosmosDB(tenantId string, account string, rq DuploAzureCosmosDB) ClientError {
	rp := ""
	err := c.postAPI(fmt.Sprintf("CreateCosmosDB(%s)", tenantId),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/cosmosDb/accounts/%s/databases", tenantId, account),
		&rq,
		&rp)
	if err != nil {
		return err
	}
	fmt.Println(rp)
	return nil
}

func (c *Client) GetCosmosDB(tenantId, account, name string) (*DuploAzureCosmosDB, ClientError) {
	rp := DuploAzureCosmosDB{}
	err := c.getAPI(fmt.Sprintf("GetCosmosDB(%s,%s,%s)", tenantId, account, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/cosmosDb/accounts/%s/databases/%s", tenantId, account, name),
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

type DuploAzureCosmosDBAccount struct {
	Name                               string                                 `json:"name"`
	Kind                               string                                 `json:"kind"`
	AccountType                        string                                 `json:"type"`
	ConsistencyPolicy                  *DuploAzureCosmosDBConsistencyPolicy   `json:"consistencyPolicy,omitempty"`
	Capabilities                       *[]DuploAzureCosmosDBCapability        `json:"-"` //Capabilities"`
	Locations                          []map[string]interface{}               `json:"Locations"`
	BackupPolicyType                   string                                 `json:"backupPolicyType,omitempty"`
	BackupIntervalInMinutes            int                                    `json:"backupIntervalInMinutes,omitempty"`
	BackupRetentionIntervalInHours     int                                    `json:"backupRetentionIntervalInHours,omitempty"`
	BackupStorageRedundancy            string                                 `json:"backupStorageRedundancy,omitempty"`
	DisableKeyBasedMetadataWriteAccess bool                                   `json:"DisableKeyBasedMetadataWriteAccess"`
	IsFreeTierEnabled                  bool                                   `json:"IsFreeTierEnabled"`
	PublicNetworkAccess                string                                 `json:"PublicNetworkAccess,omitempty"`
	CapacityMode                       string                                 `json:"CapacityMode,omitempty"`
	ProvisioningState                  string                                 `json:"ProvisioningState,omitempty"`
	ResourceType                       *DuploAzureCosmosDBAccountResourceType `json:"ResourceType,omitempty"`
	ContinuousModeTier                 string                                 `json:"ContinuousModeTier,omitempty"`
}
type DuploAzureCosmosDBAccountResourceType struct {
	Namespace string `json:"Namespace"`
	Type      string `json:"Type"`
}
type DuploAzureCosmosDB struct {
	Resource     DuploAzureCosmosDBResource     `json:"Resource"`
	Name         string                         `json:"Name"`
	ResourceType DuploAzureCosmosDBResourceType `json:"ResourceType"`
}
type DuploAzureCosmosDBResource struct {
	DatabaseName string `json:"databaseName"`
}

type DuploAzureCosmosDBResourceType struct {
	Namespace string `json:"Namespace"`
	Type      string `json:"Type"`
}

func (c *Client) CreateCosmosDBAccount(tenantId string, rq DuploAzureCosmosDBAccount) ClientError {
	rp := ""
	return c.postAPI(fmt.Sprintf("CreateCosmosDBAccount(%s)", tenantId),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/cosmosDb/accounts", tenantId),
		&rq,
		&rp)

}

func (c *Client) GetCosmosDBAccount(tenantId, name string) (*DuploAzureCosmosDBAccount, ClientError) {
	rp := DuploAzureCosmosDBAccount{}
	err := c.getAPI(fmt.Sprintf("GetCosmosDB(%s,%s)", tenantId, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/cosmosDb/accounts/%s", tenantId, name),
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

func (c *Client) UpdateCosmosDBAccount(tenantId string, name string, rq DuploAzureCosmosDBAccount) ClientError {
	rp := ""
	return c.putAPI(fmt.Sprintf("UpdateCosmosDBAccount(%s,%s)", tenantId, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/cosmosDb/accounts/%s", tenantId, name),
		&rq,
		&rp)

}

func (c *Client) DeleteCosmosDBAccount(tenantId, name string) ClientError {
	rp := ""
	return c.deleteAPI(fmt.Sprintf("DeleteCosmosDBAccount(%s,%s)", tenantId, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/cosmosDb/accounts/%s", tenantId, name), &rp)
}

func (c *Client) DeleteCosmosDB(tenantId, account, name string) ClientError {
	rp := ""
	return c.deleteAPI(fmt.Sprintf("DeleteCosmosDB(%s,%s,%s)", tenantId, account, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/cosmosDb/accounts/%s/databases/%s", tenantId, account, name), &rp)
}

type DuploAzureCosmosDBContainer struct {
	Resource     *DuploAzureCosmosDBContainerResource `json:"Resource"`
	Name         string                               `json:"Name"`
	ResourceType *DuploAzureCosmosDBResourceType      `json:"ResourceType"`
}
type DuploAzureCosmosDBContainerResource struct {
	ContainerName string                                   `json:"ContainerName"`
	PartitionKey  *DuploAzureCosmosDBContainerPartitionKey `json:"PartitionKey"`
}

type DuploAzureCosmosDBContainerPartitionKey struct {
	Paths   []string `json:"Paths"`
	Version int      `json:"version,omitempty"`
}

func (c *Client) GetCosmosDBDatabaseContainer(tenantId, account, dbName, name string) (*DuploAzureCosmosDBContainer, ClientError) {
	rp := DuploAzureCosmosDBContainer{}
	err := c.getAPI(fmt.Sprintf("GetCosmosDBDatabaseContainer(%s,%s,%s)", tenantId, account, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/cosmosDb/accounts/%s/databases/%s/containers/%s", tenantId, account, dbName, name),
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

func (c *Client) CreateCosmosDBDatabaseContainer(tenantId, account, dbName string, rq DuploAzureCosmosDBContainer) ClientError {
	rp := ""
	return c.postAPI(fmt.Sprintf("CreateCosmosDBDatabaseContainer(%s,%s,%s)", tenantId, account, dbName),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/cosmosDb/accounts/%s/databases/%s/containers", tenantId, account, dbName),
		&rq,
		&rp)
}
func (c *Client) DeleteCosmosDBDatabaseContainer(tenantId, account, dbName, name string) ClientError {
	rp := ""
	return c.deleteAPI(fmt.Sprintf("DeleteCosmosDBDatabaseContainer(%s,%s,%s)", tenantId, account, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/cosmosDb/accounts/%s/databases/%s/containers/%s", tenantId, account, dbName, name), &rp)
}

type ConsistencyLevel int

const (
	Eventual ConsistencyLevel = iota
	Session
	BoundedStaleness
	Strong
	ConsistentPrefix
)

var (
	ConsistencyLevelMap = map[float64]string{
		0: "Eventual",
		1: "Session",
		2: "BoundedStaleness",
		3: "Strong",
		4: "ConsistentPrefix",
	}
)
