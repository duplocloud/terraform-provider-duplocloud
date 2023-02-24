package duplosdk

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// DuploEcsServiceLbConfig is a Duplo SDK object that represents load balancer configuration for an ECS service
type DuploEcsServiceLbConfig struct {
	ReplicationControllerName string                              `json:"ReplicationControllerName"`
	Port                      string                              `json:"Port,omitempty"`
	BackendProtocol           string                              `json:"BeProtocolVersion,omitempty"`
	ExternalPort              int                                 `json:"ExternalPort,omitempty"`
	Protocol                  string                              `json:"Protocol,omitempty"`
	IsInternal                bool                                `json:"IsInternal,omitempty"`
	HealthCheckURL            string                              `json:"HealthCheckUrl,omitempty"`
	CertificateArn            string                              `json:"CertificateArn,omitempty"`
	LbType                    int                                 `json:"LbType,omitempty"`
	TgCount                   int                                 `json:"TgCount,omitempty"`
	HealthCheckConfig         *DuploEcsServiceLbHealthCheckConfig `json:"HealthCheckConfig,omitempty"`
	LbIndex                   int                                 `json:"LbIndex"`
	IdleTimeout               int                                 `json:"IdleTimeout,omitempty"`
}

type DuploEcsServiceLbHealthCheckConfig struct {
	HealthyThresholdCount      int    `json:"HealthyThresholdCount,omitempty"`
	UnhealthyThresholdCount    int    `json:"UnhealthyThresholdCount,omitempty"`
	HealthCheckTimeoutSeconds  int    `json:"HealthCheckTimeoutSeconds,omitempty"`
	HealthCheckIntervalSeconds int    `json:"HealthCheckIntervalSeconds,omitempty"`
	HttpSuccessCode            string `json:"HttpSuccessCode,omitempty"`
	GrpcSuccessCode            string `json:"GrpcSuccessCode,omitempty"`
}

// DuploEcsService is a Duplo SDK object that represents an ECS service
type DuploEcsServiceCapacityProviderStrategy struct {
	CapacityProvider string `json:"CapacityProvider,omitempty"`
	Weight           int    `json:"Weight"`
	Base             int    `json:"Base"`
}

