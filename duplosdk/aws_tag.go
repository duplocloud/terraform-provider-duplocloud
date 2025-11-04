package duplosdk

import "fmt"

type DuploAWSTag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

func (c *Client) CreateAWSTag(tenantId, arn string, rq *DuploAWSTag) ClientError {
	var rp interface{}
	err := c.postAPI(
		fmt.Sprintf("CreateAWSTag(%s, %s)", tenantId, arn),
		fmt.Sprintf("v3/subscriptions/%s/aws/tags/arn/%s", tenantId, arn),
		&rq,
		&rp,
	)
	return err
}

func (c *Client) GetAWSTag(tenantId, arn, key string) (*DuploAWSTag, ClientError) {
	rp := DuploAWSTag{}
	err := c.getAPI(
		fmt.Sprintf("GetAWSTag(%s, %s,%s)", tenantId, arn, key),
		fmt.Sprintf("v3/subscriptions/%s/aws/tags/arn/%s/%s", tenantId, arn, key),
		&rp,
	)
	return &rp, err
}

func (c *Client) DeleteAWSTag(tenantId, arn, key string) ClientError {
	err := c.deleteAPI(
		fmt.Sprintf("DeleteAWSTag(%s, %s,%s)", tenantId, arn, key),
		fmt.Sprintf("v3/subscriptions/%s/aws/tags/arn/%s/%s", tenantId, arn, key),
		nil,
	)
	return err
}

func (c *Client) UpdateAWSTag(tenantId, arn, key string, rq *DuploAWSTag) ClientError {
	var rp interface{}

	err := c.putAPI(
		fmt.Sprintf("CreateAWSTag(%s, %s,%s)", tenantId, arn, key),
		fmt.Sprintf("v3/subscriptions/%s/aws/tags/arn/%s/%s", tenantId, arn, key),
		&rq,
		&rp,
	)
	return err
}
