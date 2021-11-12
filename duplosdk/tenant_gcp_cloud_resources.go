package duplosdk

import "fmt"

const (
	GcpSchedulerJob_TargetType_None                int = 0
	GcpSchedulerJob_TargetType_PubsubTarget        int = 4
	GcpSchedulerJob_TargetType_AppEngineHttpTarget int = 5
	GcpSchedulerJob_TargetType_HttpTarget          int = 6

	GcpSchedulerJob_HttpMethod_Unspecified int = 0 // per Google, defaults to POST
	GcpSchedulerJob_HttpMethod_Post        int = 1
	GcpSchedulerJob_HttpMethod_Get         int = 2
	GcpSchedulerJob_HttpMethod_Head        int = 3
	GcpSchedulerJob_HttpMethod_Put         int = 4
	GcpSchedulerJob_HttpMethod_Delete      int = 5
	GcpSchedulerJob_HttpMethod_Patch       int = 6
	GcpSchedulerJob_HttpMethod_Options     int = 7

	GcpSchedulerJob_AuthorizationHeader_None       int = 0
	GcpSchedulerJob_AuthorizationHeader_OauthToken int = 5
	GcpSchedulerJob_AuthorizationHeader_OidcToken  int = 6
)

// DuploGcpPubsubTopic represents a GCP pubsub topic resource for a Duplo tenant
type DuploGcpPubsubTopic struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The ShortName field does not come from the backend - we synthesize it
	ShortName string `json:"-"`

	Name     string            `json:"Name,omitempty"`
	SelfLink string            `json:"SelfLink,omitempty"`
	Status   string            `json:"Status,omitempty"`
	Labels   map[string]string `json:"Labels,omitempty"`
}

// DuploGcpStorageBucket represents a GCP pubsub topic resource for a Duplo tenant
type DuploGcpStorageBucket struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The ShortName field does not come from the backend - we synthesize it
	ShortName string `json:"-"`

	Name              string            `json:"Name,omitempty"`
	SelfLink          string            `json:"SelfLink,omitempty"`
	Status            string            `json:"Status,omitempty"`
	EnableVersioning  bool              `json:"EnableVersioning,omitempty"`
	AllowPublicAccess bool              `json:"AllowPublicAccess,omitempty"`
	Labels            map[string]string `json:"Labels,omitempty"`
}

// DuploGcpCloudFunction represents a GCP cloud function resource for a Duplo tenant
type DuploGcpCloudFunction struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The ShortName field does not come from the backend - we synthesize it
	ShortName string `json:"-"`

	Name                      string                             `json:"Name,omitempty"`
	SelfLink                  string                             `json:"SelfLink,omitempty"`
	Status                    string                             `json:"Status,omitempty"`
	Labels                    map[string]string                  `json:"Labels,omitempty"`
	BuildID                   string                             `json:"BuildId,omitempty"`
	VersionID                 int                                `json:"VersionId,omitempty"`
	EntryPoint                string                             `json:"EntryPoint,omitempty"`
	Runtime                   string                             `json:"Runtime,omitempty"`
	Description               string                             `json:"Description,omitempty"`
	AvailableMemoryMB         int                                `json:"AvailableMemoryMb,omitempty"`
	BuildEnvironmentVariables map[string]string                  `json:"BuildEnvironmentVariables,omitempty"`
	EnvironmentVariables      map[string]string                  `json:"EnvironmentVariables,omitempty"`
	Timeout                   int                                `json:"Timeout,omitempty"`
	SourceArchiveURL          string                             `json:"SourceArchiveUrl,omitempty"`
	AllowUnauthenticated      bool                               `json:"AllowUnauthenticated"`
	RequireHTTPS              bool                               `json:"RequireHttps"`
	IngressType               int                                `json:"IngressType,omitempty"`
	VPCNetworkingType         int                                `json:"VpcNetworkingType,omitempty"`
	TriggerType               int                                `json:"TriggerType,omitempty"`
	HTTPSTrigger              *DuploGcpCloudFunctionHTTPSTrigger `json:"HttpsTrigger,omitempty"`
	EventTrigger              *DuploGcpCloudFunctionEventTrigger `json:"EventTrigger,omitempty"`
}

// DuploGcpCloudFunctionHTTPSTrigger represents a GCP cloud function resource for a Duplo tenant
type DuploGcpCloudFunctionHTTPSTrigger struct {
	SecurityLevel string `json:"securityLevel,omitempty"`
	URL           string `json:"url,omitempty"`
}

// DuploGcpCloudFunctionEventTrigger represents a GCP cloud function resource for a Duplo tenant
type DuploGcpCloudFunctionEventTrigger struct {
	EventType string `json:"EventType,omitempty"`
	Resource  string `json:"Resource,omitempty"`
	Service   string `json:"Service,omitempty"`
}

// DuploGcpSchedulerJob represents a GCP scheduler job resource for a Duplo tenant
type DuploGcpSchedulerJob struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	// NOTE: The ShortName field does not come from the backend - we synthesize it
	ShortName string `json:"-"`

	Name              string                               `json:"Name,omitempty"`
	SelfLink          string                               `json:"SelfLink,omitempty"`
	Status            string                               `json:"Status,omitempty"`
	Description       string                               `json:"Description,omitempty"`
	Schedule          string                               `json:"Schedule,omitempty"`
	TimeZone          string                               `json:"TimeZone,omitempty"`
	AttemptDeadline   string                               `json:"AttemptDeadline,omitempty"`
	TargetType        int                                  `json:"TargetType,omitempty"`
	PubsubTarget      *DuploGcpSchedulerJobPubsubTarget    `json:"PubSubTarget,omitempty"`
	PubsubTargetData  string                               `json:"PubSubTargetData"`
	HTTPTarget        *DuploGcpSchedulerJobHTTPTarget      `json:"HttpTarget,omitempty"`
	AppEngineTarget   *DuploGcpSchedulerJobAppEngineTarget `json:"AppEngineTarget,omitempty"`
	AnyHTTPTargetBody string                               `json:"AnyHttpTargetBody"`

	// Only used for create/update request body.
	PubsubTargetAttributes map[string]string `json:"PubSubTargetAttributes"`
	AnyHTTPTargetHeaders   map[string]string `json:"AnyHttpTargetHeaders"`
}

