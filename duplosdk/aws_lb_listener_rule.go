package duplosdk

import (
	"fmt"
)

type DuploTargetGroupTuple struct {
	TargetGroupArn string `json:"TargetGroupArn,omitempty"`
	Weight         int    `json:"Weight,omitempty"`
}

type DuploTargetGroupStickinessConfig struct {
	DurationSeconds int  `json:"DurationSeconds,omitempty"`
	Enabled         bool `json:"Enabled,omitempty"`
}

type DuploAwsLbListenerRuleActionFixedResponseConfig struct {
	ContentType string `json:"ContentType,omitempty"`
	MessageBody string `json:"MessageBody,omitempty"`
	StatusCode  string `json:"StatusCode,omitempty"`
}

type DuploAwsLbListenerRuleActionForwardConfig struct {
	TargetGroups                *[]DuploTargetGroupTuple          `json:"TargetGroups,omitempty"`
	TargetGroupStickinessConfig *DuploTargetGroupStickinessConfig `json:"TargetGroupStickinessConfig,omitempty"`
}

type DuploAwsLbListenerRuleActionAuthenticateOidcConfig struct {
	AuthenticationRequestExtraParams map[string]string `json:"AuthenticationRequestExtraParams,omitempty"`
	AuthorizationEndpoint            string            `json:"AuthorizationEndpoint,omitempty"`
	ClientId                         string            `json:"ClientId,omitempty"`
	ClientSecret                     string            `json:"ClientSecret,omitempty"`
	Issuer                           string            `json:"Issuer,omitempty"`
	OnUnauthenticatedRequest         *DuploStringValue `json:"OnUnauthenticatedRequest,omitempty"`
	Scope                            string            `json:"Scope,omitempty"`
	SessionCookieName                string            `json:"SessionCookieName,omitempty"`
	SessionTimeout                   int               `json:"SessionTimeout,omitempty"`
	TokenEndpoint                    string            `json:"TokenEndpoint,omitempty"`
	UseExistingClientSecret          bool              `json:"UseExistingClientSecret,omitempty"`
	UserInfoEndpoint                 string            `json:"UserInfoEndpoint,omitempty"`
}

type DuploAwsLbListenerRuleActionAuthenticateCognitoConfig struct {
	AuthenticationRequestExtraParams map[string]string `json:"AuthenticationRequestExtraParams,omitempty"`
	OnUnauthenticatedRequest         *DuploStringValue `json:"OnUnauthenticatedRequest,omitempty"`
	Scope                            string            `json:"Scope,omitempty"`
	SessionCookieName                string            `json:"SessionCookieName,omitempty"`
	SessionTimeout                   int               `json:"SessionTimeout,omitempty"`
	UserPoolArn                      string            `json:"UserPoolArn,omitempty"`
	UserPoolClientId                 string            `json:"UserPoolClientId,omitempty"`
	UserPoolDomain                   string            `json:"UserPoolDomain,omitempty"`
}

type DuploAwsLbListenerRuleActionRedirectConfig struct {
	Host       string            `json:"Host,omitempty"`
	Path       string            `json:"Path,omitempty"`
	Port       string            `json:"Port,omitempty"`
	Protocol   string            `json:"Protocol,omitempty"`
	Query      string            `json:"Query,omitempty"`
	StatusCode *DuploStringValue `json:"StatusCode,omitempty"`
}

type DuploAwsLbListenerRuleAction struct {
	RedirectConfig            *DuploAwsLbListenerRuleActionRedirectConfig            `json:"RedirectConfig,omitempty"`
	ForwardConfig             *DuploAwsLbListenerRuleActionForwardConfig             `json:"ForwardConfig,omitempty"`
	FixedResponseConfig       *DuploAwsLbListenerRuleActionFixedResponseConfig       `json:"FixedResponseConfig,omitempty"`
	AuthenticateOidcConfig    *DuploAwsLbListenerRuleActionAuthenticateOidcConfig    `json:"AuthenticateOidcConfig,omitempty"`
	AuthenticateCognitoConfig *DuploAwsLbListenerRuleActionAuthenticateCognitoConfig `json:"AuthenticateCognitoConfig,omitempty"`
	Type                      *DuploStringValue                                      `json:"Type,omitempty"`
	TargetGroupArn            string                                                 `json:"TargetGroupArn,omitempty"`
	Order                     int                                                    `json:"Order,omitempty"`
}

