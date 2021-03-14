package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	// Create the bucket via Duplo.
	return c.postAPI(
		fmt.Sprintf("TenantCreateS3Bucket(%s, %s)", tenantID, duplo.Name),
		fmt.Sprintf("subscriptions/%s/S3BucketUpdate", tenantID),
		&duplo,
		nil)
}

// TenantDeleteS3Bucket deletes an S3 bucket resource via Duplo.
func (c *Client) TenantDeleteS3Bucket(tenantID string, name string) error {

	// Get the full name of the S3 bucket
	fullName, err := c.TenantGetS3BucketFullName(tenantID, name)
	if err != nil {
		return err
	}

	// Delete the bucket via Duplo.
	return c.postAPI(
		fmt.Sprintf("TenantDeleteS3Bucket(%s, %s)", tenantID, name),
		fmt.Sprintf("subscriptions/%s/S3BucketUpdate", tenantID),
		&DuploS3BucketRequest{Type: ResourceTypeS3Bucket, Name: fullName, State: "delete"},
		nil)
}

// TenantApplyS3BucketSettings applies settings to an S3 bucket resource via Duplo.
func (c *Client) TenantApplyS3BucketSettings(tenantID string, duplo DuploS3BucketSettingsRequest) (*DuploS3Bucket, error) {
	apiName := fmt.Sprintf("TenantApplyS3BucketSettings(%s, %s)", tenantID, duplo.Name)

	// Figure out the full resource name.
	fullName, err := c.TenantGetS3BucketFullName(tenantID, duplo.Name)
	if err != nil {
		return nil, err
	}
	duplo.Name = fullName

	// Apply the settings via Duplo.
	rp := DuploS3Bucket{}
	err = c.postAPI(apiName, fmt.Sprintf("subscriptions/%s/ApplyS3BucketSettings", tenantID), &duplo, &rp)
	if err != nil {
		return nil, err
	}

	// Deal with a missing response.
	if rp.Name == "" {
		err := fmt.Errorf("%s: unexpected missing response from backend", apiName)
		log.Printf("[TRACE] %s", err)
		return nil, err
	}

	// Return the response.
	rp.TenantID = tenantID
	return &rp, nil
}

// TenantUpdateApplicationLbSettings updates an application LB resource's settings via Duplo.
func (c *Client) TenantUpdateApplicationLbSettings(tenantID string, duplo DuploAwsLbSettingsUpdateRequest) error {
	return c.postAPI("TenantUpdateApplicationLbSettings",
		fmt.Sprintf("subscriptions/%s/UpdateLbSettings", tenantID),
		&duplo,
		nil)
}

// TenantGetApplicationLbSettings updates an application LB resource's WAF association via Duplo.
func (c *Client) TenantGetApplicationLbSettings(tenantID string, loadBalancerArn string) (*DuploAwsLbSettings, error) {
	rp := DuploAwsLbSettings{}

	err := c.postAPI("TenantGetApplicationLbSettings",
		fmt.Sprintf("subscriptions/%s/GetLbSettings", tenantID),
		&DuploAwsLbSettingsRequest{LoadBalancerArn: loadBalancerArn},
		&rp)

	return &rp, err
}

// TenantCreateApplicationLB creates an application LB resource via Duplo.
func (c *Client) TenantCreateApplicationLB(tenantID string, duplo DuploAwsLBConfiguration) error {
	return c.postAPI("TenantCreateApplicationLB",
		fmt.Sprintf("subscriptions/%s/ApplicationLbUpdate", tenantID),
		&duplo,
		nil)
}

// TenantDeleteApplicationLB deletes an AWS application LB resource via Duplo.
func (c *Client) TenantDeleteApplicationLB(tenantID string, name string) error {
	// Get the full name of the ALB.
	fullName, err := c.TenantGetApplicationLbFullName(tenantID, name)
	if err != nil {
		return err
	}

	// Call the API.
	return c.postAPI("TenantDeleteApplicationLB",
		fmt.Sprintf("subscriptions/%s/ApplicationLbUpdate", tenantID),
		&DuploAwsLBConfiguration{Name: fullName, State: "delete"},
		nil)
}
