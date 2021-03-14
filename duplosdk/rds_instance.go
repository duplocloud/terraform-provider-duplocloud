package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DuploRdsInstance is a Duplo SDK object that represents an RDS instance
type DuploRdsInstance struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name"`

	Identifier                  string `json:"Identifier"`
	Arn                         string `json:"Arn"`
	Endpoint                    string `json:"Endpoint,omitempty"`
	MasterUsername              string `json:"MasterUsername,omitempty"`
	MasterPassword              string `json:"MasterPassword,omitempty"`
	Engine                      int    `json:"Engine,omitempty"`
	EngineVersion               string `json:"EngineVersion,omitempty"`
	SnapshotID                  string `json:"SnapshotId,omitempty"`
	DBParameterGroupName        string `json:"DBParameterGroupName,omitempty"`
	StoreDetailsInSecretManager bool   `json:"StoreDetailsInSecretManager,omitempty"`
	Cloud                       int    `json:"Cloud,omitempty"`
	SizeEx                      string `json:"SizeEx,omitempty"`
	EncryptStorage              bool   `json:"EncryptStorage,omitempty"`
	InstanceStatus              string `json:"InstanceStatus,omitempty"`
}

// DuploRdsInstancePasswordChange is a Duplo SDK object that represents an RDS instance password change
type DuploRdsInstancePasswordChange struct {
	Identifier     string `json:"Identifier"`
	MasterPassword string `json:"MasterPassword"`
	StorePassword  bool   `json:"StorePassword,omitempty"`
}

// DuploRdsInstanceSchema returns a Terraform resource schema for an ECS Service
func DuploRdsInstanceSchema() *map[string]*schema.Schema {
	return &map[string]*schema.Schema{
		"tenant_id": {
			Type:     schema.TypeString,
			Optional: false,
			Required: true,
			ForceNew: true, //switch tenant
		},
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"identifier": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"endpoint": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"host": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"port": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"master_username": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"master_password": {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
		},
		"engine": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"engine_version": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		"snapshot_id": {
			Type:          schema.TypeString,
			Optional:      true,
			ForceNew:      true,
			ConflictsWith: []string{"master_username"},
		},
		"parameter_group_name": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"store_details_in_secret_manager": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		},
		"cloud": {
			Type:     schema.TypeInt,
			Optional: true,
			ForceNew: true,
			Default:  0,
		},
		"size": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"encrypt_storage": {
			Type:     schema.TypeBool,
			Optional: true,
			ForceNew: true,
		},
		"instance_status": {
			Type:     schema.TypeString,
			Computed: true,
		},
	}
}

/*************************************************
 * API CALLS to duplo
 */

// RdsInstanceCreate creates an ECS service via the Duplo API.
func (c *Client) RdsInstanceCreate(tenantID string, duploObject *DuploRdsInstance) (*DuploRdsInstance, error) {
	return c.RdsInstanceCreateOrUpdate(tenantID, duploObject, false)
}

// RdsInstanceUpdate updates an ECS service via the Duplo API.
func (c *Client) RdsInstanceUpdate(tenantID string, duploObject *DuploRdsInstance) (*DuploRdsInstance, error) {
	return c.RdsInstanceCreateOrUpdate(tenantID, duploObject, true)
}

// RdsInstanceCreateOrUpdate creates or updates an RDS instance via the Duplo API.
func (c *Client) RdsInstanceCreateOrUpdate(tenantID string, duploObject *DuploRdsInstance, updating bool) (*DuploRdsInstance, error) {

	// Build the request
	verb := "POST"
	if updating {
		verb = "PUT"
	}

	// Call the API.
	rp := DuploRdsInstance{}
	err := c.doAPIWithRequest(
		verb,
		fmt.Sprintf("RdsInstanceCreateOrUpdate(%s, duplo%s)", tenantID, duploObject.Name),
		fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance", tenantID),
		&duploObject,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// RdsInstanceDelete deletes an RDS instance via the Duplo API.
func (c *Client) RdsInstanceDelete(id string) (*DuploRdsInstance, error) {
	idParts := strings.SplitN(id, "/", 5)
	tenantID := idParts[2]
	name := idParts[4]

	// Call the API.
	duploObject := DuploRdsInstance{}
	err := c.deleteAPI(
		fmt.Sprintf("RdsInstanceDelete(%s, duplo%s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/duplo%s", tenantID, name),
		&duploObject)
	if err != nil {
		return nil, err
	}

	// Tolerate an empty response from the DELETE.
	if duploObject.Name == "" {
		duploObject.Name = name
	}

	// Fill in the tenant ID and return the object
	duploObject.TenantID = tenantID
	return &duploObject, nil
}

// RdsInstanceGet retrieves an RDS instance via the Duplo API.
func (c *Client) RdsInstanceGet(id string) (*DuploRdsInstance, error) {
	idParts := strings.SplitN(id, "/", 5)
	tenantID := idParts[2]
	name := idParts[4]

	// Call the API.
	duploObject := DuploRdsInstance{}
	err := c.getAPI(
		fmt.Sprintf("RdsInstanceGet(%s, duplo%s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/RDSDBInstance/duplo%s", tenantID, name),
		&duploObject)
	if err != nil {
		return nil, err
	}

	// Fill in the tenant ID and the name and return the object
	duploObject.TenantID = tenantID
	duploObject.Name = name
	return &duploObject, nil
}

// RdsInstanceChangePassword creates or updates an RDS instance via the Duplo API.
func (c *Client) RdsInstanceChangePassword(tenantID string, duploObject DuploRdsInstancePasswordChange) error {
	// Call the API.
	return c.postAPI(
		fmt.Sprintf("RdsInstanceChangePassword(%s, %s)", tenantID, duploObject.Identifier),
		fmt.Sprintf("subscriptions/%s/RDSInstancePasswordChange", tenantID),
		&duploObject,
		nil,
	)
}

// RdsInstanceWaitUntilAvailable waits until an RDS instance is available.
//
// It should be usable both post-creation and post-modification.
func RdsInstanceWaitUntilAvailable(c *Client, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"processing", "backing-up", "backtracking", "configuring-enhanced-monitoring", "configuring-iam-database-auth", "configuring-log-exports", "creating",
			"maintenance", "modifying", "moving-to-vpc", "rebooting", "renaming",
			"resetting-master-credentials", "starting", "stopping", "storage-optimization", "upgrading",
		},
		Target:       []string{"available"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      20 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			resp, err := c.RdsInstanceGet(id)
			if err != nil {
				return 0, "", err
			}
			if resp.InstanceStatus == "" {
				resp.InstanceStatus = "processing"
			}
			return resp, resp.InstanceStatus, nil
		},
	}
	log.Printf("[DEBUG] RdsInstanceWaitUntilAvailable (%s)", id)
	_, err := stateConf.WaitForState()
	return err
}

