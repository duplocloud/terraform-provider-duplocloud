package duplosdk

import (
	"fmt"
	"log"
)

type DuploAsgProfile struct {
	MinSize             int                                `json:"MinSize"`
	MaxSize             int                                `json:"MaxSize"`
	DesiredCapacity     int                                `json:"DesiredCapacity"`
	AccountName         string                             `json:"AccountName,omitempty"`
	TenantId            string                             `json:"TenantId,omitempty"`
	FriendlyName        string                             `json:"FriendlyName,omitempty"`
	Capacity            string                             `json:"Capacity,omitempty"`
	Zone                int                                `json:"Zone"`
	IsMinion            bool                               `json:"IsMinion"`
	ImageID             string                             `json:"ImageId,omitempty"`
	Base64UserData      string                             `json:"Base64UserData,omitempty"`
	AgentPlatform       int                                `json:"AgentPlatform"`
	IsEbsOptimized      bool                               `json:"IsEbsOptimized"`
	AllocatedPublicIP   bool                               `json:"AllocatedPublicIp,omitempty"`
	Cloud               int                                `json:"Cloud"`
	KeyPairType         int                                `json:"KeyPairType,omitempty"`
	IsClusterAutoscaled bool                               `json:"IsClusterAutoscaled,omitempty"`
	EncryptDisk         bool                               `json:"EncryptDisk,omitempty"`
	Status              string                             `json:"Status,omitempty"`
	NetworkInterfaces   *[]DuploNativeHostNetworkInterface `json:"NetworkInterfaces,omitempty"`
	Volumes             *[]DuploNativeHostVolume           `json:"Volumes,omitempty"`
	MetaData            *[]DuploKeyStringValue             `json:"MetaData,omitempty"`
	Tags                *[]DuploKeyStringValue             `json:"Tags,omitempty"`
	MinionTags          *[]DuploKeyStringValue             `json:"MinionTags,omitempty"`
	UseLaunchTemplate   bool                               `json:"UseLaunchTemplate"`
	UseSpotInstances    bool                               `json:"UseSpotInstances,omitempty"`
	MaxSpotPrice        string                             `json:"SpotPrice,omitempty"`
}

type DuploAsgProfileDeleteReq struct {
	FriendlyName string `json:"FriendlyName,omitempty"`
	State        string `json:"State,omitempty"`
}

// AsgProfileGetList retrieves a list of ASG profiles via the Duplo API.
func (c *Client) AsgProfileGetList(tenantID string) (*[]DuploAsgProfile, ClientError) {
	log.Printf("[DEBUG] Duplo API - Get ASG Profile List(TenantId-%s)", tenantID)
	rp := []DuploAsgProfile{}
	err := c.getAPI(fmt.Sprintf("AsgProfileGetList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetTenantAsgProfiles", tenantID),
		&rp)
	return &rp, err
}

// AsgProfileGet retrieves an ASG profile via the Duplo API.
func (c *Client) AsgProfileGet(tenantID, friendlyName string) (*DuploAsgProfile, ClientError) {
	log.Printf("[DEBUG] Duplo API - Get ASG Profile(TenantId-%s,FriendlyName-%s)", tenantID, friendlyName)
	list, err := c.AsgProfileGetList(tenantID)
	for _, profile := range *list {
		if profile.FriendlyName == friendlyName {
			return &profile, err
		}
	}
	return nil, err
}

// AsgProfileExists checks if a ASG profile exists via the Duplo API.
func (c *Client) AsgProfileExists(tenantID, name string) (bool, ClientError) {

	// Get the list of ASG profiles
	// TODO: change the backend error to a 404
	list, err := c.AsgProfileGetList(tenantID)
	if err != nil {
		return false, err
	}

	// Check if the profile exists
	if list != nil {
		for _, profile := range *list {
			if profile.FriendlyName == name {
				return true, nil
			}
		}
	}
	return false, nil
}

// AsgProfileCreate creates an ASG profile via the Duplo API.
func (c *Client) AsgProfileCreate(rq *DuploAsgProfile) (string, ClientError) {
	return c.AsgProfileCreateOrUpdate(rq, false)
}

// AsgProfileUpdate updates an ASG profile via the Duplo API.
func (c *Client) AsgProfileUpdate(rq *DuploAsgProfile) (string, ClientError) {
	return c.AsgProfileCreateOrUpdate(rq, true)
}

// AsgProfileCreateOrUpdate creates or updates a AASG profile via the Duplo API.
func (c *Client) AsgProfileCreateOrUpdate(rq *DuploAsgProfile, updating bool) (string, ClientError) {

	// Build the request
	var verb, msg, api string

	if updating {
		verb = "POST"
		msg = fmt.Sprintf("AsgProfileCreateOrUpdate(%s, %s)", rq.TenantId, rq.FriendlyName)
		api = fmt.Sprintf("subscriptions/%s/UpdateTenantAsgProfile", rq.TenantId)
	} else {
		verb = "POST"
		msg = fmt.Sprintf("AsgProfileCreateOrUpdate(%s, %s)", rq.TenantId, rq.FriendlyName)
		api = fmt.Sprintf("subscriptions/%s/UpdateTenantAsgProfile", rq.TenantId)
	}

	// Call the API.
	rp := ""
	err := c.doAPIWithRequestBody(verb, msg, api, &rq, &rp)
	if err != nil {
		return rp, err
	}
	return rp, err
}

// AsgProfileDelete deletes a ASG profile via the Duplo API.
func (c *Client) AsgProfileDelete(tenantID, friendlyName string) ClientError {
	var rp = ""
	req := DuploAsgProfileDeleteReq{FriendlyName: friendlyName, State: "delete"}
	return c.postAPI(fmt.Sprintf("AsgProfileDelete(%s, %s)", tenantID, friendlyName),
		fmt.Sprintf("subscriptions/%s/UpdateTenantAsgProfile", tenantID), req,
		&rp)
}
