package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const (
	// ResourceTypeS3Bucket represents an S3 bucket
	ResourceTypeS3Bucket int = 1

	// ResourceTypeApplicationLB represents an AWS application LB
	ResourceTypeApplicationLB int = 16
)

// DuploAwsCloudResource represents a generic AWS cloud resource for a Duplo tenant
type DuploAwsCloudResource struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Type     int    `json:"ResourceType,omitempty"`
	Name     string `json:"Name,omitempty"`
	Arn      string `json:"Arn,omitempty"`
	MetaData string `json:"MetaData,omitempty"`

	// S3 bucket and load balancer
	EnableAccessLogs bool                   `json:"EnableAccessLogs,omitempty"`
	Tags             *[]DuploKeyStringValue `json:"Tags,omitempty"`

	// Only S3 bucket
	EnableVersioning  bool     `json:"EnableVersioning,omitempty"`
	AllowPublicAccess bool     `json:"AllowPublicAccess,omitempty"`
	DefaultEncryption string   `json:"DefaultEncryption,omitempty"`
	Policies          []string `json:"Policies,omitempty"`

	// Only Load balancer
	IsInternal bool   `json:"IsInternal,omitempty"`
	WebACLID   string `json:"WebACLID,omitempty"`
}

// DuploS3Bucket represents an S3 bucket resource for a Duplo tenant
type DuploS3Bucket struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Name              string                 `json:"Name,omitempty"`
	Arn               string                 `json:"Arn,omitempty"`
	MetaData          string                 `json:"MetaData,omitempty"`
	EnableVersioning  bool                   `json:"EnableVersioning,omitempty"`
	EnableAccessLogs  bool                   `json:"EnableAccessLogs,omitempty"`
	AllowPublicAccess bool                   `json:"AllowPublicAccess,omitempty"`
	DefaultEncryption string                 `json:"DefaultEncryption,omitempty"`
	Policies          []string               `json:"Policies,omitempty"`
	Tags              *[]DuploKeyStringValue `json:"Tags,omitempty"`
}

// DuploApplicationLB represents an AWS application load balancer resource for a Duplo tenant
type DuploApplicationLB struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Name             string                 `json:"Name,omitempty"`
	Arn              string                 `json:"Arn,omitempty"`
	DNSName          string                 `json:"MetaData,omitempty"`
	EnableAccessLogs bool                   `json:"EnableAccessLogs,omitempty"`
	IsInternal       bool                   `json:"IsInternal,omitempty"`
	WebACLID         string                 `json:"WebACLID,omitempty"`
	Tags             *[]DuploKeyStringValue `json:"Tags,omitempty"`
}

// DuploAwsLBConfiguration represents a request to create an AWS application load balancer resource
type DuploAwsLBConfiguration struct {
	Name             string `json:"Name"`
	State            string `json:"State,omitempty"`
	IsInternal       bool   `json:"IsInternal,omitempty"`
	EnableAccessLogs bool   `json:"EnableAccessLogs,omitempty"`
}

// DuploAwsLbSettings represents an AWS application load balancer's settings
type DuploAwsLbSettings struct {
	LoadBalancerArn  string `json:"LoadBalancerArn"`
	EnableAccessLogs bool   `json:"EnableAccessLogs,omitempty"`
	WebACLID         string `json:"WebACLId,omitempty"`
}

// DuploAwsLBAccessLogsRequest represents a request to retrieve an AWS application load balancer's settings.
type DuploAwsLbSettingsRequest struct {
	LoadBalancerArn string `json:"LoadBalancerArn"`
}

// DuploAwsLBAccessLogsUpdateRequest represents a request to update an AWS application load balancer's settings.
type DuploAwsLbSettingsUpdateRequest struct {
	LoadBalancerArn  string `json:"LoadBalancerArn"`
	EnableAccessLogs bool   `json:"EnableAccessLogs,omitempty"`
	WebACLID         string `json:"WebACLId,omitempty"`
}

// DuploS3BucketRequest represents a request to create an S3 bucket resource
type DuploS3BucketRequest struct {
	Type           int    `json:"ResourceType"`
	Name           string `json:"Name"`
	State          string `json:"State,omitempty"`
	InTenantRegion bool   `json:"InTenantRegion"`
}

