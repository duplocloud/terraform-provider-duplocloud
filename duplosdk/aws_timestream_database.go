package duplosdk

import (
	"fmt"
)

type DuploTimestreamDBCreateRequest struct {
	DatabaseName string                 `json:"DatabaseName"`
	KmsKeyId     string                 `json:"KmsKeyId,omitempty"`
	Tags         *[]DuploKeyStringValue `json:"Tags,omitempty"`
}

type DuploTimestreamDBUpdateRequest struct {
	DatabaseName string                 `json:"DatabaseName"`
	UpdatedTags  *[]DuploKeyStringValue `json:"UpdatedTags,omitempty"`
	DeletedTags  []string               `json:"DeletedTags,omitempty"`
}

type DuploTimestreamDBDetails struct {
	DatabaseName string                 `json:"DatabaseName"`
	KmsKeyId     string                 `json:"KmsKeyId,omitempty"`
	Arn          string                 `json:"Arn,omitempty"`
	TableCount   int                    `json:"TableCount,omitempty"`
	Tags         *[]DuploKeyStringValue `json:"Tags,omitempty"`
}

type DuploTimestreamDBTableCreateRequest struct {
	DatabaseName                 string                                              `json:"DatabaseName"`
	TableName                    string                                              `json:"TableName,omitempty"`
	RetentionProperties          *DuploTimestreamDBTableRetentionProperties          `json:"RetentionProperties,omitempty"`
	MagneticStoreWriteProperties *DuploTimestreamDBTableMagneticStoreWriteProperties `json:"MagneticStoreWriteProperties,omitempty"`
	Tags                         *[]DuploKeyStringValue                              `json:"Tags,omitempty"`
}

type DuploTimestreamDBTableUpdateRequest struct {
	DatabaseName                 string                                              `json:"DatabaseName"`
	TableName                    string                                              `json:"TableName,omitempty"`
	RetentionProperties          *DuploTimestreamDBTableRetentionProperties          `json:"RetentionProperties,omitempty"`
	MagneticStoreWriteProperties *DuploTimestreamDBTableMagneticStoreWriteProperties `json:"MagneticStoreWriteProperties,omitempty"`
	UpdatedTags                  *[]DuploKeyStringValue                              `json:"UpdatedTags,omitempty"`
	DeletedTags                  []string                                            `json:"DeletedTags,omitempty"`
}

type DuploTimestreamDBTableDetails struct {
	TenantID                     string                                              `json:"TenantID,omitempty"`
	DatabaseName                 string                                              `json:"DatabaseName"`
	TableName                    string                                              `json:"TableName,omitempty"`
	RetentionProperties          *DuploTimestreamDBTableRetentionProperties          `json:"RetentionProperties,omitempty"`
	MagneticStoreWriteProperties *DuploTimestreamDBTableMagneticStoreWriteProperties `json:"MagneticStoreWriteProperties,omitempty"`
	Arn                          string                                              `json:"Arn,omitempty"`
	TableStatus                  *DuploStringValue                                   `json:"TableStatus,omitempty"`
	Tags                         *[]DuploKeyStringValue                              `json:"Tags,omitempty"`
}

type DuploTimestreamDBTableMagneticStoreWriteProperties struct {
	EnableMagneticStoreWrites         bool                               `json:"EnableMagneticStoreWrites,omitempty"`
	MagneticStoreRejectedDataLocation *MagneticStoreRejectedDataLocation `json:"MagneticStoreRejectedDataLocation,omitempty"`
}

type MagneticStoreRejectedDataLocation struct {
	S3Configuration *MagneticStoreRejectedDataS3Configuration `json:"S3Configuration,omitempty"`
}

type MagneticStoreRejectedDataS3Configuration struct {
	BucketName       string            `json:"BucketName,omitempty"`
	ObjectKeyPrefix  string            `json:"ObjectKeyPrefix,omitempty"`
	EncryptionOption *DuploStringValue `json:"EncryptionOption,omitempty"`
	KmsKeyId         string            `json:"KmsKeyId,omitempty"`
}

type DuploTimestreamDBTableRetentionProperties struct {
	MemoryStoreRetentionPeriodInHours  int `json:"MemoryStoreRetentionPeriodInHours,omitempty"`
	MagneticStoreRetentionPeriodInDays int `json:"MagneticStoreRetentionPeriodInDays,omitempty"`
}

func (c *Client) DuploTimestreamDBCreate(tenantID string, rq *DuploTimestreamDBCreateRequest) (*DuploTimestreamDBDetails, ClientError) {
	rp := DuploTimestreamDBDetails{}
	err := c.postAPI(
		fmt.Sprintf("DuploTimestreamDBCreate(%s, %s)", tenantID, rq.DatabaseName),
		fmt.Sprintf("v3/subscriptions/%s/aws/timeStream", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploTimestreamDBUpdate(tenantID string, name string, rq *DuploTimestreamDBUpdateRequest) (*DuploTimestreamDBDetails, ClientError) {
	rp := DuploTimestreamDBDetails{}
	err := c.putAPI(
		fmt.Sprintf("DuploTimestreamDBCreate(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/timeStream/duploservices-%s", tenantID, name),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploTimestreamDBDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploTimestreamDBDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/timeStream/%s", tenantID, name),
		nil)
}

func (c *Client) DuploTimestreamDBGet(tenantID string, name string) (*DuploTimestreamDBDetails, ClientError) {
	list, err := c.DuploTimestreamDBGetList(tenantID)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, element := range *list {
			if element.DatabaseName == name {
				return &element, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) DuploTimestreamDBGetList(tenantID string) (*[]DuploTimestreamDBDetails, ClientError) {
	rp := []DuploTimestreamDBDetails{}
	err := c.getAPI(
		fmt.Sprintf("DuploTimestreamDBGetList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/timeStream", tenantID),
		&rp)
	return &rp, err
}

func (c *Client) DuploTimestreamDBTableCreate(tenantID string, rq *DuploTimestreamDBTableCreateRequest) (*DuploTimestreamDBDetails, ClientError) {
	rp := DuploTimestreamDBDetails{}
	err := c.postAPI(
		fmt.Sprintf("DuploTimestreamDBTableCreate(%s, %s)", tenantID, rq.TableName),
		fmt.Sprintf("v3/subscriptions/%s/aws/timeStreamtable", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploTimestreamDBTableUpdate(tenantID string, dbName, name string, rq *DuploTimestreamDBTableUpdateRequest) (*DuploTimestreamDBDetails, ClientError) {
	rp := DuploTimestreamDBDetails{}
	err := c.putAPI(
		fmt.Sprintf("DuploTimestreamDBTableUpdate(%s, %s,  %s)", tenantID, dbName, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/timeStreamtable/%s/%s", tenantID, dbName, name),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploTimestreamDBTableDelete(tenantID, dbName, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("DuploTimestreamDBTableDelete(%s, %s,  %s)", tenantID, dbName, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/timeStreamtable/%s/%s", tenantID, dbName, name),
		nil)
}

func (c *Client) DuploTimestreamDBTableGet(tenantID string, dbName, name string) (*DuploTimestreamDBTableDetails, ClientError) {
	rp := DuploTimestreamDBTableDetails{}
	err := c.getAPI(
		fmt.Sprintf("DuploTimestreamDBTableGet(%s, %s,  %s)", tenantID, dbName, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/timeStreamtable/%s/%s", tenantID, dbName, name),
		&rp)
	if rp.TableName == "" {
		return nil, err
	}

	//rp.TenantID = tenantID
	return &rp, err
}
