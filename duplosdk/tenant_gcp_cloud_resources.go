package duplosdk

import "fmt"

// DuploGcpPubsubTopic represents a GCP pubsub topic resource for a Duplo tenant
type DuploGcpPubsubTopic struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	// NOTE: The ShortName field does not come from the backend - we synthesize it
	ShortName string `json:"-,omitempty"`

	Name     string            `json:"Name,omitempty"`
	SelfLink string            `json:"SelfLink,omitempty"`
	Status   string            `json:"Status,omitempty"`
	Labels   map[string]string `json:"Labels,omitempty"`
}

// GcpPubsubTopicCreate creates a pubsub topic via the Duplo API.
func (c *Client) GcpPubsubTopicCreate(tenantID string, rq *DuploGcpPubsubTopic) (*DuploGcpPubsubTopic, ClientError) {
	rp := DuploGcpPubsubTopic{}
	err := c.postAPI(
		fmt.Sprintf("GcpPubsubTopicCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/google/topic", tenantID),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// GcpPubsubTopicUpdate updates a pubsub topic via the Duplo API.
func (c *Client) GcpPubsubTopicUpdate(tenantID string, rq *DuploGcpPubsubTopic) (*DuploGcpPubsubTopic, ClientError) {
	rp := DuploGcpPubsubTopic{}
	err := c.postAPI(
		fmt.Sprintf("GcpPubsubTopicUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/google/topic", tenantID),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// GcpPubsubTopicDelete deletes a pubsub topic via the Duplo API.
func (c *Client) GcpPubsubTopicDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("GcpPubsubTopicDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/topic/%s", tenantID, name),
		nil)
}

// GcpPubsubTopicGetList gets a list of pubsub topics via the Duplo API.
func (c *Client) GcpPubsubTopicGetList(tenantID string) (*[]DuploGcpPubsubTopic, ClientError) {
	prefix, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the list from Duplo
	list := []DuploGcpPubsubTopic{}
	err = c.getAPI(
		fmt.Sprintf("GcpPubsubTopicGetList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/topic", tenantID),
		&list)
	if err != nil {
		return nil, err
	}

	// Add the tenant ID and name to each element and return the list.
	for i := range list {
		list[i].TenantID = tenantID
		list[i].ShortName, _ = UnprefixName(prefix, list[i].Name)
	}
	return &list, nil
}

// GcpPubsubTopicGet gets a pubsub topic via the Duplo API.
func (c *Client) GcpPubsubTopicGet(tenantID string, name string) (*DuploGcpPubsubTopic, ClientError) {

	// Get the list from Duplo
	rp := DuploGcpPubsubTopic{}
	err := c.getAPI(
		fmt.Sprintf("GcpPubsubTopicGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/topic/%s", tenantID, name),
		&rp)
	rp.TenantID = tenantID
	rp.ShortName = name
	return &rp, err
}