type DuploEcsService struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-"`

	Name                          string                                     `json:"Name"`
	TaskDefinition                string                                     `json:"TaskDefinition,omitempty"`
	Replicas                      int                                        `json:"Replicas,omitempty"`
	HealthCheckGracePeriodSeconds int                                        `json:"HealthCheckGracePeriodSeconds,omitempty"`
	OldTaskDefinitionBufferSize   int                                        `json:"OldTaskDefinitionBufferSize,omitempty"`
	IsTargetGroupOnly             bool                                       `json:"IsTargetGroupOnly,omitempty"`
	DNSPrfx                       string                                     `json:"DnsPrfx,omitempty"`
	LBConfigurations              *[]DuploEcsServiceLbConfig                 `json:"LBConfigurations,omitempty"`
	UseIndexForLb                 bool                                       `json:"UseIndexForLb"`
	Index                         int                                        `json:"Index"`
	CapacityProviderStrategy      *[]DuploEcsServiceCapacityProviderStrategy `json:"CapacityProviderStrategy,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// EcsServiceList lists all ECS services.
func (c *Client) EcsServiceList(tenantID string) (*[]DuploEcsService, ClientError) {
	rp := []DuploEcsService{}

	err := c.getAPI(
		fmt.Sprintf("EcsServiceList(%s)", tenantID),
		fmt.Sprintf("subscriptions/%s/GetEcsServices", tenantID),
		&rp,
	)
	if err != nil {
		return nil, err
	}

	// Fill in the tenant ID.
	for i := range rp {
		rp[i].TenantID = tenantID
	}
	return &rp, nil
}

// EcsServiceGet retrieves a single ECS service.
func (c *Client) EcsServiceGet(tenantID, name string) (*DuploEcsService, ClientError) {
	allResources, err := c.EcsServiceList(tenantID)
	if err != nil {
		return nil, err
	}

	// Find and return the resource with the specific type and name.
	for _, resource := range *allResources {
		if resource.Name == name {
			return &resource, nil
		}
	}

	// No resource was found.
	return nil, nil
}

// EcsServiceCreate creates an ECS service via the Duplo API.
func (c *Client) EcsServiceCreate(tenantID string, duploObject *DuploEcsService) (*DuploEcsService, ClientError) {
	return c.EcsServiceCreateOrUpdate(tenantID, duploObject, false)
}

// EcsServiceUpdate updates an ECS service via the Duplo API.
func (c *Client) EcsServiceUpdate(tenantID string, duploObject *DuploEcsService) (*DuploEcsService, ClientError) {
	return c.EcsServiceCreateOrUpdate(tenantID, duploObject, true)
}

// EcsServiceCreateOrUpdate creates or updates an ECS service via the Duplo API.
func (c *Client) EcsServiceCreateOrUpdate(tenantID string, rq *DuploEcsService, updating bool) (*DuploEcsService, ClientError) {
	var err ClientError
	// Build the request
	verb := "POST"
	if updating {
		verb = "PUT"
	}

	// Call the API.
	rp := DuploEcsService{}
	if updating {
		err = c.doAPIWithRequestBody(
			verb,
			fmt.Sprintf("EcsServiceUpdate(%s, %s)", tenantID, rq.Name),
			fmt.Sprintf("v3/subscriptions/%s/aws/ecsService", tenantID),
			&rq,
			&rp,
		)
		if err != nil && err.PossibleMissingAPI() {

			// There is no "full" fallback for this API being missing - but older Duplos could at least update other service parameters.
			log.Printf(
				"[WARN] Remote duplo (%s) does not support updating LB configs for ECS.  Please contact your Duplo Administrator. (tenant=%s, name=%s)",
				c.HostURL,
				tenantID,
				rq.Name,
			)

			err = c.doAPIWithRequestBody(
				verb,
				fmt.Sprintf("EcsServiceUpdate(%s, %s)", tenantID, rq.Name),
				fmt.Sprintf("v2/subscriptions/%s/EcsServiceApiV2", tenantID),
				&rq,
				&rp,
			)
		}
	} else {
		err = c.doAPIWithRequestBody(
			verb,
			fmt.Sprintf("EcsServiceCreate(%s, %s)", tenantID, rq.Name),
			fmt.Sprintf("v2/subscriptions/%s/EcsServiceApiV2", tenantID),
			&rq,
			&rp,
		)
	}

	if err != nil {
		return nil, err
	}
	rp.TenantID = tenantID
	return &rp, err
}

// EcsServiceDelete deletes an ECS service via the Duplo API.
func (c *Client) EcsServiceDelete(id string) ClientError {
	idParts := strings.SplitN(id, "/", 5)
	tenantID := idParts[2]
	name := idParts[4]

	// Delete the ECS service
	return c.deleteAPI(
		fmt.Sprintf("EcsServiceDelete(%s, %s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/EcsServiceApiV2/%s", tenantID, name),
		nil)
}

// EcsServiceGetV2 retrieves an ECS service via the Duplo API.
func (c *Client) EcsServiceGetV2(id string) (*DuploEcsService, ClientError) {
	idParts := strings.SplitN(id, "/", 5)
	tenantID := idParts[2]
	name := idParts[4]

	// Retrieve the object.
	duploObject := DuploEcsService{}
	err := c.getAPI(
		fmt.Sprintf("EcsServiceGet(%s, %s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/EcsServiceApiV2/%s", tenantID, name),
		&duploObject)
	if err != nil || duploObject.Name == "" {
		return nil, err
	}

	// Fill in the tenant ID and return the object
	duploObject.TenantID = tenantID
	return &duploObject, nil
}

// EcsServiceGetTargetGroups retrieves an ECS service via the Duplo API.
func (c *Client) EcsServiceRequiredTargetGroupsCreated(tenantID string, ecs *DuploEcsService) (bool, ClientError, []string) {
	log.Printf("[TRACE] EcsServiceRequiredTargetGroupsCreated ******** start")
	targetGrpCount := 0
	tenant, err := c.TenantGet(tenantID)
	if err != nil {
		return false, err, nil
	}
	tenantFeatures, err := c.TenantFeaturesGet(tenantID)
	if err != nil {
		return false, err, nil
	}
	// Prepare taget group names
	log.Printf("[TRACE] Tenant(UseLbIndex) : %v, ECS(UseIndexForLb) : %v.", tenantFeatures.UseLbIndex, ecs.UseIndexForLb)
	isShortTargetGrpNames := tenantFeatures.UseLbIndex || ecs.UseIndexForLb
	tagetGrpNames := []string{}
	for _, lbc := range *ecs.LBConfigurations {
		targetGrpCount = lbc.TgCount + targetGrpCount
		for tg := 1; tg <= lbc.TgCount; tg++ {
			if isShortTargetGrpNames {
				if tg > 1 {
					tagetGrpNames = append(tagetGrpNames, strings.Join([]string{"duplo2", tenant.AccountName, strconv.Itoa(ecs.Index), strconv.Itoa(lbc.LbIndex), strconv.Itoa(tg)}, "-"))
				} else {
					tagetGrpNames = append(tagetGrpNames, strings.Join([]string{"duplo2", tenant.AccountName, strconv.Itoa(ecs.Index), strconv.Itoa(lbc.LbIndex)}, "-"))
				}
			} else {
				if tg > 1 {
					tagetGrpNames = append(tagetGrpNames, strings.Join([]string{"duplo2", tenant.AccountName, ecs.Name, lbc.Protocol + lbc.Port, strconv.Itoa(tg)}, "-"))
				} else {
					tagetGrpNames = append(tagetGrpNames, strings.Join([]string{"duplo2", tenant.AccountName, ecs.Name, lbc.Protocol + lbc.Port}, "-"))
				}
			}
		}
	}
	log.Printf("[TRACE] Total %v target groups to be created for ECS service %s.", targetGrpCount, ecs.Name)
	log.Printf("[TRACE] Target groups names : %s.", tagetGrpNames)

	targetGroupArns := make([]string, 0, targetGrpCount)
	targetGroups, err := c.TenantListApplicationLbTargetGroups(tenantID)

	if err != nil {
		return false, err, nil
	}
	// Check if all required target groups are created.
	counter := 0
	if targetGroups != nil && tagetGrpNames != nil {
		for _, tg := range *targetGroups {
			for _, t := range tagetGrpNames {
				if strings.EqualFold(tg.TargetGroupName, t) {
					counter++
					targetGroupArns = append(targetGroupArns, tg.TargetGroupArn)
				}
			}
		}
		log.Printf("[TRACE] Currently %v target groups are created for ECS service %s.", counter, ecs.Name)
		if counter == targetGrpCount {
			log.Printf("[TRACE] Total %v target groups are created for ECS service %s.", targetGrpCount, ecs.Name)
			log.Printf("[TRACE] EcsServiceRequiredTargetGroupsCreated ******** end")
			return true, nil, targetGroupArns
		}
	}
	log.Printf("[TRACE] EcsServiceRequiredTargetGroupsCreated ******** end")
	return false, nil, targetGroupArns
}
