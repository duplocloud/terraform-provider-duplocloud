package duplosdk

import (
	"fmt"
	"log"
)

type DuploAsgWarmPoolRequest struct {
	AutoScalingGroupName     string                       `json:"AutoScalingGroupName"`
	MaxGroupPreparedCapacity int                          `json:"MaxGroupPreparedCapacity"`
	MinSize                  int                          `json:"MinSize"`
	PoolState                string                       `json:"PoolState,omitempty"`
	InstanceReusePolicy      *DuploAsgInstanceReusePolicy `json:"InstanceReusePolicy,omitempty"`
}

type DuploAsgInstanceReusePolicy struct {
	ReuseOnScaleIn bool `json:"ReuseOnScaleIn"`
}

type DuploAsgWarmPoolResponse struct {
	Instances             []DuploAsgWarmPoolInstance     `json:"Instances,omitempty"`
	WarmPoolConfiguration *DuploAsgWarmPoolConfiguration `json:"WarmPoolConfiguration,omitempty"`
}

type DuploAsgWarmPoolConfiguration struct {
	InstanceReusePolicy      *DuploAsgInstanceReusePolicy `json:"InstanceReusePolicy,omitempty"`
	MaxGroupPreparedCapacity int                          `json:"MaxGroupPreparedCapacity"`
	MinSize                  int                          `json:"MinSize"`
	PoolState                *DuploStringValue            `json:"PoolState,omitempty"`
}

type DuploAsgWarmPoolInstance struct {
	AvailabilityZone     string                             `json:"AvailabilityZone,omitempty"`
	HealthStatus         string                             `json:"HealthStatus,omitempty"`
	InstanceId           string                             `json:"InstanceId,omitempty"`
	InstanceType         string                             `json:"InstanceType,omitempty"`
	LaunchTemplate       *DuploAsgWarmPoolLaunchTemplateRef `json:"LaunchTemplate,omitempty"`
	LifecycleState       *DuploStringValue                  `json:"LifecycleState,omitempty"`
	ProtectedFromScaleIn bool                               `json:"ProtectedFromScaleIn,omitempty"`
}

type DuploAsgWarmPoolLaunchTemplateRef struct {
	LaunchTemplateId   string `json:"LaunchTemplateId,omitempty"`
	LaunchTemplateName string `json:"LaunchTemplateName,omitempty"`
	Version            string `json:"Version,omitempty"`
}

// AsgWarmPoolGet retrieves the warm pool configuration for an ASG.
func (c *Client) AsgWarmPoolGet(tenantID, asgFullName string) (*DuploAsgWarmPoolResponse, ClientError) {
	log.Printf("[DEBUG] Duplo API - Get ASG Warm Pool(TenantId-%s, AsgName-%s)", tenantID, asgFullName)
	rp := DuploAsgWarmPoolResponse{}
	err := c.getAPI(fmt.Sprintf("AsgWarmPoolGet(%s, %s)", tenantID, asgFullName),
		fmt.Sprintf("v3/subscriptions/%s/aws/asg/%s/warmpool", tenantID, asgFullName),
		&rp)
	if err != nil {
		return nil, err
	}
	return &rp, nil
}

// AsgWarmPoolCreateOrUpdate creates or updates the warm pool configuration for an ASG.
func (c *Client) AsgWarmPoolCreateOrUpdate(tenantID string, rq *DuploAsgWarmPoolRequest) ClientError {
	log.Printf("[DEBUG] Duplo API - Upsert ASG Warm Pool(TenantId-%s, AsgName-%s)", tenantID, rq.AutoScalingGroupName)
	var rp interface{}
	return c.putAPI(fmt.Sprintf("AsgWarmPoolCreateOrUpdate(%s, %s)", tenantID, rq.AutoScalingGroupName),
		fmt.Sprintf("v3/subscriptions/%s/aws/asg/%s/warmpool", tenantID, rq.AutoScalingGroupName),
		rq, &rp)
}

// AsgWarmPoolDelete deletes the warm pool configuration for an ASG.
func (c *Client) AsgWarmPoolDelete(tenantID, asgFullName string) ClientError {
	log.Printf("[DEBUG] Duplo API - Delete ASG Warm Pool(TenantId-%s, AsgName-%s)", tenantID, asgFullName)
	var rp interface{}
	return c.deleteAPI(fmt.Sprintf("AsgWarmPoolDelete(%s, %s)", tenantID, asgFullName),
		fmt.Sprintf("v3/subscriptions/%s/aws/asg/%s/warmpool", tenantID, asgFullName),
		&rp)
}