// DuploS3BucketSettingsRequest represents a request to create an S3 bucket resource
type DuploS3BucketSettingsRequest struct {
	Name              string   `json:"Name"`
	EnableVersioning  bool     `json:"EnableVersioning,omitempty"`
	EnableAccessLogs  bool     `json:"EnableAccessLogs,omitempty"`
	AllowPublicAccess bool     `json:"AllowPublicAccess,omitempty"`
	DefaultEncryption string   `json:"DefaultEncryption,omitempty"`
	Policies          []string `json:"Policies,omitempty"`
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

// TenantGetApplicationLbFullName retrieves the full name of a pass-thru AWS application load balancer.
func (c *Client) TenantGetApplicationLbFullName(tenantID string, name string) (string, error) {

	// Figure out the full resource name.
	tenant, err := c.GetTenantForUser(tenantID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("duplo3-%s-%s", tenant.AccountName, name), nil
}

// TenantGetS3Bucket retrieves a managed S3 bucket via the Duplo API
func (c *Client) TenantGetS3Bucket(tenantID string, name string) (*DuploS3Bucket, error) {
	// Figure out the full resource name.
	fullName, err := c.TenantGetS3BucketFullName(tenantID, name)
	if err != nil {
		return nil, err
	}

	// Get the resource from Duplo.
	resource, err := c.TenantGetAwsCloudResource(tenantID, ResourceTypeS3Bucket, fullName)
	if err != nil || resource == nil {
		return nil, err
	}

	return &DuploS3Bucket{
		TenantID:          tenantID,
		Name:              resource.Name,
		Arn:               resource.Arn,
		MetaData:          resource.MetaData,
		EnableVersioning:  resource.EnableVersioning,
		AllowPublicAccess: resource.AllowPublicAccess,
		EnableAccessLogs:  resource.EnableAccessLogs,
		DefaultEncryption: resource.DefaultEncryption,
		Policies:          resource.Policies,
		Tags:              resource.Tags,
	}, nil
}

// TenantGetApplicationLB retrieves an application load balancer via the Duplo API
func (c *Client) TenantGetApplicationLB(tenantID string, name string) (*DuploApplicationLB, error) { // Figure out the full resource name.
	fullName, err := c.TenantGetApplicationLbFullName(tenantID, name)
	if err != nil {
		return nil, err
	}

	// Get the resource from Duplo.
	resource, err := c.TenantGetAwsCloudResource(tenantID, ResourceTypeApplicationLB, fullName)
	if err != nil || resource == nil {
		return nil, err
	}

	return &DuploApplicationLB{
		TenantID:         tenantID,
		Name:             resource.Name,
		Arn:              resource.Arn,
		DNSName:          resource.MetaData,
		IsInternal:       resource.IsInternal,
		EnableAccessLogs: resource.EnableAccessLogs,
		Tags:             resource.Tags,
	}, nil
}

// TenantCreateS3Bucket creates an S3 bucket resource via Duplo.
func (c *Client) TenantCreateS3Bucket(tenantID string, duplo DuploS3BucketRequest) error {
	duplo.Type = ResourceTypeS3Bucket

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
		Type:  ResourceTypeS3Bucket,
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

// TenantApplyS3BucketSettings applies settings to an S3 bucket resource via Duplo.
func (c *Client) TenantApplyS3BucketSettings(tenantID string, duplo DuploS3BucketSettingsRequest) (*DuploS3Bucket, error) {
	// Figure out the full resource name.
	fullName, err := c.TenantGetS3BucketFullName(tenantID, duplo.Name)
	if err != nil {
		return nil, err
	}
	duplo.Name = fullName

	// Build the request
	rqBody, err := json.Marshal(&duplo)
	if err != nil {
		log.Printf("[TRACE] TenantApplyS3BucketSettings 1 JSON gen : %s", err.Error())
		return nil, err
	}
	url := fmt.Sprintf("%s/subscriptions/%s/ApplyS3BucketSettings", c.HostURL, tenantID)
	log.Printf("[TRACE] TenantApplyS3BucketSettings 2 : %s <= %s", url, rqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] TenantApplyS3BucketSettings 3 HTTP builder : %s", err.Error())
		return nil, err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] TenantApplyS3BucketSettings 4 HTTP POST : %s", err.Error())
		return nil, fmt.Errorf("Tenant %s failed to update bucket %s: '%s'", tenantID, duplo.Name, err)
	}
	bodyString := string(body)

	// Return it as a resource.
	resource := DuploS3Bucket{}
	if bodyString == "" {
		log.Printf("[TRACE] TenantApplyS3BucketSettings 5 NO RESULT : %s", bodyString)
		return nil, fmt.Errorf("Tenant %s failed to update bucket %s: no result from backend", tenantID, duplo.Name)
	}
	err = json.Unmarshal(body, &resource)
	if err != nil {
		return nil, err
	}

	resource.TenantID = tenantID
	return &resource, nil
}

// TenantUpdateApplicationLbSettings updates an application LB resource's settings via Duplo.
func (c *Client) TenantUpdateApplicationLbSettings(tenantID string, duplo DuploAwsLbSettingsUpdateRequest) error {

	// Build the request
	rqBody, err := json.Marshal(&duplo)
	if err != nil {
		log.Printf("[TRACE] TenantUpdateApplicationLbSettings 1 JSON gen : %s", err.Error())
		return err
	}
	url := fmt.Sprintf("%s/subscriptions/%s/UpdateLbSettings", c.HostURL, tenantID)
	log.Printf("[TRACE] TenantUpdateApplicationLbSettings 2 : %s <= %s", url, rqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] TenantUpdateApplicationLbSettings 3 HTTP builder : %s", err.Error())
		return err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] TenantUpdateApplicationLbSettings 4 HTTP POST : %s", err.Error())
		return fmt.Errorf("Tenant %s failed to update load balancer %s: '%s'", tenantID, duplo.LoadBalancerArn, err)
	}
	bodyString := string(body)

	// Expect the response to be "null"
	if bodyString == "null" {
		return nil
	}
	return fmt.Errorf("Tenant %s failed to update load balancer %s: no result from backend", tenantID, duplo.LoadBalancerArn)
}

