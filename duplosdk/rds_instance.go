package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

// RdsInstanceCreateOrUpdate creates or updates an ECS service via the Duplo API.
func (c *Client) RdsInstanceCreateOrUpdate(tenantID string, duploObject *DuploRdsInstance, updating bool) (*DuploRdsInstance, error) {

	// Build the request
	verb := "POST"
	if updating {
		verb = "PUT"
	}
	rqBody, err := json.Marshal(&duploObject)
	if err != nil {
		log.Printf("[TRACE] RdsInstanceCreateOrUpdate 1 JSON gen : %s", err.Error())
		return nil, err
	}
	url := fmt.Sprintf("%s/v2/subscriptions/%s/RDSDBInstance", c.HostURL, tenantID)
	log.Printf("[TRACE] RdsInstanceCreate 2 : %s <= %s", url, rqBody)
	req, err := http.NewRequest(verb, url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] RdsInstanceCreateOrUpdate 3 HTTP builder : %s", err.Error())
		return nil, err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] RdsInstanceCreateOrUpdate 4 HTTP %s : %s", verb, err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] RdsInstanceCreateOrUpdate 4 HTTP RESPONSE : %s", bodyString)

	// Handle the response
	rpObject := DuploRdsInstance{}
	if bodyString == "" {
		log.Printf("[TRACE] RdsInstanceCreateOrUpdate 5 NO RESULT : %s", bodyString)
		return nil, err
	}
	err = json.Unmarshal(body, &rpObject)
	if err != nil {
		log.Printf("[TRACE] RdsInstanceCreateOrUpdate 6 JSON parse : %s", err.Error())
		return nil, err
	}
	return &rpObject, nil
}

// RdsInstanceDelete deletes an ECS service via the Duplo API.
func (c *Client) RdsInstanceDelete(id string) (*DuploRdsInstance, error) {
	idParts := strings.SplitN(id, "/", 5)
	tenantID := idParts[2]
	name := idParts[4]

	// Build the request
	url := fmt.Sprintf("%s/v2/subscriptions/%s/RDSDBInstance/duplo%s", c.HostURL, tenantID, name)
	log.Printf("[TRACE] RdsInstanceGet 1 : %s", url)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("[TRACE] RdsInstanceGet 2 HTTP builder : %s", err.Error())
		return nil, err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	bodyString := string(body)
	if err != nil {
		log.Printf("[TRACE] RdsInstanceGet 3 HTTP DELETE : %s", err.Error())
		return nil, err
	}
	log.Printf("[TRACE] RdsInstanceGet 4 HTTP RESPONSE : %s", bodyString)

	// Parse the response into a duplo object
	duploObject := DuploRdsInstance{}
	if bodyString == "" {
		// tolerate an empty response from DELETE
		duploObject.Name = name
	} else {
		err = json.Unmarshal(body, &duploObject)
		if err != nil {
			log.Printf("[TRACE] RdsInstanceGet 5 JSON PARSE : %s", bodyString)
			return nil, err
		}
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

	// Build the request
	url := fmt.Sprintf("%s/v2/subscriptions/%s/RDSDBInstance/duplo%s", c.HostURL, tenantID, name)
	log.Printf("[TRACE] RdsInstanceGet 1 : %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("[TRACE] RdsInstanceGet 2 HTTP builder : %s", err.Error())
		return nil, err
	}

	// Call the API and get the response
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] RdsInstanceGet 3 HTTP GET : %s", err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] RdsInstanceGet 4 HTTP RESPONSE : %s", bodyString)

	// Parse the response into a duplo object, detecting a missing object
	if bodyString == "null" {
		return nil, nil
	}
	duploObject := DuploRdsInstance{}
	err = json.Unmarshal(body, &duploObject)
	if err != nil {
		log.Printf("[TRACE] RdsInstanceGet 5 JSON PARSE : %s", bodyString)
		return nil, err
	}

	// Fill in the tenant ID and the name and return the object
	duploObject.TenantID = tenantID
	duploObject.Name = name
	return &duploObject, nil
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
