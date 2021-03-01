package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// DuploAwsCloudResource represents a generic AWS cloud resource for a Duplo tenant
type DuploAwsCloudResource struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Type     int    `json:"ResourceType,omitempty"`
	Name     string `json:"Name,omitempty"`
	Arn      string `json:"Arn,omitempty"`
	MetaData string `json:"MetaData,omitempty"`
}

// DuploS3Bucket represents an S3 bucket resource for a Duplo tenant
type DuploS3Bucket struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Name string `json:"Name,omitempty"`
	Arn  string `json:"Arn,omitempty"`
}

// DuploS3BucketRequest represents a request to create an S3 bucket resource
type DuploS3BucketRequest struct {
	Type           int    `json:"ResourceType"`
	Name           string `json:"Name"`
	State          string `json:"State,omitempty"`
	InTenantRegion bool   `json:"InTenantRegion"`
}

// TenantListAwsCloudResources retrieves a list of the generic AWS cloud resources for a tenant via the Duplo API.
func (c *Client) TenantListAwsCloudResources(tenantID string) (*[]DuploAwsCloudResource, error) {

	// Format the URL
	url := fmt.Sprintf("%s/subscriptions/%s/GetCloudResources", c.HostURL, tenantID)
	log.Printf("[TRACE] duplo-TenantListAwsCloudResources 1 ********: %s ", url)

	// Get the list from Duplo
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-TenantListAwsCloudResources 2 ********: %s", err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-TenantListAwsCloudResources 3 ********: %s", bodyString)

	// Return it as a list.
	list := []DuploAwsCloudResource{}
	err = json.Unmarshal(body, &list)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-TenantListAwsCloudResources 4 ********: %d items", len(list))
	for i := range list {
		list[i].TenantID = tenantID
	}
	return &list, nil
}

// TenantGetAwsCloudResource retrieves a cloud resource by type and name
func (c *Client) TenantGetAwsCloudResource(tenantID string, resourceType int, name string) (*DuploAwsCloudResource, error) {
	allResources, err := c.TenantListAwsCloudResources(tenantID)
	if err != nil {
		return nil, err
	}

	// Find and return the secret with the specific type and name.
	for _, resource := range *allResources {
		if resource.Type == resourceType && resource.Name == name {
			return &resource, nil
		}
	}

	// No resource was found.
	return nil, nil
}

// TenantGetS3BucketFullName retrieves the full name of a managed S3 bucket.
func (c *Client) TenantGetS3BucketFullName(tenantID string, name string) (string, error) {

	// Figure out the full resource name.
	accountID, err := c.TenantGetAwsAccountID(tenantID)
	if err != nil {
		return "", err
	}
	tenant, err := c.GetTenantForUser(tenantID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("duploservices-%s-%s-%s", tenant.AccountName, name, accountID), nil
}

// TenantGetS3Bucket retrieves a managed S3 bucket via the Duplo API
func (c *Client) TenantGetS3Bucket(tenantID string, name string) (*DuploS3Bucket, error) {
	// Figure out the full resource name.
	fullName, err := c.TenantGetS3BucketFullName(tenantID, name)
	if err != nil {
		return nil, err
	}

	// Get the resource from Duplo.
	resource, err := c.TenantGetAwsCloudResource(tenantID, 1, fullName)
	if err != nil || resource == nil {
		return nil, err
	}

	return &DuploS3Bucket{
		TenantID: tenantID,
		Name:     resource.Name,
		Arn:      fmt.Sprintf("arn:aws:s3:::%s", resource.Name),
	}, nil
}

// TenantCreateS3Bucket creates an S3 bucket resource via Duplo.
func (c *Client) TenantCreateS3Bucket(tenantID string, duplo DuploS3BucketRequest) error {
	duplo.Type = 1 // type of "1" signifies an S3 bucket

	// Build the request
	rqBody, err := json.Marshal(&duplo)
	if err != nil {
		log.Printf("[TRACE] TenantCreateS3Bucket 1 JSON gen : %s", err.Error())
		return err
	}
	url := fmt.Sprintf("%s/subscriptions/%s/S3BucketUpdate", c.HostURL, tenantID)
	log.Printf("[TRACE] TenantCreateS3Bucket 2 : %s <= %s", url, rqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] TenantCreateS3Bucket 3 HTTP builder : %s", err.Error())
		return err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] TenantCreateS3Bucket 4 HTTP POST : %s", err.Error())
		return fmt.Errorf("Tenant %s failed to create bucket %s: '%s'", tenantID, duplo.Name, err)
	}
	bodyString := string(body)

	// Expect the response to be "null"
	if bodyString == "null" {
		return nil
	}
	return fmt.Errorf("Tenant %s failed to create bucket %s: '%s'", tenantID, duplo.Name, bodyString)
}

// TenantDeleteS3Bucket deletes an S3 bucket resource via Duplo.
func (c *Client) TenantDeleteS3Bucket(tenantID string, name string) error {
	// Figure out the full resource name.
	fullName, err := c.TenantGetS3BucketFullName(tenantID, name)
	if err != nil {
		return err
	}

	// Build the request
	duplo := DuploS3BucketRequest{
		Type:  1, // type of "1" signifies an S3 bucket
		Name:  fullName,
		State: "delete",
	}
	rqBody, err := json.Marshal(&duplo)
	if err != nil {
		log.Printf("[TRACE] TenantDeleteS3Bucket 1 JSON gen : %s", err.Error())
		return err
	}
	url := fmt.Sprintf("%s/subscriptions/%s/S3BucketUpdate", c.HostURL, tenantID)
	log.Printf("[TRACE] TenantDeleteS3Bucket 2 : %s <= %s", url, rqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] TenantDeleteS3Bucket 3 HTTP builder : %s", err.Error())
		return err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] TenantDeleteS3Bucket 4 HTTP POST : %s", err.Error())
		return fmt.Errorf("Tenant %s failed to create bucket %s: '%s'", tenantID, duplo.Name, err)
	}
	bodyString := string(body)

	// Expect the response to be "null"
	if bodyString == "null" {
		return nil
	}
	return fmt.Errorf("Tenant %s failed to delete bucket %s: '%s'", tenantID, duplo.Name, bodyString)
}