// TenantGetApplicationLbSettings updates an application LB resource's WAF association via Duplo.
func (c *Client) TenantGetApplicationLbSettings(tenantID string, loadBalancerArn string) (*DuploAwsLbSettings, error) {

	// Build the request
	rq := DuploAwsLbSettingsRequest{LoadBalancerArn: loadBalancerArn}
	rqBody, err := json.Marshal(&rq)
	if err != nil {
		log.Printf("[TRACE] TenantGetApplicationLbSettings 1 JSON gen : %s", err.Error())
		return nil, err
	}
	url := fmt.Sprintf("%s/subscriptions/%s/GetLbSettings", c.HostURL, tenantID)
	log.Printf("[TRACE] TenantGetApplicationLbSettings 2 : %s <= %s", url, rqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] TenantGetApplicationLbSettings 3 HTTP builder : %s", err.Error())
		return nil, err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] TenantGetApplicationLbSettings 4 HTTP POST : %s", err.Error())
		return nil, fmt.Errorf("Tenant %s failed to get load balancer %s setings: '%s'", tenantID, loadBalancerArn, err)
	}
	bodyString := string(body)
	log.Printf("[TRACE] TenantGetApplicationLbSettings 5 ********: %s", bodyString)

	// Return it as an object.
	result := DuploAwsLbSettings{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// TenantCreateApplicationLB creates an application LB resource via Duplo.
func (c *Client) TenantCreateApplicationLB(tenantID string, duplo DuploAwsLBConfiguration) error {
	// Build the request
	rqBody, err := json.Marshal(&duplo)
	if err != nil {
		log.Printf("[TRACE] TenantCreateApplicationLB 1 JSON gen : %s", err.Error())
		return err
	}
	url := fmt.Sprintf("%s/subscriptions/%s/ApplicationLbUpdate", c.HostURL, tenantID)
	log.Printf("[TRACE] TenantCreateApplicationLB 2 : %s <= %s", url, rqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] TenantCreateApplicationLB 3 HTTP builder : %s", err.Error())
		return err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] TenantCreateApplicationLB 4 HTTP POST : %s", err.Error())
		return fmt.Errorf("Tenant %s failed to apply load balancer %s: '%s'", tenantID, duplo.Name, err)
	}
	bodyString := string(body)

	// Expect the response to be "null"
	if bodyString == "null" {
		return nil
	}
	return fmt.Errorf("Tenant %s failed to apply load balancer %s: '%s'", tenantID, duplo.Name, bodyString)
}

// TenantDeleteApplicationLB deletes an AWS application LB resource via Duplo.
func (c *Client) TenantDeleteApplicationLB(tenantID string, name string) error {
	fullName, err := c.TenantGetApplicationLbFullName(tenantID, name)
	if err != nil {
		return err
	}

	// Build the request
	duplo := DuploAwsLBConfiguration{
		Name:  fullName,
		State: "delete",
	}
	rqBody, err := json.Marshal(&duplo)
	if err != nil {
		log.Printf("[TRACE] TenantDeleteApplicationLB 1 JSON gen : %s", err.Error())
		return err
	}
	url := fmt.Sprintf("%s/subscriptions/%s/ApplicationLbUpdate", c.HostURL, tenantID)
	log.Printf("[TRACE] TenantDeleteApplicationLB 2 : %s <= %s", url, rqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] TenantDeleteApplicationLB 3 HTTP builder : %s", err.Error())
		return err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] TenantDeleteApplicationLB 4 HTTP POST : %s", err.Error())
		return fmt.Errorf("Tenant %s failed to delete load balancer %s: '%s'", tenantID, duplo.Name, err)
	}
	bodyString := string(body)

	// Expect the response to be "null"
	if bodyString == "null" {
		return nil
	}
	return fmt.Errorf("Tenant %s failed to delete load balancer %s: '%s'", tenantID, duplo.Name, bodyString)
}
