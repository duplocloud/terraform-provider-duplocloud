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

// DuploGcpStorageBucket represents a GCP pubsub topic resource for a Duplo tenant
type DuploGcpStorageBucket struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	// NOTE: The ShortName field does not come from the backend - we synthesize it
	ShortName string `json:"-,omitempty"`

	Name             string            `json:"Name,omitempty"`
	SelfLink         string            `json:"SelfLink,omitempty"`
	Status           string            `json:"Status,omitempty"`
	EnableVersioning bool              `json:"EnableVersioning,omitempty"`
	Labels           map[string]string `json:"Labels,omitempty"`
}

// DuploGcpCloudFunction represents a GCP cloud function resource for a Duplo tenant
type DuploGcpCloudFunction struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	// NOTE: The ShortName field does not come from the backend - we synthesize it
	ShortName string `json:"-,omitempty"`

	Name                      string                             `json:"Name,omitempty"`
	SelfLink                  string                             `json:"SelfLink,omitempty"`
	Status                    string                             `json:"Status,omitempty"`
	Labels                    map[string]string                  `json:"Labels,omitempty"`
	BuildID                   string                             `json:"BuildId,omitempty"`
	VersionID                 int                                `json:"VersionId,omitempty"`
	EntryPoint                string                             `json:"EntryPoint,omitempty"`
	Runtime                   string                             `json:"Runtime,omitempty"`
	Description               string                             `json:"Description,omitempty"`
	AvailableMemoryMb         int                                `json:"AvailableMemoryMb,omitempty"`
	BuildEnvironmentVariables map[string]string                  `json:"BuildEnvironmentVariables,omitempty"`
	EnvironmentVariables      map[string]string                  `json:"EnvironmentVariables,omitempty"`
	Timeout                   int                                `json:"Timeout,omitempty"`
	SourceArchiveUrl          string                             `json:"SourceArchiveUrl,omitempty"`
	IngressType               int                                `json:"IngressType,omitempty"`
	TriggerType               int                                `json:"TriggerType,omitempty"`
	HTTPSTrigger              *DuploGcpCloudFunctionHTTPSTrigger `json:"HttpsTrigger,omitempty"`
	EventTrigger              *DuploGcpCloudFunctionEventTrigger `json:"EventTrigger,omitempty"`
}

// DuploGcpCloudFunctionHTTPSTrigger represents a GCP cloud function resource for a Duplo tenant
type DuploGcpCloudFunctionHTTPSTrigger struct {
	SecurityLevel string `json:"SecurityLevel,omitempty"`
	Url           string `json:"Url,omitempty"`
}

// DuploGcpCloudFunctionEventTrigger represents a GCP cloud function resource for a Duplo tenant
type DuploGcpCloudFunctionEventTrigger struct {
	EventType string `json:"EventType,omitempty"`
	Resource  string `json:"Resource,omitempty"`
	Service   string `json:"Service,omitempty"`
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
	err := c.putAPI(
		fmt.Sprintf("GcpPubsubTopicUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/google/topic/%s", tenantID, rq.Name),
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

// GcpStorageBucketCreate creates a storage bucket via the Duplo API.
func (c *Client) GcpStorageBucketCreate(tenantID string, rq *DuploGcpStorageBucket) (*DuploGcpStorageBucket, ClientError) {
	rp := DuploGcpStorageBucket{}
	err := c.postAPI(
		fmt.Sprintf("GcpStorageBucketCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/google/bucket", tenantID),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// GcpStorageBucketUpdate updates a storage bucket via the Duplo API.
func (c *Client) GcpStorageBucketUpdate(tenantID string, rq *DuploGcpStorageBucket) (*DuploGcpStorageBucket, ClientError) {
	rp := DuploGcpStorageBucket{}
	err := c.putAPI(
		fmt.Sprintf("GcpStorageBucketUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/google/bucket/%s", tenantID, rq.Name),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// GcpStorageBucketDelete deletes a storage bucket via the Duplo API.
func (c *Client) GcpStorageBucketDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("GcpStorageBucketDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/bucket/%s", tenantID, name),
		nil)
}

// GcpStorageBucketGetList gets a list of storage buckets via the Duplo API.
func (c *Client) GcpStorageBucketGetList(tenantID string) (*[]DuploGcpStorageBucket, ClientError) {
	prefix, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return nil, err
	}
	projectID, err := c.TenantGetGcpProjectID(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the list from Duplo
	list := []DuploGcpStorageBucket{}
	err = c.getAPI(
		fmt.Sprintf("GcpStorageBucketGetList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/bucket", tenantID),
		&list)
	if err != nil {
		return nil, err
	}

	// Add the tenant ID and name to each element and return the list.
	for i := range list {
		list[i].TenantID = tenantID
		list[i].ShortName, _ = UnwrapName(prefix, projectID, list[i].Name)
	}
	return &list, nil
}

// GcpStorageBucketGet gets a storage bucket via the Duplo API.
func (c *Client) GcpStorageBucketGet(tenantID string, name string) (*DuploGcpStorageBucket, ClientError) {

	// Get the list from Duplo
	rp := DuploGcpStorageBucket{}
	err := c.getAPI(
		fmt.Sprintf("GcpStorageBucketGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/bucket/%s", tenantID, name),
		&rp)
	rp.TenantID = tenantID
	rp.ShortName = name
	return &rp, err
}

// GcpCloudFunctionCreate creates a cloud function via the Duplo API.
func (c *Client) GcpCloudFunctionCreate(tenantID string, rq *DuploGcpCloudFunction) (*DuploGcpCloudFunction, ClientError) {
	rp := DuploGcpCloudFunction{}
	err := c.postAPI(
		fmt.Sprintf("GcpCloudFunctionCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/google/function", tenantID),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// GcpCloudFunctionUpdate updates a cloud function via the Duplo API.
func (c *Client) GcpCloudFunctionUpdate(tenantID string, rq *DuploGcpCloudFunction) (*DuploGcpCloudFunction, ClientError) {
	rp := DuploGcpCloudFunction{}
	err := c.putAPI(
		fmt.Sprintf("GcpCloudFunctionUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/google/function/%s", tenantID, rq.Name),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// GcpCloudFunctionDelete deletes a cloud function via the Duplo API.
func (c *Client) GcpCloudFunctionDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("GcpCloudFunctionDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/function/%s", tenantID, name),
		nil)
}

// GcpCloudFunctionGetList gets a list of cloud functions via the Duplo API.
func (c *Client) GcpCloudFunctionGetList(tenantID string) (*[]DuploGcpCloudFunction, ClientError) {
	prefix, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the list from Duplo
	list := []DuploGcpCloudFunction{}
	err = c.getAPI(
		fmt.Sprintf("GcpCloudFunctionGetList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/function", tenantID),
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

// GcpCloudFunctionGet gets a cloud function via the Duplo API.
func (c *Client) GcpCloudFunctionGet(tenantID string, name string) (*DuploGcpCloudFunction, ClientError) {

	// Get the list from Duplo
	rp := DuploGcpCloudFunction{}
	err := c.getAPI(
		fmt.Sprintf("GcpCloudFunctionGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/function/%s", tenantID, name),
		&rp)
	rp.TenantID = tenantID
	rp.ShortName = name
	return &rp, err
}
