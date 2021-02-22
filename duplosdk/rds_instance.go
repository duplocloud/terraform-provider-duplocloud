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

	Identifier string `json:"Identifier"`
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
		"identifier": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
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
	identifier := idParts[4]

	// Build the request
	url := fmt.Sprintf("%s/v2/subscriptions/%s/RDSDBInstance/duplo%s", c.HostURL, tenantID, identifier)
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
		duploObject.Identifier = identifier
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
	identifier := idParts[4]

	// Build the request
	url := fmt.Sprintf("%s/v2/subscriptions/%s/RDSDBInstance/duplo%s", c.HostURL, tenantID, identifier)
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

	// Fill in the tenant ID and return the object
	duploObject.TenantID = tenantID
	return &duploObject, nil
}

/*************************************************
 * DATA CONVERSIONS to/from duplo/terraform
 */

// RdsInstanceFromState converts resource data respresenting an RDS instance to a Duplo SDK object.
func RdsInstanceFromState(d *schema.ResourceData) (*DuploRdsInstance, error) {
	duploObject := new(DuploRdsInstance)

	// First, convert things into simple scalars
	duploObject.Identifier = d.Get("identifier").(string)

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
	jo["identifier"] = duploObject.Identifier

	jsonData2, _ := json.Marshal(jo)
	log.Printf("[TRACE] duplo-RdsInstanceToState ******** 2: OUTPUT => %s ", jsonData2)

	return jo
}
