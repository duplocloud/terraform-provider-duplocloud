package duplosdk

import (
	"fmt"
)

// S3 Table Bucket

type DuploS3TableRequest struct {
	Name string `json:"Name"`
}

type DuploS3TableResource struct {
	Name           string `json:"Name"`
	Arn            string `json:"Arn"`
	OwnerAccountId string `json:"OwnerAccountId"`
	CreatedAt      string `json:"CreatedAt"`
}

// S3 Table Namespace

type DuploS3TableNamespaceRequest struct {
	Name string `json:"Name"`
}

type DuploS3TableNamespaceResource struct {
	Name           string `json:"Name"`
	TableBucketArn string `json:"TableBucketArn"`
	CreatedAt      string `json:"CreatedAt"`
	OwnerAccountId string `json:"OwnerAccountId"`
}

// S3 Table Bucket client methods

func (c *Client) S3TableCreate(tenantID string, rq *DuploS3TableRequest) (*DuploS3TableResource, ClientError) {
	rp := DuploS3TableResource{}
	err := c.postAPI(
		fmt.Sprintf("S3TableCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/s3TableBuckets", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) S3TableGet(tenantID, name string) (*DuploS3TableResource, ClientError) {
	rp := DuploS3TableResource{}
	err := c.getAPI(
		fmt.Sprintf("S3TableGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/s3TableBuckets/%s", tenantID, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) S3TableDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("S3TableDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/s3TableBuckets/%s", tenantID, name),
		nil,
	)
}

// S3 Table Namespace client methods

func (c *Client) S3TableNamespaceCreate(tenantID, bucketName string, rq *DuploS3TableNamespaceRequest) (*DuploS3TableNamespaceResource, ClientError) {
	rp := DuploS3TableNamespaceResource{}
	err := c.postAPI(
		fmt.Sprintf("S3TableNamespaceCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/aws/s3TableBuckets/%s/namespaces", tenantID, bucketName),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) S3TableNamespaceGet(tenantID, bucketName, name string) (*DuploS3TableNamespaceResource, ClientError) {
	rp := DuploS3TableNamespaceResource{}
	err := c.getAPI(
		fmt.Sprintf("S3TableNamespaceGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/s3TableBuckets/%s/namespaces/%s", tenantID, bucketName, name),
		&rp,
	)
	return &rp, err
}

func (c *Client) S3TableNamespaceDelete(tenantID, bucketName, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("S3TableNamespaceDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/aws/s3TableBuckets/%s/namespaces/%s", tenantID, bucketName, name),
		nil,
	)
}
