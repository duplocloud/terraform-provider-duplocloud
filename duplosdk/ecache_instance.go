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

// DuploEcacheInstance is a Duplo SDK object that represents an ECache instance
type DuploEcacheInstance struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	// NOTE: The Name field does not come from the backend - we synthesize it
	Name string `json:"Name"`

	Identifier          string `json:"Identifier"`
	Arn                 string `json:"Arn"`
	Endpoint            string `json:"Endpoint,omitempty"`
	CacheType           int    `json:"CacheType,omitempty"`
	Size                string `json:"Size,omitempty"`
	Replicas            int    `json:"Replicas,omitempty"`
	EncryptionAtRest    bool   `json:"EnableEncryptionAtRest,omitempty"`
	EncryptionInTransit bool   `json:"EnableEncryptionAtTransit,omitempty"`
	InstanceStatus      string `json:"InstanceStatus,omitempty"`
}

/*************************************************
 * API CALLS to duplo
 */

// EcacheInstanceCreate creates an ECache instance via the Duplo API.
func (c *Client) EcacheInstanceCreate(tenantID string, duploObject *DuploEcacheInstance) (*DuploEcacheInstance, error) {
	return c.EcacheInstanceCreateOrUpdate(tenantID, duploObject, false)
}

// EcacheInstanceUpdate updates an ECache instance via the Duplo API.
func (c *Client) EcacheInstanceUpdate(tenantID string, duploObject *DuploEcacheInstance) (*DuploEcacheInstance, error) {
	return c.EcacheInstanceCreateOrUpdate(tenantID, duploObject, true)
}

// EcacheInstanceCreateOrUpdate creates or updates an ECache instance via the Duplo API.
func (c *Client) EcacheInstanceCreateOrUpdate(tenantID string, duploObject *DuploEcacheInstance, updating bool) (*DuploEcacheInstance, error) {

	// Build the request
	verb := "POST"
	if updating {
		verb = "PUT"
	}

	// Call the API.
	rp := DuploEcacheInstance{}
	err := c.doAPIWithRequest(
		verb,
		fmt.Sprintf("EcacheInstanceCreateOrUpdate(%s, duplo-%s)", tenantID, duploObject.Name),
		fmt.Sprintf("v2/subscriptions/%s/ECacheDBInstance", tenantID),
		&duploObject,
		&rp,
	)
	if err != nil {
		return nil, err
	}
	return &rp, err
}

// EcacheInstanceDelete deletes an ECache instance via the Duplo API.
func (c *Client) EcacheInstanceDelete(id string) (*DuploEcacheInstance, error) {
	idParts := strings.SplitN(id, "/", 5)
	tenantID := idParts[2]
	name := idParts[4]

	// Call the API.
	duploObject := DuploEcacheInstance{}
	err := c.deleteAPI(
		fmt.Sprintf("EcacheInstanceDelete(%s, duplo-%s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/ECacheDBInstance/duplo-%s", tenantID, name),
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

// EcacheInstanceGet retrieves an ECache instance via the Duplo API.
func (c *Client) EcacheInstanceGet(id string) (*DuploEcacheInstance, error) {
	idParts := strings.SplitN(id, "/", 5)
	tenantID := idParts[2]
	name := idParts[4]

	// Call the API.
	duploObject := DuploEcacheInstance{}
	err := c.getAPI(
		fmt.Sprintf("EcacheInstanceGet(%s, duplo-%s)", tenantID, name),
		fmt.Sprintf("v2/subscriptions/%s/ECacheDBInstance/duplo-%s", tenantID, name),
		&duploObject)
	if err != nil {
		return nil, err
	}

	// Parse the response into a duplo object, detecting a missing object
	if duploObject.Identifier == "" {
		return nil, nil
	}

	// Fill in the tenant ID and the name and return the object
	duploObject.TenantID = tenantID
	duploObject.Name = name
	return &duploObject, nil
}

// EcacheInstanceWaitUntilAvailable waits until an ECache instance is available.
//
// It should be usable both post-creation and post-modification.
func EcacheInstanceWaitUntilAvailable(c *Client, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending:      []string{"processing", "creating", "modifying", "rebooting cluster nodes", "snapshotting"},
		Target:       []string{"available"},
		MinTimeout:   10 * time.Second,
		PollInterval: 30 * time.Second,
		Timeout:      20 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			resp, err := c.EcacheInstanceGet(id)
			if err != nil {
				return 0, "", err
			}
			if resp.InstanceStatus == "" {
				resp.InstanceStatus = "processing"
			}
			return resp, resp.InstanceStatus, nil
		},
	}
	log.Printf("[DEBUG] EcacheInstanceWaitUntilAvailable (%s)", id)
	_, err := stateConf.WaitForState()
	return err
}

/*************************************************
 * DATA CONVERSIONS to/from duplo/terraform
 */

// EcacheInstanceFromState converts resource data respresenting an ECache instance to a Duplo SDK object.
func EcacheInstanceFromState(d *schema.ResourceData) (*DuploEcacheInstance, error) {
	duploObject := new(DuploEcacheInstance)

	// First, convert things into simple scalars
	duploObject.Name = d.Get("name").(string)
	duploObject.Identifier = d.Get("identifier").(string)
	duploObject.Arn = d.Get("arn").(string)
	duploObject.Endpoint = d.Get("endpoint").(string)
	duploObject.CacheType = d.Get("cache_type").(int)
	duploObject.Size = d.Get("size").(string)
	duploObject.Replicas = d.Get("replicas").(int)
	duploObject.EncryptionAtRest = d.Get("encryption_at_rest").(bool)
	duploObject.EncryptionInTransit = d.Get("encryption_in_transit").(bool)
	duploObject.InstanceStatus = d.Get("instance_status").(string)

	return duploObject, nil
}

// EcacheInstanceToState converts a Duplo SDK object respresenting an ECache instance to terraform resource data.
func EcacheInstanceToState(duploObject *DuploEcacheInstance, d *schema.ResourceData) map[string]interface{} {
	if duploObject == nil {
		return nil
	}
	jsonData, _ := json.Marshal(duploObject)
	log.Printf("[TRACE] duplo-EcacheInstanceToState ******** 1: INPUT <= %s ", jsonData)

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
	jo["cache_type"] = duploObject.CacheType
	jo["size"] = duploObject.Size
	jo["replicas"] = duploObject.Replicas
	jo["encryption_at_rest"] = duploObject.EncryptionAtRest
	jo["encryption_in_transit"] = duploObject.EncryptionInTransit
	jo["instance_status"] = duploObject.InstanceStatus

	jsonData2, _ := json.Marshal(jo)
	log.Printf("[TRACE] duplo-EcacheInstanceToState ******** 2: OUTPUT => %s ", jsonData2)

	return jo
}
