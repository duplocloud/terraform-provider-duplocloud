package duplosdk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Client is a Duplo API client
type Client struct {
	HTTPClient *http.Client
	HostURL    string
	Token      string
	//Api        string
	//TenantId   string
}

// NewClient creates a new Duplo API client
func NewClient(host, token string) (*Client, error) {
	if (host != "") && (token != "") {
		tokenBearer := fmt.Sprintf("Bearer %s", token)
		c := Client{
			HTTPClient: &http.Client{Timeout: 20 * time.Second},
			HostURL:    host,
			Token:      tokenBearer,
		}
		return &c, nil
	}
	return nil, fmt.Errorf("missing provider config for 'duplo_token' 'duplo_host'. Not defined in environment var / main.tf")
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Authorization", c.Token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("[TRACE] duplo-doRequest ********: %s", err.Error())
		return nil, err
	}

	//allow 204
	if res.StatusCode == 200 || res.StatusCode == 204 {
		return body, err
	}

	//special case for 400/404 .. when object is deleted in backend
	if res.StatusCode == 400 {
		errMsg := fmt.Errorf("status: %d, body: %s. Please verify object exists in duplocloud. %s", res.StatusCode, body, req.URL.String())
		log.Printf("[TRACE] duplo-doRequest ********: %s", errMsg)
		return nil, errMsg
	}
	if res.StatusCode == 404 {
		errMsg := fmt.Errorf("status: %d, body: %s. Please verify object exists in duplocloud. %s", res.StatusCode, body, req.URL.String())
		log.Printf("[TRACE] duplo-doRequest ********: %s", errMsg)
		return nil, errMsg
	}
	//everything other than 200
	if res.StatusCode != http.StatusOK {
		errMsg := fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
		log.Printf("[TRACE] duplo-doRequest ********: %s", errMsg)
		return nil, errMsg
	}

	return body, err
}

func (c *Client) doRequestWithStatus(req *http.Request, statusCode int) ([]byte, error) {
	req.Header.Set("Authorization", c.Token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("[TRACE] duplo-doRequestWithStatus ********: %s", err.Error())
		return nil, err
	}

	if res.StatusCode != statusCode {
		errMsg := fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
		log.Printf("[TRACE] duplo-doRequestWithStatus ********: %s", errMsg)
		return nil, errMsg
	}

	return body, err
}

func (c *Client) doPostRequest(req *http.Request, caller string) ([]byte, error) {
	req.Header.Set("Authorization", c.Token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("[TRACE] duplo-doPostRequest ********: %s %s", caller, err.Error())
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		errMsg := fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
		log.Printf("[TRACE] duplo-doPostRequest ********: %s %s", caller, errMsg)
		return nil, errMsg
	}

	return body, err
}

// Utility method to call an API with a GET request, handling logging, etc.
func (c *Client) getAPI(apiName string, apiPath string, rp interface{}) error {

	// Build the request
	url := fmt.Sprintf("%s/%s", c.HostURL, apiPath)
	log.Printf("[TRACE] getAPI %s: prepared request: %s", apiName, url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("[TRACE] getAPI %s: cannot build request: %s", apiName, err.Error())
		return nil
	}

	// Call the API and get the response.
	body, err := c.doRequest(req)
	if err != nil {
		log.Printf("[TRACE] getAPI %s: GET failed: %s", apiName, err.Error())
		return err
	}
	bodyString := string(body)
	log.Printf("[TRACE] getAPI %s: received response: %s", apiName, bodyString)

	// Interpret the response as an object.
	err = json.Unmarshal(body, rp)
	if err != nil {
		log.Printf("[TRACE] getAPI %s: cannot unmarshal response from JSON: %s", apiName, err.Error())
		return err
	}
	return nil
}

// Utility method to call an API with a POST request, handling logging, etc.
func (c *Client) postAPI(apiName string, apiPath string, rq interface{}, rp interface{}) error {

	// Build the request
	rqBody, err := json.Marshal(rq)
	if err != nil {
		log.Printf("[TRACE] postAPI %s: cannot marshal request to JSON: %s", apiName, err.Error())
		return err
	}
	url := fmt.Sprintf("%s/%s", c.HostURL, apiPath)
	log.Printf("[TRACE] postAPI %s: prepared request: %s <= (%s)", apiName, url, rqBody)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(rqBody)))
	if err != nil {
		log.Printf("[TRACE] postAPI %s: cannot build request: %s", apiName, err.Error())
		return nil
	}

	// Call the API and get the response
	body, err := c.doPostRequest(req, apiName)
	if err != nil {
		log.Printf("[TRACE] postAPI %s: POST failed: %s", apiName, err.Error())
		return err
	}
	bodyString := string(body)
	log.Printf("[TRACE] postAPI %s: received response: %s", apiName, bodyString)

	// Check for an expected "null" response.
	if rp == nil {
		log.Printf("[TRACE] post API %s: expected null response", apiName)
		if bodyString == "null" || bodyString == "" {
			return nil
		}
		err = fmt.Errorf("post API %s: received unexpected response: %s", apiName, bodyString)
		log.Printf("[TRACE] %s", err)
		return err
	}

	// Otherwise, interpret it as an object.
	err = json.Unmarshal(body, rp)
	if err != nil {
		log.Printf("[TRACE] postAPI %s: cannot unmarshal response from JSON: %s", apiName, err.Error())
		return err
	}
	return nil
}

// StructToString converts a structure to a JSON string
func (c *Client) StructToString(structObj []map[string]interface{}) string {
	if structObj != nil {
		tags, err := json.Marshal(structObj)
		if err == nil {
			return string(tags)
		}
	}
	return ""
}

// GetID returns a terraform resource data's ID field
func (c *Client) GetID(d *schema.ResourceData, idKey string) string {
	var id = d.Id()
	if id == "" {
		id = d.Get(idKey).(string)
	}
	return id
}

// GetIDForChild returns a terraform resource data's ID field as an array of multiple components.
func (c *Client) GetIDForChild(d *schema.ResourceData) []string {
	var ids = d.Id()
	if ids != "" {
		hasChilds := strings.Index(ids, "/")
		if hasChilds != -1 {
			idArray := strings.Split(ids, "/")
			return idArray
		}
	}
	return nil
}