type DuploAwsLbListenerRuleCondition struct {
	PathPatternConfig       *DuploStringValues                                      `json:"PathPatternConfig,omitempty"`
	HostHeaderConfig        *DuploStringValues                                      `json:"HostHeaderConfig,omitempty"`
	HttpHeaderConfig        *DuploAwsLbListenerRuleConditionHttpRequestMethodConfig `json:"HttpHeaderConfig,omitempty"`
	HttpRequestMethodConfig *DuploStringValues                                      `json:"HttpRequestMethodConfig,omitempty"`
	QueryStringConfig       *DuploAwsLbListenerRuleConditionQueryStringConfig       `json:"QueryStringConfig,omitempty"`
	SourceIpConfig          *DuploStringValues                                      `json:"SourceIpConfig,omitempty"`
	Field                   string                                                  `json:"Field,omitempty"`
	Values                  []string                                                `json:"Values,omitempty"`
}

type DuploAwsLbListenerRuleConditionHttpRequestMethodConfig struct {
	HttpHeaderName string   `json:"HttpHeaderName,omitempty"`
	Values         []string `json:"Values,omitempty"`
}

type DuploAwsLbListenerRuleConditionQueryStringConfig struct {
	Values *[]DuploKeyStringValue `json:"Values,omitempty"`
}

type DuploAwsLbListenerRule struct {
	Actions     *[]DuploAwsLbListenerRuleAction    `json:"Actions,omitempty"`
	Conditions  *[]DuploAwsLbListenerRuleCondition `json:"Conditions,omitempty"`
	ListenerArn string                             `json:"ListenerArn,omitempty"`
	Priority    string                             `json:"Priority,omitempty"`
	Tags        *[]DuploKeyStringValue             `json:"Tags,omitempty"`
	RuleArn     string                             `json:"RuleArn,omitempty"`
}

type DuploAwsLbListenerRuleCreateReq struct {
	Actions     *[]DuploAwsLbListenerRuleAction    `json:"Actions,omitempty"`
	Conditions  *[]DuploAwsLbListenerRuleCondition `json:"Conditions,omitempty"`
	ListenerArn string                             `json:"ListenerArn,omitempty"`
	Priority    int                                `json:"Priority,omitempty"`
	Tags        *[]DuploKeyStringValue             `json:"Tags,omitempty"`
	RuleArn     string                             `json:"RuleArn,omitempty"`
}

func (c *Client) DuploAwsLbListenerRuleCreate(tenantID string, rq *DuploAwsLbListenerRuleCreateReq) (*DuploAwsLbListenerRule, ClientError) {
	rp := DuploAwsLbListenerRule{}
	err := c.postAPI(
		fmt.Sprintf("DuploAwsLbListenerRuleCreate(%s, %s)", tenantID, rq.ListenerArn),
		fmt.Sprintf("v3/subscriptions/%s/aws/lbListenerRule", tenantID),
		&rq,
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploAwsLbListenerRuleUpdate(tenantID, listenerArn string, rq *DuploAwsLbListenerRule) (*DuploAwsLbListenerRule, ClientError) {
	rp := DuploAwsLbListenerRule{}
	err := c.putAPI(
		fmt.Sprintf("DuploAwsLbListenerRuleUpdate(%s, %s)", tenantID, listenerArn),
		fmt.Sprintf("v3/subscriptions/%s/aws/lbListenerRule/%s", tenantID, EncodePathParam(listenerArn)),
		&rq,
		&rp,
	)
	return &rp, err
}
func (c *Client) DuploAwsLbListenerRuleList(tenantID, listenerArn string) (*[]DuploAwsLbListenerRule, ClientError) {
	rp := []DuploAwsLbListenerRule{}
	err := c.getAPI(
		fmt.Sprintf("DuploAwsLbListenerRuleGet(%s, %s)", tenantID, listenerArn),
		fmt.Sprintf("v3/subscriptions/%s/aws/lbListenerRule/%s", tenantID, EncodePathParam(listenerArn)),
		&rp,
	)
	return &rp, err
}

func (c *Client) DuploAwsLbListenerRuleGet(tenantID string, listenerArn string, ruleArn string) (*DuploAwsLbListenerRule, ClientError) {
	list, err := c.DuploAwsLbListenerRuleList(tenantID, listenerArn)
	if err != nil {
		return nil, err
	}

	if list != nil {
		for _, rule := range *list {
			if rule.RuleArn == ruleArn {
				return &rule, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) DuploAwsLbListenerRuleDelete(tenantID string, listenerArn string, ruleArn string) ClientError {
	rp := DuploAwsLbListenerRule{}
	return c.deleteAPI(
		fmt.Sprintf("DuploAwsLbListenerRuleDelete(%s, %s, %s)", tenantID, listenerArn, ruleArn),
		fmt.Sprintf("v3/subscriptions/%s/aws/lbListenerRule/%s/%s", tenantID, EncodePathParam(listenerArn), EncodePathParam(ruleArn)),
		&rp,
	)
}
