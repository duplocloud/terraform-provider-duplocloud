package duplosdk

import (
	"fmt"
	"log"
)

type DuploAsgProfile struct {
	AccountName         string                             `json:"AccountName,omitempty"`
	AgentPlatform       int                                `json:"AgentPlatform"`
	AllocatedPublicIP   bool                               `json:"AllocatedPublicIp,omitempty"`
	Base64UserData      string                             `json:"Base64UserData,omitempty"`
	PrependUserData     bool                               `json:"IsUserDataCombined,omitempty"`
	CanScaleFromZero    bool                               `json:"CanScaleFromZero,omitempty"`
	Capacity            string                             `json:"Capacity,omitempty"`
	Cloud               int                                `json:"Cloud"`
	CustomDataTags      *[]DuploKeyStringValue             `json:"CustomDataTags"`
	DesiredCapacity     int                                `json:"DesiredCapacity"`
	EncryptDisk         bool                               `json:"EncryptDisk,omitempty"`
	FriendlyName        string                             `json:"FriendlyName,omitempty"`
	ImageID             string                             `json:"ImageId,omitempty"`
	IsClusterAutoscaled bool                               `json:"IsClusterAutoscaled,omitempty"`
	IsEbsOptimized      bool                               `json:"IsEbsOptimized"`
	IsMinion            bool                               `json:"IsMinion"`
	KeyPairType         int                                `json:"KeyPairType,omitempty"`
	MaxSize             int                                `json:"MaxSize"`
	MaxSpotPrice        string                             `json:"SpotPrice,omitempty"`
	MetaData            *[]DuploKeyStringValue             `json:"MetaData,omitempty"`
	MinionTags          *[]DuploKeyStringValue             `json:"MinionTags,omitempty"`
	MinSize             int                                `json:"MinSize"`
	NetworkInterfaces   *[]DuploNativeHostNetworkInterface `json:"NetworkInterfaces,omitempty"`
	Status              string                             `json:"Status,omitempty"`
	Tags                *[]DuploKeyStringValue             `json:"Tags,omitempty"`
	TenantId            string                             `json:"TenantId,omitempty"`
	UseSpotInstances    bool                               `json:"UseSpotInstances,omitempty"`
	Volumes             *[]DuploNativeHostVolume           `json:"Volumes,omitempty"`
	Zones               []int                              `json:"Zones"`
	Zone                int                                `json:"Zone"`
	EnabledMetrics      *[]string                          `json:"EnabledMetrics,omitempty"`
	ExtraNodeLabels     *[]DuploKeyStringValue             `json:"ExtraNodeLabels,omitempty"`
	Taints              *[]DuploTaints                     `json:"Taints,omitempty"`
	Created             *bool                              `json:"Created,omitempty"`
}

type DuploAsgProfileDeleteReq struct {
	FriendlyName string `json:"FriendlyName,omitempty"`
	State        string `json:"State,omitempty"`
}

// AsgProfileGetList retrieves a list of ASG profiles via the Duplo API.
func (c *Client) AsgProfileGetList(tenantID string) (*[]DuploAsgProfile, ClientError) {
	log.Printf("[DEBUG] Duplo API - Get ASG Profile List(TenantId-%s)", tenantID)
	conf := NewRetryConf()

	rp := []DuploAsgProfile{}
	err := c.getAPIWithRetry(fmt.Sprintf("AsgProfileGetList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetTenantAsgProfiles", tenantID),
		&rp, &conf)
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
func (c *Client) AsgProfileUpdate(rq *DuploAsgProfile) ClientError {
	err := c.AsgProfileUpdateV3(rq)
	if err != nil {
		_, err := c.AsgProfileCreateOrUpdate(rq, true)

		return err
	}
	return err
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

func (c *Client) AsgProfileUpdateV3(rq *DuploAsgProfile) ClientError {

	// Build the request
	var rp interface{}
	return c.putAPI(fmt.Sprintf("AsgProfileUpdate(%s, %s)", rq.TenantId, rq.FriendlyName),
		fmt.Sprintf("v3/subscriptions/%s/aws/asg", rq.TenantId), rq,
		&rp)

}

// AsgProfileDelete deletes a ASG profile via the Duplo API.
func (c *Client) AsgProfileDelete(tenantID, friendlyName string) ClientError {
	var rp = ""
	req := DuploAsgProfileDeleteReq{FriendlyName: friendlyName, State: "delete"}
	return c.postAPI(fmt.Sprintf("AsgProfileDelete(%s, %s)", tenantID, friendlyName),
		fmt.Sprintf("subscriptions/%s/UpdateTenantAsgProfile", tenantID), req,
		&rp)
}
