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

type Client struct {
	HTTPClient *http.Client
	HostURL    string
	Token      string
	//Api        string
	//TenantId   string
}

func NewClient(host, token string) (*Client, error) {
	if (host != "") && (token != "") {
		token_bearer := fmt.Sprintf("Bearer %s", token)
		c := Client{
			HTTPClient: &http.Client{Timeout: 20 * time.Second},
			HostURL:    host,
			Token:      token_bearer,
		}
		return &c, nil
	}
	return nil, fmt.Errorf("Missing provider config for 'duplo_token' 'duplo_host'. Not defined in environment var / main.tf.")
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
		err_msg := fmt.Errorf("status: %d, body: %s. Please verify object exists in duplocloud. %s", res.StatusCode, body, req.URL.String())
		log.Printf("[TRACE] duplo-doRequest ********: %s", err_msg)
		return nil, err_msg
	}
	if res.StatusCode == 404 {
		err_msg := fmt.Errorf("status: %d, body: %s. Please verify object exists in duplocloud. %s", res.StatusCode, body, req.URL.String())
		log.Printf("[TRACE] duplo-doRequest ********: %s", err_msg)
		return nil, err_msg
	}
	//everything other than 200
	if res.StatusCode != http.StatusOK {
		err_msg := fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
		log.Printf("[TRACE] duplo-doRequest ********: %s", err_msg)
		return nil, err_msg
	}

	return body, err
}

func (c *Client) doRequestWithStatus(req *http.Request, status_code int) ([]byte, error) {
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

	if res.StatusCode != status_code {
		err_msg := fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
		log.Printf("[TRACE] duplo-doRequestWithStatus ********: %s", err_msg)
		return nil, err_msg
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
		err_msg := fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
		log.Printf("[TRACE] duplo-doPostRequest ********: %s %s", caller, err_msg)
		return nil, err_msg
	}

	return body, err
}

// helpers
func (c *Client) StructToString(structObj []map[string]interface{}) string {
	if structObj != nil {
		tags, err := json.Marshal(structObj)
		if err == nil {
			return string(tags)
		}
	}
	return ""
}

func (c *Client) GetId(d *schema.ResourceData, id_key string) string {
	var id = d.Id()
	if id == "" {
		id = d.Get(id_key).(string)
	}
	return id
}
func (c *Client) GetIdForChild(d *schema.ResourceData) []string {
	var ids = d.Id()
	if ids != "" {
		has_childs := strings.Index(ids, "/")
		if has_childs != -1 {
			id_array := strings.Split(ids, "/")
			return id_array
		}
	}
	return nil
}