// RdsInstanceWaitUntilUnavailable waits until an RDS instance is unavailable.
//
// It should be usable post-modification.
func RdsInstanceWaitUntilUnavailable(c *Client, id string) error {
	stateConf := &resource.StateChangeConf{
		Target: []string{
			"processing", "backing-up", "backtracking", "configuring-enhanced-monitoring", "configuring-iam-database-auth", "configuring-log-exports", "creating",
			"maintenance", "modifying", "moving-to-vpc", "rebooting", "renaming",
			"resetting-master-credentials", "starting", "stopping", "storage-optimization", "upgrading",
		},
		Pending:      []string{"available"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      20 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			resp, err := c.RdsInstanceGet(id)
			if err != nil {
				return 0, "", err
			}
			if resp.InstanceStatus == "" {
				resp.InstanceStatus = "available"
			}
			return resp, resp.InstanceStatus, nil
		},
	}
	log.Printf("[DEBUG] RdsInstanceWaitUntilUnavailable (%s)", id)
	_, err := stateConf.WaitForState()
	return err
}

/*************************************************
 * DATA CONVERSIONS to/from duplo/terraform
 */

// RdsInstanceFromState converts resource data respresenting an RDS instance to a Duplo SDK object.
func RdsInstanceFromState(d *schema.ResourceData) (*DuploRdsInstance, error) {
	duploObject := new(DuploRdsInstance)

	// First, convert things into simple scalars
	duploObject.Name = d.Get("name").(string)
	duploObject.Identifier = d.Get("identifier").(string)
	duploObject.Arn = d.Get("arn").(string)
	duploObject.Endpoint = d.Get("endpoint").(string)
	duploObject.MasterUsername = d.Get("master_username").(string)
	duploObject.MasterPassword = d.Get("master_password").(string)
	duploObject.Engine = d.Get("engine").(int)
	duploObject.EngineVersion = d.Get("engine_version").(string)
	duploObject.SnapshotID = d.Get("snapshot_id").(string)
	duploObject.DBParameterGroupName = d.Get("parameter_group_name").(string)
	duploObject.Cloud = d.Get("cloud").(int)
	duploObject.SizeEx = d.Get("size").(string)
	duploObject.EncryptStorage = d.Get("encrypt_storage").(bool)
	duploObject.InstanceStatus = d.Get("instance_status").(string)

	return duploObject, nil
}

// RdsInstanceToState converts a Duplo SDK object respresenting an RDS instance to terraform resource data.
func RdsInstanceToState(duploObject *DuploRdsInstance, d *schema.ResourceData) map[string]interface{} {
	if duploObject == nil {
		return nil
	}
	jsonData, _ := json.Marshal(duploObject)
	log.Printf("[TRACE] duplo-RdsInstanceToState ******** 1: INPUT <= %s ", jsonData)

	jo := make(map[string]interface{})

	// First, convert things into simple scalars
	jo["tenant_id"] = duploObject.TenantID
	jo["name"] = duploObject.Name
	jo["identifier"] = duploObject.Identifier
	jo["arn"] = duploObject.Arn
	jo["endpoint"] = duploObject.Endpoint
	if duploObject.Endpoint != "" {
		uriParts := strings.SplitN(duploObject.Endpoint, ":", 2)
		jo["host"] = uriParts[0]
		if len(uriParts) == 2 {
			jo["port"], _ = strconv.Atoi(uriParts[1])
		}
	}
	jo["master_username"] = duploObject.MasterUsername
	jo["master_password"] = duploObject.MasterPassword
	jo["engine"] = duploObject.Engine
	jo["engine_version"] = duploObject.EngineVersion
	jo["snapshot_id"] = duploObject.SnapshotID
	jo["parameter_group_name"] = duploObject.DBParameterGroupName
	jo["cloud"] = duploObject.Cloud
	jo["size"] = duploObject.SizeEx
	jo["encrypt_storage"] = duploObject.EncryptStorage
	jo["instance_status"] = duploObject.InstanceStatus

	jsonData2, _ := json.Marshal(jo)
	log.Printf("[TRACE] duplo-RdsInstanceToState ******** 2: OUTPUT => %s ", jsonData2)

	return jo
}
