package duplosdk

import (
	"fmt"
	"net/url"
)

type DuploAWSTag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// The ARN and key segments need different numbers of escape passes because
// the backend's route binders treat them differently: the ARN wildcard
// segment goes through three decode layers (route binder + explicit decode
// in the controller), while the key regular-parameter segment only goes
// through two.
func escapeAWSTagARN(s string) string {
	return url.QueryEscape(url.QueryEscape(url.QueryEscape(s)))
}

func escapeAWSTagKey(s string) string {
	return url.QueryEscape(url.QueryEscape(s))
}

func (c *Client) CreateAWSTag(tenantId, arn string, rq *DuploAWSTag) ClientError {
	var rp interface{}
	err := c.postAPI(
		fmt.Sprintf("CreateAWSTag(%s, %s)", tenantId, arn),
		fmt.Sprintf("v3/subscriptions/%s/aws/tags/arn/%s", tenantId, escapeAWSTagARN(arn)),
		&rq,
		&rp,
	)
	return err
}

func (c *Client) GetAWSTag(tenantId, arn, key string) (*DuploAWSTag, ClientError) {
	rp := DuploAWSTag{}
	err := c.getAPI(
		fmt.Sprintf("GetAWSTag(%s, %s,%s)", tenantId, arn, key),
		fmt.Sprintf("v3/subscriptions/%s/aws/tags/arn/%s/%s", tenantId, escapeAWSTagARN(arn), escapeAWSTagKey(key)),
		&rp,
	)
	return &rp, err
}

func (c *Client) DeleteAWSTag(tenantId, arn, key string) ClientError {
	err := c.deleteAPI(
		fmt.Sprintf("DeleteAWSTag(%s, %s,%s)", tenantId, arn, key),
		fmt.Sprintf("v3/subscriptions/%s/aws/tags/arn/%s/%s", tenantId, escapeAWSTagARN(arn), escapeAWSTagKey(key)),
		nil,
	)
	return err
}

func (c *Client) UpdateAWSTag(tenantId, arn, key string, rq *DuploAWSTag) ClientError {
	var rp interface{}

	err := c.putAPI(
		fmt.Sprintf("UpdateAWSTag(%s, %s,%s)", tenantId, arn, key),
		fmt.Sprintf("v3/subscriptions/%s/aws/tags/arn/%s/%s", tenantId, escapeAWSTagARN(arn), escapeAWSTagKey(key)),
		&rq,
		&rp,
	)
	return err
}
