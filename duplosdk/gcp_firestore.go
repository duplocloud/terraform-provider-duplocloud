package duplosdk

import (
	"fmt"
	"log"
)

type DuploFirestoreBody struct {
	Name                          string `json:"Name"`
	UID                           string `json:"Uid"`
	LocationId                    string `json:"LocationId"`
	VersionRetentionPeriod        string `json:"VersionRetentionPeriod,omitempty"`
	EarliestVersionTime           string `json:"EarliestVersionTime,omitempty"`
	Etag                          string `json:"Etag,omitempty"`
	Type                          string `json:"Type"`
	ConcurrencyMode               string `json:"ConcurrencyMode,omitempty"`
	PointInTimeRecoveryEnablement string `json:"PointInTimeRecoveryEnablement"`
	AppEngineIntegrationMode      string `json:"AppEngineIntegrationMode,omitempty"`
	DeleteProtectionState         string `json:"DeleteProtectionState"`
}

func (c *Client) FirestoreCreate(tenantID string, rq *DuploFirestoreBody) (*DuploFirestoreBody, ClientError) {
	log.Printf("[TRACE] \nFirestore request \n\n ******%+v\n*******", rq)
	resp := DuploFirestoreBody{}
	err := c.postAPI(
		fmt.Sprintf("FirestoreCreate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/firestore", tenantID),
		&rq,
		&resp,
	)
	return &resp, err
}

func (c *Client) FirestoreGet(tenantID string, fullname string) (*DuploFirestoreBody, ClientError) {
	rp := DuploFirestoreBody{}
	err := c.getAPI(
		fmt.Sprintf("FirestoreGet(%s, %s)", tenantID, fullname),
		fmt.Sprintf("v3/subscriptions/%s/google/firestore/%s", tenantID, fullname),
		&rp,
	)

	return &rp, err
}

func (c *Client) FirestoreList(tenantID string) (*[]DuploFirestoreBody, ClientError) {
	rp := []DuploFirestoreBody{}
	err := c.getAPI(
		fmt.Sprintf("FirestoreList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/firestore", tenantID),
		&rp,
	)

	return &rp, err
}

func (c *Client) FirestoreDelete(tenantID, fullname string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("FirestoreDelete(%s, %s)", tenantID, fullname),
		fmt.Sprintf("v3/subscriptions/%s/google/firestore/%s", tenantID, fullname),
		nil)
}

func (c *Client) FirestoreUpdate(tenantID, fullname string, rq *DuploFirestoreBody) (*DuploFirestoreBody, ClientError) {
	rp := DuploFirestoreBody{}
	err := c.putAPI(
		fmt.Sprintf("FirestoreUpdate(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/google/firestore/%s", tenantID, fullname),
		&rq,
		&rp,
	)
	return &rp, err
}
