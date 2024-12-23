package duplosdk

type DuploAzureCosmosDBRequest struct {
	Kind                               string `json:"Kind"`
	Identity                           `json:"Identity"`
	ConsistencyPolicy                  `json:"ConsistencyPolicy"`
	Locations                          string   `json:"Locations"`
	IpRules                            []string `json:"IpRules"`
	IsVirtualNetworkFilterEnabled      bool     `json:"IsVirtualNetworkFilterEnabled"`
	EnableAutomaticFailover            bool     `json:"EnableAutomaticFailover"`
	Capabilities                       `json:"Capabilities"`
	VirtualNetworkRules                `json:"VirtualNetworkRules"`
	EnableMultipleWriteLocations       bool `json:"EnableMultipleWriteLocations"`
	EnableCassandraConnector           bool `json:"EnableCassandraConnector"`
	ConnectorOffer                     `json:"ConnectorOffer"`
	DisableKeyBasedMetadataWriteAccess bool   `json:"DisableKeyBasedMetadataWriteAccess"`
	KeyVaultKeyUri                     string `json:"KeyVaultKeyUri"`
	DefaultIdentity                    string `json:"DefaultIdentity"`
	PublicNetworkAccess                `json:"PublicNetworkAccess"`
	EnableFreeTier                     bool `json:"EnableFreeTier"`
	ApiProperties                      `json:"ApiProperties"`
	EnableAnalyticalStorage            bool `json:"EnableAnalyticalStorage"`
	AnalyticalStorageConfiguration     `json:"AnalyticalStorageConfiguration"`
	CreateMode                         `json:"CreateMode"`
	BackupPolicy                       `json:"BackupPolicy"`
	Cors                               `json:"Cors"`
	NetworkAclBypass                   string                              `json:"NetworkAclBypass"`  //None AzureServices
	NetworkAclBypassResourceIds        []string                            `json:"NetworkAclBypassResourceIds"`
	DisableLocalAuth                   bool                                `json:"DisableLocalAuth"`
	RestoreParameters                  DuploAzureCosmosDBRestoreParameters `json:"RestoreParameters"`
	Capacity                           DuploAzureCosmosDBCapacity          `json:"Capacity"`
	DatabaseAccountOfferType           string                              `json:"DatabaseAccountOfferType	"`
}

type DuploAzureCosmosDBCapacity struct {
	TotalThroughputLimit int `json:"TotalThroughputLimit"`
}

type DuploAzureCosmosDBRestoreParameters struct {
	RestoreMode        string                            `json:"RestoreMode"`
	TablesToRestore    []string                          `json:"TablesToRestore"`
	DatabasesToRestore DuploAzureDatabaseRestoreResource `json:"DatabasesToRestore"`
}

type DuploAzureDatabaseRestoreResource struct {
	DatabaseName    string   `json:"DatabaseName"`
	CollectionNames []string `json:"CollectionNames"`
}

type DuploAzureCosmosDB