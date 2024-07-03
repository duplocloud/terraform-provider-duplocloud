package duplosdk

import (
	"fmt"
)

type DuploApiGatewayEvent struct {
	APIGatewayID      string                           `json:"ApiGatewayId"`
	Path              string                           `json:"Path,omitempty"`
	Method            string                           `json:"Method,omitempty"`
	Cors              bool                             `json:"Cors,omitempty"`
	ApiKeyRequired    bool                             `json:"ApiKeyRequired,omitempty"`
	AuthorizerId      string                           `json:"AuthorizerId,omitempty"`
	AuthorizationType string                           `json:"AuthorizationType,omitempty"`
	Integration       *DuploApiGatewayEventIntegration `json:"Integration"`
}

type DuploApiGatewayEventIntegration struct {
	Type    string `json:"Type,omitempty"`
	URI     string `json:"Uri,omitempty"`
	Timeout int    `json:"Timeout,omitempty"`
}

func (c *Client) ApiGatewayEventCreate(tenantID string, rq *DuploApiGatewayEvent) (*DuploApiGatewayEvent, ClientError) {
	rp := DuploApiGatewayEvent{}
	err := c.postAPI(
		fmt.Sprintf("ApiGatewayEventCreate(%s, %s)", tenantID, rq.APIGatewayID+"--"+rq.Method+"--"+rq.Path),
		fmt.Sprintf("v3/subscriptions/%s/aws/apigateway/events", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) ApiGatewayEventUpdate(tenantID string, rq *DuploApiGatewayEvent) (*DuploApiGatewayEvent, ClientError) {
	rp := DuploApiGatewayEvent{}
	err := c.putAPI(
		fmt.Sprintf("ApiGatewayEventUpdate(%s, %s)", tenantID, rq.APIGatewayID+"--"+rq.Method+"--"+rq.Path),
		fmt.Sprintf("v3/subscriptions/%s/aws/apigateway/events/%s/%s/%s", tenantID, rq.APIGatewayID, rq.Method, EncodePathParam(rq.Path)),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) ApiGatewayEventDelete(tenantID, apiGatewayID, method, path string) ClientError {
	return c.deleteAPI(
		fmt.Sprintf("ApiGatewayEventDelete(%s, %s)", tenantID, apiGatewayID+"--"+method+"--"+path),
		fmt.Sprintf("v3/subscriptions/%s/aws/apigateway/events/%s/%s/%s", tenantID, apiGatewayID, method, EncodePathParam(path)),
		nil)
}

func (c *Client) ApiGatewayEventGet(tenantID string, apiGatewayID, method, path string) (*DuploApiGatewayEvent, ClientError) {
	rp := DuploApiGatewayEvent{}
	err := c.getAPI(
		fmt.Sprintf("ApiGatewayEventGet(%s, %s)", tenantID, apiGatewayID+"--"+method+"--"+path),
		fmt.Sprintf("v3/subscriptions/%s/aws/apigateway/events/%s/%s/%s", tenantID, apiGatewayID, method, EncodePathParam(path)),
		&rp)

	return &rp, err
}

func (c *Client) ApiGatewayEventList(tenantID, apiGatewayID string) (*[]DuploApiGatewayEvent, ClientError) {
	var list []DuploApiGatewayEvent
	err := c.getAPI(
		fmt.Sprintf("ApiGatewayEventList(%s)", tenantID),
		fmt.Sprintf("v3/subscriptions/%s/aws/apigateway/events/%s", tenantID, apiGatewayID),
		&list)
	if err != nil {
		return nil, err
	}

	return &list, nil
}