// DuploGcpSchedulerJobPubsubTarget represents a GCP scheduler job pubsub target for a Duplo tenant
type DuploGcpSchedulerJobPubsubTarget struct {
	TopicName  string            `json:"TopicName,omitempty"`
	Attributes map[string]string `json:"Attributes,omitempty"`
}

// DuploGcpSchedulerJobHTTPTarget represents a GCP scheduler job http target for a Duplo tenant
type DuploGcpSchedulerJobHTTPTarget struct {
	HTTPMethod              int                             `json:"HttpMethod,omitempty"`
	Headers                 map[string]string               `json:"Headers,omitempty"`
	URI                     string                          `json:"Uri,omitempty"`
	OidcToken               *DuploGcpSchedulerJobOidcToken  `json:"OidcToken"`
	OAuthToken              *DuploGcpSchedulerJobOAuthToken `json:"OAuthToken"`
	AuthorizationHeaderCase int                             `json:"AuthorizationHeaderCase,omitempty"`
}

// DuploGcpSchedulerJobAppEngineTarget represents a GCP scheduler job app engine target for a Duplo tenant
type DuploGcpSchedulerJobAppEngineTarget struct {
	HTTPMethod       int                                   `json:"HttpMethod,omitempty"`
	Headers          map[string]string                     `json:"Headers,omitempty"`
	RelativeURI      string                                `json:"RelativeUri,omitempty"`
	AppEngineRouting *DuploGcpSchedulerJobAppEngineRouting `json:"AppEngineRouting,omitempty"`
}

// DuploGcpSchedulerJobOidcToken represents a GCP scheduler job OIDC token for a Duplo tenant
type DuploGcpSchedulerJobOidcToken struct {
	Audience            string `json:"Audience"`
	ServiceAccountEmail string `json:"ServiceAccountEmail,omitempty"`
}

// DuploGcpSchedulerJobOAuthToken represents a GCP scheduler job OAuth token for a Duplo tenant
type DuploGcpSchedulerJobOAuthToken struct {
	Scope               string `json:"Scope"`
	ServiceAccountEmail string `json:"ServiceAccountEmail,omitempty"`
}

// DuploGcpSchedulerJobAppEngineRouting represents a GCP scheduler job app engine routing for a Duplo tenant
type DuploGcpSchedulerJobAppEngineRouting struct {
	Host     string `json:"Host,omitempty"`
	Instance string `json:"Instance,omitempty"`
	Service  string `json:"Service,omitempty"`
	Version  string `json:"Version,omitempty"`
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

// GcpSchedulerJobCreate creates a scheduler job via the Duplo API.
func (c *Client) GcpSchedulerJobCreate(tenantID string, rq *DuploGcpSchedulerJob) (*DuploGcpSchedulerJob, ClientError) {
	rp := DuploGcpSchedulerJob{}
	err := c.postAPI(
		fmt.Sprintf("GcpSchedulerJobCreate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/google/schedulerJob", tenantID),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// GcpSchedulerJobUpdate updates a scheduler job via the Duplo API.
func (c *Client) GcpSchedulerJobUpdate(tenantID string, rq *DuploGcpSchedulerJob) (*DuploGcpSchedulerJob, ClientError) {
	rp := DuploGcpSchedulerJob{}
	err := c.putAPI(
		fmt.Sprintf("GcpSchedulerJobUpdate(%s, %s)", tenantID, rq.Name),
		fmt.Sprintf("v3/subscriptions/%s/google/schedulerJob/%s", tenantID, rq.Name),
		&rq,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// GcpSchedulerJobDelete deletes a scheduler job via the Duplo API.
func (c *Client) GcpSchedulerJobDelete(tenantID, name string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("GcpSchedulerJobDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/schedulerJob/%s", tenantID, name),
		nil)
}

// GcpSchedulerJobGetList gets a list of scheduler jobs via the Duplo API.
func (c *Client) GcpSchedulerJobGetList(tenantID string) (*[]DuploGcpSchedulerJob, ClientError) {
	prefix, err := c.GetDuploServicesPrefix(tenantID)
	if err != nil {
		return nil, err
	}

	// Get the list from Duplo
	list := []DuploGcpSchedulerJob{}
	err = c.getAPI(
		fmt.Sprintf("GcpSchedulerJobGetList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/schedulerJob", tenantID),
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

// GcpSchedulerJobGet gets a scheduler job via the Duplo API.
func (c *Client) GcpSchedulerJobGet(tenantID string, name string) (*DuploGcpSchedulerJob, ClientError) {

	// Get the list from Duplo
	rp := DuploGcpSchedulerJob{}
	err := c.getAPI(
		fmt.Sprintf("GcpSchedulerJobGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v3/subscriptions/%s/google/schedulerJob/%s", tenantID, name),
		&rp)
	rp.TenantID = tenantID
	rp.ShortName = name
	return &rp, err
}
