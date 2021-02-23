package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// GetAwsAccountID retrieves the AWS account ID via the Duplo API.
func (c *Client) GetAwsAccountID() (string, error) {

	// Format the URL
	url := fmt.Sprintf("%s/admin/GetAwsAccountId", c.HostURL)
	log.Printf("[TRACE] duplo-GetAwsAccountId 1 ********: %s ", url)

	// Get the AWS region from Duplo
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-GetAwsAccountId 2 ********: %s", err.Error())
		return "", err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-GetAwsAccountId 3 ********: %s", bodyString)

	// Return it as a string.
	awsAccount := ""
	err = json.Unmarshal(body, &awsAccount)
	if err != nil {
		return "", err
	}
	log.Printf("[TRACE] duplo-GetAwsAccountId 4 ********: %s", awsAccount)

	return awsAccount, nil
}
