package duplosdk

import "fmt"

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
	MaxStalenessPrefix      float64 `json:"maxStalenessPrefix"`
	MaxIntervalInSeconds    int     `json:"maxIntervalInSeconds"`
	DefaultConsistencyLevel string  `json:"defaultConsistencyLevel"` //ENUM: Eventual,Session,BoundedStaleness,Strong,

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

func (c *Client) CreateCosmosDB(tenantId string, rq DuploAzureCosmosDBRequest) ClientError {
	rp := make(map[string]interface{})
	err := c.postAPI(fmt.Sprintf("CreateCosmosDB(%s)", tenantId),
		fmt.Sprintf("v3/subscriptions/%s/azure/cosmosDb/account", tenantId),
		&rq,
		&rp)
	if err != nil {
		return err
	}
	fmt.Println(rp)
	return nil
}

func (c *Client) GetCosmosDB(tenantId, name string) (*DuploAzureCosmosDBResponse, ClientError) {
	rp := DuploAzureCosmosDBResponse{}
	err := c.getAPI(fmt.Sprintf("GetCosmosDB(%s,%s)", tenantId, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/cosmosDb/account/%s", tenantId, name),
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

type DuploAzureCosmosDBAccount struct {
	Name                               string                               `json:"name"`
	Kind                               string                               `json:"kind"`
	AccountType                        string                               `json:"type"`
	ConsistencyPolicy                  *DuploAzureCosmosDBConsistencyPolicy `json:"consistencyPolicy"`
	Capabilities                       *[]DuploAzureCosmosDBCapability      `json:"Capabilities"`
	Locations                          []string                             `json:"locations"`
	BackupPolicyType                   string                               `json:"backupPolicyType,omitempty"`
	BackupIntervalInMinutes            int                                  `json:"backupIntervalInMinutes,omitempty"`
	BackupRetentionIntervalInHours     int                                  `json:"backupRetentionIntervalInHours,omitempty"`
	BackupStorageRedundancy            string                               `json:"backupStorageRedundancy,omitempty"`
	DisableKeyBasedMetadataWriteAccess bool                                 `json:"DisableKeyBasedMetadataWriteAccess,omitempty"`
	IsFreeTierEnabled                  bool                                 `json:"IsFreeTierEnabled,omitempty"`
	PublicNetworkAccess                string                               `json:"PublicNetworkAccess,omitempty"`
	CapacityMode                       string                               `json:"CapacityMode,omitempty"`
}

func (c *Client) CreateCosmosDBAccount(tenantId string, rq DuploAzureCosmosDBAccount) ClientError {
	rp := make(map[string]interface{})
	return c.postAPI(fmt.Sprintf("CreateCosmosDBAccount(%s)", tenantId),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/cosmosDb/accounts", tenantId),
		&rq,
		&rp)

}

func (c *Client) GetCosmosDBAccount(tenantId, name string) (*DuploAzureCosmosDBAccount, ClientError) {
	rp := DuploAzureCosmosDBAccount{}
	err := c.getAPI(fmt.Sprintf("GetCosmosDB(%s,%s)", tenantId, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/cosmosDb/account/%s", tenantId, name),
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

func (c *Client) UpdateCosmosDBAccount(tenantId string, name string, rq DuploAzureCosmosDBAccount) ClientError {
	rp := make(map[string]interface{})
	return c.postAPI(fmt.Sprintf("UpdateCosmosDBAccount(%s,%s)", tenantId, name),
		fmt.Sprintf("v3/subscriptions/%s/azure/arm/cosmosDb/accounts/%s", tenantId, name),
		&rq,
		&rp)

}
