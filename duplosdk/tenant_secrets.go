package duplosdk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// DuploTenantSecret represents a managed secret for a Duplo tenant
type DuploTenantSecret struct {
	// NOTE: The TenantID field does not come from the backend - we synthesize it
	TenantID string `json:"-,omitempty"`

	Arn                    string              `json:"ARN"`
	Name                   string              `json:"Name"`
	RotationEnabled        bool                `json:"RotationEnabled,omitempty"`
	SecretVersionsToStages map[string][]string `json:"SecretVersionsToStages,omitempty"`
	Tags                   *[]DuploKeyValue    `json:"Tags,omitempty"`
	CreatedDate            string              `json:"CreatedDate,omitempty"`
	DeletedDate            string              `json:"DeletedDate,omitempty"`
	LastAccessedDate       string              `json:"LastAccessedDate,omitempty"`
	LastChangedDate        string              `json:"LastChangedDate,omitempty"`
	LastRotatedDate        string              `json:"LastRotatedDate,omitempty"`
}

// TenantListSecrets retrieves a list of managed secrets
func (c *Client) TenantListSecrets(tenantID string) (*[]DuploTenantSecret, error) {

	// Format the URL
	url := fmt.Sprintf("%s/subscriptions/%s/ListTenantSecrets", c.HostURL, tenantID)
	log.Printf("[TRACE] duplo-TenantListSecrets 1 ********: %s ", url)

	// Get the AWS region from Duplo
	req2, _ := http.NewRequest("GET", url, nil)
	body, err := c.doRequest(req2)
	if err != nil {
		log.Printf("[TRACE] duplo-TenantListSecrets 2 ********: %s", err.Error())
		return nil, err
	}
	bodyString := string(body)
	log.Printf("[TRACE] duplo-TenantListSecrets 3 ********: %s", bodyString)

	// Return it as an object.
	duploObjects := make([]DuploTenantSecret, 0)
	err = json.Unmarshal(body, &duploObjects)
	if err != nil {
		return nil, err
	}
	log.Printf("[TRACE] duplo-TenantGetAwsCredentials 4 ********")
	for _, duploObject := range duploObjects {
		duploObject.TenantID = tenantID
	}
	return &duploObjects, nil
}
